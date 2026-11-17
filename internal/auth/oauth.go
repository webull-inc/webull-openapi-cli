// Copyright 2024-2026 Webull Inc. All rights reserved.
// Licensed under the Apache License, Version 2.0.
// See LICENSE file in the project root for full license information.

package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/webull-inc/webull-openapi-cli/internal/config"
)

const (
	tokenRefreshPath = "/api/userauth/oauth/token/token"
)

// Compile-time interface check.
var _ CredentialProvider = (*OAuthProvider)(nil)

// OAuthProvider implements CredentialProvider for OAuth Bearer token authentication.
type OAuthProvider struct {
	profile *config.Profile
	pm      *config.ProfileManager
}

// NewOAuthProvider creates a new OAuth credential provider.
func NewOAuthProvider(profile *config.Profile, pm *config.ProfileManager) *OAuthProvider {
	return &OAuthProvider{
		profile: profile,
		pm:      pm,
	}
}

// Authenticate adds Bearer token to the request.
// If the access token is expired but refresh token is valid, it refreshes first.
func (p *OAuthProvider) Authenticate(req *http.Request, body []byte) error {
	oauth := p.profile.OAuth
	if oauth == nil {
		return fmt.Errorf("OAuth credentials not configured. Run 'webull auth login'")
	}

	// Check if access token is expired
	if p.isAccessTokenExpired() {
		if p.isRefreshTokenExpired() {
			return fmt.Errorf("session expired. Please run 'webull auth login' to re-authorize")
		}
		// Attempt refresh
		if err := p.refreshToken(); err != nil {
			return fmt.Errorf("token refresh failed: %w. Please run 'webull auth login' to re-authorize", err)
		}
	}

	// Load fresh token from credential store
	cred, err := config.LoadOAuthCredential()
	if err != nil {
		return fmt.Errorf("failed to load OAuth token: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+cred.AccessToken)
	return nil
}

// HandleAuthError handles 401 by attempting to refresh the token.
func (p *OAuthProvider) HandleAuthError(resp *http.Response) (bool, error) {
	if resp.StatusCode != http.StatusUnauthorized {
		return false, nil
	}

	if p.isRefreshTokenExpired() {
		return false, fmt.Errorf("session expired. Please run 'webull auth login' to re-authorize")
	}

	if err := p.refreshToken(); err != nil {
		return false, fmt.Errorf("token refresh failed: %w", err)
	}

	return true, nil
}

func (p *OAuthProvider) isAccessTokenExpired() bool {
	if p.profile.OAuth == nil || p.profile.OAuth.ExpiresAt == 0 {
		return true
	}
	// Expired if current time >= expiresAt
	return time.Now().Unix() >= p.profile.OAuth.ExpiresAt
}

func (p *OAuthProvider) isRefreshTokenExpired() bool {
	if p.profile.OAuth == nil || p.profile.OAuth.RefreshExpiresAt == 0 {
		return true
	}
	return time.Now().Unix() >= p.profile.OAuth.RefreshExpiresAt
}

func (p *OAuthProvider) refreshToken() error {
	cred, err := config.LoadOAuthCredential()
	if err != nil {
		return err
	}

	oauthEndpoints := config.ResolveOAuthEndpoints(p.profile.Env, &config.OAuthEndpoints{
		Token: p.profile.Endpoints.Token,
		H5:    p.profile.Endpoints.H5,
	})

	// Call refresh endpoint (application/x-www-form-urlencoded)
	data := url.Values{}
	data.Set("client_id", p.profile.OAuth.ClientID)
	data.Set("grant_type", "refresh_token")
	data.Set("refresh_token", cred.RefreshToken)

	refreshURL := fmt.Sprintf("https://%s%s", oauthEndpoints.Token, tokenRefreshPath)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, refreshURL, strings.NewReader(data.Encode()))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("refresh request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		var errResp struct {
			Error       string `json:"error"`
			Description string `json:"error_description"`
		}
		if json.Unmarshal(respBody, &errResp) == nil && errResp.Error != "" {
			return fmt.Errorf("%s: %s", errResp.Error, errResp.Description)
		}
		return fmt.Errorf("refresh failed (HTTP %d): %s", resp.StatusCode, string(respBody))
	}

	var tokenResp OAuthToken
	if err := json.Unmarshal(respBody, &tokenResp); err != nil {
		return fmt.Errorf("failed to parse refresh response: %w", err)
	}

	// Update credentials in store
	newCred := &config.OAuthCredentialData{
		AccessToken:  tokenResp.AccessToken,
		RefreshToken: tokenResp.RefreshToken,
	}
	if err := config.SaveOAuthCredential(newCred); err != nil {
		return fmt.Errorf("failed to persist refreshed token: %w", err)
	}

	// Update profile metadata
	now := time.Now().Unix()
	p.profile.OAuth.ExpiresAt = now + int64(tokenResp.ExpiresIn)
	rtExpiresIn := tokenResp.RtExpiresIn
	if rtExpiresIn <= 0 {
		rtExpiresIn = 30 * 24 * 60 * 60
	}
	p.profile.OAuth.RefreshExpiresAt = now + int64(rtExpiresIn)
	if tokenResp.Scope != "" {
		p.profile.OAuth.Scope = tokenResp.Scope
	}

	// Persist profile changes
	if p.pm != nil {
		cfg, err := p.pm.Load()
		if err == nil && cfg.Profile != nil && cfg.Profile.OAuth != nil {
			cfg.Profile.OAuth.ExpiresAt = p.profile.OAuth.ExpiresAt
			cfg.Profile.OAuth.RefreshExpiresAt = p.profile.OAuth.RefreshExpiresAt
			cfg.Profile.OAuth.Scope = p.profile.OAuth.Scope
			if err := p.pm.Save(cfg); err != nil {
				fmt.Fprintf(os.Stderr, "warning: token refreshed but failed to save session metadata: %v\n", err)
			}
		}
	}

	return nil
}
