// Copyright 2024-2026 Webull Inc. All rights reserved.
// Licensed under the Apache License, Version 2.0.
// See LICENSE file in the project root for full license information.

package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/webull-inc/webull-openapi-cli/internal/client"
	"github.com/webull-inc/webull-openapi-cli/internal/output"
)

var instrumentCmd = &cobra.Command{
	Use:   "instrument",
	Short: "Query instrument/symbol information",
}

// Top-level init: register sub-commands only
func init() {
	instrumentCmd.AddCommand(instrumentStockCmd)
	instrumentCmd.AddCommand(instrumentCryptoCmd)
	instrumentCmd.AddCommand(instrumentOptionCmd)
	instrumentCmd.AddCommand(instrumentFuturesCmd)
	instrumentCmd.AddCommand(instrumentFuturesProductsCmd)
	instrumentCmd.AddCommand(instrumentFuturesClassCmd)
	instrumentCmd.AddCommand(instrumentEventCmd)
	instrumentCmd.AddCommand(instrumentEventCategoriesCmd)
	instrumentCmd.AddCommand(instrumentEventSeriesCmd)
	instrumentCmd.AddCommand(instrumentEventEventsCmd)
	instrumentCmd.AddCommand(instrumentCompanyProfileCmd)
	instrumentCmd.AddCommand(instrumentAnalystPriceCmd)
	instrumentCmd.AddCommand(instrumentAnalystRatingCmd)
}

var instrumentStockCmd = &cobra.Command{
	Use: "stock", Short: "Query stock/ETF instrument information",
	RunE: func(cmd *cobra.Command, args []string) error {
		symbol, _ := cmd.Flags().GetString("symbol")
		category, _ := cmd.Flags().GetString("category")
		status, _ := cmd.Flags().GetString("status")
		lastID, _ := cmd.Flags().GetString("last-instrument-id")
		pageSize, _ := cmd.Flags().GetInt("page-size")
		q := map[string]string{"category": category, "page_size": fmt.Sprintf("%d", pageSize)}
		if symbol != "" {
			q["symbols"] = symbol
		}
		if status != "" {
			q["status"] = status
		}
		if lastID != "" {
			q["last_instrument_id"] = lastID
		}
		return doInstrumentGet("/openapi/instrument/stock/list", q)
	},
}

func init() {
	instrumentStockCmd.Flags().String("symbol", "", "Symbol(s), comma-separated")
	instrumentStockCmd.Flags().String("category", "US_STOCK", "Category: US_STOCK")
	instrumentStockCmd.Flags().String("status", "", "Status: OC, CO, NT")
	instrumentStockCmd.Flags().String("last-instrument-id", "", "Last instrument ID for pagination")
	instrumentStockCmd.Flags().Int("page-size", 1000, "Page size (max 1000)")
}

var instrumentCryptoCmd = &cobra.Command{
	Use: "crypto", Short: "Query crypto instrument information",
	RunE: func(cmd *cobra.Command, args []string) error {
		symbol, _ := cmd.Flags().GetString("symbol")
		category, _ := cmd.Flags().GetString("category")
		status, _ := cmd.Flags().GetString("status")
		lastID, _ := cmd.Flags().GetString("last-instrument-id")
		pageSize, _ := cmd.Flags().GetInt("page-size")
		q := map[string]string{"category": category, "page_size": fmt.Sprintf("%d", pageSize)}
		if symbol != "" {
			q["symbols"] = symbol
		}
		if status != "" {
			q["status"] = status
		}
		if lastID != "" {
			q["last_instrument_id"] = lastID
		}
		return doInstrumentGet("/openapi/instrument/crypto/list", q)
	},
}

func init() {
	instrumentCryptoCmd.Flags().String("symbol", "", "Crypto symbol(s), comma-separated")
	instrumentCryptoCmd.Flags().String("category", "US_CRYPTO", "Category: US_CRYPTO")
	instrumentCryptoCmd.Flags().String("status", "", "Status: OC, CO, NT")
	instrumentCryptoCmd.Flags().String("last-instrument-id", "", "Last instrument ID for pagination")
	instrumentCryptoCmd.Flags().Int("page-size", 1000, "Page size (max 1000)")
}

var instrumentOptionCmd = &cobra.Command{
	Use: "option", Short: "Query option contract information",
	RunE: func(cmd *cobra.Command, args []string) error {
		category, _ := cmd.Flags().GetString("category")
		underlyingSymbols, _ := cmd.Flags().GetString("underlying-symbols")
		status, _ := cmd.Flags().GetString("status")
		startDate, _ := cmd.Flags().GetString("start-date")
		endDate, _ := cmd.Flags().GetString("end-date")
		rootSymbol, _ := cmd.Flags().GetString("root-symbol")
		optionSymbol, _ := cmd.Flags().GetString("option-symbol")
		optionType, _ := cmd.Flags().GetString("option-type")
		style, _ := cmd.Flags().GetString("style")
		strikePriceGte, _ := cmd.Flags().GetString("strike-price-gte")
		strikePriceLte, _ := cmd.Flags().GetString("strike-price-lte")
		ppind, _ := cmd.Flags().GetString("ppind")
		showDeliverables, _ := cmd.Flags().GetString("show-deliverables")
		lastID, _ := cmd.Flags().GetString("last-instrument-id")
		pageSize, _ := cmd.Flags().GetInt("page-size")

		q := map[string]string{"category": category, "page_size": fmt.Sprintf("%d", pageSize)}
		if underlyingSymbols != "" {
			q["underlying_symbols"] = underlyingSymbols
		}
		if status != "" {
			q["status"] = status
		}
		if startDate != "" {
			q["start_date"] = startDate
		}
		if endDate != "" {
			q["end_date"] = endDate
		}
		if rootSymbol != "" {
			q["root_symbol"] = rootSymbol
		}
		if optionSymbol != "" {
			q["option_symbol"] = optionSymbol
		}
		if optionType != "" {
			q["option_type"] = optionType
		}
		if style != "" {
			q["style"] = style
		}
		if strikePriceGte != "" {
			q["strike_price_gte"] = strikePriceGte
		}
		if strikePriceLte != "" {
			q["strike_price_lte"] = strikePriceLte
		}
		if ppind != "" {
			q["ppind"] = ppind
		}
		if showDeliverables != "" {
			q["show_deliverables"] = showDeliverables
		}
		if lastID != "" {
			q["last_instrument_id"] = lastID
		}
		return doInstrumentGet("/openapi/instrument/option/contracts", q)
	},
}

func init() {
	instrumentOptionCmd.Flags().String("underlying-symbols", "", "Underlying symbols, comma-separated (e.g. AAPL,MSFT)")
	instrumentOptionCmd.Flags().String("category", "US_OPTION", "Category: US_OPTION")
	instrumentOptionCmd.Flags().String("status", "", "Contract status: LISTING, DELISTING (default LISTING)")
	instrumentOptionCmd.Flags().String("start-date", "", "Exact expiration date (YYYY-MM-DD)")
	instrumentOptionCmd.Flags().String("end-date", "", "Expiration date lower bound inclusive (YYYY-MM-DD)")
	instrumentOptionCmd.Flags().String("root-symbol", "", "Root symbol / series symbol (e.g. SPXW)")
	instrumentOptionCmd.Flags().String("option-symbol", "", "Option symbol (e.g. AAPL250620C00150000)")
	instrumentOptionCmd.Flags().String("option-type", "", "Contract type: CALL, PUT")
	instrumentOptionCmd.Flags().String("style", "", "Exercise style: AMERICAN, EUROPEAN")
	instrumentOptionCmd.Flags().String("strike-price-gte", "", "Strike price lower bound (inclusive)")
	instrumentOptionCmd.Flags().String("strike-price-lte", "", "Strike price upper bound (inclusive)")
	instrumentOptionCmd.Flags().String("ppind", "", "Penny Program Indicator: true/false")
	instrumentOptionCmd.Flags().String("show-deliverables", "", "Return deliverables in response: true/false (default false)")
	instrumentOptionCmd.Flags().String("last-instrument-id", "", "Last instrument ID for pagination")
	instrumentOptionCmd.Flags().Int("page-size", 100, "Page size (max 1000, default 100)")
}

var instrumentFuturesCmd = &cobra.Command{
	Use: "futures", Short: "Query futures contract information",
	RunE: func(cmd *cobra.Command, args []string) error {
		symbol, _ := cmd.Flags().GetString("symbol")
		category, _ := cmd.Flags().GetString("category")
		code, _ := cmd.Flags().GetString("code")
		status, _ := cmd.Flags().GetString("status")
		q := map[string]string{"category": category}
		if symbol != "" {
			q["symbols"] = symbol
		}
		if code != "" {
			q["code"] = code
		}
		if status != "" {
			q["status"] = status
		}
		return doInstrumentGet("/openapi/instrument/futures/list", q)
	},
}

func init() {
	instrumentFuturesCmd.Flags().String("symbol", "", "Futures symbol(s), comma-separated")
	instrumentFuturesCmd.Flags().String("category", "US_FUTURES", "Category: US_FUTURES")
	instrumentFuturesCmd.Flags().String("code", "", "Futures product code (e.g. ES)")
	instrumentFuturesCmd.Flags().String("status", "", "Status: OC, CO, NT")
}

var instrumentFuturesProductsCmd = &cobra.Command{
	Use: "futures-products", Short: "Query futures product codes",
	RunE: func(cmd *cobra.Command, args []string) error {
		category, _ := cmd.Flags().GetString("category")
		classID, _ := cmd.Flags().GetInt("product-class-id")
		q := map[string]string{"category": category}
		if classID > 0 {
			q["product_class_id"] = fmt.Sprintf("%d", classID)
		}
		return doInstrumentGet("/openapi/instrument/futures/products", q)
	},
}

func init() {
	instrumentFuturesProductsCmd.Flags().String("category", "US_FUTURES", "Category: US_FUTURES")
	instrumentFuturesProductsCmd.Flags().Int("product-class-id", 0, "Product class ID (optional)")
}

var instrumentFuturesClassCmd = &cobra.Command{
	Use: "futures-class", Short: "Query futures product classification groups",
	RunE: func(cmd *cobra.Command, args []string) error {
		category, _ := cmd.Flags().GetString("category")
		return doInstrumentGet("/openapi/instrument/futures/product-classes", map[string]string{"category": category})
	},
}

func init() {
	instrumentFuturesClassCmd.Flags().String("category", "US_FUTURES", "Category: US_FUTURES")
}

var instrumentEventCmd = &cobra.Command{
	Use: "event", Short: "Query event contract instruments",
	RunE: func(cmd *cobra.Command, args []string) error {
		seriesSymbol, _ := cmd.Flags().GetString("series-symbol")
		eventSymbol, _ := cmd.Flags().GetString("event-symbol")
		symbol, _ := cmd.Flags().GetString("symbol")
		expAfter, _ := cmd.Flags().GetString("expiration-date-after")
		lastID, _ := cmd.Flags().GetString("last-instrument-id")
		pageSize, _ := cmd.Flags().GetInt("page-size")
		q := map[string]string{"series_symbol": seriesSymbol, "page_size": fmt.Sprintf("%d", pageSize)}
		if eventSymbol != "" {
			q["event_symbol"] = eventSymbol
		}
		if symbol != "" {
			q["symbols"] = symbol
		}
		if expAfter != "" {
			q["expiration_date_after"] = expAfter
		}
		if lastID != "" {
			q["last_instrument_id"] = lastID
		}
		return doInstrumentGet("/openapi/instrument/event/market/list", q)
	},
}

func init() {
	instrumentEventCmd.Flags().String("series-symbol", "", "Series symbol (required)")
	instrumentEventCmd.Flags().String("event-symbol", "", "Event symbol")
	instrumentEventCmd.Flags().String("symbol", "", "Market symbol(s), comma-separated")
	instrumentEventCmd.Flags().String("expiration-date-after", "", "Filter expiration after date (YYYY-MM-DD)")
	instrumentEventCmd.Flags().String("last-instrument-id", "", "Last instrument ID for pagination")
	instrumentEventCmd.Flags().Int("page-size", 500, "Page size")
	instrumentEventCmd.MarkFlagRequired("series-symbol")
}

var instrumentEventCategoriesCmd = &cobra.Command{
	Use: "event-categories", Short: "List event contract categories",
	RunE: func(cmd *cobra.Command, args []string) error {
		return doInstrumentGet("/openapi/instrument/event/categories", nil)
	},
}

var instrumentEventSeriesCmd = &cobra.Command{
	Use: "event-series", Short: "List event contract series",
	RunE: func(cmd *cobra.Command, args []string) error {
		category, _ := cmd.Flags().GetString("category")
		symbols, _ := cmd.Flags().GetString("symbol")
		lastID, _ := cmd.Flags().GetString("last-series-id")
		pageSize, _ := cmd.Flags().GetInt("page-size")
		q := map[string]string{"page_size": fmt.Sprintf("%d", pageSize)}
		if category != "" {
			q["category"] = category
		}
		if symbols != "" {
			q["symbols"] = symbols
		}
		if lastID != "" {
			q["last_series_id"] = lastID
		}
		return doInstrumentGet("/openapi/instrument/event/series/list", q)
	},
}

func init() {
	instrumentEventSeriesCmd.Flags().String("category", "", "Category: ECONOMICS,FINANCIALS,POLITICS,ENTERTAINMENT,etc.")
	instrumentEventSeriesCmd.Flags().String("symbol", "", "Series symbol(s), comma-separated")
	instrumentEventSeriesCmd.Flags().String("last-series-id", "", "Last series ID for pagination")
	instrumentEventSeriesCmd.Flags().Int("page-size", 500, "Page size")
}

var instrumentEventEventsCmd = &cobra.Command{
	Use: "event-events", Short: "List events by series symbol",
	RunE: func(cmd *cobra.Command, args []string) error {
		seriesSymbol, _ := cmd.Flags().GetString("series-symbol")
		symbols, _ := cmd.Flags().GetString("symbol")
		status, _ := cmd.Flags().GetString("status")
		q := map[string]string{"series_symbol": seriesSymbol}
		if symbols != "" {
			q["symbols"] = symbols
		}
		if status != "" {
			q["status"] = status
		}
		return doInstrumentGet("/openapi/instrument/event/events", q)
	},
}

func init() {
	instrumentEventEventsCmd.Flags().String("series-symbol", "", "Series symbol (required)")
	instrumentEventEventsCmd.Flags().String("symbol", "", "Event symbol(s), comma-separated")
	instrumentEventEventsCmd.Flags().String("status", "", "Status: ACTIVE, INACTIVE")
	instrumentEventEventsCmd.MarkFlagRequired("series-symbol")
}

var instrumentCompanyProfileCmd = &cobra.Command{
	Use:        "company-profile",
	Short:      "Get company profile information",
	Deprecated: "use 'webull data fundamentals company-profile' instead",
	RunE: func(cmd *cobra.Command, args []string) error {
		symbol, _ := cmd.Flags().GetString("symbol")
		category, _ := cmd.Flags().GetString("category")
		return doInstrumentGet("/openapi/instrument/company/profile", map[string]string{"symbol": symbol, "category": category})
	},
}

func init() {
	instrumentCompanyProfileCmd.Flags().String("symbol", "", "Symbol (required)")
	instrumentCompanyProfileCmd.Flags().String("category", "US_STOCK", "Category: US_STOCK")
	instrumentCompanyProfileCmd.MarkFlagRequired("symbol")
}

var instrumentAnalystPriceCmd = &cobra.Command{
	Use:        "analyst-price",
	Short:      "Get analyst target price",
	Deprecated: "use 'webull data fundamentals analyst-price' instead",
	RunE: func(cmd *cobra.Command, args []string) error {
		symbol, _ := cmd.Flags().GetString("symbol")
		category, _ := cmd.Flags().GetString("category")
		return doInstrumentGet("/openapi/instrument/analyst/target-price", map[string]string{"symbol": symbol, "category": category})
	},
}

func init() {
	instrumentAnalystPriceCmd.Flags().String("symbol", "", "Symbol (required)")
	instrumentAnalystPriceCmd.Flags().String("category", "US_STOCK", "Category: US_STOCK")
	instrumentAnalystPriceCmd.MarkFlagRequired("symbol")
}

var instrumentAnalystRatingCmd = &cobra.Command{
	Use:        "analyst-rating",
	Short:      "Get analyst rating",
	Deprecated: "use 'webull data fundamentals analyst-rating' instead",
	RunE: func(cmd *cobra.Command, args []string) error {
		symbol, _ := cmd.Flags().GetString("symbol")
		category, _ := cmd.Flags().GetString("category")
		return doInstrumentGet("/openapi/instrument/analyst/rating", map[string]string{"symbol": symbol, "category": category})
	},
}

func init() {
	instrumentAnalystRatingCmd.Flags().String("symbol", "", "Symbol (required)")
	instrumentAnalystRatingCmd.Flags().String("category", "US_STOCK", "Category: US_STOCK")
	instrumentAnalystRatingCmd.MarkFlagRequired("symbol")
}

func doInstrumentGet(path string, query map[string]string) error {
	cl, err := newAPIClient()
	if err != nil {
		return err
	}
	opts := &client.RequestOptions{}
	if query != nil {
		opts.Query = query
	}
	resp, err := cl.Do(context.Background(), "GET", path, opts)
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
