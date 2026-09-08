// Copyright 2024-2026 Webull Inc. All rights reserved.
// Licensed under the Apache License, Version 2.0.
// See LICENSE file in the project root for full license information.

package config

// Region identifies a Webull deployment region.
type Region string

const (
	RegionUS Region = "us"
	RegionHK Region = "hk"
	RegionJP Region = "jp"
	RegionSG Region = "sg"
	RegionMY Region = "my"
	RegionUK Region = "uk"
	RegionBR Region = "br"
	RegionMX Region = "mx"
	RegionZA Region = "za"
)

// ValidRegions is the list of currently supported regions.
//
// Note: TH, EU, and AU are not supported yet due to compliance
// requirements, so they are intentionally omitted here and fail IsValid.
// To enable a region later, add its constant back to this list.
var ValidRegions = []Region{
	RegionUS, RegionHK, RegionJP, RegionSG,
	RegionMY, RegionUK, RegionBR, RegionMX, RegionZA,
}

// IsValid checks if a region string is supported.
func (r Region) IsValid() bool {
	for _, v := range ValidRegions {
		if r == v {
			return true
		}
	}
	return false
}

// ApiModule identifies an API service type.
type ApiModule string

const (
	ModuleAPI    ApiModule = "api"
	ModuleQuotes ApiModule = "quotes-api"
	ModuleEvents ApiModule = "events-api"
)

// Environment identifies a deployment environment.
type Environment string

const (
	EnvProd Environment = "prod"
	EnvUAT  Environment = "uat"
)

// RegionEndpoints holds the three endpoints for a region.
type RegionEndpoints struct {
	API    string
	Quotes string
	Events string
}

// defaultEndpoints maps region to production endpoints.
var defaultEndpoints = map[Region]RegionEndpoints{
	RegionUS: {API: "api.webull.com", Quotes: "data-api.webull.com", Events: "events-api.webull.com"},
	RegionHK: {API: "api.webull.hk", Quotes: "data-api.webull.hk", Events: "events-api.webull.hk"},
	RegionJP: {API: "api.webull.co.jp", Quotes: "data-api.webull.co.jp", Events: "events-api.webull.co.jp"},
	RegionSG: {API: "api.webull.com.sg", Quotes: "data-api.webull.com.sg", Events: "events-api.webull.com.sg"},
	RegionMY: {API: "api.webull.com.my", Quotes: "data-api.webull.com.my", Events: "events-api.webull.com.my"},
	RegionUK: {API: "api.webull-uk.com", Quotes: "data-api.webull-uk.com", Events: "events-api.webull-uk.com"},
	RegionBR: {API: "api.webull.com", Quotes: "data-api.webull.com", Events: "events-api.webull.com"},
	RegionMX: {API: "api.webull.com", Quotes: "data-api.webull.com", Events: "events-api.webull.com"},
	RegionZA: {API: "api.webull.com.au", Quotes: "data-api.webull.com.au", Events: "events-api.webull.com.au"},
}

// GetEndpoints resolves endpoints for a region, applying overrides.
func GetEndpoints(region Region, overrides *RegionEndpoints) RegionEndpoints {
	ep := defaultEndpoints[region]
	if overrides != nil {
		if overrides.API != "" {
			ep.API = overrides.API
		}
		if overrides.Quotes != "" {
			ep.Quotes = overrides.Quotes
		}
		if overrides.Events != "" {
			ep.Events = overrides.Events
		}
	}
	return ep
}
