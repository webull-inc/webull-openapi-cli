// Copyright 2024-2026 Webull Inc. All rights reserved.
// Licensed under the Apache License, Version 2.0.
// See LICENSE file in the project root for full license information.

package cmd

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/spf13/cobra"
	"github.com/webull-inc/webull-openapi-cli/internal/auth"
	"github.com/webull-inc/webull-openapi-cli/internal/client"
	"github.com/webull-inc/webull-openapi-cli/internal/config"
	"github.com/webull-inc/webull-openapi-cli/internal/output"
	"golang.org/x/term"
)

var authCmd = &cobra.Command{
	Use:   "auth",
	Short: "Authentication management",
}

var authLoginCmd = &cobra.Command{
	Use:   "login",
	Short: "Login with OAuth (default) or AK/SK credentials",
	Long: `Login with OAuth (default) or AK/SK credentials.

The CLI stores a single authentication record. Each successful login overwrites
the previous one, whether it was OAuth or AK/SK.`,
	RunE: runAuthLogin,
}

var authLogoutCmd = &cobra.Command{
	Use:   "logout",
	Short: "Clear local credentials and logout",
	RunE:  runAuthLogout,
}

func init() {
	authLoginCmd.Flags().String("method", "oauth", "Auth method: oauth (default) or credentials (AK/SK)")
	authLoginCmd.Flags().String("env", "prod", "Environment: prod, sandbox, or uat (sandbox and uat are equivalent)")
	authLoginCmd.Flags().String("region", "", "Region (required for --method credentials: us, hk, jp, sg, th, au, my, uk, br, mx, za)")

	authCmd.AddCommand(authLoginCmd)
	authCmd.AddCommand(authLogoutCmd)
	authCmd.AddCommand(authStatusCmd)
}

func runAuthLogin(cmd *cobra.Command, args []string) error {
	method, _ := cmd.Flags().GetString("method")

	switch method {
	case "oauth":
		return runOAuthLogin(cmd, args)
	case "credentials":
		return runCredentialsLogin(cmd, args)
	default:
		output.PrintError("INVALID_METHOD", fmt.Sprintf("Invalid method %q. Valid: oauth, credentials", method), 0)
		return errSilent
	}
}

// parseEnv maps the user-facing env flag to an internal Environment.
// "prod" -> prod; "sandbox"/"uat" -> uat (a single non-production environment).
func parseEnv(envStr string) (config.Environment, bool) {
	switch envStr {
	case "prod":
		return config.EnvProd, true
	case "sandbox", "uat":
		return config.EnvUAT, true
	default:
		return "", false
	}
}

// --- OAuth Login ---

func runOAuthLogin(cmd *cobra.Command, args []string) error {
	envStr, _ := cmd.Flags().GetString("env")

	// OAuth needs no --region: the region is chosen on the authorization page and
	// determined by the region returned with the token.
	env, ok := parseEnv(envStr)
	if !ok {
		output.PrintError("INVALID_ENV", fmt.Sprintf("Invalid env %q. Valid: prod, sandbox, uat", envStr), 0)
		return errSilent
	}

	// Load any existing profile so we can reuse an internal endpoint override
	// (uat only) that may have been pre-seeded for testing.
	pm := config.NewProfileManager("")
	existing, err := pm.Load()
	if err != nil {
		output.PrintError("CONFIG_ERROR", err.Error(), 0)
		return errSilent
	}
	var oauthOverride *config.OAuthEndpoints
	if existing.Profile != nil {
		oauthOverride = &config.OAuthEndpoints{
			Token: existing.Profile.Endpoints.Token,
			H5:    existing.Profile.Endpoints.H5,
		}
	}
	oauthEndpoints := config.ResolveOAuthEndpoints(env, oauthOverride)

	// Cancel the flow gracefully on Ctrl+C (SIGINT) or SIGTERM.
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	dcm := auth.NewDeviceCodeManager(oauthEndpoints, flagQuiet)

	// Step 1: Obtain a client_id. Reuse the one cached for this environment on
	// this machine; only register a new client when none exists yet. Different
	// environments keep separate registrations (see client_store.go).
	clientID := config.LoadClientID(env)
	if clientID == "" {
		if !flagQuiet {
			fmt.Fprintln(os.Stderr, "Registering client...")
		}
		regResp, err := dcm.RegisterClient(ctx)
		if err != nil {
			output.PrintError("REGISTRATION_ERROR", err.Error(), 0)
			return errSilent
		}
		clientID = regResp.ClientID
		if err := config.SaveClientID(env, clientID); err != nil {
			output.PrintError("CONFIG_ERROR", err.Error(), 0)
			return errSilent
		}
	}

	// Step 2: Generate PKCE pair (verifier kept local, challenge sent to server)
	pkce, err := auth.GeneratePKCE()
	if err != nil {
		output.PrintError("PKCE_ERROR", err.Error(), 0)
		return errSilent
	}

	// Step 3: Generate device_code + user_code
	dcResp, err := dcm.GenerateDeviceCode(ctx, clientID, pkce)
	if err != nil {
		output.PrintError("DEVICE_CODE_ERROR", err.Error(), 0)
		return errSilent
	}

	// Step 4: Display authorization info (only user_code is shown to the user;
	// device_code stays local for polling).
	authURL := dcm.BuildAuthorizationURL(clientID, dcResp.UserCode)

	fmt.Fprintln(os.Stderr)
	fmt.Fprintln(os.Stderr, "Open the following URL in your browser to authorize:")
	fmt.Fprintln(os.Stderr, authURL)
	fmt.Fprintln(os.Stderr)
	fmt.Fprintf(os.Stderr, "Enter the code: %s\n", dcResp.UserCode)
	fmt.Fprintln(os.Stderr)

	// Attempt to open browser
	if auth.OpenBrowser(authURL) {
		fmt.Fprintln(os.Stderr, "Browser opened. Waiting for authorization...")
	} else {
		fmt.Fprintln(os.Stderr, "Waiting for authorization...")
	}
	fmt.Fprintln(os.Stderr)

	// Step 5: Poll for token using device_code + PKCE verifier
	token, err := dcm.PollForToken(ctx, clientID, dcResp.DeviceCode, pkce.Verifier, dcResp.Interval)
	if err != nil {
		switch {
		case errors.Is(err, context.Canceled):
			output.PrintError("AUTH_CANCELLED", "Authentication cancelled.", 0)
		case errors.Is(err, auth.ErrDeviceCodeExpired):
			output.PrintError("AUTH_TIMEOUT",
				"Authentication failed: Device authorization timed out — please try again", 0)
		case errors.Is(err, auth.ErrAccessDenied):
			output.PrintError("AUTH_DENIED", "Authorization failed: access_denied", 0)
		default:
			output.PrintError("AUTH_FAILED", "Authorization failed: "+err.Error(), 0)
		}
		return errSilent
	}

	// Determine region from the token response; fall back to US when the token
	// does not carry a region yet.
	region := config.RegionUS
	tokenRegion := config.Region(token.Region)
	if tokenRegion.IsValid() {
		region = tokenRegion
	} else if !flagQuiet {
		fmt.Fprintf(os.Stderr,
			"warning: the server did not return a valid region (got %q); defaulting to %q. "+
				"If your account is in another region, API requests will go to the wrong host.\n",
			token.Region, region)
	}

	// Step 6: Save credentials to the system keychain.
	oauthCred := &config.OAuthCredentialData{
		AccessToken:  token.AccessToken,
		RefreshToken: token.RefreshToken,
	}
	if err := config.SaveOAuthCredential(oauthCred); err != nil {
		output.PrintError("KEYCHAIN_ERROR", err.Error(), 0)
		return errSilent
	}

	// access_token / refresh_token lifetimes come from the server response.
	now := time.Now().Unix()
	rtExpiresIn := token.RtExpiresIn
	if rtExpiresIn <= 0 {
		rtExpiresIn = 30 * 24 * 60 * 60
	}

	// Step 7: Persist the single profile (overwrites any existing record).
	profile := &config.Profile{
		AuthMethod: config.AuthMethodOAuth,
		Region:     region,
		Env:        env,
		Endpoints:  existingEndpoints(existing),
		OAuth: &config.OAuthCredentials{
			ClientID:         clientID,
			Scope:            token.Scope,
			IdentityID:       token.IdentityID,
			ExpiresAt:        now + int64(token.ExpiresIn),
			RefreshExpiresAt: now + int64(rtExpiresIn),
		},
	}

	if err := pm.SaveProfile(profile); err != nil {
		return fmt.Errorf("failed to save profile: %w", err)
	}

	fmt.Fprintln(os.Stderr)
	fmt.Fprintln(os.Stderr, "Successfully authenticated.")
	return nil
}

// existingEndpoints preserves any pre-seeded endpoint override from the previous
// profile so a testing backdoor survives re-login. Returns a zero value when the
// previous profile carried no override.
func existingEndpoints(cfg *config.ConfigFile) config.Endpoints {
	if cfg != nil && cfg.Profile != nil {
		return cfg.Profile.Endpoints
	}
	return config.Endpoints{}
}

// --- Credentials Login (AK/SK + 2FA) ---

func runCredentialsLogin(cmd *cobra.Command, args []string) error {
	regionStr, _ := cmd.Flags().GetString("region")
	envStr, _ := cmd.Flags().GetString("env")

	region := config.Region(regionStr)
	if !region.IsValid() {
		output.PrintError("INVALID_REGION", fmt.Sprintf("Invalid region %q. Valid: us, hk, jp, sg, th, au, my, uk, br, mx, za", regionStr), 0)
		return errSilent
	}

	env, ok := parseEnv(envStr)
	if !ok {
		output.PrintError("INVALID_ENV", fmt.Sprintf("Invalid env %q. Valid: prod, sandbox, uat", envStr), 0)
		return errSilent
	}

	pm := config.NewProfileManager("")
	existing, err := pm.Load()
	if err != nil {
		return err
	}

	// Resolve the API host: prod is hardcoded; uat may reuse a pre-seeded override.
	override := ""
	if existing.Profile != nil {
		override = existing.Profile.Endpoints.API
	}
	apiHost := config.ResolveAPIEndpoint(region, env, override)

	// Read AK/SK interactively
	reader := bufio.NewReader(os.Stdin)

	if !flagQuiet {
		fmt.Fprint(os.Stderr, "Enter App Key: ")
	}
	appKey, _ := reader.ReadString('\n')
	appKey = strings.TrimSpace(appKey)

	if !flagQuiet {
		fmt.Fprint(os.Stderr, "Enter App Secret: ")
	}
	oldState, _ := term.GetState(int(os.Stdin.Fd()))
	appSecretBytes, err := term.ReadPassword(int(os.Stdin.Fd()))
	if err != nil {
		if oldState != nil {
			term.Restore(int(os.Stdin.Fd()), oldState)
		}
		fmt.Fprintln(os.Stderr)
		return fmt.Errorf("failed to read secret: %w", err)
	}
	appSecret := strings.TrimSpace(string(appSecretBytes))
	if !flagQuiet {
		fmt.Fprintln(os.Stderr)
	}

	if appKey == "" || appSecret == "" {
		output.PrintError("MISSING_CREDENTIALS", "App Key and App Secret are required", 0)
		return errSilent
	}

	// Create token manager and check config
	tm := auth.NewTokenManager(appKey, appSecret, apiHost, flagQuiet)

	if !flagQuiet {
		fmt.Fprintf(os.Stderr, "Checking 2FA configuration... (host: %s)\n", apiHost)
	}

	ctx := context.Background()
	configResp, err := tm.GetConfig(ctx)
	if err != nil {
		output.PrintError("CONFIG_ERROR", err.Error(), 0)
		return errSilent
	}

	// Build profile
	profile := &config.Profile{
		AuthMethod: config.AuthMethodAkSk,
		Region:     region,
		Env:        env,
		Endpoints:  config.Endpoints{API: override},
		AkSk: &config.AkSkCredentials{
			TokenCheckEnabled: configResp.TokenCheckEnabled,
		},
	}

	// Handle 2FA if enabled
	var token string
	if configResp.TokenCheckEnabled {
		tokenResp, err := tm.AcquireToken(ctx)
		if err != nil {
			output.PrintError("2FA_ERROR", err.Error(), 0)
			return errSilent
		}
		token = tokenResp.Token
		profile.AkSk.TokenExpires = tokenResp.Expires
		profile.AkSk.TokenStatus = tokenResp.Status
	} else {
		if !flagQuiet {
			fmt.Fprintln(os.Stderr, "2FA is not enabled for this app.")
		}
	}

	// Save credential to system keychain
	cred := &config.Credential{
		AppKey:    appKey,
		AppSecret: appSecret,
		Token:     token,
	}
	if err := config.SaveCredential(cred); err != nil {
		output.PrintError("KEYCHAIN_ERROR", err.Error(), 0)
		return errSilent
	}

	// Persist the single profile (overwrites any existing record).
	if err := pm.SaveProfile(profile); err != nil {
		return fmt.Errorf("failed to save profile: %w", err)
	}

	if !flagQuiet {
		fmt.Fprintf(os.Stderr, "Successfully authenticated (region: %s, auth: ak_sk, 2fa: %v)\n",
			region, configResp.TokenCheckEnabled)
	}
	return nil
}

// --- Auth Logout ---

func runAuthLogout(cmd *cobra.Command, args []string) error {
	pm := config.NewProfileManager("")
	profile, err := pm.GetActive()
	if err != nil {
		output.PrintError("NO_PROFILE", "Not authenticated.", 0)
		return errSilent
	}

	// Delete credentials from secure storage based on auth method.
	var deleteErr error
	switch profile.AuthMethod {
	case config.AuthMethodOAuth:
		deleteErr = config.DeleteOAuthCredential()
	case config.AuthMethodAkSk:
		deleteErr = config.DeleteCredential()
	}
	if deleteErr != nil {
		// Keep the profile so the user can retry logout.
		output.PrintError("LOGOUT_INCOMPLETE",
			fmt.Sprintf("Credentials may still be stored on this machine: %v. Fix the problem and run 'webull auth logout' again.", deleteErr), 0)
		return errSilent
	}

	// Clear the stored profile record.
	if err := pm.Clear(); err != nil {
		return fmt.Errorf("failed to clear config: %w", err)
	}

	if !flagQuiet {
		fmt.Fprintln(os.Stderr, "Logged out. Local credentials cleared.")
	}
	return nil
}

// --- Auth Status ---

var authStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Check authentication status",
	RunE: func(cmd *cobra.Command, args []string) error {
		pm := config.NewProfileManager("")
		profile, err := pm.GetActive()
		if err != nil {
			output.PrintError("NO_PROFILE", err.Error(), 0)
			return errSilent
		}

		status := map[string]interface{}{
			"auth_method": profile.AuthMethod,
			"region":      profile.Region,
			"env":         profile.Env,
		}

		switch profile.AuthMethod {
		case config.AuthMethodOAuth:
			if profile.OAuth != nil {
				status["scope"] = profile.OAuth.Scope
				status["identity_id"] = profile.OAuth.IdentityID

				// Local token-expiry metadata (from the last token response). These
				// reflect only the locally recorded lifetimes, not whether the server
				// still accepts the token.
				now := time.Now().Unix()
				if profile.OAuth.ExpiresAt > 0 {
					if now < profile.OAuth.ExpiresAt {
						remaining := time.Duration(profile.OAuth.ExpiresAt-now) * time.Second
						status["access_token_status"] = "valid"
						status["access_token_expires_in"] = remaining.String()
					} else {
						status["access_token_status"] = "expired"
					}
				}
				if profile.OAuth.RefreshExpiresAt > 0 {
					if now < profile.OAuth.RefreshExpiresAt {
						remaining := time.Duration(profile.OAuth.RefreshExpiresAt-now) * time.Second
						status["refresh_token_status"] = "valid"
						status["refresh_token_expires_in"] = remaining.String()
					} else {
						status["refresh_token_status"] = "expired"
					}
				}

				// Authoritative check: hit the server to see whether this session is
				// still accepted. This catches revocation / being kicked by a login
				// on another machine, which the local expiry timestamps cannot detect.
				sessionStatus, msg := probeOAuthSession()
				status["session_status"] = sessionStatus
				if msg != "" {
					status["message"] = msg
				}
			}

		case config.AuthMethodAkSk:
			if profile.AkSk != nil {
				status["token_check_enabled"] = profile.AkSk.TokenCheckEnabled
				if cred, err := config.LoadCredential(); err == nil {
					status["app_key"] = output.Mask(cred.AppKey)
					if cred.Token != "" {
						status["2fa_token"] = output.Mask(cred.Token)
					}
				}
			}
		}

		return output.PrintJSON(status)
	},
}

// probeOAuthSession makes one lightweight authenticated request to determine
// whether the OAuth session is still accepted by the server. It returns a
// status string and an optional user-facing message.
//
// Possible status values:
//   - "active":      server accepted the request, session is valid
//   - "invalid":     server rejected auth (401/403, e.g. revoked or kicked by a
//     login on another machine) — local timestamps cannot detect this
//   - "unknown":     could not determine (network error, or non-auth error);
//     the local token_status fields still apply
func probeOAuthSession() (string, string) {
	c, err := newAPIClient()
	if err != nil {
		// Refresh/credential load failed (e.g. refresh token expired). Treat as
		// invalid and guide re-login.
		return "invalid", "Session invalid. Please run 'webull auth login' to re-authorize."
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	resp, err := c.Do(ctx, "GET", "/trading/accounts/list", &client.RequestOptions{})
	if err != nil {
		return "unknown", ""
	}

	switch {
	case resp.StatusCode == 200:
		return "active", ""
	case resp.StatusCode == 401 || resp.StatusCode == 403:
		return "invalid", "Session invalid (revoked or signed in elsewhere). Please run 'webull auth login' to re-authorize."
	default:
		// Non-auth error (5xx, etc.): we can't confirm invalidity.
		return "unknown", ""
	}
}
