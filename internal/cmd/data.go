// Copyright 2024-2026 Webull Inc. All rights reserved.
// Licensed under the Apache License, Version 2.0.
// See LICENSE file in the project root for full license information.

package cmd

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/webull-inc/webull-openapi-cli/internal/auth"
	"github.com/webull-inc/webull-openapi-cli/internal/client"
	"github.com/webull-inc/webull-openapi-cli/internal/config"
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
var data13FCmd = &cobra.Command{Use: "13f", Short: "US 13F data (institutions, holdings, transactions)"}

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
	data13FCmd.AddCommand(data13FInstitutionsSearchCmd)
	data13FCmd.AddCommand(data13FHoldingsCmd)
	data13FCmd.AddCommand(data13FTransactionsCmd)

	dataScreenerCmd.AddCommand(dataScreenerMoversCmd)
	dataScreenerCmd.AddCommand(dataScreenerActivesCmd)
	dataScreenerCmd.AddCommand(dataScreenerMarketSectorsCmd)
	dataScreenerCmd.AddCommand(dataScreenerMarketSectorsDetailCmd)
	dataScreenerCmd.AddCommand(dataScreenerHighDividendCmd)
	dataScreenerCmd.AddCommand(dataScreener52WHLCmd)

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

var dataMarketSentimentCmd = &cobra.Command{
	Use: "market-sentiment", Short: "US market fear and greed sentiment index",
	RunE: func(cmd *cobra.Command, args []string) error {
		return doUSDataGet("/market-data/screeners/fear-greed-indexes/get", map[string]string{
			"category": "US_STOCK",
		})
	},
}

var dataUnusualActivitiesCmd = &cobra.Command{
	Use: "unusual-activities", Short: "US stock unusual activity alerts",
	RunE: func(cmd *cobra.Command, args []string) error {
		symbol, _ := cmd.Flags().GetString("symbol")
		count, _ := cmd.Flags().GetInt("count")
		alertTypes, _ := cmd.Flags().GetString("alert-types")
		extendedHours, _ := cmd.Flags().GetBool("extended-hours")
		if count < 1 || count > 30 {
			return fmt.Errorf("count must be between 1 and 30")
		}
		q := map[string]string{
			"category": "US_STOCK", "count": fmt.Sprintf("%d", count),
			"extended_hours_required": fmt.Sprintf("%t", extendedHours),
		}
		if symbol != "" {
			q["symbol"] = symbol
		}
		if alertTypes != "" {
			q["alert_types"] = alertTypes
		}
		return doUSDataGet("/market-data/screeners/unusual-activities/list", q)
	},
}

var data13FInstitutionsSearchCmd = &cobra.Command{
	Use: "search", Short: "Search US institutional investors",
	RunE: func(cmd *cobra.Command, args []string) error {
		keyword, _ := cmd.Flags().GetString("keyword")
		keyword = strings.TrimSpace(keyword)
		if keyword == "" {
			return fmt.Errorf("keyword must not be empty")
		}
		cursor, _ := cmd.Flags().GetString("pagination-key")
		q := map[string]string{"keyword": keyword}
		if cursor != "" {
			q["pagination_key"] = cursor
		}
		return doUSDataGet("/market-data/fundamentals/institutions/list", q)
	},
}

var data13FHoldingsCmd = &cobra.Command{
	Use: "holdings", Short: "Get US institution 13F holdings",
	RunE: func(cmd *cobra.Command, args []string) error {
		cik, _ := cmd.Flags().GetString("cik")
		cik = strings.TrimSpace(cik)
		if cik == "" {
			return fmt.Errorf("cik must not be empty")
		}
		cursor, _ := cmd.Flags().GetString("pagination-key")
		q := map[string]string{"cik": cik}
		if cursor != "" {
			q["pagination_key"] = cursor
		}
		return doUSDataGet("/market-data/fundamentals/institutions/holdings/list", q)
	},
}

var data13FTransactionsCmd = &cobra.Command{
	Use: "transactions", Short: "Get US institution holding changes",
	RunE: func(cmd *cobra.Command, args []string) error {
		cik, _ := cmd.Flags().GetString("cik")
		cik = strings.TrimSpace(cik)
		if cik == "" {
			return fmt.Errorf("cik must not be empty")
		}
		position, _ := cmd.Flags().GetString("position")
		position = strings.ToUpper(strings.TrimSpace(position))
		cursor, _ := cmd.Flags().GetString("pagination-key")
		if position != "OPEN" && position != "CLOSE" && position != "CHANGE" {
			return fmt.Errorf("position must be one of OPEN, CLOSE, CHANGE")
		}
		q := map[string]string{"cik": cik, "position": position}
		if cursor != "" {
			q["pagination_key"] = cursor
		}
		return doUSDataGet("/market-data/fundamentals/institutions/holding-changes/list", q)
	},
}

var dataStockInsiderTransactionsCmd = &cobra.Command{
	Use: "insider-transactions", Short: "Get US insider transactions",
	RunE: func(cmd *cobra.Command, args []string) error {
		symbol, _ := cmd.Flags().GetString("symbol")
		symbol = strings.TrimSpace(symbol)
		if symbol == "" {
			return fmt.Errorf("symbol must not be empty")
		}
		side, _ := cmd.Flags().GetString("side")
		side = strings.ToUpper(strings.TrimSpace(side))
		cursor, _ := cmd.Flags().GetString("pagination-key")
		if side != "" && side != "BUY" && side != "SELL" {
			return fmt.Errorf("side must be BUY or SELL")
		}
		q := map[string]string{"symbol": symbol, "category": "US_STOCK"}
		if side != "" {
			q["side"] = side
		}
		if cursor != "" {
			q["pagination_key"] = cursor
		}
		return doUSDataGet("/market-data/fundamentals/insider-transactions/list", q)
	},
}

var dataStockInsiderHoldingsOverviewCmd = &cobra.Command{
	Use: "insider-holdings-overview", Short: "Get US insider ownership overview",
	RunE: func(cmd *cobra.Command, args []string) error {
		symbol, _ := cmd.Flags().GetString("symbol")
		symbol = strings.TrimSpace(symbol)
		if symbol == "" {
			return fmt.Errorf("symbol must not be empty")
		}
		compareMonths, _ := cmd.Flags().GetInt("compare-months")
		if compareMonths != 1 && compareMonths != 3 {
			return fmt.Errorf("compare-months must be 1 or 3")
		}
		return doUSDataGet("/market-data/fundamentals/insider-holdings/get", map[string]string{
			"symbol": symbol, "compare_months": fmt.Sprintf("%d", compareMonths), "category": "US_STOCK",
		})
	},
}

var dataEconomicCalendarCmd = &cobra.Command{
	Use: "economic-calendar", Short: "Get the US macroeconomic calendar",
	RunE: func(cmd *cobra.Command, args []string) error {
		startDate, _ := cmd.Flags().GetString("start-date")
		endDate, _ := cmd.Flags().GetString("end-date")
		q := map[string]string{"region": "US"}
		if startDate != "" {
			q["start_date"] = startDate
		}
		if endDate != "" {
			q["end_date"] = endDate
		}
		return doUSDataGet("/market-data/fundamentals/economic-calendars/list", q)
	},
}

var dataNewsCmd = &cobra.Command{Use: "news", Short: "US market news (search, detail, stock news)"}

var dataNewsSearchCmd = &cobra.Command{
	Use: "search", Short: "Search US market news by keyword",
	RunE: func(cmd *cobra.Command, args []string) error {
		keyword := strings.TrimSpace(cmd.Flags().Lookup("keyword").Value.String())
		startDate, _ := cmd.Flags().GetString("start-date")
		endDate, _ := cmd.Flags().GetString("end-date")
		cursor, _ := cmd.Flags().GetString("pagination-key")
		if keyword == "" {
			return fmt.Errorf("keyword must not be empty")
		}
		if len(keyword) > 100 {
			return fmt.Errorf("keyword must not exceed 100 characters")
		}
		q := map[string]string{"keyword": keyword}
		if startDate != "" {
			q["start_date"] = startDate
		}
		if endDate != "" {
			q["end_date"] = endDate
		}
		if cursor != "" {
			q["pagination_key"] = cursor
		}
		return doUSDataGet("/market-data/news/search/list", q)
	},
}

var dataNewsDetailCmd = &cobra.Command{
	Use: "detail", Short: "Get full US news content by ID",
	RunE: func(cmd *cobra.Command, args []string) error {
		newsID := strings.TrimSpace(cmd.Flags().Lookup("news-id").Value.String())
		if newsID == "" {
			return fmt.Errorf("news-id must not be empty")
		}
		return doUSDataGet("/market-data/news/detail/get", map[string]string{"news_id": newsID})
	},
}

var dataStockNewsCmd = &cobra.Command{
	Use: "stock-news", Short: "Get US stock-related news list",
	RunE: func(cmd *cobra.Command, args []string) error {
		symbol := strings.TrimSpace(cmd.Flags().Lookup("symbol").Value.String())
		sentiment := strings.ToUpper(strings.TrimSpace(cmd.Flags().Lookup("sentiment").Value.String()))
		cursor, _ := cmd.Flags().GetString("pagination-key")
		if symbol == "" {
			return fmt.Errorf("symbol must not be empty")
		}
		if sentiment != "" && sentiment != "POSITIVE" && sentiment != "NEGATIVE" && sentiment != "NEUTRAL" {
			return fmt.Errorf("sentiment must be one of POSITIVE, NEGATIVE, NEUTRAL")
		}
		q := map[string]string{"symbol": symbol, "category": "US_STOCK"}
		if sentiment != "" {
			q["sentiment"] = sentiment
		}
		if cursor != "" {
			q["pagination_key"] = cursor
		}
		return doUSDataGet("/market-data/news/stock-news/list", q)
	},
}

func init() {
	dataStockCmd.AddCommand(dataUnusualActivitiesCmd)
	dataStockCmd.AddCommand(dataStockInsiderTransactionsCmd)
	dataStockCmd.AddCommand(dataStockInsiderHoldingsOverviewCmd)
	dataStockCmd.AddCommand(dataEconomicCalendarCmd)
	dataCmd.AddCommand(data13FCmd)
	dataCmd.AddCommand(dataMarketSentimentCmd)
	dataCmd.AddCommand(dataNewsCmd)
	dataNewsCmd.AddCommand(dataNewsSearchCmd)
	dataNewsCmd.AddCommand(dataNewsDetailCmd)
	dataNewsCmd.AddCommand(dataStockNewsCmd)

	dataUnusualActivitiesCmd.Flags().String("symbol", "", "Symbol")
	dataUnusualActivitiesCmd.Flags().Int("count", 10, "Number of activities (1-30)")
	dataUnusualActivitiesCmd.Flags().String("alert-types", "", "Comma-separated alert types")
	dataUnusualActivitiesCmd.Flags().Bool("extended-hours", false, "Include pre-market and after-hours alerts")
	data13FInstitutionsSearchCmd.Flags().String("keyword", "", "Institution name or CIK (required)")
	data13FInstitutionsSearchCmd.Flags().String("pagination-key", "", "Pagination cursor")
	data13FInstitutionsSearchCmd.MarkFlagRequired("keyword")
	data13FHoldingsCmd.Flags().String("cik", "", "Institution CIK (required)")
	data13FHoldingsCmd.Flags().String("pagination-key", "", "Pagination cursor")
	data13FHoldingsCmd.MarkFlagRequired("cik")
	data13FTransactionsCmd.Flags().String("cik", "", "Institution CIK (required)")
	data13FTransactionsCmd.Flags().String("position", "", "Position: OPEN, CLOSE, CHANGE (required)")
	data13FTransactionsCmd.Flags().String("pagination-key", "", "Pagination cursor")
	data13FTransactionsCmd.MarkFlagRequired("cik")
	data13FTransactionsCmd.MarkFlagRequired("position")
	dataStockInsiderTransactionsCmd.Flags().String("symbol", "", "Symbol (required)")
	dataStockInsiderTransactionsCmd.Flags().String("side", "", "Side: BUY or SELL")
	dataStockInsiderTransactionsCmd.Flags().String("pagination-key", "", "Pagination cursor")
	dataStockInsiderTransactionsCmd.MarkFlagRequired("symbol")
	dataStockInsiderHoldingsOverviewCmd.Flags().String("symbol", "", "Symbol (required)")
	dataStockInsiderHoldingsOverviewCmd.Flags().Int("compare-months", 1, "Comparison period: 1 or 3 months")
	dataStockInsiderHoldingsOverviewCmd.MarkFlagRequired("symbol")
	dataEconomicCalendarCmd.Flags().String("start-date", "", "Start date yyyy-MM-dd")
	dataEconomicCalendarCmd.Flags().String("end-date", "", "End date yyyy-MM-dd")
	dataNewsSearchCmd.Flags().String("keyword", "", "Search keyword, max 100 chars (required)")
	dataNewsSearchCmd.Flags().String("start-date", "", "Start date yyyy-MM-dd")
	dataNewsSearchCmd.Flags().String("end-date", "", "End date yyyy-MM-dd")
	dataNewsSearchCmd.Flags().String("pagination-key", "", "Pagination cursor")
	dataNewsSearchCmd.MarkFlagRequired("keyword")
	dataNewsDetailCmd.Flags().String("news-id", "", "News unique identifier (required)")
	dataNewsDetailCmd.MarkFlagRequired("news-id")
	dataStockNewsCmd.Flags().String("symbol", "", "Symbol (required)")
	dataStockNewsCmd.Flags().String("sentiment", "", "Sentiment: POSITIVE, NEGATIVE, NEUTRAL")
	dataStockNewsCmd.Flags().String("pagination-key", "", "Pagination cursor")
	dataStockNewsCmd.MarkFlagRequired("symbol")
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

func newUSDataClient() (*client.APIClient, error) {
	pm := config.NewProfileManager("")
	profile, err := pm.GetActive()
	if err != nil {
		return nil, err
	}

	region := profile.Region
	if flagRegion != "" {
		region = config.Region(strings.ToLower(strings.TrimSpace(flagRegion)))
	}
	if !region.IsValid() {
		return nil, fmt.Errorf("unsupported region %q", region)
	}
	if region != config.RegionUS {
		return nil, fmt.Errorf("this market-data action is only supported in region 'us'")
	}

	effectiveProfile := *profile
	effectiveProfile.Region = region

	var appKey, appSecret, token string
	if envKey, envSecret, ok := config.EnvCredentials(); ok {
		appKey = envKey
		appSecret = envSecret
	} else {
		cred, err := config.LoadCredential(effectiveProfile.Name)
		if err != nil {
			return nil, fmt.Errorf("failed to load credentials: %w. Run 'webull auth login' to re-authenticate", err)
		}
		appKey = cred.AppKey
		appSecret = cred.AppSecret
		token = cred.Token
	}

	envOverrides := config.EnvEndpointOverrides()
	var overrides *config.RegionEndpoints
	if effectiveProfile.Endpoints.API != "" {
		overrides = &config.RegionEndpoints{API: effectiveProfile.Endpoints.API}
	}
	endpoints := config.GetEndpoints(region, overrides)
	if envOverrides != nil && envOverrides.API != "" {
		endpoints.API = envOverrides.API
	}

	provider := auth.NewAkSkProvider(&effectiveProfile, pm, appKey, appSecret, token)
	apiClient := client.NewAPIClient(provider, endpoints, time.Duration(flagTimeout)*time.Second)
	apiClient.SetVerbose(flagVerbose)
	apiClient.SetDebug(flagDebug)
	return apiClient, nil
}

func doUSDataGet(path string, query map[string]string) error {
	cl, err := newUSDataClient()
	if err != nil {
		output.PrintError("CONFIG_ERROR", err.Error(), 0)
		return err
	}
	return doDataGetWithClient(cl, path, query)
}

func doDataGet(path string, query map[string]string) error {
	cl, err := newAPIClient()
	if err != nil {
		return err
	}
	return doDataGetWithClient(cl, path, query)
}

func doDataGetWithClient(cl *client.APIClient, path string, query map[string]string) error {
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
