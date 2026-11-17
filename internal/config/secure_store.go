// Copyright 2024-2026 Webull Inc. All rights reserved.
// Licensed under the Apache License, Version 2.0.
// See LICENSE file in the project root for full license information.

package config

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"sync"

	"github.com/zalando/go-keyring"
)

// Secure credential storage with keychain-first, encrypted-file fallback.
//
// Primary store: the OS keychain (go-keyring). When the keychain is unavailable
// (headless Linux without Secret Service, CI containers, locked keychain), the
// store transparently falls back to an AES-256-GCM encrypted file and prints a
// one-line notice to stderr so the user knows credentials are no longer in the
// OS keychain.
//
// Encryption key derivation combines three independent parts so that no single
// artifact is enough to decrypt the file:
//   - a 32-byte random key material file (~/.config/webull/key_material, 0600)
//   - an optional environment variable (WEBULL_KEY_SALT), injectable in CI
//   - a machine fingerprint (hostname + current user)
//
// This is obfuscation-grade protection against casual disclosure (accidental
// git commit, backup leaks, another user browsing the filesystem). It is not a
// substitute for the OS keychain and cannot defend against an attacker with
// full read access to the user's home directory and process environment.

// envKeySalt is an optional, user/CI supplied component of the file encryption
// key. Keep it out of the home directory for defense in depth.
const envKeySalt = "WEBULL_KEY_SALT"

// fallbackNoticeOnce ensures the degraded-storage notice is printed at most once
// per process.
var fallbackNoticeOnce sync.Once

func keyMaterialPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config", "webull", "key_material")
}

func encryptedCredentialPath(service string) string {
	home, _ := os.UserHomeDir()
	// One file per service so AK/SK and OAuth credentials stay separate, mirroring
	// the keychain's per-service layout.
	return filepath.Join(home, ".config", "webull", "credentials-"+service+".enc")
}

// secureSet stores a value, preferring the keychain and falling back to an
// encrypted file when the keychain cannot be written.
func secureSet(service, account, value string) error {
	if err := keyring.Set(service, account, value); err == nil {
		return nil
	}
	noticeFallback()
	return fileSet(service, value)
}

// secureGet retrieves a value. It prefers the keychain, then falls back to the
// encrypted file when the keychain read fails (unavailable or key absent).
func secureGet(service, account string) (string, error) {
	if v, err := keyring.Get(service, account); err == nil {
		return v, nil
	}
	return fileGet(service)
}

// secureDelete removes a value from both backends. A miss in either backend is
// not an error. A failure is reported only when the credential is verifiably
// still present afterwards, so a keychain that was never usable (for example
// headless Linux) does not make logout fail.
func secureDelete(service, account string) error {
	delErr := keyring.Delete(service, account)
	if delErr != nil && !errors.Is(delErr, keyring.ErrNotFound) {
		if _, getErr := keyring.Get(service, account); getErr == nil {
			return fmt.Errorf("credential is still present in the system keychain: %w", delErr)
		}
	}
	if path := encryptedCredentialPath(service); fileExists(path) {
		if err := os.Remove(path); err != nil {
			return fmt.Errorf("failed to delete encrypted credential file: %w", err)
		}
	}
	return nil
}

func noticeFallback() {
	fallbackNoticeOnce.Do(func() {
		fmt.Fprintln(os.Stderr,
			"notice: system keychain unavailable, storing credentials in an encrypted local file")
	})
}

// --- encrypted file backend ---

func fileSet(service, value string) error {
	key, err := deriveFileKey()
	if err != nil {
		return err
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return fmt.Errorf("failed to generate nonce: %w", err)
	}
	ciphertext := gcm.Seal(nonce, nonce, []byte(value), nil)

	path := encryptedCredentialPath(service)
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}
	encoded := base64.StdEncoding.EncodeToString(ciphertext)
	if err := os.WriteFile(path, []byte(encoded), 0600); err != nil {
		return fmt.Errorf("failed to write encrypted credential file: %w", err)
	}
	return nil
}

func fileGet(service string) (string, error) {
	path := encryptedCredentialPath(service)
	raw, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return "", fmt.Errorf("no stored credential found")
		}
		return "", fmt.Errorf("failed to read encrypted credential file: %w", err)
	}

	ciphertext, err := base64.StdEncoding.DecodeString(string(raw))
	if err != nil {
		return "", fmt.Errorf("corrupt encrypted credential file: %w", err)
	}

	key, err := deriveFileKey()
	if err != nil {
		return "", err
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	if len(ciphertext) < gcm.NonceSize() {
		return "", fmt.Errorf("corrupt encrypted credential file: too short")
	}
	nonce, enc := ciphertext[:gcm.NonceSize()], ciphertext[gcm.NonceSize():]
	plaintext, err := gcm.Open(nil, nonce, enc, nil)
	if err != nil {
		return "", fmt.Errorf("failed to decrypt credential (key material changed or file tampered): %w", err)
	}
	return string(plaintext), nil
}

// deriveFileKey derives a 32-byte AES key from three independent parts via
// HMAC-SHA256, so possessing only one part (the file, the env var, or the
// machine) is insufficient to reconstruct the key.
func deriveFileKey() ([]byte, error) {
	material, err := loadOrCreateKeyMaterial()
	if err != nil {
		return nil, err
	}

	mac := hmac.New(sha256.New, material)
	mac.Write([]byte(os.Getenv(envKeySalt)))
	mac.Write([]byte(machineFingerprint()))
	return mac.Sum(nil), nil
}

// loadOrCreateKeyMaterial reads the 32-byte random key material file, creating
// it on first use with 0600 permissions.
func loadOrCreateKeyMaterial() ([]byte, error) {
	path := keyMaterialPath()
	if data, err := os.ReadFile(path); err == nil {
		decoded, derr := base64.StdEncoding.DecodeString(string(data))
		if derr == nil && len(decoded) == 32 {
			return decoded, nil
		}
		return nil, fmt.Errorf("corrupt key material file: %s", path)
	}

	material := make([]byte, 32)
	if _, err := rand.Read(material); err != nil {
		return nil, fmt.Errorf("failed to generate key material: %w", err)
	}
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return nil, fmt.Errorf("failed to create config directory: %w", err)
	}
	encoded := base64.StdEncoding.EncodeToString(material)
	if err := os.WriteFile(path, []byte(encoded), 0600); err != nil {
		return nil, fmt.Errorf("failed to write key material file: %w", err)
	}
	return material, nil
}

// machineFingerprint binds the key to the current host and user, raising the
// cost of lifting the credential file to another machine or account.
func machineFingerprint() string {
	host, _ := os.Hostname()
	name := ""
	if u, err := user.Current(); err == nil {
		name = u.Username
	}
	return host + "/" + name
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
