// Copyright 2024-2026 Webull Inc. All rights reserved.
// Licensed under the Apache License, Version 2.0.
// See LICENSE file in the project root for full license information.

package config

import (
	"encoding/json"
	"fmt"

	"github.com/zalando/go-keyring"
)

const keyringService = "webull-cli"

// Credential holds sensitive authentication data stored in the system keychain.
type Credential struct {
	AppKey    string `json:"app_key"`
	AppSecret string `json:"app_secret"`
	Token     string `json:"token,omitempty"`
}

// SaveCredential stores AK/SK and token in the system keychain.
func SaveCredential(profileName string, cred *Credential) error {
	data, err := json.Marshal(cred)
	if err != nil {
		return fmt.Errorf("failed to serialize credential: %w", err)
	}
	if err := keyring.Set(keyringService, profileName, string(data)); err != nil {
		return fmt.Errorf("failed to save credential to keychain: %w", err)
	}
	return nil
}

// LoadCredential retrieves AK/SK and token from the system keychain.
func LoadCredential(profileName string) (*Credential, error) {
	data, err := keyring.Get(keyringService, profileName)
	if err != nil {
		return nil, fmt.Errorf("failed to load credential from keychain: %w", err)
	}
	var cred Credential
	if err := json.Unmarshal([]byte(data), &cred); err != nil {
		return nil, fmt.Errorf("failed to parse credential: %w", err)
	}
	return &cred, nil
}

// DeleteCredential removes a credential from the system keychain.
func DeleteCredential(profileName string) error {
	err := keyring.Delete(keyringService, profileName)
	if err != nil {
		return fmt.Errorf("failed to delete credential from keychain: %w", err)
	}
	return nil
}
