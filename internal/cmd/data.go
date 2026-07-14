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

var dataCmd = &cobra.Command{
	Use:   "data",
	Short: "Market data (snapshots, bars, quotes)",
}

var dataStockCmd = &cobra.Command{Use: "stock", Short: "Stock market data"}
var dataOptionCmd = &cobra.Command{Use: "option", Short: "Option market data"}
var dataFuturesCmd = &cobra.Command{Use: "futures", Short: "Futures market data"}
var dataCryptoCmd = &cobra.Command{Use: "crypto", Short: "Crypto market data"}
var dataEventCmd = &cobra.Command{Use: "event", Short: "Event contract market data"}
var dataScreenerCmd = &cobra.Command{Use: "screener", Short: "Stock screener (gainers/losers, most active)"}

// Top-level init: register sub-commands only
func init() {
	dataCmd.AddCommand(dataStockCmd)
	dataCmd.AddCommand(dataOptionCmd)
	dataCmd.AddCommand(dataFuturesCmd)
	dataCmd.AddCommand(dataCryptoCmd)
	dataCmd.AddCommand(dataEventCmd)
	dataCmd.AddCommand(dataScreenerCmd)

	dataStockCmd.AddCommand(dataStockSnapshotCmd)
	dataStockCmd.AddCommand(dataStockBarsCmd)
	dataStockCmd.AddCommand(dataStockBatchBarsCmd)
	dataStockCmd.AddCommand(dataStockQuotesCmd)
	dataStockCmd.AddCommand(dataStockTickCmd)
	dataStockCmd.AddCommand(dataStockFootprintCmd)
	dataStockCmd.AddCommand(dataStockNoiiBarsCmd)
	dataStockCmd.AddCommand(dataStockNoiiSnapshotCmd)

	dataScreenerCmd.AddCommand(dataScreenerMoversCmd)
	dataScreenerCmd.AddCommand(dataScreenerActivesCmd)

	dataOptionCmd.AddCommand(dataOptionSnapshotCmd)
	dataOptionCmd.AddCommand(dataOptionBarsCmd)
	dataOptionCmd.AddCommand(dataOptionTickCmd)

	dataFuturesCmd.AddCommand(dataFuturesSnapshotCmd)
	dataFuturesCmd.AddCommand(dataFuturesBarsCmd)
	dataFuturesCmd.AddCommand(dataFuturesTickCmd)
	dataFuturesCmd.AddCommand(dataFuturesFootprintCmd)
	dataFuturesCmd.AddCommand(dataFuturesDepthCmd)

	dataCryptoCmd.AddCommand(dataCryptoSnapshotCmd)
	dataCryptoCmd.AddCommand(dataCryptoBarsCmd)

	dataEventCmd.AddCommand(dataEventSnapshotCmd)
	dataEventCmd.AddCommand(dataEventBarsCmd)
	dataEventCmd.AddCommand(dataEventTickCmd)
	dataEventCmd.AddCommand(dataEventDepthCmd)
}

// ==================== Stock Data ====================

var dataStockSnapshotCmd = &cobra.Command{
	Use: "snapshot", Short: "Real-time stock snapshot",
	RunE: func(cmd *cobra.Command, args []string) error {
		symbol, _ := cmd.Flags().GetString("symbol")
		category, _ := cmd.Flags().GetString("category")
		extHour, _ := cmd.Flags().GetBool("extend-hour")
		overnight, _ := cmd.Flags().GetBool("overnight")
		q := map[string]string{"symbols": symbol, "category": category}
		if extHour {
			q["extend_hour_required"] = "true"
		} else {
			q["extend_hour_required"] = "false"
		}
		if overnight {
			q["overnight_required"] = "true"
		} else {
			q["overnight_required"] = "false"
		}
		return doDataGet("/openapi/market-data/stock/snapshot", q)
	},
}

func init() {
	dataStockSnapshotCmd.Flags().String("symbol", "", "Symbol(s), comma-separated (required)")
	dataStockSnapshotCmd.Flags().String("category", "US_STOCK", "Category: US_STOCK, US_ETF, HK_STOCK, CN_STOCK")
	dataStockSnapshotCmd.Flags().Bool("extend-hour", false, "Include pre/post market data")
	dataStockSnapshotCmd.Flags().Bool("overnight", false, "Include overnight data")
	dataStockSnapshotCmd.MarkFlagRequired("symbol")
}

var dataStockBarsCmd = &cobra.Command{
	Use: "bars", Short: "Historical candlestick bars",
	RunE: func(cmd *cobra.Command, args []string) error {
		symbol, _ := cmd.Flags().GetString("symbol")
		category, _ := cmd.Flags().GetString("category")
		timespan, _ := cmd.Flags().GetString("timespan")
		count, _ := cmd.Flags().GetInt("count")
		realTime, _ := cmd.Flags().GetBool("real-time")
		sessions, _ := cmd.Flags().GetString("trading-sessions")
		q := map[string]string{
			"symbol": symbol, "category": category,
			"timespan": timespan, "count": fmt.Sprintf("%d", count),
		}
		if realTime {
			q["real_time_required"] = "true"
		} else {
			q["real_time_required"] = "false"
		}
		if sessions != "" {
			q["trading_sessions"] = sessions
		}
		return doDataGet("/openapi/market-data/stock/bars", q)
	},
}

func init() {
	dataStockBarsCmd.Flags().String("symbol", "", "Symbol (required)")
	dataStockBarsCmd.Flags().String("category", "US_STOCK", "Category: US_STOCK, US_ETF, HK_STOCK, CN_STOCK")
	dataStockBarsCmd.Flags().String("timespan", "D", "Timespan: M1,M5,M15,M30,M60,M120,M240,D,W,M,Y")
	dataStockBarsCmd.Flags().Int("count", 200, "Number of bars (max 1200)")
	dataStockBarsCmd.Flags().Bool("real-time", true, "Include latest unfinalized data")
	dataStockBarsCmd.Flags().String("trading-sessions", "", "Trading sessions: PRE,RTH,ATH,OVN")
	dataStockBarsCmd.MarkFlagRequired("symbol")
}

var dataStockBatchBarsCmd = &cobra.Command{
	Use: "batch-bars", Short: "Historical bars for multiple symbols (batch)",
	RunE: func(cmd *cobra.Command, args []string) error {
		symbols, _ := cmd.Flags().GetString("symbol")
		category, _ := cmd.Flags().GetString("category")
		timespan, _ := cmd.Flags().GetString("timespan")
		count, _ := cmd.Flags().GetInt("count")
		realTime, _ := cmd.Flags().GetBool("real-time")
		sessions, _ := cmd.Flags().GetString("trading-sessions")

		// Convert comma-separated symbols to array
		symbolList := strings.Split(symbols, ",")
		for i := range symbolList {
			symbolList[i] = strings.TrimSpace(symbolList[i])
		}

		body := map[string]interface{}{
			"symbols":  symbolList,
			"category": category,
			"timespan": timespan,
			"count":    fmt.Sprintf("%d", count),
		}
		if realTime {
			body["real_time_required"] = "true"
		} else {
			body["real_time_required"] = "false"
		}
		if sessions != "" {
			body["trading_sessions"] = sessions
		}

		cl, err := newAPIClient()
		if err != nil {
			return err
		}
		resp, err := cl.Do(context.Background(), "POST", "/openapi/market-data/stock/batch-bars", &client.RequestOptions{
			Body: body,
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

func init() {
	dataStockBatchBarsCmd.Flags().String("symbol", "", "Symbol(s), comma-separated (required)")
	dataStockBatchBarsCmd.Flags().String("category", "US_STOCK", "Category: US_STOCK, US_ETF, HK_STOCK, CN_STOCK")
	dataStockBatchBarsCmd.Flags().String("timespan", "D", "Timespan: M1,M5,M15,M30,M60,M120,M240,D,W,M,Y")
	dataStockBatchBarsCmd.Flags().Int("count", 200, "Number of bars (max 1200)")
	dataStockBatchBarsCmd.Flags().Bool("real-time", true, "Include latest unfinalized data")
	dataStockBatchBarsCmd.Flags().String("trading-sessions", "", "Trading sessions: PRE,RTH,ATH,OVN")
	dataStockBatchBarsCmd.MarkFlagRequired("symbol")
}

var dataStockQuotesCmd = &cobra.Command{
	Use: "quotes", Short: "Stock depth quotes (bid/ask)",
	RunE: func(cmd *cobra.Command, args []string) error {
		symbol, _ := cmd.Flags().GetString("symbol")
		category, _ := cmd.Flags().GetString("category")
		depth, _ := cmd.Flags().GetString("depth")
		overnight, _ := cmd.Flags().GetBool("overnight")
		q := map[string]string{"symbol": symbol, "category": category, "depth": depth}
		if overnight {
			q["overnight_required"] = "true"
		} else {
			q["overnight_required"] = "false"
		}
		return doDataGet("/openapi/market-data/stock/quotes", q)
	},
}

func init() {
	dataStockQuotesCmd.Flags().String("symbol", "", "Symbol (required)")
	dataStockQuotesCmd.Flags().String("category", "US_STOCK", "Category: US_STOCK, US_ETF, HK_STOCK, CN_STOCK")
	dataStockQuotesCmd.Flags().String("depth", "10", "Market depth (L1=1, L2=10)")
	dataStockQuotesCmd.Flags().Bool("overnight", false, "Include overnight data")
	dataStockQuotesCmd.MarkFlagRequired("symbol")
}

var dataStockTickCmd = &cobra.Command{
	Use: "tick", Short: "Tick-by-tick trade data",
	RunE: func(cmd *cobra.Command, args []string) error {
		symbol, _ := cmd.Flags().GetString("symbol")
		category, _ := cmd.Flags().GetString("category")
		count, _ := cmd.Flags().GetInt("count")
		sessions, _ := cmd.Flags().GetString("trading-sessions")
		q := map[string]string{
			"symbol": symbol, "category": category,
			"count": fmt.Sprintf("%d", count), "trading_sessions": sessions,
		}
		return doDataGet("/openapi/market-data/stock/tick", q)
	},
}

func init() {
	dataStockTickCmd.Flags().String("symbol", "", "Symbol (required)")
	dataStockTickCmd.Flags().String("category", "US_STOCK", "Category: US_STOCK, US_ETF, HK_STOCK, CN_STOCK")
	dataStockTickCmd.Flags().Int("count", 30, "Number of ticks (max 1000)")
	dataStockTickCmd.Flags().String("trading-sessions", "RTH", "Trading sessions: PRE,RTH,ATH,OVN (comma-separated)")
	dataStockTickCmd.MarkFlagRequired("symbol")
}

var dataStockFootprintCmd = &cobra.Command{
	Use: "footprint", Short: "Large order footprint data",
	RunE: func(cmd *cobra.Command, args []string) error {
		symbol, _ := cmd.Flags().GetString("symbol")
		category, _ := cmd.Flags().GetString("category")
		timespan, _ := cmd.Flags().GetString("timespan")
		count, _ := cmd.Flags().GetInt("count")
		realTime, _ := cmd.Flags().GetBool("real-time")
		sessions, _ := cmd.Flags().GetString("trading-sessions")
		q := map[string]string{
			"symbols": symbol, "category": category,
			"timespan": timespan, "count": fmt.Sprintf("%d", count),
		}
		if realTime {
			q["real_time_required"] = "true"
		} else {
			q["real_time_required"] = "false"
		}
		if sessions != "" {
			q["trading_sessions"] = sessions
		}
		return doDataGet("/openapi/market-data/stock/footprint", q)
	},
}

func init() {
	dataStockFootprintCmd.Flags().String("symbol", "", "Symbol(s), comma-separated (required)")
	dataStockFootprintCmd.Flags().String("category", "US_STOCK", "Category: US_STOCK")
	dataStockFootprintCmd.Flags().String("timespan", "M1", "Timespan: S5, S15, M1, M5, M30")
	dataStockFootprintCmd.Flags().Int("count", 200, "Number of records (max 1200)")
	dataStockFootprintCmd.Flags().Bool("real-time", false, "Include latest unfinalized data")
	dataStockFootprintCmd.Flags().String("trading-sessions", "", "Trading sessions: PRE,RTH,ATH (OVN not supported)")
	dataStockFootprintCmd.MarkFlagRequired("symbol")
}

var dataStockNoiiBarsCmd = &cobra.Command{
	Use: "noii-bars", Short: "NOII (Net Order Imbalance) bars",
	RunE: func(cmd *cobra.Command, args []string) error {
		symbol, _ := cmd.Flags().GetString("symbol")
		category, _ := cmd.Flags().GetString("category")
		actionType, _ := cmd.Flags().GetString("action-type")
		return doDataGet("/openapi/market-data/stock/noii/bars", map[string]string{
			"symbol": symbol, "category": category, "imbalance_action_type": actionType,
		})
	},
}

func init() {
	dataStockNoiiBarsCmd.Flags().String("symbol", "", "Symbol (required)")
	dataStockNoiiBarsCmd.Flags().String("category", "US_STOCK", "Category: US_STOCK")
	dataStockNoiiBarsCmd.Flags().String("action-type", "PRE_OPEN", "Imbalance type: PRE_OPEN, PRE_CLOSE")
	dataStockNoiiBarsCmd.MarkFlagRequired("symbol")
}

var dataStockNoiiSnapshotCmd = &cobra.Command{
	Use: "noii-snapshot", Short: "NOII (Net Order Imbalance) snapshot",
	RunE: func(cmd *cobra.Command, args []string) error {
		symbol, _ := cmd.Flags().GetString("symbol")
		category, _ := cmd.Flags().GetString("category")
		actionType, _ := cmd.Flags().GetString("action-type")
		return doDataGet("/openapi/market-data/stock/noii/snapshot", map[string]string{
			"symbol": symbol, "category": category, "imbalance_action_type": actionType,
		})
	},
}

func init() {
	dataStockNoiiSnapshotCmd.Flags().String("symbol", "", "Symbol (required)")
	dataStockNoiiSnapshotCmd.Flags().String("category", "US_STOCK", "Category: US_STOCK")
	dataStockNoiiSnapshotCmd.Flags().String("action-type", "PRE_OPEN", "Imbalance type: PRE_OPEN, PRE_CLOSE")
	dataStockNoiiSnapshotCmd.MarkFlagRequired("symbol")
}

// ==================== Screener ====================

var dataScreenerMoversCmd = &cobra.Command{
	Use: "gainers-losers", Short: "Top gainers and losers",
	RunE: func(cmd *cobra.Command, args []string) error {
		rankType, _ := cmd.Flags().GetString("rank-type")
		category, _ := cmd.Flags().GetString("category")
		sortBy, _ := cmd.Flags().GetString("sort-by")
		direction, _ := cmd.Flags().GetString("direction")
		pageSize, _ := cmd.Flags().GetInt("page-size")
		pageIndex, _ := cmd.Flags().GetInt("page-index")
		q := map[string]string{
			"rank_type": rankType, "category": category, "sort_by": sortBy,
			"direction": direction, "page_size": fmt.Sprintf("%d", pageSize),
			"page_index": fmt.Sprintf("%d", pageIndex),
		}
		return doDataGet("/openapi/market-data/screener/gainers-losers", q)
	},
}

func init() {
	dataScreenerMoversCmd.Flags().String("rank-type", "DAY_1", "Rank type: PRE_MARKET,AFTER_MARKET,MIN_3,MIN_5,DAY_1,DAY_5,MONTH_1,MONTH_3,WEEK_52")
	dataScreenerMoversCmd.Flags().String("category", "US_STOCK", "Category: US_STOCK")
	dataScreenerMoversCmd.Flags().String("sort-by", "CHANGE_RATIO", "Sort by: CHANGE_RATIO,VOLUME,MARKET_VALUE,etc.")
	dataScreenerMoversCmd.Flags().String("direction", "DESC", "Direction: DESC (Gainers), ASC (Losers)")
	dataScreenerMoversCmd.Flags().Int("page-size", 20, "Page size")
	dataScreenerMoversCmd.Flags().Int("page-index", 1, "Page number")
}

var dataScreenerActivesCmd = &cobra.Command{
	Use: "actives", Short: "Most active stocks",
	RunE: func(cmd *cobra.Command, args []string) error {
		rankType, _ := cmd.Flags().GetString("rank-type")
		category, _ := cmd.Flags().GetString("category")
		sortBy, _ := cmd.Flags().GetString("sort-by")
		direction, _ := cmd.Flags().GetString("direction")
		pageSize, _ := cmd.Flags().GetInt("page-size")
		pageIndex, _ := cmd.Flags().GetInt("page-index")
		q := map[string]string{
			"rank_type": rankType, "category": category, "sort_by": sortBy,
			"direction": direction, "page_size": fmt.Sprintf("%d", pageSize),
			"page_index": fmt.Sprintf("%d", pageIndex),
		}
		return doDataGet("/openapi/market-data/screener/top-active", q)
	},
}

func init() {
	dataScreenerActivesCmd.Flags().String("rank-type", "VOLUME", "Rank type: VOLUME,RELATIVE_VOLUME_10D,TURNOVER,TURNOVER_RATE,AMPLITUDE")
	dataScreenerActivesCmd.Flags().String("category", "US_STOCK", "Category: US_STOCK")
	dataScreenerActivesCmd.Flags().String("sort-by", "VOLUME", "Sort by: VOLUME,CHANGE_RATIO,MARKET_VALUE,etc.")
	dataScreenerActivesCmd.Flags().String("direction", "DESC", "Direction: ASC, DESC")
	dataScreenerActivesCmd.Flags().Int("page-size", 20, "Page size")
	dataScreenerActivesCmd.Flags().Int("page-index", 1, "Page number")
}

// ==================== Option Data ====================

var dataOptionSnapshotCmd = &cobra.Command{
	Use: "snapshot", Short: "Option real-time snapshot",
	RunE: func(cmd *cobra.Command, args []string) error {
		symbol, _ := cmd.Flags().GetString("symbol")
		category, _ := cmd.Flags().GetString("category")
		return doDataGet("/openapi/market-data/option/snapshot", map[string]string{
			"symbols": symbol, "category": category,
		})
	},
}

func init() {
	dataOptionSnapshotCmd.Flags().String("symbol", "", "Option symbol(s), comma-separated (required)")
	dataOptionSnapshotCmd.Flags().String("category", "US_OPTION", "Category: US_OPTION")
	dataOptionSnapshotCmd.MarkFlagRequired("symbol")
}

var dataOptionBarsCmd = &cobra.Command{
	Use: "bars", Short: "Option historical bars",
	RunE: func(cmd *cobra.Command, args []string) error {
		symbol, _ := cmd.Flags().GetString("symbol")
		category, _ := cmd.Flags().GetString("category")
		timespan, _ := cmd.Flags().GetString("timespan")
		count, _ := cmd.Flags().GetInt("count")
		realTime, _ := cmd.Flags().GetBool("real-time")
		q := map[string]string{
			"symbols": symbol, "category": category,
			"timespan": timespan, "count": fmt.Sprintf("%d", count),
		}
		if realTime {
			q["real_time_required"] = "true"
		} else {
			q["real_time_required"] = "false"
		}
		return doDataGet("/openapi/market-data/option/bars", q)
	},
}

func init() {
	dataOptionBarsCmd.Flags().String("symbol", "", "Option symbol(s), comma-separated (required)")
	dataOptionBarsCmd.Flags().String("category", "US_OPTION", "Category: US_OPTION")
	dataOptionBarsCmd.Flags().String("timespan", "D", "Timespan: M1,M5,M15,M30,H1,D,W")
	dataOptionBarsCmd.Flags().Int("count", 200, "Number of bars (max 1200)")
	dataOptionBarsCmd.Flags().Bool("real-time", true, "Include latest in-progress bar")
	dataOptionBarsCmd.MarkFlagRequired("symbol")
}

var dataOptionTickCmd = &cobra.Command{
	Use: "tick", Short: "Option tick data",
	RunE: func(cmd *cobra.Command, args []string) error {
		symbol, _ := cmd.Flags().GetString("symbol")
		category, _ := cmd.Flags().GetString("category")
		count, _ := cmd.Flags().GetInt("count")
		return doDataGet("/openapi/market-data/option/tick", map[string]string{
			"symbol": symbol, "category": category, "count": fmt.Sprintf("%d", count),
		})
	},
}

func init() {
	dataOptionTickCmd.Flags().String("symbol", "", "Option symbol (required)")
	dataOptionTickCmd.Flags().String("category", "US_OPTION", "Category: US_OPTION")
	dataOptionTickCmd.Flags().Int("count", 50, "Number of ticks")
	dataOptionTickCmd.MarkFlagRequired("symbol")
}

// ==================== Futures Data ====================

var dataFuturesSnapshotCmd = &cobra.Command{
	Use: "snapshot", Short: "Futures real-time snapshot",
	RunE: func(cmd *cobra.Command, args []string) error {
		symbol, _ := cmd.Flags().GetString("symbol")
		category, _ := cmd.Flags().GetString("category")
		return doDataGet("/openapi/market-data/futures/snapshot", map[string]string{
			"symbols": symbol, "category": category,
		})
	},
}

func init() {
	dataFuturesSnapshotCmd.Flags().String("symbol", "", "Futures symbol(s), comma-separated (required)")
	dataFuturesSnapshotCmd.Flags().String("category", "US_FUTURES", "Category: US_FUTURES")
	dataFuturesSnapshotCmd.MarkFlagRequired("symbol")
}

var dataFuturesBarsCmd = &cobra.Command{
	Use: "bars", Short: "Futures historical bars",
	RunE: func(cmd *cobra.Command, args []string) error {
		symbol, _ := cmd.Flags().GetString("symbol")
		category, _ := cmd.Flags().GetString("category")
		timespan, _ := cmd.Flags().GetString("timespan")
		count, _ := cmd.Flags().GetInt("count")
		return doDataGet("/openapi/market-data/futures/bars", map[string]string{
			"symbols": symbol, "category": category, "timespan": timespan, "count": fmt.Sprintf("%d", count),
		})
	},
}

func init() {
	dataFuturesBarsCmd.Flags().String("symbol", "", "Futures symbol (required)")
	dataFuturesBarsCmd.Flags().String("category", "US_FUTURES", "Category: US_FUTURES")
	dataFuturesBarsCmd.Flags().String("timespan", "D", "Timespan")
	dataFuturesBarsCmd.Flags().Int("count", 30, "Number of bars")
	dataFuturesBarsCmd.MarkFlagRequired("symbol")
}

var dataFuturesTickCmd = &cobra.Command{
	Use: "tick", Short: "Futures tick data",
	RunE: func(cmd *cobra.Command, args []string) error {
		symbol, _ := cmd.Flags().GetString("symbol")
		category, _ := cmd.Flags().GetString("category")
		count, _ := cmd.Flags().GetInt("count")
		return doDataGet("/openapi/market-data/futures/tick", map[string]string{
			"symbol": symbol, "category": category, "count": fmt.Sprintf("%d", count),
		})
	},
}

func init() {
	dataFuturesTickCmd.Flags().String("symbol", "", "Futures symbol (required)")
	dataFuturesTickCmd.Flags().String("category", "US_FUTURES", "Category: US_FUTURES")
	dataFuturesTickCmd.Flags().Int("count", 200, "Number of ticks (max 1200)")
	dataFuturesTickCmd.MarkFlagRequired("symbol")
}

var dataFuturesFootprintCmd = &cobra.Command{
	Use: "footprint", Short: "Futures footprint data",
	RunE: func(cmd *cobra.Command, args []string) error {
		symbol, _ := cmd.Flags().GetString("symbol")
		category, _ := cmd.Flags().GetString("category")
		timespan, _ := cmd.Flags().GetString("timespan")
		count, _ := cmd.Flags().GetInt("count")
		realTime, _ := cmd.Flags().GetBool("real-time")
		sessions, _ := cmd.Flags().GetString("trading-sessions")
		q := map[string]string{
			"symbols": symbol, "category": category,
			"timespan": timespan, "count": fmt.Sprintf("%d", count),
		}
		if realTime {
			q["real_time_required"] = "true"
		} else {
			q["real_time_required"] = "false"
		}
		if sessions != "" {
			q["trading_sessions"] = sessions
		}
		return doDataGet("/openapi/market-data/futures/footprint", q)
	},
}

func init() {
	dataFuturesFootprintCmd.Flags().String("symbol", "", "Futures symbol(s), comma-separated (required)")
	dataFuturesFootprintCmd.Flags().String("category", "US_FUTURES", "Category: US_FUTURES")
	dataFuturesFootprintCmd.Flags().String("timespan", "M1", "Timespan: S5, S15, M1, M5, M30")
	dataFuturesFootprintCmd.Flags().Int("count", 200, "Number of records (max 1200)")
	dataFuturesFootprintCmd.Flags().Bool("real-time", false, "Include latest unfinalized data")
	dataFuturesFootprintCmd.Flags().String("trading-sessions", "", "Trading sessions: PRE,RTH,ATH")
	dataFuturesFootprintCmd.MarkFlagRequired("symbol")
}

var dataFuturesDepthCmd = &cobra.Command{
	Use: "depth", Short: "Futures depth of book (L2)",
	RunE: func(cmd *cobra.Command, args []string) error {
		symbol, _ := cmd.Flags().GetString("symbol")
		category, _ := cmd.Flags().GetString("category")
		depth, _ := cmd.Flags().GetString("depth")
		return doDataGet("/openapi/market-data/futures/depth", map[string]string{
			"symbol": symbol, "category": category, "depth": depth,
		})
	},
}

func init() {
	dataFuturesDepthCmd.Flags().String("symbol", "", "Futures symbol (required)")
	dataFuturesDepthCmd.Flags().String("category", "US_FUTURES", "Category: US_FUTURES")
	dataFuturesDepthCmd.Flags().String("depth", "10", "Depth levels 1-10")
	dataFuturesDepthCmd.MarkFlagRequired("symbol")
}

// ==================== Crypto Data ====================

var dataCryptoSnapshotCmd = &cobra.Command{
	Use: "snapshot", Short: "Crypto real-time snapshot",
	RunE: func(cmd *cobra.Command, args []string) error {
		symbol, _ := cmd.Flags().GetString("symbol")
		category, _ := cmd.Flags().GetString("category")
		return doDataGet("/openapi/market-data/crypto/snapshot", map[string]string{
			"symbols": symbol, "category": category,
		})
	},
}

func init() {
	dataCryptoSnapshotCmd.Flags().String("symbol", "", "Crypto symbol(s), comma-separated (required)")
	dataCryptoSnapshotCmd.Flags().String("category", "US_CRYPTO", "Category: US_CRYPTO")
	dataCryptoSnapshotCmd.MarkFlagRequired("symbol")
}

var dataCryptoBarsCmd = &cobra.Command{
	Use: "bars", Short: "Crypto historical bars",
	RunE: func(cmd *cobra.Command, args []string) error {
		symbol, _ := cmd.Flags().GetString("symbol")
		category, _ := cmd.Flags().GetString("category")
		timespan, _ := cmd.Flags().GetString("timespan")
		count, _ := cmd.Flags().GetInt("count")
		realTime, _ := cmd.Flags().GetBool("real-time")
		q := map[string]string{
			"symbols": symbol, "category": category,
			"timespan": timespan, "count": fmt.Sprintf("%d", count),
		}
		if realTime {
			q["real_time_required"] = "true"
		} else {
			q["real_time_required"] = "false"
		}
		return doDataGet("/openapi/market-data/crypto/bars", q)
	},
}

func init() {
	dataCryptoBarsCmd.Flags().String("symbol", "", "Crypto symbol(s), comma-separated (required)")
	dataCryptoBarsCmd.Flags().String("category", "US_CRYPTO", "Category: US_CRYPTO")
	dataCryptoBarsCmd.Flags().String("timespan", "D", "Timespan: M1,M5,M15,M30,M60,M120,M240,D,W,M,Y")
	dataCryptoBarsCmd.Flags().Int("count", 200, "Number of bars (max 1200)")
	dataCryptoBarsCmd.Flags().Bool("real-time", true, "Include latest in-progress bar")
	dataCryptoBarsCmd.MarkFlagRequired("symbol")
}

// ==================== Event Data ====================

var dataEventSnapshotCmd = &cobra.Command{
	Use: "snapshot", Short: "Event contract snapshot",
	RunE: func(cmd *cobra.Command, args []string) error {
		symbol, _ := cmd.Flags().GetString("symbol")
		category, _ := cmd.Flags().GetString("category")
		return doDataGet("/openapi/market-data/event/snapshot", map[string]string{
			"symbols": symbol, "category": category,
		})
	},
}

func init() {
	dataEventSnapshotCmd.Flags().String("symbol", "", "Event symbol (required)")
	dataEventSnapshotCmd.Flags().String("category", "US_EVENT", "Category: US_EVENT")
	dataEventSnapshotCmd.MarkFlagRequired("symbol")
}

var dataEventBarsCmd = &cobra.Command{
	Use: "bars", Short: "Event contract bars",
	RunE: func(cmd *cobra.Command, args []string) error {
		symbol, _ := cmd.Flags().GetString("symbol")
		category, _ := cmd.Flags().GetString("category")
		timespan, _ := cmd.Flags().GetString("timespan")
		count, _ := cmd.Flags().GetInt("count")
		realTime, _ := cmd.Flags().GetBool("real-time")
		q := map[string]string{
			"symbols": symbol, "category": category,
			"timespan": timespan, "count": fmt.Sprintf("%d", count),
		}
		if realTime {
			q["real_time_required"] = "true"
		} else {
			q["real_time_required"] = "false"
		}
		return doDataGet("/openapi/market-data/event/bars", q)
	},
}

func init() {
	dataEventBarsCmd.Flags().String("symbol", "", "Event symbol(s), comma-separated (required)")
	dataEventBarsCmd.Flags().String("category", "US_EVENT", "Category: US_EVENT")
	dataEventBarsCmd.Flags().String("timespan", "M5", "Timespan: M1,M5,M15,M30,M60,M120,M240,D")
	dataEventBarsCmd.Flags().Int("count", 200, "Number of bars (max 1200)")
	dataEventBarsCmd.Flags().Bool("real-time", false, "Include latest unfinalized data")
	dataEventBarsCmd.MarkFlagRequired("symbol")
}

var dataEventTickCmd = &cobra.Command{
	Use: "tick", Short: "Event contract tick data",
	RunE: func(cmd *cobra.Command, args []string) error {
		symbol, _ := cmd.Flags().GetString("symbol")
		category, _ := cmd.Flags().GetString("category")
		count, _ := cmd.Flags().GetInt("count")
		return doDataGet("/openapi/market-data/event/tick", map[string]string{
			"symbol": symbol, "category": category, "count": fmt.Sprintf("%d", count),
		})
	},
}

func init() {
	dataEventTickCmd.Flags().String("symbol", "", "Event symbol (required)")
	dataEventTickCmd.Flags().String("category", "US_EVENT", "Category: US_EVENT")
	dataEventTickCmd.Flags().Int("count", 30, "Number of ticks (max 1200)")
	dataEventTickCmd.MarkFlagRequired("symbol")
}

var dataEventDepthCmd = &cobra.Command{
	Use: "depth", Short: "Event contract order book depth",
	RunE: func(cmd *cobra.Command, args []string) error {
		symbol, _ := cmd.Flags().GetString("symbol")
		category, _ := cmd.Flags().GetString("category")
		depth, _ := cmd.Flags().GetString("depth")
		return doDataGet("/openapi/market-data/event/depth", map[string]string{
			"symbol": symbol, "category": category, "depth": depth,
		})
	},
}

func init() {
	dataEventDepthCmd.Flags().String("symbol", "", "Event symbol (required)")
	dataEventDepthCmd.Flags().String("category", "US_EVENT", "Category: US_EVENT")
	dataEventDepthCmd.Flags().String("depth", "10", "Depth levels")
	dataEventDepthCmd.MarkFlagRequired("symbol")
}

// ==================== Helper ====================

func doDataGet(path string, query map[string]string) error {
	cl, err := newAPIClient()
	if err != nil {
		return err
	}
	resp, err := cl.Do(context.Background(), "GET", path, &client.RequestOptions{
		Query: query,
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
}
