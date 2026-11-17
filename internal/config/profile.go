// Copyright 2024-2026 Webull Inc. All rights reserved.
// Licensed under the Apache License, Version 2.0.
// See LICENSE file in the project root for full license information.

package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

// configVersion is the current profiles.json schema version.
const configVersion = "2.0"

// corruptConfigNoticeOnce limits the unreadable-config warning to once per process.
var corruptConfigNoticeOnce sync.Once

// AuthMethodOAuth / AuthMethodAkSk identify the two supported auth methods.
const (
	AuthMethodOAuth = "oauth"
	AuthMethodAkSk  = "ak_sk"
)

// AkSkCredentials holds non-sensitive AK/SK metadata.
// Sensitive data (app_key, app_secret, token) is stored in the system keychain.
type AkSkCredentials struct {
	TokenCheckEnabled bool   `json:"token_check_enabled"`
	TokenExpires      int64  `json:"token_expires,omitempty"`
	TokenStatus       string `json:"token_status,omitempty"`
}

// OAuthCredentials holds non-sensitive OAuth metadata stored in profiles.json.
// Sensitive data (access_token, refresh_token) is stored in the system keychain.
type OAuthCredentials struct {
	ClientID         string `json:"client_id"`
	Scope            string `json:"scope,omitempty"`
	IdentityID       string `json:"identity_id,omitempty"`
	ExpiresAt        int64  `json:"expires_at"`
	RefreshExpiresAt int64  `json:"refresh_expires_at"`
}

// Endpoints holds optional endpoint overrides. Overrides only take effect in the
// uat environment (see ResolveAPIEndpoint / ResolveOAuthEndpoints); prod is
// always locked to the hardcoded values.
//
// AK/SK only uses API. OAuth uses API + H5 + Token.
type Endpoints struct {
	API   string `json:"api,omitempty"`
	H5    string `json:"h5,omitempty"`
	Token string `json:"token,omitempty"`
}

// Profile represents the single authentication record. The CLI stores exactly
// one profile: either OAuth or AK/SK. Each login overwrites the previous record.
type Profile struct {
	AuthMethod string            `json:"auth_method"` // "oauth" or "ak_sk"
	Region     Region            `json:"region"`
	Env        Environment       `json:"env"`
	Endpoints  Endpoints         `json:"endpoints,omitempty"`
	OAuth      *OAuthCredentials `json:"oauth,omitempty"`
	AkSk       *AkSkCredentials  `json:"ak_sk,omitempty"`
}

// ConfigFile represents the full profiles.json structure (single record, v2.0).
type ConfigFile struct {
	Version string   `json:"version"`
	Profile *Profile `json:"profile,omitempty"`
}

// DefaultConfigPath returns the default path to profiles.json.
func DefaultConfigPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config", "webull", "profiles.json")
}

// ProfileManager handles reading and writing the single profile record.
type ProfileManager struct {
	configPath string
}

// NewProfileManager creates a ProfileManager with the given config path.
func NewProfileManager(configPath string) *ProfileManager {
	if configPath == "" {
		configPath = DefaultConfigPath()
	}
	return &ProfileManager{configPath: configPath}
}

// Load reads and parses the config file.
//
// Any file that does not conform to the current v2.0 schema (missing file,
// legacy multi-profile layout, or an empty profile) is treated as "not
// authenticated": Load returns an empty ConfigFile and no error, and the caller
// prompts the user to log in again. The CLI is an Alpha Preview and does not
// migrate legacy layouts.
func (pm *ProfileManager) Load() (*ConfigFile, error) {
	empty := &ConfigFile{Version: configVersion}

	data, err := os.ReadFile(pm.configPath)
	if err != nil {
		if os.IsNotExist(err) {
			return empty, nil
		}
		return nil, fmt.Errorf("failed to read config: %w", err)
	}

	var cfg ConfigFile
	if err := json.Unmarshal(data, &cfg); err != nil {
		// Unparseable / legacy layout → treat as not authenticated, but tell the
		// user the file could not be read so a damaged or hand-edited file is not
		// mistaken for "never logged in".
		corruptConfigNoticeOnce.Do(func() {
			fmt.Fprintf(os.Stderr, "warning: %s is not valid JSON (%v); ignoring it and treating this machine as not logged in\n",
				pm.configPath, err)
		})
		return empty, nil
	}
	if cfg.Version != configVersion || cfg.Profile == nil || cfg.Profile.AuthMethod == "" {
		return empty, nil
	}
	return &cfg, nil
}

// Save writes the config file with restricted permissions.
func (pm *ProfileManager) Save(cfg *ConfigFile) error {
	dir := filepath.Dir(pm.configPath)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to serialize config: %w", err)
	}

	// Write with flock for concurrency safety.
	return writeFileWithLock(pm.configPath, data, 0600)
}

// GetActive returns the single authentication profile, or an error when the CLI
// has not been authenticated yet.
func (pm *ProfileManager) GetActive() (*Profile, error) {
	cfg, err := pm.Load()
	if err != nil {
		return nil, err
	}
	if cfg.Profile == nil {
		return nil, fmt.Errorf("not authenticated. Run: webull auth login")
	}
	return cfg.Profile, nil
}

// SaveProfile persists a single profile, overwriting any existing record.
func (pm *ProfileManager) SaveProfile(profile *Profile) error {
	return pm.Save(&ConfigFile{
		Version: configVersion,
		Profile: profile,
	})
}

// Clear removes the stored profile record, leaving an empty config file.
func (pm *ProfileManager) Clear() error {
	return pm.Save(&ConfigFile{Version: configVersion})
}
