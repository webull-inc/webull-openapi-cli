// Copyright 2024-2026 Webull Inc. All rights reserved.
// Licensed under the Apache License, Version 2.0.
// See LICENSE file in the project root for full license information.

package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/webull-inc/webull-openapi-cli/internal/config"
)

func optionLeg(side, optionType, strike string) map[string]interface{} {
	return map[string]interface{}{
		"side":               side,
		"quantity":           "1",
		"symbol":             "AAPL",
		"strike_price":       strike,
		"option_expire_date": "2026-12-18",
		"instrument_type":    "OPTION",
		"option_type":        optionType,
		"market":             "US",
	}
}

func equityLeg() map[string]interface{} {
	return map[string]interface{}{
		"side":            "BUY",
		"quantity":        "100",
		"symbol":          "AAPL",
		"market":          "US",
		"instrument_type": "EQUITY",
	}
}

func optionOrderBody(strategy string, legs ...interface{}) map[string]interface{} {
	return map[string]interface{}{
		"account_id": "ACC1",
		"new_orders": []interface{}{map[string]interface{}{
			"client_order_id": "cli1",
			"combo_type":      "NORMAL",
			"option_strategy": strategy,
			"instrument_type": "OPTION",
			"market":          "US",
			"entrust_type":    "QTY",
			"symbol":          "AAPL",
			"side":            "BUY",
			"quantity":        "1",
			"order_type":      "LIMIT",
			"limit_price":     "1.5",
			"time_in_force":   "DAY",
			"legs":            legs,
		}},
	}
}

func singleLegBody() map[string]interface{} {
	return optionOrderBody("SINGLE", optionLeg("BUY", "CALL", "200"))
}

func verticalBody() map[string]interface{} {
	return optionOrderBody("VERTICAL", optionLeg("BUY", "CALL", "200"), optionLeg("SELL", "CALL", "210"))
}

func coveredStockBody() map[string]interface{} {
	return optionOrderBody("COVERED_STOCK", equityLeg(), optionLeg("SELL", "CALL", "210"))
}

func ironCondorBody() map[string]interface{} {
	return optionOrderBody("IRON_CONDOR",
		optionLeg("BUY", "PUT", "180"), optionLeg("SELL", "PUT", "190"),
		optionLeg("SELL", "CALL", "210"), optionLeg("BUY", "CALL", "220"))
}

func firstOrder(body map[string]interface{}) map[string]interface{} {
	return body["new_orders"].([]interface{})[0].(map[string]interface{})
}

func TestValidateOptionStrategyOrderBody(t *testing.T) {
	// US bodies are forwarded as-is, so shapes the CLI used to reject must now pass.
	usNumericFields := verticalBody()
	firstOrder(usNumericFields)["quantity"] = 1
	firstOrder(usNumericFields)["limit_price"] = 1.5
	usTwoOrders := singleLegBody()
	usTwoOrders["new_orders"] = append(usTwoOrders["new_orders"].([]interface{}), firstOrder(verticalBody()))
	usIOC := verticalBody()
	firstOrder(usIOC)["time_in_force"] = "IOC"
	singleTwoLegs := optionOrderBody("SINGLE", optionLeg("BUY", "CALL", "200"), optionLeg("SELL", "CALL", "210"))

	tests := []struct {
		name    string
		region  config.Region
		body    map[string]interface{}
		wantErr string
	}{
		{name: "us single leg", region: config.RegionUS, body: singleLegBody()},
		{name: "us vertical", region: config.RegionUS, body: verticalBody()},
		{name: "us covered stock", region: config.RegionUS, body: coveredStockBody()},
		{name: "us iron condor", region: config.RegionUS, body: ironCondorBody()},
		{name: "us unknown strategy passes through", region: config.RegionUS, body: optionOrderBody("FOO", optionLeg("BUY", "CALL", "200"))},
		{name: "us numeric fields pass through", region: config.RegionUS, body: usNumericFields},
		{name: "us multiple new_orders pass through", region: config.RegionUS, body: usTwoOrders},
		{name: "us IOC passes through", region: config.RegionUS, body: usIOC},
		{name: "us single with 2 legs passes through", region: config.RegionUS, body: singleTwoLegs},
		{name: "hk single leg", region: config.RegionHK, body: singleLegBody()},
		{name: "hk single with 2 legs", region: config.RegionHK, body: singleTwoLegs, wantErr: "legs must contain exactly 1 item"},
		{name: "hk vertical", region: config.RegionHK, body: verticalBody(), wantErr: "only supported in region 'us'"},
		{name: "th vertical", region: config.RegionTH, body: verticalBody(), wantErr: "only supported in region 'us'"},
		{name: "br vertical", region: config.RegionBR, body: verticalBody(), wantErr: "only supported in region 'us'"},
		{name: "mx vertical", region: config.RegionMX, body: verticalBody(), wantErr: "only supported in region 'us'"},
		{name: "jp vertical", region: config.RegionJP, body: verticalBody(), wantErr: "only supported in region 'us'"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateOptionStrategyOrderBody(tt.body, tt.region)
			checkError(t, err, tt.wantErr)
		})
	}
}

func replaceLegBody(legCount int) map[string]interface{} {
	order := map[string]interface{}{"client_order_id": "cli1"}
	if legCount >= 0 {
		legs := make([]interface{}, 0, legCount)
		for i := 0; i < legCount; i++ {
			legs = append(legs, map[string]interface{}{"id": "L", "quantity": "1"})
		}
		order["legs"] = legs
	}
	return map[string]interface{}{"account_id": "ACC1", "modify_orders": []interface{}{order}}
}

func TestValidateOptionReplaceLegScope(t *testing.T) {
	noLegs := map[string]interface{}{"account_id": "ACC1", "modify_orders": []interface{}{map[string]interface{}{"client_order_id": "cli1"}}}
	tests := []struct {
		name    string
		region  config.Region
		body    map[string]interface{}
		wantErr string
	}{
		{name: "us multi-leg passes through", region: config.RegionUS, body: replaceLegBody(4)},
		{name: "us single leg", region: config.RegionUS, body: replaceLegBody(1)},
		{name: "hk single leg", region: config.RegionHK, body: replaceLegBody(1)},
		{name: "hk no legs", region: config.RegionHK, body: noLegs},
		{name: "hk two legs rejected", region: config.RegionHK, body: replaceLegBody(2), wantErr: "only supported in region 'us'"},
		{name: "th two legs rejected", region: config.RegionTH, body: replaceLegBody(2), wantErr: "only supported in region 'us'"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateOptionReplaceLegScope(tt.body, tt.region)
			checkError(t, err, tt.wantErr)
		})
	}
}

func TestActiveOptionRegion(t *testing.T) {
	for _, tt := range []struct {
		region  string
		want    config.Region
		wantErr bool
	}{
		{region: "us", want: config.RegionUS},
		{region: "US", want: config.RegionUS},
		{region: " Us ", want: config.RegionUS},
		{region: "hk", want: config.RegionHK},
		{region: "HK", want: config.RegionHK},
		{region: "au", wantErr: true},
		{region: "AU", wantErr: true},
		{region: "za", wantErr: true},
		{region: "ZA", wantErr: true},
	} {
		t.Run(tt.region, func(t *testing.T) {
			home := t.TempDir()
			t.Setenv("HOME", home)
			dir := filepath.Join(home, ".config", "webull")
			if err := os.MkdirAll(dir, 0700); err != nil {
				t.Fatal(err)
			}
			profiles := `{"version":"2.0","profile":{"auth_method":"ak_sk","region":"` + tt.region + `","env":"prod"}}`
			if err := os.WriteFile(filepath.Join(dir, "profiles.json"), []byte(profiles), 0600); err != nil {
				t.Fatal(err)
			}
			got, err := activeOptionRegion()
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error for region %q, got nil", tt.region)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.want {
				t.Fatalf("got region %q, want %q", got, tt.want)
			}
		})
	}
}

func checkError(t *testing.T, err error, wantErr string) {
	t.Helper()
	if wantErr == "" {
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		return
	}
	if err == nil {
		t.Fatalf("expected error containing %q, got nil", wantErr)
	}
	if !strings.Contains(err.Error(), wantErr) {
		t.Fatalf("expected error containing %q, got %q", wantErr, err.Error())
	}
}
