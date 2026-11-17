// Copyright 2024-2026 Webull Inc. All rights reserved.
// Licensed under the Apache License, Version 2.0.
// See LICENSE file in the project root for full license information.

package auth

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"github.com/webull-inc/webull-openapi-cli/internal/config"
)

const (
	clientRegPath  = "/api/userauth/oauth/client/register"
	deviceCodePath = "/api/userauth/oauth/device/code"
	tokenPath      = "/api/userauth/oauth/token/token"

	// grantTypeDeviceCode is the RFC 8628 device code grant type.
	grantTypeDeviceCode = "urn:ietf:params:oauth:grant-type:device_code"
	grantTypeRefresh    = "refresh_token"

	// cliClientName is the fixed client name for CLI OAuth registration.
	cliClientName = "OPENAPI_CLI"

	// authPageLanguage is the fixed display language requested from the H5
	// authorization page (query parameter "hl"), applied in every environment.
	authPageLanguage = "en"

	// defaultPollInterval is the fallback polling interval when the server
	// does not return one.
	defaultPollInterval = 5 * time.Second
	// devicePollTimeout bounds the whole polling loop.
	devicePollTimeout = 300 * time.Second
)

// Sentinel errors for the device code flow. Callers use errors.Is to classify.
var (
	// ErrDeviceCodeExpired indicates the device code expired or does not exist.
	ErrDeviceCodeExpired = errors.New("device code expired")
	// ErrAccessDenied indicates the user denied the authorization request.
	ErrAccessDenied = errors.New("authorization denied by user")
)

// ClientRegisterResponse is the response from the DCR endpoint.
type ClientRegisterResponse struct {
	ClientID   string `json:"client_id"`
	ClientName string `json:"client_name"`
}

// DeviceCodeResponse is the response from the device code generation endpoint.
type DeviceCodeResponse struct {
	DeviceCode string `json:"deviceCode"`
	UserCode   string `json:"userCode"`
	ExpiresIn  int    `json:"expireSeconds"`
	Interval   int    `json:"interval"`
}

// OAuthToken holds the token data returned by the token endpoint.
type OAuthToken struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
	RtExpiresIn  int    `json:"rt_expires_in"`
	Scope        string `json:"scope"`
	Region       string `json:"region"`
	IdentityID   string `json:"identity_id"`
}

// DeviceCodeManager handles the OAuth 2.1 Device Code (RFC 8628) flow.
type DeviceCodeManager struct {
	authHost string
	h5Base   string
	quiet    bool
}

// NewDeviceCodeManager creates a new DeviceCodeManager from resolved endpoints.
func NewDeviceCodeManager(endpoints config.OAuthEndpoints, quiet bool) *DeviceCodeManager {
	return &DeviceCodeManager{
		authHost: endpoints.Token,
		h5Base:   endpoints.H5,
		quiet:    quiet,
	}
}

// RegisterClient registers a new OAuth client with the authorization server as
// a public device-code client. Returns the client_id. Called once per machine.
func (m *DeviceCodeManager) RegisterClient(ctx context.Context) (*ClientRegisterResponse, error) {
	body := map[string]interface{}{
		"client_name":                cliClientName,
		"grant_types":                []string{grantTypeDeviceCode, grantTypeRefresh},
		"response_types":             []string{"code"},
		"token_endpoint_auth_method": "none",
	}

	resp, err := m.doPostJSON(ctx, clientRegPath, body)
	if err != nil {
		return nil, fmt.Errorf("client registration failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("client registration failed (HTTP %d): %s", resp.StatusCode, string(respBody))
	}

	var result ClientRegisterResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("failed to parse registration response: %w", err)
	}

	if result.ClientID == "" {
		return nil, fmt.Errorf("empty client_id in registration response")
	}

	return &result, nil
}

// GenerateDeviceCode requests a device_code + user_code pair, submitting the
// PKCE code_challenge that will be verified when exchanging for a token.
func (m *DeviceCodeManager) GenerateDeviceCode(ctx context.Context, clientID string, pkce *PKCE) (*DeviceCodeResponse, error) {
	body := map[string]interface{}{
		"client_id":             clientID,
		"code_challenge":        pkce.Challenge,
		"code_challenge_method": pkce.Method,
	}

	resp, err := m.doPostJSON(ctx, deviceCodePath, body)
	if err != nil {
		return nil, fmt.Errorf("device code request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("device code request failed (HTTP %d): %s", resp.StatusCode, string(respBody))
	}

	var result DeviceCodeResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("failed to parse device code response: %w", err)
	}

	if result.DeviceCode == "" || result.UserCode == "" {
		return nil, fmt.Errorf("empty device_code or user_code in response")
	}

	return &result, nil
}

// PollForToken polls the standard token endpoint with the device_code until the
// user completes authorization (RFC 8628). It honors the server-suggested
// interval and backs off on slow_down.
func (m *DeviceCodeManager) PollForToken(ctx context.Context, clientID, deviceCode, codeVerifier string, interval int) (*OAuthToken, error) {
	pollInterval := time.Duration(interval) * time.Second
	if pollInterval <= 0 {
		pollInterval = defaultPollInterval
	}

	timeout := time.After(devicePollTimeout)
	timer := time.NewTimer(pollInterval)
	defer timer.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-timeout:
			return nil, ErrDeviceCodeExpired
		case <-timer.C:
			token, errCode, err := m.exchangeToken(ctx, clientID, deviceCode, codeVerifier)
			if err != nil {
				return nil, err
			}
			if token != nil {
				return token, nil
			}
			// Not ready yet — handle RFC 8628 error codes.
			switch errCode {
			case "authorization_pending":
				// keep current interval
			case "slow_down":
				// RFC 8628: increase interval by 5 seconds
				pollInterval += 5 * time.Second
			case "access_denied":
				return nil, ErrAccessDenied
			case "expired_token":
				return nil, ErrDeviceCodeExpired
			default:
				return nil, fmt.Errorf("unexpected authorization error: %s", errCode)
			}
			timer.Reset(pollInterval)
		}
	}
}

// exchangeToken makes one token request. Returns (token, "", nil) on success,
// (nil, errorCode, nil) for a pending/retryable RFC 8628 error, or (nil, "", err)
// on a hard failure.
func (m *DeviceCodeManager) exchangeToken(ctx context.Context, clientID, deviceCode, codeVerifier string) (*OAuthToken, string, error) {
	form := url.Values{}
	form.Set("grant_type", grantTypeDeviceCode)
	form.Set("client_id", clientID)
	form.Set("device_code", deviceCode)
	form.Set("code_verifier", codeVerifier)

	resp, err := m.doPostForm(ctx, tokenPath, form)
	if err != nil {
		return nil, "", fmt.Errorf("token request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, "", err
	}

	if resp.StatusCode == http.StatusOK {
		var token OAuthToken
		if err := json.Unmarshal(respBody, &token); err != nil {
			return nil, "", fmt.Errorf("failed to parse token response: %w", err)
		}
		if token.AccessToken == "" {
			return nil, "", fmt.Errorf("empty access_token in token response")
		}
		return &token, "", nil
	}

	// Non-200: parse the standard OAuth error body.
	var errResp struct {
		Error       string `json:"error"`
		Description string `json:"error_description"`
	}
	if json.Unmarshal(respBody, &errResp) == nil && errResp.Error != "" {
		return nil, errResp.Error, nil
	}

	return nil, "", fmt.Errorf("token request failed (HTTP %d): %s", resp.StatusCode, string(respBody))
}

// BuildAuthorizationURL constructs the H5 authorization page URL. It carries
// client_id plus a fixed display language (hl); the user_code is shown
// separately on the console for the user to enter on the page.
// h5Base is stored without a scheme (consistent with the token host); default to
// https unless the configured value already carries an explicit scheme (e.g. an
// internal http:// override).
func (m *DeviceCodeManager) BuildAuthorizationURL(clientID, userCode string) string {
	base := strings.TrimSpace(m.h5Base)
	if !strings.HasPrefix(base, "http://") && !strings.HasPrefix(base, "https://") {
		base = "https://" + base
	}
	u, err := url.Parse(base)
	if err != nil {
		return fmt.Sprintf("%s?client_id=%s&hl=%s", base, url.QueryEscape(clientID), authPageLanguage)
	}
	// Keep any query already present on the configured base, then add ours.
	q := u.Query()
	q.Set("client_id", clientID)
	q.Set("hl", authPageLanguage)
	u.RawQuery = q.Encode()
	return u.String()
}

// OpenBrowser attempts to open the URL in the default browser.
// Returns true if the browser was successfully opened.
func OpenBrowser(target string) bool {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", target)
	case "linux":
		cmd = exec.Command("xdg-open", target)
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", target)
	default:
		return false
	}
	return cmd.Start() == nil
}

func (m *DeviceCodeManager) doPostJSON(ctx context.Context, path string, body interface{}) (*http.Response, error) {
	bodyBytes, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}

	target := fmt.Sprintf("https://%s%s", m.authHost, path)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, target, bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 30 * time.Second}
	return client.Do(req)
}

func (m *DeviceCodeManager) doPostForm(ctx context.Context, path string, form url.Values) (*http.Response, error) {
	target := fmt.Sprintf("https://%s%s", m.authHost, path)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, target, strings.NewReader(form.Encode()))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{Timeout: 30 * time.Second}
	return client.Do(req)
}
