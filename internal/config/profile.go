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

// AkSkCredentials holds AK/SK authentication data.
// AkSkCredentials holds non-sensitive AK/SK metadata.
// Sensitive data (app_key, app_secret, token) is stored in system keychain.
type AkSkCredentials struct {
	TokenCheckEnabled bool   `json:"token_check_enabled"`
	TokenExpires      int64  `json:"token_expires,omitempty"`
	TokenStatus       string `json:"token_status,omitempty"`
}

// OAuthCredentials holds OAuth authentication data.
// Reserved for Phase 2 - not used in current release.
type OAuthCredentials struct {
	ClientID         string `json:"client_id"`
	AccessToken      string `json:"access_token"`
	RefreshToken     string `json:"refresh_token"`
	Scope            string `json:"scope"`
	ExpiresAt        string `json:"expires_at"`
	RefreshExpiresAt string `json:"refresh_expires_at"`
}

// Endpoints holds custom endpoint overrides.
type Endpoints struct {
	API    string `json:"api,omitempty"`
	Quotes string `json:"quotes_api,omitempty"`
	Events string `json:"events_api,omitempty"`
}

// Profile represents a single named configuration.
type Profile struct {
	Name       string           `json:"name"`
	AuthMethod string           `json:"auth_method"` // "ak_sk" (Phase 2 will add "oauth")
	Region     Region           `json:"region"`
	Env        Environment      `json:"env"`
	Endpoints  Endpoints        `json:"endpoints"`
	AkSk       *AkSkCredentials `json:"ak_sk,omitempty"`
	// OAuth is reserved for Phase 2.
	OAuth *OAuthCredentials `json:"oauth,omitempty"`
}

// ConfigFile represents the full profiles.json structure.
type ConfigFile struct {
	Version       string              `json:"version"`
	ActiveProfile string              `json:"active_profile"`
	Profiles      map[string]*Profile `json:"profiles"`
}

// DefaultConfigPath returns the default path to profiles.json.
func DefaultConfigPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config", "webull", "profiles.json")
}

// ProfileManager handles CRUD operations on profiles.
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
func (pm *ProfileManager) Load() (*ConfigFile, error) {
	data, err := os.ReadFile(pm.configPath)
	if err != nil {
		if os.IsNotExist(err) {
			return &ConfigFile{
				Version:  "1.0",
				Profiles: make(map[string]*Profile),
			}, nil
		}
		return nil, fmt.Errorf("failed to read config: %w", err)
	}

	var cfg ConfigFile
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}
	if cfg.Profiles == nil {
		cfg.Profiles = make(map[string]*Profile)
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

	// Write with flock for concurrency safety
	return writeFileWithLock(pm.configPath, data, 0600)
}

// GetActive returns the currently active profile.
func (pm *ProfileManager) GetActive() (*Profile, error) {
	cfg, err := pm.Load()
	if err != nil {
		return nil, err
	}
	if cfg.ActiveProfile == "" {
		return nil, fmt.Errorf("no active profile. Run: webull auth login --region <region>")
	}
	p, ok := cfg.Profiles[cfg.ActiveProfile]
	if !ok {
		return nil, fmt.Errorf("active profile %q not found in config", cfg.ActiveProfile)
	}
	return p, nil
}
