// Copyright 2024-2026 Webull Inc. All rights reserved.
// Licensed under the Apache License, Version 2.0.
// See LICENSE file in the project root for full license information.

package cmd

import (
	"github.com/spf13/cobra"
)

// ==================== New Screener Commands ====================

var dataScreenerMarketSectorsCmd = &cobra.Command{
	Use: "market-sectors", Short: "Market sector overview and performance rankings",
	RunE: func(cmd *cobra.Command, args []string) error {
		category, _ := cmd.Flags().GetString("category")
		aggType, _ := cmd.Flags().GetString("agg-type")
		period, _ := cmd.Flags().GetString("period")
		cursor, _ := cmd.Flags().GetString("pagination-key")
		direction, _ := cmd.Flags().GetString("direction")
		q := map[string]string{
			"category":  category,
			"agg_type":  aggType,
			"period":    period,
			"direction": direction,
		}
		if cursor != "" {
			q["pagination_key"] = cursor
		}
		return doDataGet("/market-data/screeners/market-sectors/list", q)
	},
}

func init() {
	dataScreenerMarketSectorsCmd.Flags().String("category", "US_STOCK", "Category: US_STOCK")
	dataScreenerMarketSectorsCmd.Flags().String("agg-type", "MARKET_VALUE", "Statistics type: MARKET_VALUE, VOLUME")
	dataScreenerMarketSectorsCmd.Flags().String("period", "D1", "Period: D1, D5, MO1, MO3")
	dataScreenerMarketSectorsCmd.Flags().String("pagination-key", "", "Pagination cursor")
	dataScreenerMarketSectorsCmd.Flags().String("direction", "ASC", "Direction: ASC, DESC")
}

var dataScreenerMarketSectorsDetailCmd = &cobra.Command{
	Use: "market-sectors-detail", Short: "Stock list and statistics for a specific sector",
	RunE: func(cmd *cobra.Command, args []string) error {
		sectorID, _ := cmd.Flags().GetString("sector-id")
		category, _ := cmd.Flags().GetString("category")
		period, _ := cmd.Flags().GetString("period")
		cursor, _ := cmd.Flags().GetString("pagination-key")
		sortBy, _ := cmd.Flags().GetString("sort-by")
		direction, _ := cmd.Flags().GetString("direction")
		q := map[string]string{
			"sector_id": sectorID,
			"category":  category,
			"period":    period,
			"sort_by":   sortBy,
			"direction": direction,
		}
		if cursor != "" {
			q["pagination_key"] = cursor
		}
		return doDataGet("/market-data/screeners/market-sectors/get", q)
	},
}

func init() {
	dataScreenerMarketSectorsDetailCmd.Flags().String("sector-id", "", "Sector ID (required)")
	dataScreenerMarketSectorsDetailCmd.Flags().String("category", "US_STOCK", "Category: US_STOCK")
	dataScreenerMarketSectorsDetailCmd.Flags().String("period", "D1", "Period: D1, D5, MO1, MO3")
	dataScreenerMarketSectorsDetailCmd.Flags().String("pagination-key", "", "Pagination cursor")
	dataScreenerMarketSectorsDetailCmd.Flags().String("sort-by", "CHANGE_RATIO", "Sort by: CHANGE_RATIO,VOLUME,MARKET_VALUE,etc.")
	dataScreenerMarketSectorsDetailCmd.Flags().String("direction", "ASC", "Direction: ASC, DESC")
	dataScreenerMarketSectorsDetailCmd.MarkFlagRequired("sector-id")
}

var dataScreenerHighDividendCmd = &cobra.Command{
	Use: "high-dividend", Short: "High dividend stocks ranking",
	RunE: func(cmd *cobra.Command, args []string) error {
		category, _ := cmd.Flags().GetString("category")
		sortBy, _ := cmd.Flags().GetString("sort-by")
		direction, _ := cmd.Flags().GetString("direction")
		q := map[string]string{
			"category":  category,
			"sort_by":   sortBy,
			"direction": direction,
		}
		return doDataGet("/market-data/screeners/high-dividend-ranks/list", q)
	},
}

func init() {
	dataScreenerHighDividendCmd.Flags().String("category", "US_STOCK", "Category: US_STOCK")
	dataScreenerHighDividendCmd.Flags().String("sort-by", "YIELD", "Sort by: YIELD,CHANGE_RATIO,MARKET_VALUE,etc.")
	dataScreenerHighDividendCmd.Flags().String("direction", "ASC", "Direction: ASC, DESC")
}

var dataScreener52WHLCmd = &cobra.Command{
	Use: "52whl", Short: "52-week high/low stocks ranking",
	RunE: func(cmd *cobra.Command, args []string) error {
		category, _ := cmd.Flags().GetString("category")
		rankType, _ := cmd.Flags().GetString("rank-type")
		sortBy, _ := cmd.Flags().GetString("sort-by")
		direction, _ := cmd.Flags().GetString("direction")
		q := map[string]string{
			"category":  category,
			"direction": direction,
		}
		if rankType != "" {
			q["rank_type"] = rankType
		}
		if sortBy != "" {
			q["sort_by"] = sortBy
		}
		return doDataGet("/market-data/screeners/week52-high-low/list", q)
	},
}

func init() {
	dataScreener52WHLCmd.Flags().String("category", "US_STOCK", "Category: US_STOCK")
	dataScreener52WHLCmd.Flags().String("rank-type", "NEW_HIGH", "Rank type: NEW_HIGH, NEAR_HIGH, NEW_LOW, NEAR_LOW")
	dataScreener52WHLCmd.Flags().String("sort-by", "CHANGE_RATIO_52W", "Sort by: CHANGE_RATIO_52W,VOLUME,MARKET_VALUE,etc.")
	dataScreener52WHLCmd.Flags().String("direction", "ASC", "Direction: ASC, DESC")
}
