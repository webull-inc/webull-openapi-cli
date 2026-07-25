// Copyright 2024-2026 Webull Inc. All rights reserved.
// Licensed under the Apache License, Version 2.0.
// See LICENSE file in the project root for full license information.

package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"
	"github.com/webull-inc/webull-openapi-cli/internal/client"
	"github.com/webull-inc/webull-openapi-cli/internal/output"
)

// ==================== Top-level order command ====================

var orderCmd = &cobra.Command{
	Use:   "order",
	Short: "Trade orders (place, modify, cancel, query)",
}

// --- Per-asset sub-groups ---
var orderStockCmd = &cobra.Command{Use: "stock", Short: "Stock/ETF order operations"}
var orderOptionCmd = &cobra.Command{Use: "option", Short: "Option order operations"}
var orderFuturesCmd = &cobra.Command{Use: "futures", Short: "Futures order operations"}
var orderCryptoCmd = &cobra.Command{Use: "crypto", Short: "Crypto order operations"}
var orderEventCmd = &cobra.Command{Use: "event", Short: "Event contract order operations"}

// Top-level init: register sub-commands only
func init() {
	orderCmd.AddCommand(orderStockCmd)
	orderCmd.AddCommand(orderOptionCmd)
	orderCmd.AddCommand(orderFuturesCmd)
	orderCmd.AddCommand(orderCryptoCmd)
	orderCmd.AddCommand(orderEventCmd)
	orderCmd.AddCommand(orderCancelCmd)
	orderCmd.AddCommand(orderListCmd)
	orderCmd.AddCommand(orderDetailCmd)
	orderCmd.AddCommand(orderHistoryCmd)

	orderStockCmd.AddCommand(orderStockSubmitCmd)
	orderStockCmd.AddCommand(orderStockPreviewCmd)
	orderStockCmd.AddCommand(orderStockReplaceCmd)
	orderStockCmd.AddCommand(orderStockComboCmd)

	orderOptionCmd.AddCommand(orderOptionSubmitCmd)
	orderOptionCmd.AddCommand(orderOptionPreviewCmd)
	orderOptionCmd.AddCommand(orderOptionReplaceCmd)
	orderOptionCmd.AddCommand(orderOptionStrategyCmd)

	orderFuturesCmd.AddCommand(orderFuturesSubmitCmd)
	orderFuturesCmd.AddCommand(orderFuturesReplaceCmd)

	orderCryptoCmd.AddCommand(orderCryptoSubmitCmd)

	orderEventCmd.AddCommand(orderEventSubmitCmd)
	orderEventCmd.AddCommand(orderEventReplaceCmd)
}

// ==================== Stock Order Commands ====================

var orderStockSubmitCmd = &cobra.Command{
	Use:   "submit",
	Short: "Place a stock order",
	RunE: func(cmd *cobra.Command, args []string) error {
		accountID, _ := cmd.Flags().GetString("account-id")
		symbol, _ := cmd.Flags().GetString("symbol")
		side, _ := cmd.Flags().GetString("side")
		qty, _ := cmd.Flags().GetString("qty")
		orderType, _ := cmd.Flags().GetString("type")
		limitPrice, _ := cmd.Flags().GetString("limit-price")
		stopPrice, _ := cmd.Flags().GetString("stop-price")
		tif, _ := cmd.Flags().GetString("time-in-force")
		expireDate, _ := cmd.Flags().GetString("expire-date")
		market, _ := cmd.Flags().GetString("market")
		session, _ := cmd.Flags().GetString("session")
		clientOrderID, _ := cmd.Flags().GetString("client-order-id")
		trailingType, _ := cmd.Flags().GetString("trailing-type")
		trailingStep, _ := cmd.Flags().GetString("trailing-stop-step")
		accountTaxType, _ := cmd.Flags().GetString("account-tax-type")
		marginType, _ := cmd.Flags().GetString("margin-type")
		positionIntent, _ := cmd.Flags().GetString("position-intent")
		closeContracts, _ := cmd.Flags().GetString("close-contracts")
		algoType, _ := cmd.Flags().GetString("algo-type")
		targetVolPercent, _ := cmd.Flags().GetString("target-vol-percent")
		maxTargetPercent, _ := cmd.Flags().GetString("max-target-percent")
		algoStartTime, _ := cmd.Flags().GetString("algo-start-time")
		algoEndTime, _ := cmd.Flags().GetString("algo-end-time")
		entrustType, _ := cmd.Flags().GetString("entrust-type")
		totalCashAmount, _ := cmd.Flags().GetString("total-cash-amount")
		dryRun, _ := cmd.Flags().GetBool("dry-run")

		if clientOrderID == "" {
			clientOrderID = fmt.Sprintf("cli_stock_%d", time.Now().UnixMilli())
		}

		ob := map[string]interface{}{
			"client_order_id": clientOrderID, "symbol": symbol, "side": side,
			"order_type": orderType, "time_in_force": tif,
			"market": market, "instrument_type": "EQUITY", "entrust_type": entrustType, "combo_type": "NORMAL",
		}
		if qty != "" {
			ob["quantity"] = qty
		}
		if totalCashAmount != "" {
			ob["total_cash_amount"] = totalCashAmount
		}
		if limitPrice != "" {
			ob["limit_price"] = limitPrice
		}
		if stopPrice != "" {
			ob["stop_price"] = stopPrice
		}
		if session != "" {
			ob["support_trading_session"] = session
		}
		if expireDate != "" {
			ob["expire_date"] = expireDate
		}
		if trailingType != "" {
			ob["trailing_type"] = trailingType
		}
		if trailingStep != "" {
			ob["trailing_stop_step"] = trailingStep
		}
		trailingLimitOffset, _ := cmd.Flags().GetString("trailing-limit-price-offset")
		if trailingLimitOffset != "" {
			ob["trailing_limit_price_offset"] = trailingLimitOffset
		}
		triggerPriceType, _ := cmd.Flags().GetString("trigger-price-type")
		if triggerPriceType != "" {
			ob["trigger_price_type"] = triggerPriceType
		}
		if accountTaxType != "" {
			ob["account_tax_type"] = accountTaxType
		}
		if marginType != "" {
			ob["margin_type"] = marginType
		}
		if positionIntent != "" {
			ob["position_intent"] = positionIntent
		}
		if closeContracts != "" {
			var contracts []interface{}
			if err := json.Unmarshal([]byte(closeContracts), &contracts); err != nil {
				output.PrintError("INVALID_JSON", fmt.Sprintf("invalid --close-contracts JSON: %v", err), 0)
				return errSilent
			}
			ob["close_contracts"] = contracts
		}
		if algoType != "" {
			ob["algo_type"] = algoType
		}
		if targetVolPercent != "" {
			ob["target_vol_percent"] = targetVolPercent
		}
		if maxTargetPercent != "" {
			ob["max_target_percent"] = maxTargetPercent
		}
		if algoStartTime != "" {
			ob["algo_start_time"] = algoStartTime
		}
		if algoEndTime != "" {
			ob["algo_end_time"] = algoEndTime
		}

		body := map[string]interface{}{"account_id": accountID, "new_orders": []interface{}{ob}}

		if dryRun {
			return output.PrintJSON(map[string]interface{}{"dry_run": true, "request": map[string]interface{}{"method": "POST", "path": "/openapi/trade/order/place", "body": body}})
		}
		if !flagYes && !flagQuiet {
			fmt.Fprintf(os.Stderr, "Submit order: %s %s %s %s", side, qty, symbol, orderType)
			if limitPrice != "" {
				fmt.Fprintf(os.Stderr, " @ %s", limitPrice)
			}
			fmt.Fprint(os.Stderr, "? [y/N]: ")
			var c string
			fmt.Scanln(&c)
			if c != "y" && c != "Y" {
				fmt.Fprintln(os.Stderr, "Order cancelled.")
				return nil
			}
		}
		return doPlaceOrder(body)
	},
}

func init() {
	orderStockSubmitCmd.Flags().String("account-id", "", "Account ID (required)")
	orderStockSubmitCmd.Flags().String("symbol", "", "Symbol (required)")
	orderStockSubmitCmd.Flags().String("side", "", "Order side: BUY, SELL (required)")
	orderStockSubmitCmd.Flags().String("qty", "", "Quantity (required for QTY entrust type)")
	orderStockSubmitCmd.Flags().String("type", "", "Order type (required): MARKET, LIMIT, STOP_LOSS, STOP_LOSS_LIMIT, etc. See developer docs for full list by market")
	orderStockSubmitCmd.Flags().String("entrust-type", "QTY", "Entrust type: QTY (by quantity), AMOUNT (by cash amount)")
	orderStockSubmitCmd.Flags().String("total-cash-amount", "", "Total cash amount (for fractional shares, when entrust-type=AMOUNT)")
	orderStockSubmitCmd.Flags().String("limit-price", "", "Limit price")
	orderStockSubmitCmd.Flags().String("stop-price", "", "Stop price")
	orderStockSubmitCmd.Flags().String("time-in-force", "DAY", "Time in force: DAY, GTC, GTD")
	orderStockSubmitCmd.Flags().String("expire-date", "", "GTD order expire date (yyyy-MM-dd, required when time-in-force=GTD)")
	orderStockSubmitCmd.Flags().String("market", "US", "Market: US, HK, JP, CN, etc. See developer docs for full list by region")
	orderStockSubmitCmd.Flags().String("session", "", "Trading session: CORE, ALL, NIGHT, ALL_DAY")
	orderStockSubmitCmd.Flags().String("client-order-id", "", "Custom order ID (auto-generated if empty)")
	orderStockSubmitCmd.Flags().String("trailing-type", "", "Trailing type: AMOUNT, PERCENTAGE")
	orderStockSubmitCmd.Flags().String("trailing-stop-step", "", "Trailing stop step value")
	orderStockSubmitCmd.Flags().String("trailing-limit-price-offset", "", "Offset between stop price and limit price (for TRAILING_STOP_LOSS_LIMIT)")
	orderStockSubmitCmd.Flags().String("trigger-price-type", "", "Trigger price type: PRICE, PRICE_BID, PRICE_ASK")
	orderStockSubmitCmd.Flags().String("account-tax-type", "", "Account tax type (JP only): SPECIFIC, GENERAL")
	orderStockSubmitCmd.Flags().String("margin-type", "", "Margin type (JP only): CASH, MARGIN_STANDARD, MARGIN_NEGOTIABLE, MARGIN_DAY")
	orderStockSubmitCmd.Flags().String("position-intent", "", "Position intent: BUY_TO_OPEN, BUY_TO_CLOSE, SELL_TO_OPEN, SELL_TO_CLOSE")
	orderStockSubmitCmd.Flags().String("close-contracts", "", "Close contracts JSON array (JP only), e.g. '[{\"contract_id\":\"x\",\"quantity\":\"10\"}]'")
	orderStockSubmitCmd.Flags().String("algo-type", "", "Algorithm type: TWAP, VWAP, POV (algo orders only)")
	orderStockSubmitCmd.Flags().String("target-vol-percent", "", "Target volume participation % (1-20, POV only)")
	orderStockSubmitCmd.Flags().String("max-target-percent", "", "Max participation % (1-20, TWAP/VWAP)")
	orderStockSubmitCmd.Flags().String("algo-start-time", "", "Algo start time HH:mm:ss ET")
	orderStockSubmitCmd.Flags().String("algo-end-time", "", "Algo end time HH:mm:ss ET")
	orderStockSubmitCmd.Flags().Bool("dry-run", false, "Print request payload without submitting")
	orderStockSubmitCmd.MarkFlagRequired("account-id")
	orderStockSubmitCmd.MarkFlagRequired("symbol")
	orderStockSubmitCmd.MarkFlagRequired("side")
	orderStockSubmitCmd.MarkFlagRequired("type")
}

var orderStockPreviewCmd = &cobra.Command{
	Use: "preview", Short: "Preview stock order fees",
	RunE: func(cmd *cobra.Command, args []string) error {
		accountID, _ := cmd.Flags().GetString("account-id")
		symbol, _ := cmd.Flags().GetString("symbol")
		side, _ := cmd.Flags().GetString("side")
		qty, _ := cmd.Flags().GetString("qty")
		orderType, _ := cmd.Flags().GetString("type")
		limitPrice, _ := cmd.Flags().GetString("limit-price")
		stopPrice, _ := cmd.Flags().GetString("stop-price")
		tif, _ := cmd.Flags().GetString("time-in-force")
		market, _ := cmd.Flags().GetString("market")
		session, _ := cmd.Flags().GetString("session")
		clientOrderID, _ := cmd.Flags().GetString("client-order-id")
		accountTaxType, _ := cmd.Flags().GetString("account-tax-type")
		marginType, _ := cmd.Flags().GetString("margin-type")
		closeContracts, _ := cmd.Flags().GetString("close-contracts")
		if clientOrderID == "" {
			clientOrderID = fmt.Sprintf("cli_stock_preview_%d", time.Now().UnixMilli())
		}

		ob := map[string]interface{}{
			"client_order_id": clientOrderID, "combo_type": "NORMAL", "symbol": symbol,
			"instrument_type": "EQUITY", "market": market, "side": side,
			"order_type": orderType, "quantity": qty, "time_in_force": tif, "entrust_type": "QTY",
		}
		if limitPrice != "" {
			ob["limit_price"] = limitPrice
		}
		if stopPrice != "" {
			ob["stop_price"] = stopPrice
		}
		if session != "" {
			ob["support_trading_session"] = session
		}
		if accountTaxType != "" {
			ob["account_tax_type"] = accountTaxType
		}
		if marginType != "" {
			ob["margin_type"] = marginType
		}
		if closeContracts != "" {
			var contracts []interface{}
			if err := json.Unmarshal([]byte(closeContracts), &contracts); err != nil {
				output.PrintError("INVALID_JSON", fmt.Sprintf("invalid --close-contracts JSON: %v", err), 0)
				return errSilent
			}
			ob["close_contracts"] = contracts
		}

		return doPreviewOrder(map[string]interface{}{"account_id": accountID, "new_orders": []interface{}{ob}})
	},
}

func init() {
	orderStockPreviewCmd.Flags().String("account-id", "", "Account ID (required)")
	orderStockPreviewCmd.Flags().String("symbol", "", "Symbol (required)")
	orderStockPreviewCmd.Flags().String("side", "", "Order side (required)")
	orderStockPreviewCmd.Flags().String("qty", "", "Quantity (required)")
	orderStockPreviewCmd.Flags().String("type", "", "Order type (required)")
	orderStockPreviewCmd.Flags().String("limit-price", "", "Limit price")
	orderStockPreviewCmd.Flags().String("stop-price", "", "Stop price")
	orderStockPreviewCmd.Flags().String("time-in-force", "DAY", "Time in force")
	orderStockPreviewCmd.Flags().String("market", "US", "Market: US, HK, JP, CN, etc. See developer docs for full list by region")
	orderStockPreviewCmd.Flags().String("session", "", "Trading session")
	orderStockPreviewCmd.Flags().String("client-order-id", "", "Custom order ID")
	orderStockPreviewCmd.Flags().String("account-tax-type", "", "Account tax type (JP only): SPECIFIC, GENERAL")
	orderStockPreviewCmd.Flags().String("margin-type", "", "Margin type (JP only)")
	orderStockPreviewCmd.Flags().String("close-contracts", "", "Close contracts JSON array (JP only)")
	orderStockPreviewCmd.MarkFlagRequired("account-id")
	orderStockPreviewCmd.MarkFlagRequired("symbol")
	orderStockPreviewCmd.MarkFlagRequired("side")
	orderStockPreviewCmd.MarkFlagRequired("qty")
	orderStockPreviewCmd.MarkFlagRequired("type")
}

var orderStockReplaceCmd = &cobra.Command{
	Use: "replace", Short: "Modify an existing stock order",
	RunE: func(cmd *cobra.Command, args []string) error {
		return doReplaceOrder(cmd, "stock")
	},
}

func init() {
	orderStockReplaceCmd.Flags().String("account-id", "", "Account ID (required)")
	orderStockReplaceCmd.Flags().String("client-order-id", "", "Client order ID (required)")
	orderStockReplaceCmd.Flags().String("qty", "", "New quantity")
	orderStockReplaceCmd.Flags().String("limit-price", "", "New limit price")
	orderStockReplaceCmd.Flags().String("stop-price", "", "New stop price")
	orderStockReplaceCmd.Flags().String("order-type", "", "New order type")
	orderStockReplaceCmd.Flags().String("time-in-force", "", "New time in force")
	orderStockReplaceCmd.Flags().String("close-contracts", "", "Close contracts JSON array (JP only)")
	orderStockReplaceCmd.Flags().String("trailing-type", "", "Trailing type: AMOUNT, PERCENTAGE")
	orderStockReplaceCmd.Flags().String("trailing-stop-step", "", "Trailing stop spread (e.g. 1 for $1, 0.01 for 1%)")
	orderStockReplaceCmd.Flags().String("trailing-limit-price-offset", "", "Offset between stop price and limit price (for TRAILING_STOP_LOSS_LIMIT)")
	orderStockReplaceCmd.Flags().String("trigger-price-type", "", "Trigger price type: PRICE, PRICE_BID, PRICE_ASK")
	orderStockReplaceCmd.Flags().String("target-vol-percent", "", "Target volume participation % (1-20, POV algo)")
	orderStockReplaceCmd.Flags().String("max-target-percent", "", "Max participation % (1-20, TWAP/VWAP algo)")
	orderStockReplaceCmd.Flags().String("algo-start-time", "", "Algo start time HH:mm:ss ET")
	orderStockReplaceCmd.Flags().String("algo-end-time", "", "Algo end time HH:mm:ss ET")
	orderStockReplaceCmd.MarkFlagRequired("account-id")
	orderStockReplaceCmd.MarkFlagRequired("client-order-id")
}

var orderStockComboCmd = &cobra.Command{
	Use: "combo", Short: "Place a combo order (OTO/OCO/OTOCO) via JSON",
	RunE: func(cmd *cobra.Command, args []string) error { return doJSONOrder(cmd, "combo") },
}

func init() {
	orderStockComboCmd.Flags().String("body", "", "JSON request body")
	orderStockComboCmd.Flags().String("file", "", "Path to JSON file")
}

// ==================== Option Order Commands ====================

func buildOptionOrderBody(cmd *cobra.Command) map[string]interface{} {
	symbol, _ := cmd.Flags().GetString("symbol")
	side, _ := cmd.Flags().GetString("side")
	qty, _ := cmd.Flags().GetString("qty")
	optionType, _ := cmd.Flags().GetString("option-type")
	strike, _ := cmd.Flags().GetString("strike")
	expiry, _ := cmd.Flags().GetString("expiry")
	orderType, _ := cmd.Flags().GetString("order-type")
	limitPrice, _ := cmd.Flags().GetString("limit-price")
	stopPrice, _ := cmd.Flags().GetString("stop-price")
	tif, _ := cmd.Flags().GetString("time-in-force")
	market, _ := cmd.Flags().GetString("market")
	clientOrderID, _ := cmd.Flags().GetString("client-order-id")
	if clientOrderID == "" {
		clientOrderID = fmt.Sprintf("cli_option_%d", time.Now().UnixMilli())
	}

	leg := map[string]interface{}{
		"symbol": symbol, "side": side, "quantity": qty,
		"strike_price": strike, "option_expire_date": expiry,
		"option_type": optionType, "instrument_type": "OPTION", "market": market,
	}
	order := map[string]interface{}{
		"client_order_id": clientOrderID, "symbol": symbol, "side": side,
		"quantity": qty, "order_type": orderType, "time_in_force": tif,
		"market": market, "instrument_type": "OPTION", "combo_type": "NORMAL",
		"option_strategy": "SINGLE", "entrust_type": "QTY", "legs": []interface{}{leg},
	}
	if limitPrice != "" {
		order["limit_price"] = limitPrice
	}
	if stopPrice != "" {
		order["stop_price"] = stopPrice
	}
	if positionIntent, _ := cmd.Flags().GetString("position-intent"); positionIntent != "" {
		order["position_intent"] = positionIntent
	}
	return order
}

var orderOptionSubmitCmd = &cobra.Command{
	Use: "submit", Short: "Place a single-leg option order",
	RunE: func(cmd *cobra.Command, args []string) error {
		accountID, _ := cmd.Flags().GetString("account-id")
		order := buildOptionOrderBody(cmd)
		body := map[string]interface{}{"account_id": accountID, "new_orders": []interface{}{order}}

		if !flagYes && !flagQuiet {
			symbol, _ := cmd.Flags().GetString("symbol")
			side, _ := cmd.Flags().GetString("side")
			ot, _ := cmd.Flags().GetString("option-type")
			strike, _ := cmd.Flags().GetString("strike")
			expiry, _ := cmd.Flags().GetString("expiry")
			fmt.Fprintf(os.Stderr, "Submit option: %s %s %s $%s exp %s? [y/N]: ", side, symbol, ot, strike, expiry)
			var c string
			fmt.Scanln(&c)
			if c != "y" && c != "Y" {
				fmt.Fprintln(os.Stderr, "Order cancelled.")
				return nil
			}
		}
		return doPlaceOrder(body)
	},
}

func init() {
	orderOptionSubmitCmd.Flags().String("account-id", "", "Account ID (required)")
	orderOptionSubmitCmd.Flags().String("symbol", "", "Underlying symbol (required)")
	orderOptionSubmitCmd.Flags().String("side", "", "Order side: BUY, SELL (required)")
	orderOptionSubmitCmd.Flags().String("qty", "", "Number of contracts (required)")
	orderOptionSubmitCmd.Flags().String("option-type", "", "Option type: CALL, PUT (required)")
	orderOptionSubmitCmd.Flags().String("strike", "", "Strike price (required)")
	orderOptionSubmitCmd.Flags().String("expiry", "", "Expiration date YYYY-MM-DD (required)")
	orderOptionSubmitCmd.Flags().String("order-type", "LIMIT", "Order type: LIMIT, MARKET, STOP_LOSS, STOP_LOSS_LIMIT")
	orderOptionSubmitCmd.Flags().String("limit-price", "", "Limit price")
	orderOptionSubmitCmd.Flags().String("stop-price", "", "Stop price")
	orderOptionSubmitCmd.Flags().String("time-in-force", "DAY", "Time in force: DAY, GTC")
	orderOptionSubmitCmd.Flags().String("market", "US", "Market: US, HK, JP, CN, etc. See developer docs for full list by region")
	orderOptionSubmitCmd.Flags().String("client-order-id", "", "Custom order ID")
	orderOptionSubmitCmd.Flags().String("position-intent", "", "Position intent: BUY_TO_OPEN, BUY_TO_CLOSE, SELL_TO_OPEN, SELL_TO_CLOSE")
	orderOptionSubmitCmd.MarkFlagRequired("account-id")
	orderOptionSubmitCmd.MarkFlagRequired("symbol")
	orderOptionSubmitCmd.MarkFlagRequired("side")
	orderOptionSubmitCmd.MarkFlagRequired("qty")
	orderOptionSubmitCmd.MarkFlagRequired("option-type")
	orderOptionSubmitCmd.MarkFlagRequired("strike")
	orderOptionSubmitCmd.MarkFlagRequired("expiry")
}

var orderOptionPreviewCmd = &cobra.Command{
	Use: "preview", Short: "Preview option order fees",
	RunE: func(cmd *cobra.Command, args []string) error {
		accountID, _ := cmd.Flags().GetString("account-id")
		order := buildOptionOrderBody(cmd)
		return doPreviewOrder(map[string]interface{}{"account_id": accountID, "new_orders": []interface{}{order}})
	},
}

func init() {
	orderOptionPreviewCmd.Flags().String("account-id", "", "Account ID (required)")
	orderOptionPreviewCmd.Flags().String("symbol", "", "Underlying symbol (required)")
	orderOptionPreviewCmd.Flags().String("side", "", "Order side (required)")
	orderOptionPreviewCmd.Flags().String("qty", "", "Number of contracts (required)")
	orderOptionPreviewCmd.Flags().String("option-type", "", "Option type: CALL, PUT (required)")
	orderOptionPreviewCmd.Flags().String("strike", "", "Strike price (required)")
	orderOptionPreviewCmd.Flags().String("expiry", "", "Expiration date (required)")
	orderOptionPreviewCmd.Flags().String("order-type", "LIMIT", "Order type")
	orderOptionPreviewCmd.Flags().String("limit-price", "", "Limit price")
	orderOptionPreviewCmd.Flags().String("stop-price", "", "Stop price")
	orderOptionPreviewCmd.Flags().String("time-in-force", "DAY", "Time in force")
	orderOptionPreviewCmd.Flags().String("market", "US", "Market: US, HK, JP, CN, etc. See developer docs for full list by region")
	orderOptionPreviewCmd.Flags().String("client-order-id", "", "Custom order ID")
	orderOptionPreviewCmd.MarkFlagRequired("account-id")
	orderOptionPreviewCmd.MarkFlagRequired("symbol")
	orderOptionPreviewCmd.MarkFlagRequired("side")
	orderOptionPreviewCmd.MarkFlagRequired("qty")
	orderOptionPreviewCmd.MarkFlagRequired("option-type")
	orderOptionPreviewCmd.MarkFlagRequired("strike")
	orderOptionPreviewCmd.MarkFlagRequired("expiry")
}

var orderOptionReplaceCmd = &cobra.Command{
	Use: "replace", Short: "Modify an existing option order",
	Long: `Modify an option order. Use flags for single-leg, or --body/--file for multi-leg JSON.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		bodyStr, _ := cmd.Flags().GetString("body")
		fileStr, _ := cmd.Flags().GetString("file")
		if bodyStr != "" || fileStr != "" {
			return doJSONReplace(cmd)
		}

		accountID, _ := cmd.Flags().GetString("account-id")
		clientOrderID, _ := cmd.Flags().GetString("client-order-id")
		qty, _ := cmd.Flags().GetString("qty")
		limitPrice, _ := cmd.Flags().GetString("limit-price")
		stopPrice, _ := cmd.Flags().GetString("stop-price")
		orderType, _ := cmd.Flags().GetString("order-type")
		tif, _ := cmd.Flags().GetString("time-in-force")
		legID, _ := cmd.Flags().GetString("leg-id")
		legQty, _ := cmd.Flags().GetString("leg-qty")

		mo := map[string]interface{}{"client_order_id": clientOrderID}
		if qty != "" {
			mo["quantity"] = qty
		}
		if limitPrice != "" {
			mo["limit_price"] = limitPrice
		}
		if stopPrice != "" {
			mo["stop_price"] = stopPrice
		}
		if orderType != "" {
			mo["order_type"] = orderType
		}
		if tif != "" {
			mo["time_in_force"] = tif
		}
		if legID != "" {
			leg := map[string]interface{}{"id": legID}
			if legQty != "" {
				leg["quantity"] = legQty
			}
			mo["legs"] = []interface{}{leg}
		}

		body := map[string]interface{}{"account_id": accountID, "modify_orders": []interface{}{mo}}
		return doConfirmAndReplace(body, clientOrderID, "option")
	},
}

func init() {
	orderOptionReplaceCmd.Flags().String("account-id", "", "Account ID (required)")
	orderOptionReplaceCmd.Flags().String("client-order-id", "", "Client order ID (required)")
	orderOptionReplaceCmd.Flags().String("qty", "", "New quantity")
	orderOptionReplaceCmd.Flags().String("limit-price", "", "New limit price")
	orderOptionReplaceCmd.Flags().String("stop-price", "", "New stop price")
	orderOptionReplaceCmd.Flags().String("order-type", "", "New order type")
	orderOptionReplaceCmd.Flags().String("time-in-force", "", "New time in force")
	orderOptionReplaceCmd.Flags().String("leg-id", "", "Leg ID (US/BR/MX)")
	orderOptionReplaceCmd.Flags().String("leg-qty", "", "New leg quantity (US/BR/MX)")
	orderOptionReplaceCmd.Flags().String("body", "", "JSON body (multi-leg)")
	orderOptionReplaceCmd.Flags().String("file", "", "JSON file (multi-leg)")
}

var orderOptionStrategyCmd = &cobra.Command{
	Use: "strategy", Short: "Place a multi-leg option strategy order via JSON",
	RunE: func(cmd *cobra.Command, args []string) error { return doJSONOrder(cmd, "strategy") },
}

func init() {
	orderOptionStrategyCmd.Flags().String("body", "", "JSON request body")
	orderOptionStrategyCmd.Flags().String("file", "", "Path to JSON file")
}

// ==================== Futures Order Commands ====================

var orderFuturesSubmitCmd = &cobra.Command{
	Use: "submit", Short: "Place a futures order",
	RunE: func(cmd *cobra.Command, args []string) error {
		accountID, _ := cmd.Flags().GetString("account-id")
		symbol, _ := cmd.Flags().GetString("symbol")
		side, _ := cmd.Flags().GetString("side")
		qty, _ := cmd.Flags().GetString("qty")
		orderType, _ := cmd.Flags().GetString("type")
		limitPrice, _ := cmd.Flags().GetString("limit-price")
		stopPrice, _ := cmd.Flags().GetString("stop-price")
		tif, _ := cmd.Flags().GetString("time-in-force")
		market, _ := cmd.Flags().GetString("market")
		trailingType, _ := cmd.Flags().GetString("trailing-type")
		trailingStep, _ := cmd.Flags().GetString("trailing-stop-step")
		clientOrderID, _ := cmd.Flags().GetString("client-order-id")

		if clientOrderID == "" {
			clientOrderID = fmt.Sprintf("cli_futures_%d", time.Now().UnixMilli())
		}

		ob := map[string]interface{}{
			"client_order_id": clientOrderID, "symbol": symbol, "side": side, "quantity": qty, "order_type": orderType,
			"time_in_force": tif, "market": market, "instrument_type": "FUTURES", "combo_type": "NORMAL", "entrust_type": "QTY",
		}
		if limitPrice != "" {
			ob["limit_price"] = limitPrice
		}
		if stopPrice != "" {
			ob["stop_price"] = stopPrice
		}
		if trailingType != "" {
			ob["trailing_type"] = trailingType
		}
		if trailingStep != "" {
			ob["trailing_stop_step"] = trailingStep
		}

		body := map[string]interface{}{"account_id": accountID, "new_orders": []interface{}{ob}}
		if !flagYes && !flagQuiet {
			fmt.Fprintf(os.Stderr, "Submit futures order: %s %s %s %s? [y/N]: ", side, qty, symbol, orderType)
			var c string
			fmt.Scanln(&c)
			if c != "y" && c != "Y" {
				fmt.Fprintln(os.Stderr, "Order cancelled.")
				return nil
			}
		}
		return doPlaceOrder(body)
	},
}

func init() {
	orderFuturesSubmitCmd.Flags().String("account-id", "", "Account ID (required)")
	orderFuturesSubmitCmd.Flags().String("symbol", "", "Futures symbol (required)")
	orderFuturesSubmitCmd.Flags().String("side", "", "Order side: BUY, SELL (required)")
	orderFuturesSubmitCmd.Flags().String("qty", "", "Quantity (required)")
	orderFuturesSubmitCmd.Flags().String("type", "", "Order type: MARKET, LIMIT, STOP_LOSS, STOP_LOSS_LIMIT, TRAILING_STOP_LOSS (required)")
	orderFuturesSubmitCmd.Flags().String("limit-price", "", "Limit price")
	orderFuturesSubmitCmd.Flags().String("stop-price", "", "Stop price")
	orderFuturesSubmitCmd.Flags().String("time-in-force", "DAY", "Time in force")
	orderFuturesSubmitCmd.Flags().String("market", "US", "Market: US, HK, JP, CN, etc. See developer docs for full list by region")
	orderFuturesSubmitCmd.Flags().String("trailing-type", "", "Trailing type: AMOUNT, PERCENTAGE")
	orderFuturesSubmitCmd.Flags().String("trailing-stop-step", "", "Trailing stop spread (e.g. 1 for $1, 0.01 for 1%)")
	orderFuturesSubmitCmd.Flags().String("client-order-id", "", "Custom order ID (auto-generated if empty)")
	orderFuturesSubmitCmd.MarkFlagRequired("account-id")
	orderFuturesSubmitCmd.MarkFlagRequired("symbol")
	orderFuturesSubmitCmd.MarkFlagRequired("side")
	orderFuturesSubmitCmd.MarkFlagRequired("qty")
	orderFuturesSubmitCmd.MarkFlagRequired("type")
}

var orderFuturesReplaceCmd = &cobra.Command{
	Use: "replace", Short: "Modify a futures order",
	RunE: func(cmd *cobra.Command, args []string) error { return doReplaceOrder(cmd, "futures") },
}

func init() {
	orderFuturesReplaceCmd.Flags().String("account-id", "", "Account ID (required)")
	orderFuturesReplaceCmd.Flags().String("client-order-id", "", "Client order ID (required)")
	orderFuturesReplaceCmd.Flags().String("qty", "", "New quantity")
	orderFuturesReplaceCmd.Flags().String("limit-price", "", "New limit price")
	orderFuturesReplaceCmd.Flags().String("stop-price", "", "New stop price")
	orderFuturesReplaceCmd.Flags().String("order-type", "", "New order type")
	orderFuturesReplaceCmd.Flags().String("trailing-type", "", "Trailing type: AMOUNT, PERCENTAGE")
	orderFuturesReplaceCmd.Flags().String("trailing-stop-step", "", "Trailing stop spread")
	orderFuturesReplaceCmd.MarkFlagRequired("account-id")
	orderFuturesReplaceCmd.MarkFlagRequired("client-order-id")
}

// ==================== Crypto Order Commands ====================

var orderCryptoSubmitCmd = &cobra.Command{
	Use: "submit", Short: "Place a crypto order",
	RunE: func(cmd *cobra.Command, args []string) error {
		accountID, _ := cmd.Flags().GetString("account-id")
		symbol, _ := cmd.Flags().GetString("symbol")
		side, _ := cmd.Flags().GetString("side")
		qty, _ := cmd.Flags().GetString("qty")
		orderType, _ := cmd.Flags().GetString("type")
		limitPrice, _ := cmd.Flags().GetString("limit-price")
		stopPrice, _ := cmd.Flags().GetString("stop-price")
		tif, _ := cmd.Flags().GetString("time-in-force")
		market, _ := cmd.Flags().GetString("market")
		clientOrderID, _ := cmd.Flags().GetString("client-order-id")

		if clientOrderID == "" {
			clientOrderID = fmt.Sprintf("cli_crypto_%d", time.Now().UnixMilli())
		}

		ob := map[string]interface{}{
			"client_order_id": clientOrderID, "symbol": symbol, "side": side, "quantity": qty, "order_type": orderType,
			"time_in_force": tif, "market": market, "instrument_type": "CRYPTO", "combo_type": "NORMAL", "entrust_type": "QTY",
		}
		if limitPrice != "" {
			ob["limit_price"] = limitPrice
		}
		if stopPrice != "" {
			ob["stop_price"] = stopPrice
		}

		body := map[string]interface{}{"account_id": accountID, "new_orders": []interface{}{ob}}
		if !flagYes && !flagQuiet {
			fmt.Fprintf(os.Stderr, "Submit crypto order: %s %s %s %s? [y/N]: ", side, qty, symbol, orderType)
			var c string
			fmt.Scanln(&c)
			if c != "y" && c != "Y" {
				fmt.Fprintln(os.Stderr, "Order cancelled.")
				return nil
			}
		}
		return doPlaceOrder(body)
	},
}

func init() {
	orderCryptoSubmitCmd.Flags().String("account-id", "", "Account ID (required)")
	orderCryptoSubmitCmd.Flags().String("symbol", "", "Crypto symbol e.g. BTCUSD (required)")
	orderCryptoSubmitCmd.Flags().String("side", "", "Order side: BUY, SELL (required)")
	orderCryptoSubmitCmd.Flags().String("qty", "", "Quantity (required)")
	orderCryptoSubmitCmd.Flags().String("type", "", "Order type: MARKET, LIMIT, STOP_LOSS_LIMIT (required)")
	orderCryptoSubmitCmd.Flags().String("limit-price", "", "Limit price")
	orderCryptoSubmitCmd.Flags().String("stop-price", "", "Stop price")
	orderCryptoSubmitCmd.Flags().String("time-in-force", "GTC", "Time in force: DAY, GTC, IOC")
	orderCryptoSubmitCmd.Flags().String("market", "US", "Market: US, HK, JP, CN, etc. See developer docs for full list by region")
	orderCryptoSubmitCmd.Flags().String("client-order-id", "", "Custom order ID (auto-generated if empty)")
	orderCryptoSubmitCmd.MarkFlagRequired("account-id")
	orderCryptoSubmitCmd.MarkFlagRequired("symbol")
	orderCryptoSubmitCmd.MarkFlagRequired("side")
	orderCryptoSubmitCmd.MarkFlagRequired("qty")
	orderCryptoSubmitCmd.MarkFlagRequired("type")
}

// ==================== Event Order Commands ====================

var orderEventSubmitCmd = &cobra.Command{
	Use: "submit", Short: "Place an event contract order",
	RunE: func(cmd *cobra.Command, args []string) error {
		accountID, _ := cmd.Flags().GetString("account-id")
		symbol, _ := cmd.Flags().GetString("symbol")
		side, _ := cmd.Flags().GetString("side")
		qty, _ := cmd.Flags().GetString("qty")
		limitPrice, _ := cmd.Flags().GetString("limit-price")
		tif, _ := cmd.Flags().GetString("time-in-force")
		market, _ := cmd.Flags().GetString("market")
		eventOutcome, _ := cmd.Flags().GetString("event-outcome")
		clientOrderID, _ := cmd.Flags().GetString("client-order-id")

		if clientOrderID == "" {
			clientOrderID = fmt.Sprintf("cli_event_%d", time.Now().UnixMilli())
		}

		ob := map[string]interface{}{
			"client_order_id": clientOrderID, "symbol": symbol, "side": side, "quantity": qty, "order_type": "LIMIT",
			"limit_price": limitPrice, "time_in_force": tif, "market": market,
			"instrument_type": "EVENT", "combo_type": "NORMAL", "entrust_type": "QTY",
		}
		if eventOutcome != "" {
			ob["event_outcome"] = eventOutcome
		}
		body := map[string]interface{}{"account_id": accountID, "new_orders": []interface{}{ob}}
		if !flagYes && !flagQuiet {
			fmt.Fprintf(os.Stderr, "Submit event order: %s %s %s @ %s? [y/N]: ", side, qty, symbol, limitPrice)
			var c string
			fmt.Scanln(&c)
			if c != "y" && c != "Y" {
				fmt.Fprintln(os.Stderr, "Order cancelled.")
				return nil
			}
		}
		return doPlaceOrder(body)
	},
}

func init() {
	orderEventSubmitCmd.Flags().String("account-id", "", "Account ID (required)")
	orderEventSubmitCmd.Flags().String("symbol", "", "Event contract symbol (required)")
	orderEventSubmitCmd.Flags().String("side", "", "Order side: BUY, SELL (required)")
	orderEventSubmitCmd.Flags().String("qty", "", "Quantity (required)")
	orderEventSubmitCmd.Flags().String("limit-price", "", "Limit price (required)")
	orderEventSubmitCmd.Flags().String("time-in-force", "DAY", "Time in force")
	orderEventSubmitCmd.Flags().String("market", "US", "Market: US, HK, JP, CN, etc. See developer docs for full list by region")
	orderEventSubmitCmd.Flags().String("event-outcome", "", "Event outcome: yes, no (required for event orders)")
	orderEventSubmitCmd.Flags().String("client-order-id", "", "Custom order ID (auto-generated if empty)")
	orderEventSubmitCmd.MarkFlagRequired("account-id")
	orderEventSubmitCmd.MarkFlagRequired("symbol")
	orderEventSubmitCmd.MarkFlagRequired("side")
	orderEventSubmitCmd.MarkFlagRequired("qty")
	orderEventSubmitCmd.MarkFlagRequired("limit-price")
}

var orderEventReplaceCmd = &cobra.Command{
	Use: "replace", Short: "Modify an event contract order",
	RunE: func(cmd *cobra.Command, args []string) error {
		accountID, _ := cmd.Flags().GetString("account-id")
		clientOrderID, _ := cmd.Flags().GetString("client-order-id")
		qty, _ := cmd.Flags().GetString("qty")
		limitPrice, _ := cmd.Flags().GetString("limit-price")

		mo := map[string]interface{}{"client_order_id": clientOrderID}
		if qty != "" {
			mo["quantity"] = qty
		}
		if limitPrice != "" {
			mo["limit_price"] = limitPrice
		}
		body := map[string]interface{}{"account_id": accountID, "modify_orders": []interface{}{mo}}
		return doConfirmAndReplace(body, clientOrderID, "event")
	},
}

func init() {
	orderEventReplaceCmd.Flags().String("account-id", "", "Account ID (required)")
	orderEventReplaceCmd.Flags().String("client-order-id", "", "Client order ID (required)")
	orderEventReplaceCmd.Flags().String("qty", "", "New quantity")
	orderEventReplaceCmd.Flags().String("limit-price", "", "New limit price")
	orderEventReplaceCmd.MarkFlagRequired("account-id")
	orderEventReplaceCmd.MarkFlagRequired("client-order-id")
}

// ==================== Generic Order Commands ====================

var orderCancelCmd = &cobra.Command{
	Use: "cancel", Short: "Cancel an order",
	RunE: func(cmd *cobra.Command, args []string) error {
		accountID, _ := cmd.Flags().GetString("account-id")
		clientOrderID, _ := cmd.Flags().GetString("client-order-id")
		if !flagYes && !flagQuiet {
			fmt.Fprintf(os.Stderr, "Cancel order %s? [y/N]: ", clientOrderID)
			var c string
			fmt.Scanln(&c)
			if c != "y" && c != "Y" {
				fmt.Fprintln(os.Stderr, "Cancelled.")
				return nil
			}
		}
		cl, err := newAPIClient()
		if err != nil {
			return err
		}
		resp, err := cl.Do(context.Background(), "POST", "/openapi/trade/order/cancel", &client.RequestOptions{
			Body: map[string]interface{}{"account_id": accountID, "client_order_id": clientOrderID},
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

func init() {
	orderCancelCmd.Flags().String("account-id", "", "Account ID (required)")
	orderCancelCmd.Flags().String("client-order-id", "", "Client order ID (required)")
	orderCancelCmd.MarkFlagRequired("account-id")
	orderCancelCmd.MarkFlagRequired("client-order-id")
}

var orderListCmd = &cobra.Command{
	Use: "open", Short: "List open orders",
	RunE: func(cmd *cobra.Command, args []string) error {
		accountID, _ := cmd.Flags().GetString("account-id")
		pageSize, _ := cmd.Flags().GetInt("page-size")
		lastID, _ := cmd.Flags().GetString("last-client-order-id")
		q := map[string]string{"account_id": accountID, "page_size": fmt.Sprintf("%d", pageSize)}
		if lastID != "" {
			q["last_client_order_id"] = lastID
		}
		cl, err := newAPIClient()
		if err != nil {
			return err
		}
		resp, err := cl.Do(context.Background(), "GET", "/openapi/trade/order/open", &client.RequestOptions{
			Query: q,
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

func init() {
	orderListCmd.Flags().String("account-id", "", "Account ID (required)")
	orderListCmd.Flags().Int("page-size", 10, "Page size (max 200)")
	orderListCmd.Flags().String("last-client-order-id", "", "Last client order ID for pagination")
	orderListCmd.MarkFlagRequired("account-id")
}

var orderDetailCmd = &cobra.Command{
	Use: "detail", Short: "Get order detail",
	RunE: func(cmd *cobra.Command, args []string) error {
		accountID, _ := cmd.Flags().GetString("account-id")
		clientOrderID, _ := cmd.Flags().GetString("client-order-id")
		cl, err := newAPIClient()
		if err != nil {
			return err
		}
		resp, err := cl.Do(context.Background(), "GET", "/openapi/trade/order/detail", &client.RequestOptions{
			Query: map[string]string{"account_id": accountID, "client_order_id": clientOrderID},
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

func init() {
	orderDetailCmd.Flags().String("account-id", "", "Account ID (required)")
	orderDetailCmd.Flags().String("client-order-id", "", "Client order ID (required)")
	orderDetailCmd.MarkFlagRequired("account-id")
	orderDetailCmd.MarkFlagRequired("client-order-id")
}

var orderHistoryCmd = &cobra.Command{
	Use: "history", Short: "Query order history",
	RunE: func(cmd *cobra.Command, args []string) error {
		accountID, _ := cmd.Flags().GetString("account-id")
		start, _ := cmd.Flags().GetString("start")
		end, _ := cmd.Flags().GetString("end")
		limit, _ := cmd.Flags().GetInt("limit")
		lastID, _ := cmd.Flags().GetString("last-client-order-id")
		q := map[string]string{"account_id": accountID, "page_size": fmt.Sprintf("%d", limit)}
		if start != "" {
			q["start_date"] = start
		}
		if end != "" {
			q["end_date"] = end
		}
		if lastID != "" {
			q["last_client_order_id"] = lastID
		}
		cl, err := newAPIClient()
		if err != nil {
			return err
		}
		resp, err := cl.Do(context.Background(), "GET", "/openapi/trade/order/history", &client.RequestOptions{Query: q})
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

func init() {
	orderHistoryCmd.Flags().String("account-id", "", "Account ID (required)")
	orderHistoryCmd.Flags().String("start", "", "Start date (YYYY-MM-DD)")
	orderHistoryCmd.Flags().String("end", "", "End date (YYYY-MM-DD)")
	orderHistoryCmd.Flags().Int("limit", 20, "Max results")
	orderHistoryCmd.Flags().String("last-client-order-id", "", "Last client order ID for pagination")
	orderHistoryCmd.MarkFlagRequired("account-id")
}

// ==================== Helpers ====================

func doPlaceOrder(body map[string]interface{}) error {
	cl, err := newAPIClient()
	if err != nil {
		output.PrintError("CONFIG_ERROR", err.Error(), 0)
		return err
	}
	resp, err := cl.Do(context.Background(), "POST", "/openapi/trade/order/place", &client.RequestOptions{Body: body})
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

func doPreviewOrder(body map[string]interface{}) error {
	cl, err := newAPIClient()
	if err != nil {
		return err
	}
	resp, err := cl.Do(context.Background(), "POST", "/openapi/trade/order/preview", &client.RequestOptions{Body: body})
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

func doReplaceOrder(cmd *cobra.Command, asset string) error {
	accountID, _ := cmd.Flags().GetString("account-id")
	clientOrderID, _ := cmd.Flags().GetString("client-order-id")
	qty, _ := cmd.Flags().GetString("qty")
	limitPrice, _ := cmd.Flags().GetString("limit-price")
	stopPrice, _ := cmd.Flags().GetString("stop-price")
	orderType, _ := cmd.Flags().GetString("order-type")

	mo := map[string]interface{}{"client_order_id": clientOrderID}
	if qty != "" {
		mo["quantity"] = qty
	}
	if limitPrice != "" {
		mo["limit_price"] = limitPrice
	}
	if stopPrice != "" {
		mo["stop_price"] = stopPrice
	}
	if orderType != "" {
		mo["order_type"] = orderType
	}

	// time-in-force (optional, may not be defined on all replace commands)
	if tif, err := cmd.Flags().GetString("time-in-force"); err == nil && tif != "" {
		mo["time_in_force"] = tif
	}

	// close-contracts (JP only, optional)
	if closeContracts, err := cmd.Flags().GetString("close-contracts"); err == nil && closeContracts != "" {
		var contracts []interface{}
		if err := json.Unmarshal([]byte(closeContracts), &contracts); err != nil {
			output.PrintError("INVALID_JSON", fmt.Sprintf("invalid --close-contracts JSON: %v", err), 0)
			return errSilent
		}
		mo["close_contracts"] = contracts
	}

	// Trailing stop fields
	if v, err := cmd.Flags().GetString("trailing-type"); err == nil && v != "" {
		mo["trailing_type"] = v
	}
	if v, err := cmd.Flags().GetString("trailing-stop-step"); err == nil && v != "" {
		mo["trailing_stop_step"] = v
	}
	if v, err := cmd.Flags().GetString("trailing-limit-price-offset"); err == nil && v != "" {
		mo["trailing_limit_price_offset"] = v
	}
	if v, err := cmd.Flags().GetString("trigger-price-type"); err == nil && v != "" {
		mo["trigger_price_type"] = v
	}

	// Algo replace fields
	if v, err := cmd.Flags().GetString("target-vol-percent"); err == nil && v != "" {
		mo["target_vol_percent"] = v
	}
	if v, err := cmd.Flags().GetString("max-target-percent"); err == nil && v != "" {
		mo["max_target_percent"] = v
	}
	if v, err := cmd.Flags().GetString("algo-start-time"); err == nil && v != "" {
		mo["algo_start_time"] = v
	}
	if v, err := cmd.Flags().GetString("algo-end-time"); err == nil && v != "" {
		mo["algo_end_time"] = v
	}

	body := map[string]interface{}{"account_id": accountID, "modify_orders": []interface{}{mo}}
	return doConfirmAndReplace(body, clientOrderID, asset)
}

func doConfirmAndReplace(body map[string]interface{}, clientOrderID, asset string) error {
	if !flagYes && !flagQuiet {
		fmt.Fprintf(os.Stderr, "Replace %s order %s? [y/N]: ", asset, clientOrderID)
		var c string
		fmt.Scanln(&c)
		if c != "y" && c != "Y" {
			fmt.Fprintln(os.Stderr, "Cancelled.")
			return nil
		}
	}
	cl, err := newAPIClient()
	if err != nil {
		return err
	}
	resp, err := cl.Do(context.Background(), "POST", "/openapi/trade/order/replace", &client.RequestOptions{Body: body})
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

func doJSONOrder(cmd *cobra.Command, label string) error {
	body, err := readJSONInput(cmd)
	if err != nil {
		output.PrintError("INPUT_ERROR", err.Error(), 0)
		return err
	}
	if !flagYes && !flagQuiet {
		fmt.Fprintf(os.Stderr, "Submit %s order? [y/N]: ", label)
		var c string
		fmt.Scanln(&c)
		if c != "y" && c != "Y" {
			fmt.Fprintln(os.Stderr, "Order cancelled.")
			return nil
		}
	}
	return doPlaceOrder(body)
}

func doJSONReplace(cmd *cobra.Command) error {
	body, err := readJSONInput(cmd)
	if err != nil {
		output.PrintError("INPUT_ERROR", err.Error(), 0)
		return err
	}
	if !flagYes && !flagQuiet {
		fmt.Fprint(os.Stderr, "Replace order (JSON)? [y/N]: ")
		var c string
		fmt.Scanln(&c)
		if c != "y" && c != "Y" {
			fmt.Fprintln(os.Stderr, "Cancelled.")
			return nil
		}
	}
	cl, err := newAPIClient()
	if err != nil {
		return err
	}
	resp, err := cl.Do(context.Background(), "POST", "/openapi/trade/order/replace", &client.RequestOptions{Body: body})
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
