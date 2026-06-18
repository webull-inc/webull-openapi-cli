// Copyright 2024-2026 Webull Inc. All rights reserved.
// Licensed under the Apache License, Version 2.0.
// See LICENSE file in the project root for full license information.

package auth

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/webull-inc/webull-openapi-cli/internal/config"
)

const (
	tokenCreatePath = "/openapi/auth/token/create"
	tokenCheckPath  = "/openapi/auth/token/check"
	configPath      = "/openapi/config"

	pollInterval = 5 * time.Second
	pollTimeout  = 300 * time.Second
)

// TokenManager handles the 2FA token lifecycle including creation,
// polling for approval, and validation of existing tokens.
type TokenManager struct {
	appKey    string
	appSecret string
	host      string
	quiet     bool
}

// NewTokenManager creates a new TokenManager with the given credentials and host.
func NewTokenManager(appKey, appSecret, host string, quiet bool) *TokenManager {
	return &TokenManager{
		appKey:    appKey,
		appSecret: appSecret,
		host:      host,
		quiet:     quiet,
	}
}

// ConfigResponse represents the /openapi/config response.
type ConfigResponse struct {
	TokenCheckEnabled bool `json:"token_check_enabled"`
}

// TokenResponse represents the token create/check response.
type TokenResponse struct {
	Token   string `json:"token"`
	Status  string `json:"status"`
	Expires int64  `json:"expires"`
}

// GetConfig calls /openapi/config to check if 2FA is enabled.
func (tm *TokenManager) GetConfig(ctx context.Context) (*ConfigResponse, error) {
	resp, err := tm.doRequest(ctx, "GET", configPath, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get config: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("config request failed (HTTP %d): %s", resp.StatusCode, string(body))
	}

	// Debug: print raw response for troubleshooting
	if !tm.quiet {
		fmt.Fprintf(os.Stderr, "[debug] GET %s -> %d: %s\n", configPath, resp.StatusCode, string(body))
	}

	var cfg ConfigResponse
	if err := json.Unmarshal(body, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config response: %w", err)
	}
	return &cfg, nil
}

// AcquireToken performs the full 2FA token acquisition flow:
// createToken -> poll checkToken -> return verified token.
func (tm *TokenManager) AcquireToken(ctx context.Context) (*TokenResponse, error) {
	// Step 1: Create token
	if !tm.quiet {
		fmt.Fprintln(os.Stderr, "2FA verification required.")
		fmt.Fprintln(os.Stderr, "Please complete verification in Webull App.")
	}

	createResp, err := tm.createToken(ctx)
	if err != nil {
		return nil, err
	}

	if createResp.Status == "NORMAL" {
		// Token already valid (rare case)
		return createResp, nil
	}

	if !tm.quiet {
		fmt.Fprintln(os.Stderr, "Waiting for approval...")
	}

	// Step 2: Poll checkToken
	token := createResp.Token
	ticker := time.NewTicker(pollInterval)
	defer ticker.Stop()
	timeout := time.After(pollTimeout)

	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-timeout:
			return nil, fmt.Errorf("verification timeout after %v. Please retry", pollTimeout)
		case <-ticker.C:
			checkResp, err := tm.checkToken(ctx, token)
			if err != nil {
				return nil, err
			}

			switch checkResp.Status {
			case "NORMAL":
				if !tm.quiet {
					fmt.Fprintln(os.Stderr, "✓ Verification successful.")
				}
				return checkResp, nil
			case "PENDING":
				if !tm.quiet {
					fmt.Fprintf(os.Stderr, "Still waiting for approval...\n")
				}
				continue
			default:
				return nil, fmt.Errorf("unexpected token status: %s", checkResp.Status)
			}
		}
	}
}

// CheckExistingToken checks if an existing token is still valid.
func (tm *TokenManager) CheckExistingToken(ctx context.Context, token string) (*TokenResponse, error) {
	return tm.checkToken(ctx, token)
}

func (tm *TokenManager) createToken(ctx context.Context) (*TokenResponse, error) {
	resp, err := tm.doRequest(ctx, "POST", tokenCreatePath, map[string]interface{}{})
	if err != nil {
		return nil, fmt.Errorf("failed to create token: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("create token failed (HTTP %d): %s", resp.StatusCode, string(body))
	}

	var tokenResp TokenResponse
	if err := json.Unmarshal(body, &tokenResp); err != nil {
		return nil, fmt.Errorf("failed to parse token response: %w", err)
	}

	if tokenResp.Token == "" {
		return nil, fmt.Errorf("empty token in response: %s", string(body))
	}

	return &tokenResp, nil
}

func (tm *TokenManager) checkToken(ctx context.Context, token string) (*TokenResponse, error) {
	resp, err := tm.doRequest(ctx, "POST", tokenCheckPath, map[string]interface{}{
		"token": token,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to check token: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("check token failed (HTTP %d): %s", resp.StatusCode, string(body))
	}

	var tokenResp TokenResponse
	if err := json.Unmarshal(body, &tokenResp); err != nil {
		return nil, fmt.Errorf("failed to parse check token response: %w", err)
	}

	return &tokenResp, nil
}

func (tm *TokenManager) doRequest(ctx context.Context, method, path string, body interface{}) (*http.Response, error) {
	var bodyBytes []byte
	var bodyReader io.Reader

	if body != nil {
		var err error
		bodyBytes, err = json.Marshal(body)
		if err != nil {
			return nil, err
		}
		bodyReader = bytes.NewReader(bodyBytes)
	}

	url := fmt.Sprintf("https://%s%s", tm.host, path)
	req, err := http.NewRequestWithContext(ctx, method, url, bodyReader)
	if err != nil {
		return nil, err
	}

	if bodyBytes != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	// Sign the request
	query := make(map[string]string)
	for k, v := range req.URL.Query() {
		if len(v) > 0 {
			query[k] = v[0]
		}
	}

	result := Sign(SignRequest{
		AppKey:    tm.appKey,
		AppSecret: tm.appSecret,
		Host:      tm.host,
		Path:      path,
		Query:     query,
		Body:      bodyBytes,
		Algorithm: resolveAlgorithm(tm.host),
	})

	for k, v := range result.Headers {
		req.Header.Set(k, v)
	}
	req.Header.Set("x-version", "v2")
	req.Header.Set("x-webull-client-source", "cli")

	client := &http.Client{Timeout: 30 * time.Second}
	return client.Do(req)
}

// ResolveAPIHost returns the API host for a given profile.
// It uses the profile's endpoint override if set, otherwise falls back to the
// default production endpoint for the region.
func ResolveAPIHost(profile *config.Profile) string {
	if profile.Endpoints.API != "" {
		return profile.Endpoints.API
	}
	ep := config.GetEndpoints(profile.Region, nil)
	return ep.API
}
