// Copyright 2024-2026 Webull Inc. All rights reserved.
// Licensed under the Apache License, Version 2.0.
// See LICENSE file in the project root for full license information.

package config

import (
	"encoding/json"
	"fmt"
)

const (
	// oauthKeyringService is the keychain service for OAuth tokens.
	oauthKeyringService = "webull-cli-oauth"
	// keyringAccount is the fixed keychain account. The CLI stores a single
	// credential, so the account name is constant rather than per-profile.
	keyringAccount = "default"
)

// OAuthCredentialData holds sensitive OAuth tokens stored in the system keychain.
type OAuthCredentialData struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

// SaveOAuthCredential stores OAuth tokens in the system keychain.
func SaveOAuthCredential(cred *OAuthCredentialData) error {
	data, err := json.Marshal(cred)
	if err != nil {
		return fmt.Errorf("failed to serialize OAuth credential: %w", err)
	}
	if err := secureSet(oauthKeyringService, keyringAccount, string(data)); err != nil {
		return fmt.Errorf("failed to save OAuth credential: %w", err)
	}
	return nil
}

// LoadOAuthCredential retrieves OAuth tokens from secure storage.
func LoadOAuthCredential() (*OAuthCredentialData, error) {
	data, err := secureGet(oauthKeyringService, keyringAccount)
	if err != nil {
		return nil, fmt.Errorf("failed to load OAuth credential: %w", err)
	}
	var cred OAuthCredentialData
	if err := json.Unmarshal([]byte(data), &cred); err != nil {
		return nil, fmt.Errorf("failed to parse OAuth credential: %w", err)
	}
	return &cred, nil
}

// DeleteOAuthCredential removes OAuth tokens from secure storage.
func DeleteOAuthCredential() error {
	if err := secureDelete(oauthKeyringService, keyringAccount); err != nil {
		return fmt.Errorf("failed to delete OAuth credential: %w", err)
	}
	return nil
}
