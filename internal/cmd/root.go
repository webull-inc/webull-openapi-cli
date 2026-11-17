// Copyright 2024-2026 Webull Inc. All rights reserved.
// Licensed under the Apache License, Version 2.0.
// See LICENSE file in the project root for full license information.

package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/webull-inc/webull-openapi-cli/internal/config"
	"github.com/webull-inc/webull-openapi-cli/internal/output"
	"github.com/webull-inc/webull-openapi-cli/pkg/version"
)

var (
	flagCSV     bool
	flagQuiet   bool
	flagVerbose bool
	flagDebug   bool
	flagTimeout int
	flagYes     bool
)

var rootCmd = &cobra.Command{
	Use:   "webull",
	Short: "Webull OpenAPI CLI",
	Long: `Webull OpenAPI CLI - Trade stocks, options, futures, crypto, and event contracts,
access market data, and manage your account from the terminal.

Designed for developers, quantitative teams, and AI agents.`,
	SilenceUsage:  true,
	SilenceErrors: true,
}

func init() {
	// Global flags
	rootCmd.PersistentFlags().BoolVar(&flagCSV, "csv", false, "Output in CSV format")
	rootCmd.PersistentFlags().BoolVarP(&flagQuiet, "quiet", "q", false, "Suppress non-data output")
	rootCmd.PersistentFlags().BoolVar(&flagVerbose, "verbose", false, "Show request summaries on stderr")
	rootCmd.PersistentFlags().BoolVar(&flagDebug, "debug", false, "Show full request/response on stderr")
	rootCmd.PersistentFlags().IntVar(&flagTimeout, "timeout", 30, "HTTP timeout in seconds")
	rootCmd.PersistentFlags().BoolVarP(&flagYes, "yes", "y", false, "Skip confirmation prompts")

	// Register sub-commands
	rootCmd.AddCommand(authCmd)
	rootCmd.AddCommand(accountCmd)
	rootCmd.AddCommand(orderCmd)
	rootCmd.AddCommand(dataCmd)
	rootCmd.AddCommand(instrumentCmd)
	rootCmd.AddCommand(watchlistCmd)
	rootCmd.AddCommand(communityCmd)
	rootCmd.AddCommand(versionCmd)
	rootCmd.AddCommand(doctorCmd)
}

// Execute runs the root command.
func Execute() error {
	err := rootCmd.Execute()
	if err != nil {
		// silentError means error was already printed by the command
		if _, ok := err.(silentError); ok {
			return err
		}
		// All other errors (cobra flag errors, unknown commands, etc.) → JSON to stderr
		output.PrintError("CLI_ERROR", err.Error(), 0)
	}
	return err
}

// --- Version command ---
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print version information",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Fprintf(os.Stdout, `{"version":"%s","git_commit":"%s","build_time":"%s","go_version":"%s"}`+"\n",
			version.Version, version.GitCommit, version.BuildTime, version.GoVersion)
	},
}

// --- Doctor command ---
var doctorCmd = &cobra.Command{
	Use:   "doctor",
	Short: "Diagnose configuration and connectivity",
	RunE: func(cmd *cobra.Command, args []string) error {
		result := map[string]interface{}{}

		// Check config file
		pm := config.NewProfileManager("")
		cfg, err := pm.Load()
		if err != nil {
			result["config_file"] = map[string]interface{}{"status": "error", "message": err.Error()}
		} else {
			result["config_file"] = map[string]interface{}{"status": "ok", "path": config.DefaultConfigPath()}
		}

		// Check the stored profile
		if cfg != nil && cfg.Profile != nil {
			profile := cfg.Profile
			result["auth"] = buildDoctorAuth(profile)

			// Resolved API host
			apiHost := config.ResolveAPIEndpoint(profile.Region, profile.Env, profile.Endpoints.API)
			result["network"] = map[string]interface{}{"api_host": apiHost}
		} else {
			result["auth"] = map[string]interface{}{"status": "none", "message": "not authenticated"}
		}

		// Check env vars
		envVars := map[string]interface{}{}
		if _, _, ok := config.EnvCredentials(); ok {
			envVars["WEBULL_APP_KEY"] = "set"
			envVars["WEBULL_APP_SECRET"] = "set"
		}
		if len(envVars) > 0 {
			result["environment"] = envVars
		}

		return output.PrintJSON(result)
	},
}

// buildDoctorAuth summarizes the stored profile for the doctor command.
func buildDoctorAuth(profile *config.Profile) map[string]interface{} {
	auth := map[string]interface{}{
		"status":      "ok",
		"auth_method": profile.AuthMethod,
		"region":      profile.Region,
		"env":         profile.Env,
	}

	if profile.AuthMethod == config.AuthMethodAkSk && profile.AkSk != nil {
		auth["token_check_enabled"] = profile.AkSk.TokenCheckEnabled
		if cred, err := config.LoadCredential(); err == nil {
			auth["app_key"] = output.Mask(cred.AppKey)
			if cred.Token != "" {
				auth["token"] = output.Mask(cred.Token)
				auth["token_status"] = profile.AkSk.TokenStatus
			}
		}
	}
	return auth
}
