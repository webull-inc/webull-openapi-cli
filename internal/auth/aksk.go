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
	"time"

	"github.com/webull-inc/webull-openapi-cli/internal/config"
)

// Compile-time interface check.
var _ CredentialProvider = (*AkSkProvider)(nil)

// AkSkProvider implements CredentialProvider for AK/SK + 2FA authentication.
type AkSkProvider struct {
	profile   *config.Profile
	pm        *config.ProfileManager
	appKey    string
	appSecret string
	token     string
}

// NewAkSkProvider creates a new AK/SK credential provider.
func NewAkSkProvider(profile *config.Profile, pm *config.ProfileManager, appKey, appSecret, token string) *AkSkProvider {
	return &AkSkProvider{
		profile:   profile,
		pm:        pm,
		appKey:    appKey,
		appSecret: appSecret,
		token:     token,
	}
}

// Authenticate adds AK/SK signature headers to the request.
func (p *AkSkProvider) Authenticate(req *http.Request, body []byte) error {
	if p.appKey == "" {
		return fmt.Errorf("AK/SK credentials not configured in profile %q", p.profile.Name)
	}

	host := req.URL.Host
	path := req.URL.Path

	query := make(map[string]string)
	for k, v := range req.URL.Query() {
		if len(v) > 0 {
			query[k] = v[0]
		}
	}

	result := Sign(SignRequest{
		AppKey:    p.appKey,
		AppSecret: p.appSecret,
		Host:      host,
		Path:      path,
		Query:     query,
		Body:      body,
		Algorithm: resolveAlgorithm(host),
	})

	for k, v := range result.Headers {
		req.Header.Set(k, v)
	}

	// Add version header
	req.Header.Set("x-version", "v2")

	// Add 2FA token if available and not expired
	if p.profile.AkSk != nil && p.profile.AkSk.TokenCheckEnabled && p.token != "" {
		if p.profile.AkSk.TokenExpires == 0 || time.Now().UnixMilli() < p.profile.AkSk.TokenExpires {
			req.Header.Set("x-access-token", p.token)
		}
	}

	return nil
}

// HandleAuthError handles INVALID_TOKEN by triggering 2FA refresh.
func (p *AkSkProvider) HandleAuthError(resp *http.Response) (bool, error) {
	if resp.StatusCode != http.StatusUnauthorized {
		return false, nil
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return false, err
	}

	var errResp struct {
		ErrorCode string `json:"error_code"`
	}
	if err := json.Unmarshal(body, &errResp); err != nil {
		return false, nil
	}

	if errResp.ErrorCode == "INVALID_TOKEN" {
		// Trigger 2FA token acquisition
		host := ResolveAPIHost(p.profile)
		tm := NewTokenManager(p.appKey, p.appSecret, host, false)

		ctx := context.Background()
		tokenResp, err := tm.AcquireToken(ctx)
		if err != nil {
			return false, fmt.Errorf("2FA verification failed: %w", err)
		}

		// Update in-memory token
		p.token = tokenResp.Token

		// Persist non-sensitive fields to disk
		if p.pm != nil {
			cfg, err := p.pm.Load()
			if err == nil {
				if prof, ok := cfg.Profiles[p.profile.Name]; ok {
					prof.AkSk.TokenExpires = tokenResp.Expires
					prof.AkSk.TokenCheckEnabled = true
					prof.AkSk.TokenStatus = tokenResp.Status
					p.pm.Save(cfg)
				}
			}
		}

		// Update token in system keychain
		cred, err := config.LoadCredential(p.profile.Name)
		if err == nil {
			cred.Token = tokenResp.Token
			config.SaveCredential(p.profile.Name, cred)
		}

		return true, nil
	}

	return false, nil
}

// sha1Hosts lists hosts that still require HMAC-SHA1 (legacy, not yet upgraded).
var sha1Hosts = map[string]bool{
	"api.webull.com":                             true,
	"events-api.webull.com":                      true,
	"api.webull.hk":                              true,
	"events-api.webull.hk":                       true,
	"pre-openapi-us-alb.webullbroker.com":        true,
	"pre-openapi-us-events.webullbroker.com":     true,
	"pre-openapi-alb.webullbroker.com":           true,
	"pre-openapi-events.webullbroker.com":        true,
	"us-openapi-alb.uat.webullbroker.com":        true,
	"us-openapi-events.uat.webullbroker.com":     true,
	"hk-openapi.uat.webullbroker.com":            true,
	"hk-openapi-events-api.uat.webullbroker.com": true,
	"api.sandbox.webull.hk":                      true,
	"events-api.sandbox.webull.hk":               true,
}

// resolveAlgorithm determines the signing algorithm based on host.
// Hosts that have NOT been upgraded still use HMAC-SHA1.
// All other hosts use HMAC-SHA256.
func resolveAlgorithm(host string) SignAlgorithm {
	if sha1Hosts[host] {
		return AlgoHMACSHA1
	}
	return AlgoHMACSHA256
}
