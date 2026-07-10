// Copyright 2024-2026 Webull Inc. All rights reserved.
// Licensed under the Apache License, Version 2.0.
// See LICENSE file in the project root for full license information.

package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var dataFundamentalsCmd = &cobra.Command{
	Use:   "fundamentals",
	Short: "Fundamentals data (company profile, financials, fund data)",
}

func init() {
	dataCmd.AddCommand(dataFundamentalsCmd)

	dataFundamentalsCmd.AddCommand(fundCompanyProfileCmd)
	dataFundamentalsCmd.AddCommand(fundAnalystPriceCmd)
	dataFundamentalsCmd.AddCommand(fundAnalystRatingCmd)
	dataFundamentalsCmd.AddCommand(fundForecastEPSCmd)
	dataFundamentalsCmd.AddCommand(fundFilingsCmd)
	dataFundamentalsCmd.AddCommand(fundEarningsCalendarCmd)
	dataFundamentalsCmd.AddCommand(fundDividendCalendarCmd)
	dataFundamentalsCmd.AddCommand(fundCapitalFlowCmd)
	dataFundamentalsCmd.AddCommand(fundIndustryComparisonCmd)
	dataFundamentalsCmd.AddCommand(fundIndicatorsCmd)
	dataFundamentalsCmd.AddCommand(fundIncomeCmd)
	dataFundamentalsCmd.AddCommand(fundCashflowCmd)
	dataFundamentalsCmd.AddCommand(fundBalancesheetCmd)
	dataFundamentalsCmd.AddCommand(fundAlertCmd)
	dataFundamentalsCmd.AddCommand(fundBriefCmd)
	dataFundamentalsCmd.AddCommand(fundPerformanceCmd)
	dataFundamentalsCmd.AddCommand(fundNetValueCmd)
	dataFundamentalsCmd.AddCommand(fundHoldingsCmd)
	dataFundamentalsCmd.AddCommand(fundDividendsCmd)
	dataFundamentalsCmd.AddCommand(fundRatingCmd)
	dataFundamentalsCmd.AddCommand(fundSplitsCmd)
	dataFundamentalsCmd.AddCommand(fundFilesCmd)
	dataFundamentalsCmd.AddCommand(fundAllocationCmd)
}

// ==================== Company/Analyst (instrument path) ====================

var fundCompanyProfileCmd = &cobra.Command{
	Use: "company-profile", Short: "Get company profile information",
	RunE: func(cmd *cobra.Command, args []string) error {
		symbol, _ := cmd.Flags().GetString("symbol")
		category, _ := cmd.Flags().GetString("category")
		return doInstrumentGet("/openapi/instrument/company/profile", map[string]string{
			"symbol": symbol, "category": category,
		})
	},
}

func init() {
	fundCompanyProfileCmd.Flags().String("symbol", "", "Symbol (required)")
	fundCompanyProfileCmd.Flags().String("category", "US_STOCK", "Category: US_STOCK, HK_STOCK, US_ETF")
	fundCompanyProfileCmd.MarkFlagRequired("symbol")
}

var fundAnalystPriceCmd = &cobra.Command{
	Use: "analyst-price", Short: "Get analyst target price",
	RunE: func(cmd *cobra.Command, args []string) error {
		symbol, _ := cmd.Flags().GetString("symbol")
		category, _ := cmd.Flags().GetString("category")
		return doInstrumentGet("/openapi/instrument/analyst/target-price", map[string]string{
			"symbol": symbol, "category": category,
		})
	},
}

func init() {
	fundAnalystPriceCmd.Flags().String("symbol", "", "Symbol (required)")
	fundAnalystPriceCmd.Flags().String("category", "US_STOCK", "Category: US_STOCK, HK_STOCK, US_ETF")
	fundAnalystPriceCmd.MarkFlagRequired("symbol")
}

var fundAnalystRatingCmd = &cobra.Command{
	Use: "analyst-rating", Short: "Get analyst rating",
	RunE: func(cmd *cobra.Command, args []string) error {
		symbol, _ := cmd.Flags().GetString("symbol")
		category, _ := cmd.Flags().GetString("category")
		return doInstrumentGet("/openapi/instrument/analyst/rating", map[string]string{
			"symbol": symbol, "category": category,
		})
	},
}

func init() {
	fundAnalystRatingCmd.Flags().String("symbol", "", "Symbol (required)")
	fundAnalystRatingCmd.Flags().String("category", "US_STOCK", "Category: US_STOCK, HK_STOCK, US_ETF")
	fundAnalystRatingCmd.MarkFlagRequired("symbol")
}

// ==================== Stock Fundamentals ====================

var fundForecastEPSCmd = &cobra.Command{
	Use: "forecast-eps", Short: "Get forecast EPS data (most recent 5 quarters)",
	RunE: func(cmd *cobra.Command, args []string) error {
		symbol, _ := cmd.Flags().GetString("symbol")
		category, _ := cmd.Flags().GetString("category")
		return doInstrumentGet("/openapi/fundamentals/stock/forecast-eps", map[string]string{
			"symbol": symbol, "category": category,
		})
	},
}

func init() {
	fundForecastEPSCmd.Flags().String("symbol", "", "Symbol (required)")
	fundForecastEPSCmd.Flags().String("category", "US_STOCK", "Category: US_STOCK, HK_STOCK, CN_STOCK")
	fundForecastEPSCmd.MarkFlagRequired("symbol")
}

var fundFilingsCmd = &cobra.Command{
	Use: "filings", Short: "Get most recent SEC filings for a stock",
	RunE: func(cmd *cobra.Command, args []string) error {
		symbol, _ := cmd.Flags().GetString("symbol")
		category, _ := cmd.Flags().GetString("category")
		return doInstrumentGet("/openapi/fundamentals/stock/filings", map[string]string{
			"symbol": symbol, "category": category,
		})
	},
}

func init() {
	fundFilingsCmd.Flags().String("symbol", "", "Symbol (required)")
	fundFilingsCmd.Flags().String("category", "US_STOCK", "Category: US_STOCK, HK_STOCK, CN_STOCK")
	fundFilingsCmd.MarkFlagRequired("symbol")
}

var fundEarningsCalendarCmd = &cobra.Command{
	Use: "earnings-calendar", Short: "Get earnings calendar (most recent 6 quarters)",
	RunE: func(cmd *cobra.Command, args []string) error {
		symbol, _ := cmd.Flags().GetString("symbol")
		category, _ := cmd.Flags().GetString("category")
		return doInstrumentGet("/openapi/fundamentals/stock/earnings-calendar", map[string]string{
			"symbol": symbol, "category": category,
		})
	},
}

func init() {
	fundEarningsCalendarCmd.Flags().String("symbol", "", "Symbol (required)")
	fundEarningsCalendarCmd.Flags().String("category", "US_STOCK", "Category: US_STOCK, HK_STOCK, CN_STOCK, JP_STOCK")
	fundEarningsCalendarCmd.MarkFlagRequired("symbol")
}

var fundDividendCalendarCmd = &cobra.Command{
	Use: "dividend-calendar", Short: "Get most recent 6 dividend calendar records",
	RunE: func(cmd *cobra.Command, args []string) error {
		symbol, _ := cmd.Flags().GetString("symbol")
		category, _ := cmd.Flags().GetString("category")
		return doInstrumentGet("/openapi/fundamentals/stock/dividend-calendar", map[string]string{
			"symbol": symbol, "category": category,
		})
	},
}

func init() {
	fundDividendCalendarCmd.Flags().String("symbol", "", "Symbol (required)")
	fundDividendCalendarCmd.Flags().String("category", "US_STOCK", "Category: US_STOCK, HK_STOCK, CN_STOCK, JP_STOCK")
	fundDividendCalendarCmd.MarkFlagRequired("symbol")
}

var fundCapitalFlowCmd = &cobra.Command{
	Use: "capital-flow", Short: "Get capital flow distribution for a stock",
	RunE: func(cmd *cobra.Command, args []string) error {
		symbol, _ := cmd.Flags().GetString("symbol")
		category, _ := cmd.Flags().GetString("category")
		count, _ := cmd.Flags().GetInt("count")
		q := map[string]string{"symbol": symbol, "category": category}
		if count > 0 {
			q["count"] = fmt.Sprintf("%d", count)
		}
		return doInstrumentGet("/openapi/fundamentals/stock/capital-flow", q)
	},
}

func init() {
	fundCapitalFlowCmd.Flags().String("symbol", "", "Symbol (required)")
	fundCapitalFlowCmd.Flags().String("category", "US_STOCK", "Category: US_STOCK, HK_STOCK, CN_STOCK")
	fundCapitalFlowCmd.Flags().Int("count", 5, "Number of trading days (1-5)")
	fundCapitalFlowCmd.MarkFlagRequired("symbol")
}

var fundIndustryComparisonCmd = &cobra.Command{
	Use: "industry-comparison", Short: "Get industry peer comparison (up to 20 symbols)",
	RunE: func(cmd *cobra.Command, args []string) error {
		symbol, _ := cmd.Flags().GetString("symbol")
		category, _ := cmd.Flags().GetString("category")
		sortBy, _ := cmd.Flags().GetString("sort-by")
		q := map[string]string{"symbol": symbol, "category": category}
		if sortBy != "" {
			q["sort_by"] = sortBy
		}
		return doInstrumentGet("/openapi/fundamentals/stock/industry-comparison", q)
	},
}

func init() {
	fundIndustryComparisonCmd.Flags().String("symbol", "", "Symbol (required)")
	fundIndustryComparisonCmd.Flags().String("category", "US_STOCK", "Category: US_STOCK, HK_STOCK, CN_STOCK")
	fundIndustryComparisonCmd.Flags().String("sort-by", "EPS_TTM", "Sort metric: EPS_TTM, etc.")
	fundIndustryComparisonCmd.MarkFlagRequired("symbol")
}

// ==================== Financial Statements ====================

var fundIndicatorsCmd = &cobra.Command{
	Use: "indicators", Short: "Get financial indicators (ROA, ROE, EPS, etc.)",
	RunE: func(cmd *cobra.Command, args []string) error {
		symbol, _ := cmd.Flags().GetString("symbol")
		category, _ := cmd.Flags().GetString("category")
		fType, _ := cmd.Flags().GetString("type")
		count, _ := cmd.Flags().GetInt("count")
		q := map[string]string{"symbol": symbol, "category": category}
		if fType != "" {
			q["type"] = fType
		}
		if count > 0 {
			q["count"] = fmt.Sprintf("%d", count)
		}
		return doInstrumentGet("/openapi/fundamentals/financial/indicators", q)
	},
}

func init() {
	fundIndicatorsCmd.Flags().String("symbol", "", "Symbol (required)")
	fundIndicatorsCmd.Flags().String("category", "US_STOCK", "Category: US_STOCK")
	fundIndicatorsCmd.Flags().String("type", "QUARTERLY", "Financial type: ANNUAL or QUARTERLY")
	fundIndicatorsCmd.Flags().Int("count", 5, "Number of periods (max 20)")
	fundIndicatorsCmd.MarkFlagRequired("symbol")
}

var fundIncomeCmd = &cobra.Command{
	Use: "income", Short: "Get income statement data",
	RunE: func(cmd *cobra.Command, args []string) error {
		symbol, _ := cmd.Flags().GetString("symbol")
		category, _ := cmd.Flags().GetString("category")
		fType, _ := cmd.Flags().GetString("type")
		count, _ := cmd.Flags().GetInt("count")
		q := map[string]string{"symbol": symbol, "category": category}
		if fType != "" {
			q["type"] = fType
		}
		if count > 0 {
			q["count"] = fmt.Sprintf("%d", count)
		}
		return doInstrumentGet("/openapi/fundamentals/financial/income", q)
	},
}

func init() {
	fundIncomeCmd.Flags().String("symbol", "", "Symbol (required)")
	fundIncomeCmd.Flags().String("category", "US_STOCK", "Category: US_STOCK")
	fundIncomeCmd.Flags().String("type", "QUARTERLY", "Financial type: ANNUAL or QUARTERLY")
	fundIncomeCmd.Flags().Int("count", 5, "Number of periods (max 20)")
	fundIncomeCmd.MarkFlagRequired("symbol")
}

var fundCashflowCmd = &cobra.Command{
	Use: "cashflow", Short: "Get cash flow statement data",
	RunE: func(cmd *cobra.Command, args []string) error {
		symbol, _ := cmd.Flags().GetString("symbol")
		category, _ := cmd.Flags().GetString("category")
		fType, _ := cmd.Flags().GetString("type")
		count, _ := cmd.Flags().GetInt("count")
		q := map[string]string{"symbol": symbol, "category": category}
		if fType != "" {
			q["type"] = fType
		}
		if count > 0 {
			q["count"] = fmt.Sprintf("%d", count)
		}
		return doInstrumentGet("/openapi/fundamentals/financial/cash-flow", q)
	},
}

func init() {
	fundCashflowCmd.Flags().String("symbol", "", "Symbol (required)")
	fundCashflowCmd.Flags().String("category", "US_STOCK", "Category: US_STOCK")
	fundCashflowCmd.Flags().String("type", "QUARTERLY", "Financial type: ANNUAL or QUARTERLY")
	fundCashflowCmd.Flags().Int("count", 5, "Number of periods (max 20)")
	fundCashflowCmd.MarkFlagRequired("symbol")
}

var fundBalancesheetCmd = &cobra.Command{
	Use: "balancesheet", Short: "Get balance sheet data",
	RunE: func(cmd *cobra.Command, args []string) error {
		symbol, _ := cmd.Flags().GetString("symbol")
		category, _ := cmd.Flags().GetString("category")
		fType, _ := cmd.Flags().GetString("type")
		count, _ := cmd.Flags().GetInt("count")
		q := map[string]string{"symbol": symbol, "category": category}
		if fType != "" {
			q["type"] = fType
		}
		if count > 0 {
			q["count"] = fmt.Sprintf("%d", count)
		}
		return doInstrumentGet("/openapi/fundamentals/financial/balance-sheet", q)
	},
}

func init() {
	fundBalancesheetCmd.Flags().String("symbol", "", "Symbol (required)")
	fundBalancesheetCmd.Flags().String("category", "US_STOCK", "Category: US_STOCK")
	fundBalancesheetCmd.Flags().String("type", "QUARTERLY", "Financial type: ANNUAL or QUARTERLY")
	fundBalancesheetCmd.Flags().Int("count", 5, "Number of periods (max 20)")
	fundBalancesheetCmd.MarkFlagRequired("symbol")
}

var fundAlertCmd = &cobra.Command{
	Use: "alert", Short: "Get upcoming earnings release alert",
	RunE: func(cmd *cobra.Command, args []string) error {
		symbol, _ := cmd.Flags().GetString("symbol")
		category, _ := cmd.Flags().GetString("category")
		return doInstrumentGet("/openapi/fundamentals/financial/alert", map[string]string{
			"symbol": symbol, "category": category,
		})
	},
}

func init() {
	fundAlertCmd.Flags().String("symbol", "", "Symbol (required)")
	fundAlertCmd.Flags().String("category", "US_STOCK", "Category: US_STOCK")
	fundAlertCmd.MarkFlagRequired("symbol")
}

// ==================== Fund Data ====================

var fundBriefCmd = &cobra.Command{
	Use: "fund-brief", Short: "Get fund brief information",
	RunE: func(cmd *cobra.Command, args []string) error {
		symbol, _ := cmd.Flags().GetString("symbol")
		category, _ := cmd.Flags().GetString("category")
		return doInstrumentGet("/openapi/fundamentals/fund/brief", map[string]string{
			"symbol": symbol, "category": category,
		})
	},
}

func init() {
	fundBriefCmd.Flags().String("symbol", "", "Symbol (required)")
	fundBriefCmd.Flags().String("category", "US_STOCK", "Category: US_STOCK")
	fundBriefCmd.MarkFlagRequired("symbol")
}

var fundPerformanceCmd = &cobra.Command{
	Use: "fund-performance", Short: "Get fund performance returns",
	RunE: func(cmd *cobra.Command, args []string) error {
		symbol, _ := cmd.Flags().GetString("symbol")
		category, _ := cmd.Flags().GetString("category")
		return doInstrumentGet("/openapi/fundamentals/fund/performance", map[string]string{
			"symbol": symbol, "category": category,
		})
	},
}

func init() {
	fundPerformanceCmd.Flags().String("symbol", "", "Symbol (required)")
	fundPerformanceCmd.Flags().String("category", "US_STOCK", "Category: US_STOCK")
	fundPerformanceCmd.MarkFlagRequired("symbol")
}

var fundNetValueCmd = &cobra.Command{
	Use: "fund-net-value", Short: "Get fund net asset value (NAV) history",
	RunE: func(cmd *cobra.Command, args []string) error {
		symbol, _ := cmd.Flags().GetString("symbol")
		category, _ := cmd.Flags().GetString("category")
		count, _ := cmd.Flags().GetInt("count")
		q := map[string]string{"symbol": symbol, "category": category}
		if count > 0 {
			q["count"] = fmt.Sprintf("%d", count)
		}
		return doInstrumentGet("/openapi/fundamentals/fund/net-value", q)
	},
}

func init() {
	fundNetValueCmd.Flags().String("symbol", "", "Symbol (required)")
	fundNetValueCmd.Flags().String("category", "US_STOCK", "Category: US_STOCK")
	fundNetValueCmd.Flags().Int("count", 20, "Number of records")
	fundNetValueCmd.MarkFlagRequired("symbol")
}

var fundHoldingsCmd = &cobra.Command{
	Use: "fund-holdings", Short: "Get fund top 10 holdings",
	RunE: func(cmd *cobra.Command, args []string) error {
		symbol, _ := cmd.Flags().GetString("symbol")
		category, _ := cmd.Flags().GetString("category")
		return doInstrumentGet("/openapi/fundamentals/fund/holdings", map[string]string{
			"symbol": symbol, "category": category,
		})
	},
}

func init() {
	fundHoldingsCmd.Flags().String("symbol", "", "Symbol (required)")
	fundHoldingsCmd.Flags().String("category", "US_STOCK", "Category: US_STOCK")
	fundHoldingsCmd.MarkFlagRequired("symbol")
}

var fundDividendsCmd = &cobra.Command{
	Use: "fund-dividends", Short: "Get fund dividend history",
	RunE: func(cmd *cobra.Command, args []string) error {
		symbol, _ := cmd.Flags().GetString("symbol")
		category, _ := cmd.Flags().GetString("category")
		pageIndex, _ := cmd.Flags().GetInt("page-index")
		pageSize, _ := cmd.Flags().GetInt("page-size")
		q := map[string]string{"symbol": symbol, "category": category}
		if pageIndex > 0 {
			q["page_index"] = fmt.Sprintf("%d", pageIndex)
		}
		if pageSize > 0 {
			q["page_size"] = fmt.Sprintf("%d", pageSize)
		}
		return doInstrumentGet("/openapi/fundamentals/fund/dividends", q)
	},
}

func init() {
	fundDividendsCmd.Flags().String("symbol", "", "Symbol (required)")
	fundDividendsCmd.Flags().String("category", "US_STOCK", "Category: US_STOCK")
	fundDividendsCmd.Flags().Int("page-index", 1, "Page index (default 1)")
	fundDividendsCmd.Flags().Int("page-size", 10, "Number of entries per page (default 10, max 20)")
	fundDividendsCmd.MarkFlagRequired("symbol")
}

var fundRatingCmd = &cobra.Command{
	Use: "fund-rating", Short: "Get fund rating",
	RunE: func(cmd *cobra.Command, args []string) error {
		symbol, _ := cmd.Flags().GetString("symbol")
		category, _ := cmd.Flags().GetString("category")
		return doInstrumentGet("/openapi/fundamentals/fund/rating", map[string]string{
			"symbol": symbol, "category": category,
		})
	},
}

func init() {
	fundRatingCmd.Flags().String("symbol", "", "Symbol (required)")
	fundRatingCmd.Flags().String("category", "US_STOCK", "Category: US_STOCK")
	fundRatingCmd.MarkFlagRequired("symbol")
}

var fundSplitsCmd = &cobra.Command{
	Use: "fund-splits", Short: "Get fund split history",
	RunE: func(cmd *cobra.Command, args []string) error {
		symbol, _ := cmd.Flags().GetString("symbol")
		category, _ := cmd.Flags().GetString("category")
		return doInstrumentGet("/openapi/fundamentals/fund/splits", map[string]string{
			"symbol": symbol, "category": category,
		})
	},
}

func init() {
	fundSplitsCmd.Flags().String("symbol", "", "Symbol (required)")
	fundSplitsCmd.Flags().String("category", "US_STOCK", "Category: US_STOCK")
	fundSplitsCmd.MarkFlagRequired("symbol")
}

var fundFilesCmd = &cobra.Command{
	Use: "fund-files", Short: "Get fund documents and files",
	RunE: func(cmd *cobra.Command, args []string) error {
		symbol, _ := cmd.Flags().GetString("symbol")
		category, _ := cmd.Flags().GetString("category")
		return doInstrumentGet("/openapi/fundamentals/fund/files", map[string]string{
			"symbol": symbol, "category": category,
		})
	},
}

func init() {
	fundFilesCmd.Flags().String("symbol", "", "Symbol (required)")
	fundFilesCmd.Flags().String("category", "US_STOCK", "Category: US_STOCK")
	fundFilesCmd.MarkFlagRequired("symbol")
}

var fundAllocationCmd = &cobra.Command{
	Use: "fund-allocation", Short: "Get fund asset allocation",
	RunE: func(cmd *cobra.Command, args []string) error {
		symbol, _ := cmd.Flags().GetString("symbol")
		category, _ := cmd.Flags().GetString("category")
		return doInstrumentGet("/openapi/fundamentals/fund/allocation", map[string]string{
			"symbol": symbol, "category": category,
		})
	},
}

func init() {
	fundAllocationCmd.Flags().String("symbol", "", "Symbol (required)")
	fundAllocationCmd.Flags().String("category", "US_STOCK", "Category: US_STOCK")
	fundAllocationCmd.MarkFlagRequired("symbol")
}
