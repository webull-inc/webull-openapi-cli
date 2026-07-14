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
)

var profileCmd = &cobra.Command{
	Use:   "profile",
	Short: "Manage authentication profiles",
}

func init() {
	profileCmd.AddCommand(profileListCmd)
	profileCmd.AddCommand(profileSwitchCmd)
	profileCmd.AddCommand(profileLogoutCmd)
	profileCmd.AddCommand(profileSetEndpointCmd)

	profileSetEndpointCmd.Flags().String("api-endpoint", "", "Override API endpoint")
}

var profileListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all profiles",
	RunE: func(cmd *cobra.Command, args []string) error {
		pm := config.NewProfileManager("")
		cfg, err := pm.Load()
		if err != nil {
			return err
		}

		profiles := make([]map[string]interface{}, 0)
		for _, p := range cfg.Profiles {
			profiles = append(profiles, map[string]interface{}{
				"name":        p.Name,
				"region":      p.Region,
				"env":         p.Env,
				"auth_method": p.AuthMethod,
				"active":      p.Name == cfg.ActiveProfile,
			})
		}

		return output.PrintJSON(profiles)
	},
}

var profileSwitchCmd = &cobra.Command{
	Use:   "switch <name>",
	Short: "Switch active profile",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]
		pm := config.NewProfileManager("")
		cfg, err := pm.Load()
		if err != nil {
			return err
		}

		if _, ok := cfg.Profiles[name]; !ok {
			output.PrintError("PROFILE_NOT_FOUND", fmt.Sprintf("Profile %q not found", name), 0)
			return fmt.Errorf("profile not found")
		}

		cfg.ActiveProfile = name
		if err := pm.Save(cfg); err != nil {
			return err
		}

		if !flagQuiet {
			p := cfg.Profiles[name]
			fmt.Fprintf(os.Stderr, "✓ Switched to profile %q (region: %s, auth: %s)\n", name, p.Region, p.AuthMethod)
		}
		return nil
	},
}

var profileLogoutCmd = &cobra.Command{
	Use:   "logout [name]",
	Short: "Remove a profile",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		pm := config.NewProfileManager("")
		cfg, err := pm.Load()
		if err != nil {
			return err
		}

		name := cfg.ActiveProfile
		if len(args) > 0 {
			name = args[0]
		}

		if _, ok := cfg.Profiles[name]; !ok {
			output.PrintError("PROFILE_NOT_FOUND", fmt.Sprintf("Profile %q not found", name), 0)
			return fmt.Errorf("profile not found")
		}

		delete(cfg.Profiles, name)
		// Remove credential from system keychain
		config.DeleteCredential(name)
		if cfg.ActiveProfile == name {
			cfg.ActiveProfile = ""
			// Set first remaining profile as active
			for k := range cfg.Profiles {
				cfg.ActiveProfile = k
				break
			}
		}

		if err := pm.Save(cfg); err != nil {
			return err
		}

		if !flagQuiet {
			fmt.Fprintf(os.Stderr, "✓ Profile %q removed\n", name)
		}
		return nil
	},
}

var profileSetEndpointCmd = &cobra.Command{
	Use:   "set-endpoint",
	Short: "Set custom endpoint override for the active profile",
	Long: `Set custom API endpoint override for the active profile.

Examples:
  webull profile set-endpoint --api-endpoint us-openapi-alb.uat.webullbroker.com
  webull profile set-endpoint --api-endpoint api.webull.com`,
	RunE: func(cmd *cobra.Command, args []string) error {
		apiEp, _ := cmd.Flags().GetString("api-endpoint")

		if apiEp == "" {
			output.PrintError("MISSING_PARAMETER", "--api-endpoint is required", 0)
			return fmt.Errorf("no endpoint specified")
		}

		pm := config.NewProfileManager("")
		cfg, err := pm.Load()
		if err != nil {
			return err
		}

		if cfg.ActiveProfile == "" {
			output.PrintError("NO_PROFILE", "no active profile", 0)
			return fmt.Errorf("no active profile")
		}

		profile, ok := cfg.Profiles[cfg.ActiveProfile]
		if !ok {
			return fmt.Errorf("active profile %q not found", cfg.ActiveProfile)
		}

		profile.Endpoints.API = apiEp

		if err := pm.Save(cfg); err != nil {
			return err
		}

		if !flagQuiet {
			fmt.Fprintf(os.Stderr, "✓ Endpoint updated for profile %q\n", profile.Name)
			fmt.Fprintf(os.Stderr, "  api: %s\n", apiEp)
		}
		return nil
	},
}
