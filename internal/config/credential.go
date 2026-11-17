// Copyright 2024-2026 Webull Inc. All rights reserved.
// Licensed under the Apache License, Version 2.0.
// See LICENSE file in the project root for full license information.

package config

import (
	"encoding/json"
	"fmt"
)

// keyringService is the keychain service for AK/SK credentials. The CLI stores a
// single credential, so it shares the fixed keyringAccount used for OAuth.
const keyringService = "webull-cli"

// Credential holds sensitive authentication data stored in the system keychain.
type Credential struct {
	AppKey    string `json:"app_key"`
	AppSecret string `json:"app_secret"`
	Token     string `json:"token,omitempty"`
}

// SaveCredential stores AK/SK and token in the system keychain.
func SaveCredential(cred *Credential) error {
	data, err := json.Marshal(cred)
	if err != nil {
		return fmt.Errorf("failed to serialize credential: %w", err)
	}
	if err := secureSet(keyringService, keyringAccount, string(data)); err != nil {
		return fmt.Errorf("failed to save credential: %w", err)
	}
	return nil
}

// LoadCredential retrieves AK/SK and token from secure storage.
func LoadCredential() (*Credential, error) {
	data, err := secureGet(keyringService, keyringAccount)
	if err != nil {
		return nil, fmt.Errorf("failed to load credential: %w", err)
	}
	var cred Credential
	if err := json.Unmarshal([]byte(data), &cred); err != nil {
		return nil, fmt.Errorf("failed to parse credential: %w", err)
	}
	return &cred, nil
}

// DeleteCredential removes the AK/SK credential from secure storage.
func DeleteCredential() error {
	if err := secureDelete(keyringService, keyringAccount); err != nil {
		return fmt.Errorf("failed to delete credential: %w", err)
	}
	return nil
}
