// Copyright 2024-2026 Webull Inc. All rights reserved.
// Licensed under the Apache License, Version 2.0.
// See LICENSE file in the project root for full license information.

package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// OAuth client registrations are cached per environment so a machine reuses its
// client_id instead of registering a new one on every login. Different
// environments (prod vs sandbox/uat) register separate clients, since a
// client_id issued by one authorization server is not valid on another.
//
// This is non-sensitive data (client_id only, no secret), kept in a plain file
// separate from the single-record profiles.json so switching the active auth
// does not wipe the cached registrations.

// clientRegistration holds the cached registration for one environment.
type clientRegistration struct {
	ClientID string `json:"client_id"`
}

// clientStoreFile is keyed by environment ("prod" / "uat").
type clientStoreFile struct {
	Version string                         `json:"version"`
	Clients map[string]*clientRegistration `json:"clients"`
}

// ClientStorePath returns the path to clients.json.
func ClientStorePath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config", "webull", "clients.json")
}

// LoadClientID returns the cached client_id for an environment, or "" when none
// is registered yet.
func LoadClientID(env Environment) string {
	data, err := os.ReadFile(ClientStorePath())
	if err != nil {
		return ""
	}
	var store clientStoreFile
	if err := json.Unmarshal(data, &store); err != nil {
		return ""
	}
	if reg, ok := store.Clients[string(env)]; ok && reg != nil {
		return reg.ClientID
	}
	return ""
}

// SaveClientID persists the client_id for an environment, preserving any other
// environment's registration already in the file.
func SaveClientID(env Environment, clientID string) error {
	path := ClientStorePath()

	store := clientStoreFile{Version: configVersion, Clients: map[string]*clientRegistration{}}
	if data, err := os.ReadFile(path); err == nil {
		// Ignore parse errors: treat an unreadable file as empty and overwrite.
		_ = json.Unmarshal(data, &store)
		if store.Clients == nil {
			store.Clients = map[string]*clientRegistration{}
		}
	}

	store.Clients[string(env)] = &clientRegistration{ClientID: clientID}

	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}
	out, err := json.MarshalIndent(&store, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to serialize client store: %w", err)
	}
	if err := os.WriteFile(path, out, 0600); err != nil {
		return fmt.Errorf("failed to write client store: %w", err)
	}
	return nil
}
