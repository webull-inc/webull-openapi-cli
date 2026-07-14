// Copyright 2024-2026 Webull Inc. All rights reserved.
// Licensed under the Apache License, Version 2.0.
// See LICENSE file in the project root for full license information.

package cmd

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"
	"github.com/webull-inc/webull-openapi-cli/internal/auth"
	"github.com/webull-inc/webull-openapi-cli/internal/client"
	"github.com/webull-inc/webull-openapi-cli/internal/config"
	"github.com/webull-inc/webull-openapi-cli/internal/output"
)

var accountCmd = &cobra.Command{
	Use:   "account",
	Short: "Account information and positions",
}

func init() {
	accountCmd.AddCommand(accountListCmd)
	accountCmd.AddCommand(accountBalanceCmd)
	accountCmd.AddCommand(accountPositionsCmd)
	accountCmd.AddCommand(accountPositionDetailCmd)

	accountBalanceCmd.Flags().String("account-id", "", "Account ID (required)")
	accountBalanceCmd.MarkFlagRequired("account-id")

	accountPositionsCmd.Flags().String("account-id", "", "Account ID (required)")
	accountPositionsCmd.MarkFlagRequired("account-id")

	accountPositionDetailCmd.Flags().String("account-id", "", "Account ID (required)")
	accountPositionDetailCmd.Flags().String("instrument-id", "", "Instrument ID (required)")
	accountPositionDetailCmd.Flags().String("last-id", "", "Last position ID for pagination")
	accountPositionDetailCmd.Flags().Int("page-size", 10, "Page size (max 100)")
	accountPositionDetailCmd.MarkFlagRequired("account-id")
	accountPositionDetailCmd.MarkFlagRequired("instrument-id")
}

func newAPIClient() (*client.APIClient, error) {
	pm := config.NewProfileManager("")
	profile, err := pm.GetActive()
	if err != nil {
		return nil, err
	}

	var appKey, appSecret, token string

	// Environment variable credentials override profile
	if envKey, envSecret, ok := config.EnvCredentials(); ok {
		appKey = envKey
		appSecret = envSecret
	} else {
		// Load credentials from system keychain
		cred, err := config.LoadCredential(profile.Name)
		if err != nil {
			return nil, fmt.Errorf("failed to load credentials: %w. Run 'webull auth login' to re-authenticate", err)
		}
		appKey = cred.AppKey
		appSecret = cred.AppSecret
		token = cred.Token
	}

	// Resolve endpoints: env overrides > profile overrides > defaults
	envOverrides := config.EnvEndpointOverrides()
	var overrides *config.RegionEndpoints
	if profile.Endpoints.API != "" {
		overrides = &config.RegionEndpoints{
			API: profile.Endpoints.API,
		}
	}
	endpoints := config.GetEndpoints(profile.Region, overrides)
	// Env overrides take highest priority
	if envOverrides != nil {
		if envOverrides.API != "" {
			endpoints.API = envOverrides.API
		}
	}

	// Create credential provider
	provider := auth.NewAkSkProvider(profile, pm, appKey, appSecret, token)

	timeout := time.Duration(flagTimeout) * time.Second
	apiClient := client.NewAPIClient(provider, endpoints, timeout)
	apiClient.SetVerbose(flagVerbose)
	apiClient.SetDebug(flagDebug)

	return apiClient, nil
}

var accountListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all trading accounts",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newAPIClient()
		if err != nil {
			output.PrintError("CONFIG_ERROR", err.Error(), 0)
			return err
		}

		resp, err := c.Do(context.Background(), "GET", "/openapi/account/list", nil)
		if err != nil {
			output.PrintError("REQUEST_ERROR", err.Error(), 0)
			return err
		}

		if resp.StatusCode != 200 {
			fmt.Fprintln(os.Stderr, string(resp.Body))
			return errSilent
		}

		if flagCSV {
			return output.PrintCSV(resp.Body)
		}
		output.PrintRawJSON(resp.Body)
		return nil
	},
}

var accountBalanceCmd = &cobra.Command{
	Use:   "balance",
	Short: "Account balance (equity, buying power, cash, P&L)",
	RunE: func(cmd *cobra.Command, args []string) error {
		accountID, _ := cmd.Flags().GetString("account-id")

		c, err := newAPIClient()
		if err != nil {
			output.PrintError("CONFIG_ERROR", err.Error(), 0)
			return err
		}

		resp, err := c.Do(context.Background(), "GET", "/openapi/assets/balance", &client.RequestOptions{
			Query: map[string]string{"account_id": accountID},
		})
		if err != nil {
			output.PrintError("REQUEST_ERROR", err.Error(), 0)
			return err
		}

		if resp.StatusCode != 200 {
			fmt.Fprintln(os.Stderr, string(resp.Body))
			return errSilent
		}

		if flagCSV {
			return output.PrintCSV(resp.Body)
		}
		output.PrintRawJSON(resp.Body)
		return nil
	},
}

var accountPositionsCmd = &cobra.Command{
	Use:   "positions",
	Short: "Account positions",
	RunE: func(cmd *cobra.Command, args []string) error {
		accountID, _ := cmd.Flags().GetString("account-id")

		c, err := newAPIClient()
		if err != nil {
			output.PrintError("CONFIG_ERROR", err.Error(), 0)
			return err
		}

		resp, err := c.Do(context.Background(), "GET", "/openapi/assets/positions", &client.RequestOptions{
			Query: map[string]string{"account_id": accountID},
		})
		if err != nil {
			output.PrintError("REQUEST_ERROR", err.Error(), 0)
			return err
		}

		if resp.StatusCode != 200 {
			fmt.Fprintln(os.Stderr, string(resp.Body))
			return errSilent
		}

		if flagCSV {
			return output.PrintCSV(resp.Body)
		}
		output.PrintRawJSON(resp.Body)
		return nil
	},
}

var accountPositionDetailCmd = &cobra.Command{
	Use:   "position-detail",
	Short: "Position details by contract (JP only)",
	Long:  `Query position details at the contract level. Only available for JP region.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		accountID, _ := cmd.Flags().GetString("account-id")
		instrumentID, _ := cmd.Flags().GetString("instrument-id")
		lastID, _ := cmd.Flags().GetString("last-id")
		pageSize, _ := cmd.Flags().GetInt("page-size")

		c, err := newAPIClient()
		if err != nil {
			output.PrintError("CONFIG_ERROR", err.Error(), 0)
			return errSilent
		}

		q := map[string]string{
			"account_id":    accountID,
			"instrument_id": instrumentID,
			"page_size":     fmt.Sprintf("%d", pageSize),
		}
		if lastID != "" {
			q["last_id"] = lastID
		}

		resp, err := c.Do(context.Background(), "GET", "/openapi/assets/position/details", &client.RequestOptions{
			Query: q,
		})
		if err != nil {
			output.PrintError("REQUEST_ERROR", err.Error(), 0)
			return errSilent
		}

		if resp.StatusCode != 200 {
			fmt.Fprintln(os.Stderr, string(resp.Body))
			return errSilent
		}

		if flagCSV {
			return output.PrintCSV(resp.Body)
		}
		output.PrintRawJSON(resp.Body)
		return nil
	},
}
