// Licensed under the Apache License, Version 2.0.
// See LICENSE file in the project root for full license information.

package cmd

import (
	"fmt"
	"math/big"
	"regexp"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/webull-inc/webull-openapi-cli/internal/config"
)

var (
	decimalPattern       = regexp.MustCompile(`^(?:0|[1-9][0-9]*)(?:\.[0-9]+)?$`)
	datePattern          = regexp.MustCompile(`^[0-9]{4}-[0-9]{2}-[0-9]{2}$`)
	timestampPattern     = regexp.MustCompile(`^[0-9]{4}-[0-9]{2}-[0-9]{2}T[0-9]{2}:[0-9]{2}:[0-9]{2}\.[0-9]{3}Z$`)
	clientOrderIDPattern = regexp.MustCompile(`^[A-Za-z0-9_-]+$`)
)

func validateEnum(name, value string, allowed ...string) error {
	for _, candidate := range allowed {
		if value == candidate {
			return nil
		}
	}
	return fmt.Errorf("invalid %s %q; allowed values: %s", name, value, strings.Join(allowed, ", "))
}

func validatePositiveDecimal(name, value string, integer bool) error {
	if !decimalPattern.MatchString(value) {
		return fmt.Errorf("%s must be a positive decimal number", name)
	}
	number, ok := new(big.Rat).SetString(value)
	if !ok || number.Sign() <= 0 {
		return fmt.Errorf("%s must be greater than zero", name)
	}
	if integer && !number.IsInt() {
		return fmt.Errorf("%s must be a positive whole number", name)
	}
	return nil
}

func validateDate(name, value string) error {
	if !datePattern.MatchString(value) {
		return fmt.Errorf("%s must use YYYY-MM-DD format", name)
	}
	parsed, err := time.Parse("2006-01-02", value)
	if err != nil || parsed.Format("2006-01-02") != value {
		return fmt.Errorf("%s must be a valid date in YYYY-MM-DD format", name)
	}
	return nil
}

func validateUTCTimestamp(name, value string) error {
	if !timestampPattern.MatchString(value) {
		return fmt.Errorf("%s must use YYYY-MM-DDTHH:mm:ss.SSSZ UTC format", name)
	}
	parsed, err := time.Parse("2006-01-02T15:04:05.000Z", value)
	if err != nil || parsed.UTC().Format("2006-01-02T15:04:05.000Z") != value {
		return fmt.Errorf("%s must be a valid UTC timestamp in YYYY-MM-DDTHH:mm:ss.SSSZ format", name)
	}
	return nil
}

func validateSymbols(name, value string, maximum int) error {
	symbols := strings.Split(value, ",")
	if len(symbols) > maximum {
		return fmt.Errorf("%s accepts at most %d symbols", name, maximum)
	}
	for _, symbol := range symbols {
		if strings.TrimSpace(symbol) == "" {
			return fmt.Errorf("%s must contain non-empty comma-separated symbols", name)
		}
	}
	return nil
}

func validatePaginationKey(value string) error {
	if len(value) > 500 {
		return fmt.Errorf("pagination key must not exceed 500 characters")
	}
	return nil
}

// activeOptionRegion returns the active profile region if it may trade US-listed options.
func activeOptionRegion() (config.Region, error) {
	profile, err := config.NewProfileManager("").GetActive()
	if err != nil {
		return "", err
	}
	// Match case-insensitively so profiles storing e.g. "US" keep working.
	region := config.Region(strings.ToLower(strings.TrimSpace(string(profile.Region))))
	switch region {
	case config.RegionUS, config.RegionHK, config.RegionJP, config.RegionSG,
		config.RegionTH, config.RegionMY, config.RegionUK, config.RegionBR, config.RegionMX:
		return region, nil
	default:
		return "", fmt.Errorf("US-listed options are not available for investor region %q", profile.Region)
	}
}

func validateOptionAccess() error {
	_, err := activeOptionRegion()
	return err
}

func validateClientOrderID(value string) error {
	if value == "" {
		return nil
	}
	if len(value) > 32 || !clientOrderIDPattern.MatchString(value) {
		return fmt.Errorf("client order ID must be at most 32 characters and contain only letters, digits, underscores, or hyphens")
	}
	return nil
}

func validateOptionOrderFlags(cmd *cobra.Command) error {
	if err := validateOptionAccess(); err != nil {
		return err
	}
	market, _ := cmd.Flags().GetString("market")
	if err := validateEnum("market", market, "US"); err != nil {
		return err
	}
	symbol, _ := cmd.Flags().GetString("symbol")
	if err := validateSymbols("symbol", symbol, 1); err != nil {
		return err
	}
	side, _ := cmd.Flags().GetString("side")
	if err := validateEnum("side", side, "BUY", "SELL"); err != nil {
		return err
	}
	quantity, _ := cmd.Flags().GetString("qty")
	if err := validatePositiveDecimal("quantity", quantity, true); err != nil {
		return err
	}
	optionType, _ := cmd.Flags().GetString("option-type")
	if err := validateEnum("option type", optionType, "CALL", "PUT"); err != nil {
		return err
	}
	strike, _ := cmd.Flags().GetString("strike")
	if err := validatePositiveDecimal("strike price", strike, false); err != nil {
		return err
	}
	expiry, _ := cmd.Flags().GetString("expiry")
	if err := validateDate("expiration date", expiry); err != nil {
		return err
	}
	orderType, _ := cmd.Flags().GetString("order-type")
	if err := validateEnum("order type", orderType, "LIMIT", "MARKET", "STOP_LOSS", "STOP_LOSS_LIMIT"); err != nil {
		return err
	}
	timeInForce, _ := cmd.Flags().GetString("time-in-force")
	if err := validateEnum("time in force", timeInForce, "DAY", "GTC"); err != nil {
		return err
	}
	clientOrderID, _ := cmd.Flags().GetString("client-order-id")
	if err := validateClientOrderID(clientOrderID); err != nil {
		return err
	}
	if positionIntent, err := cmd.Flags().GetString("position-intent"); err == nil && positionIntent != "" {
		if err := validateEnum("position intent", positionIntent, "BUY_TO_OPEN", "BUY_TO_CLOSE", "SELL_TO_OPEN", "SELL_TO_CLOSE"); err != nil {
			return err
		}
	}
	limitPrice, _ := cmd.Flags().GetString("limit-price")
	stopPrice, _ := cmd.Flags().GetString("stop-price")
	if limitPrice != "" {
		if err := validatePositiveDecimal("limit price", limitPrice, false); err != nil {
			return err
		}
	}
	if stopPrice != "" {
		if err := validatePositiveDecimal("stop price", stopPrice, false); err != nil {
			return err
		}
	}
	if (orderType == "LIMIT" || orderType == "STOP_LOSS_LIMIT") && limitPrice == "" {
		return fmt.Errorf("limit price is required for %s orders", orderType)
	}
	if (orderType == "STOP_LOSS" || orderType == "STOP_LOSS_LIMIT") && stopPrice == "" {
		return fmt.Errorf("stop price is required for %s orders", orderType)
	}
	return nil
}

func validateOptionReplaceFlags(cmd *cobra.Command) error {
	if err := validateOptionAccess(); err != nil {
		return err
	}
	accountID, _ := cmd.Flags().GetString("account-id")
	if accountID == "" {
		return fmt.Errorf("account ID is required")
	}
	clientOrderID, _ := cmd.Flags().GetString("client-order-id")
	if clientOrderID == "" {
		return fmt.Errorf("client order ID is required")
	}
	if err := validateClientOrderID(clientOrderID); err != nil {
		return err
	}
	for _, field := range []struct {
		flag string
		name string
		int  bool
	}{
		{flag: "qty", name: "quantity", int: true},
		{flag: "limit-price", name: "limit price"},
		{flag: "stop-price", name: "stop price"},
		{flag: "leg-qty", name: "leg quantity", int: true},
	} {
		value, _ := cmd.Flags().GetString(field.flag)
		if value != "" {
			if err := validatePositiveDecimal(field.name, value, field.int); err != nil {
				return err
			}
		}
	}
	if orderType, _ := cmd.Flags().GetString("order-type"); orderType != "" {
		if err := validateEnum("order type", orderType, "LIMIT", "MARKET", "STOP_LOSS", "STOP_LOSS_LIMIT"); err != nil {
			return err
		}
	}
	if timeInForce, _ := cmd.Flags().GetString("time-in-force"); timeInForce != "" {
		if err := validateEnum("time in force", timeInForce, "DAY", "GTC"); err != nil {
			return err
		}
	}
	legID, _ := cmd.Flags().GetString("leg-id")
	legQuantity, _ := cmd.Flags().GetString("leg-qty")
	if legQuantity != "" && legID == "" {
		return fmt.Errorf("leg ID is required when leg quantity is provided")
	}
	return nil
}

func requireStringField(value map[string]interface{}, field string) (string, error) {
	result, ok := value[field].(string)
	if !ok || result == "" {
		return "", fmt.Errorf("%s must be a non-empty string", field)
	}
	return result, nil
}

func requireObjectList(value map[string]interface{}, field string, count int) ([]interface{}, error) {
	result, ok := value[field].([]interface{})
	if !ok || len(result) != count {
		return nil, fmt.Errorf("%s must contain exactly %d item", field, count)
	}
	return result, nil
}

func optionalStringField(value map[string]interface{}, field string) (string, error) {
	entry, exists := value[field]
	if !exists {
		return "", nil
	}
	result, ok := entry.(string)
	if !ok || result == "" {
		return "", fmt.Errorf("%s must be a non-empty string when provided", field)
	}
	return result, nil
}

func validateOptionOrderBody(body map[string]interface{}) error {
	if _, err := requireStringField(body, "account_id"); err != nil {
		return err
	}
	orders, err := requireObjectList(body, "new_orders", 1)
	if err != nil {
		return err
	}
	order, ok := orders[0].(map[string]interface{})
	if !ok {
		return fmt.Errorf("new_orders item must be an object")
	}
	for field, expected := range map[string]string{
		"combo_type":      "NORMAL",
		"option_strategy": "SINGLE",
		"instrument_type": "OPTION",
		"market":          "US",
		"entrust_type":    "QTY",
	} {
		value, err := requireStringField(order, field)
		if err != nil {
			return err
		}
		if value != expected {
			return fmt.Errorf("%s must be %s for a single-leg US option order", field, expected)
		}
	}
	symbol, err := requireStringField(order, "symbol")
	if err != nil {
		return err
	}
	if err := validateSymbols("symbol", symbol, 1); err != nil {
		return err
	}
	side, err := requireStringField(order, "side")
	if err != nil {
		return err
	}
	if err := validateEnum("side", side, "BUY", "SELL"); err != nil {
		return err
	}
	quantity, err := requireStringField(order, "quantity")
	if err != nil {
		return err
	}
	if err := validatePositiveDecimal("quantity", quantity, true); err != nil {
		return err
	}
	orderType, err := requireStringField(order, "order_type")
	if err != nil {
		return err
	}
	if err := validateEnum("order type", orderType, "LIMIT", "MARKET", "STOP_LOSS", "STOP_LOSS_LIMIT"); err != nil {
		return err
	}
	timeInForce, err := requireStringField(order, "time_in_force")
	if err != nil {
		return err
	}
	if err := validateEnum("time in force", timeInForce, "DAY", "GTC"); err != nil {
		return err
	}
	clientOrderID, err := requireStringField(order, "client_order_id")
	if err != nil {
		return err
	}
	if err := validateClientOrderID(clientOrderID); err != nil {
		return err
	}
	limitPrice, err := optionalStringField(order, "limit_price")
	if err != nil {
		return err
	}
	stopPrice, err := optionalStringField(order, "stop_price")
	if err != nil {
		return err
	}
	if limitPrice != "" {
		if err := validatePositiveDecimal("limit price", limitPrice, false); err != nil {
			return err
		}
	}
	if stopPrice != "" {
		if err := validatePositiveDecimal("stop price", stopPrice, false); err != nil {
			return err
		}
	}
	if (orderType == "LIMIT" || orderType == "STOP_LOSS_LIMIT") && limitPrice == "" {
		return fmt.Errorf("limit_price is required for %s orders", orderType)
	}
	if (orderType == "STOP_LOSS" || orderType == "STOP_LOSS_LIMIT") && stopPrice == "" {
		return fmt.Errorf("stop_price is required for %s orders", orderType)
	}
	if positionIntent, err := optionalStringField(order, "position_intent"); err != nil {
		return err
	} else if positionIntent != "" {
		if err := validateEnum("position intent", positionIntent, "BUY_TO_OPEN", "BUY_TO_CLOSE", "SELL_TO_OPEN", "SELL_TO_CLOSE"); err != nil {
			return err
		}
	}
	legs, err := requireObjectList(order, "legs", 1)
	if err != nil {
		return err
	}
	leg, ok := legs[0].(map[string]interface{})
	if !ok {
		return fmt.Errorf("option leg must be an object")
	}
	for field, expected := range map[string]string{"instrument_type": "OPTION", "market": "US"} {
		value, err := requireStringField(leg, field)
		if err != nil {
			return err
		}
		if value != expected {
			return fmt.Errorf("option leg %s must be %s", field, expected)
		}
	}
	legSymbol, err := requireStringField(leg, "symbol")
	if err != nil {
		return err
	}
	if err := validateSymbols("option leg symbol", legSymbol, 1); err != nil {
		return err
	}
	legSide, err := requireStringField(leg, "side")
	if err != nil {
		return err
	}
	if err := validateEnum("option leg side", legSide, "BUY", "SELL"); err != nil {
		return err
	}
	legQuantity, err := requireStringField(leg, "quantity")
	if err != nil {
		return err
	}
	if err := validatePositiveDecimal("option leg quantity", legQuantity, true); err != nil {
		return err
	}
	optionType, err := requireStringField(leg, "option_type")
	if err != nil {
		return err
	}
	if err := validateEnum("option type", optionType, "CALL", "PUT"); err != nil {
		return err
	}
	strikePrice, err := requireStringField(leg, "strike_price")
	if err != nil {
		return err
	}
	if err := validatePositiveDecimal("strike price", strikePrice, false); err != nil {
		return err
	}
	expirationDate, err := requireStringField(leg, "option_expire_date")
	if err != nil {
		return err
	}
	return validateDate("option expiration date", expirationDate)
}

// validateOptionStrategyOrderBody validates a JSON option order body. In region 'us' the body
// is forwarded as-is (single-leg and multi-leg) and validated by the server; other option
// regions accept single-leg orders only.
func validateOptionStrategyOrderBody(body map[string]interface{}, region config.Region) error {
	if region == config.RegionUS {
		return nil
	}
	orders, err := requireObjectList(body, "new_orders", 1)
	if err != nil {
		return err
	}
	order, ok := orders[0].(map[string]interface{})
	if !ok {
		return fmt.Errorf("new_orders item must be an object")
	}
	strategy, _ := order["option_strategy"].(string)
	if strategy != "" && strategy != "SINGLE" {
		return fmt.Errorf("multi-leg option strategy %q is only supported in region 'us'; region %q supports single-leg options only", strategy, region)
	}
	return validateOptionOrderBody(body)
}
