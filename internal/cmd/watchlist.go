// Copyright 2024-2026 Webull Inc. All rights reserved.
// Licensed under the Apache License, Version 2.0.
// See LICENSE file in the project root for full license information.

package cmd

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/webull-inc/webull-openapi-cli/internal/client"
	"github.com/webull-inc/webull-openapi-cli/internal/output"
)

var watchlistCmd = &cobra.Command{
	Use:   "watchlist",
	Short: "Manage watchlists",
}

func init() {
	watchlistCmd.AddCommand(watchlistListCmd)
	watchlistCmd.AddCommand(watchlistGetCmd)
	watchlistCmd.AddCommand(watchlistCreateCmd)
	watchlistCmd.AddCommand(watchlistUpdateCmd)
	watchlistCmd.AddCommand(watchlistDeleteCmd)
	watchlistCmd.AddCommand(watchlistAddCmd)
	watchlistCmd.AddCommand(watchlistRemoveCmd)
	watchlistCmd.AddCommand(watchlistSortCmd)

	watchlistGetCmd.Flags().String("id", "", "Watchlist ID (required)")
	watchlistGetCmd.MarkFlagRequired("id")

	watchlistCreateCmd.Flags().String("name", "", "Watchlist name (required)")
	watchlistCreateCmd.Flags().Int("sort", 0, "Sort order (optional)")
	watchlistCreateCmd.MarkFlagRequired("name")

	watchlistUpdateCmd.Flags().String("id", "", "Watchlist ID (required)")
	watchlistUpdateCmd.Flags().String("name", "", "New name")
	watchlistUpdateCmd.Flags().Int("sort", 0, "New sort order")
	watchlistUpdateCmd.MarkFlagRequired("id")

	watchlistDeleteCmd.Flags().String("id", "", "Watchlist ID (required)")
	watchlistDeleteCmd.MarkFlagRequired("id")

	watchlistAddCmd.Flags().String("id", "", "Watchlist ID (required)")
	watchlistAddCmd.Flags().String("body", "", "JSON body with instruments array")
	watchlistAddCmd.Flags().String("file", "", "Path to JSON file")
	watchlistAddCmd.MarkFlagRequired("id")

	watchlistRemoveCmd.Flags().String("id", "", "Watchlist ID (required)")
	watchlistRemoveCmd.Flags().StringSlice("symbol", nil, "Symbol(s) to remove, format: SYMBOL:CATEGORY")
	watchlistRemoveCmd.Flags().String("category", "US_STOCK", "Default category: US_STOCK, HK_STOCK")
	watchlistRemoveCmd.Flags().String("body", "", "JSON body with instruments array")
	watchlistRemoveCmd.Flags().String("file", "", "Path to JSON file")
	watchlistRemoveCmd.MarkFlagRequired("id")

	watchlistSortCmd.Flags().String("id", "", "Watchlist ID (required)")
	watchlistSortCmd.Flags().String("body", "", "JSON body with instruments sort order")
	watchlistSortCmd.Flags().String("file", "", "Path to JSON file")
	watchlistSortCmd.MarkFlagRequired("id")
}

var watchlistListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all watchlists",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newAPIClient()
		if err != nil {
			return err
		}
		resp, err := c.Do(context.Background(), "GET", "/openapi/market-data/watchlist/list", nil)
		if err != nil {
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

var watchlistGetCmd = &cobra.Command{
	Use:   "get",
	Short: "Get watchlist instruments",
	RunE: func(cmd *cobra.Command, args []string) error {
		id, _ := cmd.Flags().GetString("id")
		c, err := newAPIClient()
		if err != nil {
			return err
		}
		resp, err := c.Do(context.Background(), "GET", "/openapi/market-data/watchlist/instruments/list", &client.RequestOptions{
			Query: map[string]string{"watchlist_id": id},
		})
		if err != nil {
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

var watchlistCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a watchlist",
	RunE: func(cmd *cobra.Command, args []string) error {
		name, _ := cmd.Flags().GetString("name")
		sort, _ := cmd.Flags().GetInt("sort")
		c, err := newAPIClient()
		if err != nil {
			return err
		}
		body := map[string]interface{}{"name": name}
		if sort > 0 {
			body["sort"] = sort
		}
		resp, err := c.Do(context.Background(), "POST", "/openapi/market-data/watchlist/create", &client.RequestOptions{Body: body})
		if err != nil {
			return err
		}
		if resp.StatusCode != 200 {
			fmt.Fprintln(os.Stderr, string(resp.Body))
			return errSilent
		}
		output.PrintRawJSON(resp.Body)
		return nil
	},
}

var watchlistUpdateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update a watchlist",
	RunE: func(cmd *cobra.Command, args []string) error {
		id, _ := cmd.Flags().GetString("id")
		name, _ := cmd.Flags().GetString("name")
		sort, _ := cmd.Flags().GetInt("sort")
		c, err := newAPIClient()
		if err != nil {
			return err
		}
		body := map[string]interface{}{"watchlist_id": id}
		if name != "" {
			body["name"] = name
		}
		if sort > 0 {
			body["sort"] = sort
		}
		resp, err := c.Do(context.Background(), "POST", "/openapi/market-data/watchlist/update", &client.RequestOptions{Body: body})
		if err != nil {
			return err
		}
		if resp.StatusCode != 200 {
			fmt.Fprintln(os.Stderr, string(resp.Body))
			return errSilent
		}
		output.PrintRawJSON(resp.Body)
		return nil
	},
}

var watchlistDeleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete a watchlist",
	RunE: func(cmd *cobra.Command, args []string) error {
		id, _ := cmd.Flags().GetString("id")
		if !flagYes && !flagQuiet {
			fmt.Fprintf(os.Stderr, "Delete watchlist %s? [y/N]: ", id)
			var confirm string
			fmt.Scanln(&confirm)
			if confirm != "y" && confirm != "Y" {
				return nil
			}
		}
		c, err := newAPIClient()
		if err != nil {
			return err
		}
		resp, err := c.Do(context.Background(), "POST", "/openapi/market-data/watchlist/delete", &client.RequestOptions{
			Body: map[string]interface{}{"watchlist_id": id},
		})
		if err != nil {
			return err
		}
		if resp.StatusCode != 200 {
			fmt.Fprintln(os.Stderr, string(resp.Body))
			return errSilent
		}
		output.PrintRawJSON(resp.Body)
		return nil
	},
}

var watchlistAddCmd = &cobra.Command{
	Use:   "add",
	Short: "Add instruments to a watchlist",
	Long: `Add instruments via JSON input (sort is required per instrument).

Example:
  webull watchlist add --id 12345 --body '{"instruments":[{"symbol":"AAPL","category":"US_STOCK","sort":1},{"symbol":"00700","category":"HK_STOCK","sort":2}]}'
  webull watchlist add --id 12345 --file instruments.json`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return doWatchlistInstruments(cmd, "/openapi/market-data/watchlist/instruments/add")
	},
}

var watchlistRemoveCmd = &cobra.Command{
	Use:   "remove",
	Short: "Remove instruments from a watchlist",
	Long: `Remove instruments via flags or JSON input.

Flag mode (simple):
  webull watchlist remove --id 12345 --symbol AAPL:US_STOCK,00700:HK_STOCK

JSON mode:
  webull watchlist remove --id 12345 --body '{"instruments":[{"symbol":"AAPL","category":"US_STOCK"}]}'`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Check if JSON mode
		bodyStr, _ := cmd.Flags().GetString("body")
		fileStr, _ := cmd.Flags().GetString("file")
		if bodyStr != "" || fileStr != "" {
			return doWatchlistInstruments(cmd, "/openapi/market-data/watchlist/instruments/remove")
		}

		// Flag mode
		id, _ := cmd.Flags().GetString("id")
		symbols, _ := cmd.Flags().GetStringSlice("symbol")
		defaultCategory, _ := cmd.Flags().GetString("category")

		if len(symbols) == 0 {
			return fmt.Errorf("provide --symbol flags or use --body/--file for JSON input")
		}

		instruments := buildInstrumentsList(symbols, defaultCategory)

		c, err := newAPIClient()
		if err != nil {
			return err
		}
		resp, err := c.Do(context.Background(), "POST", "/openapi/market-data/watchlist/instruments/remove", &client.RequestOptions{
			Body: map[string]interface{}{"watchlist_id": id, "instruments": instruments},
		})
		if err != nil {
			return err
		}
		if resp.StatusCode != 200 {
			fmt.Fprintln(os.Stderr, string(resp.Body))
			return errSilent
		}
		output.PrintRawJSON(resp.Body)
		return nil
	},
}

var watchlistSortCmd = &cobra.Command{
	Use:   "sort",
	Short: "Update instruments sort order in a watchlist",
	Long: `Update sort order via JSON input (sort is required per instrument).

Example:
  webull watchlist sort --id 12345 --body '{"instruments":[{"symbol":"AAPL","category":"US_STOCK","sort":1},{"symbol":"00700","category":"HK_STOCK","sort":2}]}'`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return doWatchlistInstruments(cmd, "/openapi/market-data/watchlist/instruments/update")
	},
}

// doWatchlistInstruments handles add/sort/update via JSON input.
func doWatchlistInstruments(cmd *cobra.Command, path string) error {
	id, _ := cmd.Flags().GetString("id")

	body, err := readJSONInput(cmd)
	if err != nil {
		output.PrintError("INPUT_ERROR", err.Error(), 0)
		return err
	}
	// Ensure watchlist_id is set
	body["watchlist_id"] = id

	c, err := newAPIClient()
	if err != nil {
		return err
	}
	resp, err := c.Do(context.Background(), "POST", path, &client.RequestOptions{Body: body})
	if err != nil {
		return err
	}
	if resp.StatusCode != 200 {
		fmt.Fprintln(os.Stderr, string(resp.Body))
		return errSilent
	}
	output.PrintRawJSON(resp.Body)
	return nil
}

// buildInstrumentsList parses symbols with optional :CATEGORY suffix.
// e.g. "AAPL" uses default, "0700:HK_STOCK" uses HK_STOCK.
func buildInstrumentsList(symbols []string, defaultCategory string) []interface{} {
	instruments := make([]interface{}, 0, len(symbols))
	for _, s := range symbols {
		sym := s
		cat := defaultCategory
		if idx := strings.LastIndex(s, ":"); idx > 0 {
			sym = s[:idx]
			cat = s[idx+1:]
		}
		instruments = append(instruments, map[string]interface{}{"symbol": sym, "category": cat})
	}
	return instruments
}
