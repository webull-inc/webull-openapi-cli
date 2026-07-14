// Copyright 2024-2026 Webull Inc. All rights reserved.
// Licensed under the Apache License, Version 2.0.
// See LICENSE file in the project root for full license information.

package cmd

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/webull-inc/webull-openapi-cli/internal/auth"
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
	Short: "Login with AK/SK credentials",
	RunE:  runAuthLogin,
}

func init() {
	authLoginCmd.Flags().String("region", "", "Region (required: us, hk, jp)")
	authLoginCmd.Flags().String("env", "prod", "Environment (prod/uat)")
	authLoginCmd.Flags().String("name", "", "Profile name (default: 'default')")
	authLoginCmd.Flags().String("api-endpoint", "", "API endpoint host (required for uat)")
	authLoginCmd.MarkFlagRequired("region")

	authCmd.AddCommand(authLoginCmd)
	authCmd.AddCommand(authStatusCmd)
}

func runAuthLogin(cmd *cobra.Command, args []string) error {
	regionStr, _ := cmd.Flags().GetString("region")
	envStr, _ := cmd.Flags().GetString("env")
	name, _ := cmd.Flags().GetString("name")

	region := config.Region(regionStr)
	if !region.IsValid() {
		output.PrintError("INVALID_REGION", fmt.Sprintf("Invalid region %q. Valid: us, hk, jp", regionStr), 0)
		return errSilent
	}
	if !region.IsSupported() {
		output.PrintError("UNSUPPORTED_REGION", fmt.Sprintf("Region %q is not yet supported. Currently available: us, hk, jp", regionStr), 0)
		return errSilent
	}

	env := config.Environment(envStr)
	if env != config.EnvProd && env != config.EnvUAT {
		output.PrintError("INVALID_ENV", fmt.Sprintf("Invalid env %q. Valid: prod, uat", envStr), 0)
		return errSilent
	}

	if name == "" {
		name = "default"
	}

	// Check if profile already exists
	pm := config.NewProfileManager("")
	cfg, err := pm.Load()
	if err != nil {
		return err
	}
	if _, exists := cfg.Profiles[name]; exists {
		output.PrintError("PROFILE_EXISTS", fmt.Sprintf("Profile %q already exists. Use --name to specify a different name.", name), 0)
		return errSilent
	}

	// Resolve endpoints early so we can fail fast before asking for credentials
	apiEndpoint, _ := cmd.Flags().GetString("api-endpoint")

	// UAT requires --api-endpoint
	if env == config.EnvUAT {
		if apiEndpoint == "" {
			output.PrintError("MISSING_ENDPOINT",
				"UAT environment requires --api-endpoint", 0)
			return errSilent
		}
	}

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
	// Save terminal state before ReadPassword so we can restore on interrupt
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
		fmt.Fprintln(os.Stderr) // newline after hidden input
	}

	if appKey == "" || appSecret == "" {
		output.PrintError("MISSING_CREDENTIALS", "App Key and App Secret are required", 0)
		return errSilent
	}

	// Resolve API host
	tempProfile := &config.Profile{
		Region: region,
		Env:    env,
		Endpoints: config.Endpoints{
			API: apiEndpoint,
		},
	}
	apiHost := auth.ResolveAPIHost(tempProfile)

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

	// Build profile (sensitive fields go to keychain, not to file)
	profile := &config.Profile{
		Name:       name,
		AuthMethod: "ak_sk",
		Region:     region,
		Env:        env,
		Endpoints: config.Endpoints{
			API: apiEndpoint,
		},
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
	if err := config.SaveCredential(name, cred); err != nil {
		output.PrintError("KEYCHAIN_ERROR", err.Error(), 0)
		return errSilent
	}

	// Save profile (non-sensitive data only)
	cfg.Profiles[name] = profile
	cfg.ActiveProfile = name

	if err := pm.Save(cfg); err != nil {
		return fmt.Errorf("failed to save profile: %w", err)
	}

	if !flagQuiet {
		fmt.Fprintf(os.Stderr, "✓ Profile %q created (region: %s, auth: ak_sk, 2fa: %v)\n",
			name, region, configResp.TokenCheckEnabled)
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
			"profile":     profile.Name,
			"auth_method": profile.AuthMethod,
			"region":      profile.Region,
			"env":         profile.Env,
		}

		if profile.AkSk != nil {
			status["token_check_enabled"] = profile.AkSk.TokenCheckEnabled
			// Load credential from keychain
			if cred, err := config.LoadCredential(profile.Name); err == nil {
				status["app_key"] = output.Mask(cred.AppKey)
				if cred.Token != "" {
					status["2fa_token"] = output.Mask(cred.Token)
				}
			}
		}

		return output.PrintJSON(status)
	},
}
