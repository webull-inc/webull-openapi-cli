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
	RegionTH Region = "th"
	RegionMY Region = "my"
	RegionUK Region = "uk"
	RegionBR Region = "br"
	RegionMX Region = "mx"
	RegionZA Region = "za"
)

// ValidRegions is the list of currently supported regions.
//
// Note: EU and AU are not supported yet due to compliance
// requirements, so they are intentionally omitted here and fail IsValid.
// To enable a region later, add its constant back to this list.
var ValidRegions = []Region{
	RegionUS, RegionHK, RegionJP, RegionSG, RegionTH,
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
	ModuleAPI ApiModule = "api"
)

// Environment identifies a deployment environment.
//
// prod maps to each region's production Trading API host.
// uat maps to the test host (sandbox for regions that have one, uat otherwise);
// for the CLI both are treated as the single non-production environment.
type Environment string

const (
	EnvProd Environment = "prod"
	EnvUAT  Environment = "uat"
)

// prodAPIEndpoints maps region to its production Trading API host.
var prodAPIEndpoints = map[Region]string{
	RegionUS: "api.webull.com",
	RegionHK: "api.webull.hk",
	RegionJP: "api.webull.co.jp",
	RegionSG: "api.webull.com.sg",
	RegionTH: "api.webull.co.th",
	// RegionAU: "api.webull.com.au", // not supported yet (compliance); re-add the region constant to enable
	RegionMY: "api.webull.com.my",
	RegionUK: "api.webull-uk.com",
	RegionBR: "api.webull.com",
	RegionMX: "api.webull.com",
	// RegionEU: "api.webull.eu",     // not supported yet (compliance); re-add the region constant to enable
	RegionZA: "api.webull.com.au",
}

// uatAPIEndpoints maps region to its test Trading API host.
// US/HK expose a dedicated sandbox; the remaining regions use uat.
var uatAPIEndpoints = map[Region]string{
	RegionUS: "api.sandbox.webull.com",
	RegionHK: "api.sandbox.webull.hk",
	RegionJP: "jp-openapi-alb.uat.webullbroker.com",
	RegionSG: "sg-api.uat.webullbroker.com",
	RegionTH: "th-api.uat.webullbroker.com",
	// RegionAU: "au-api.uat.webullbroker.com", // not supported yet (compliance); re-add the region constant to enable
	RegionMY: "my-api.uat.webullbroker.com",
	RegionUK: "uk-api.uat.webullbroker.com",
	RegionBR: "us-openapi-alb.uat.webullbroker.com",
	RegionMX: "us-openapi-alb.uat.webullbroker.com",
	// RegionEU: "eu-api.uat.webullbroker.com", // not supported yet (compliance); re-add the region constant to enable
	RegionZA: "au-api.uat.webullbroker.com",
}

// defaultAPIEndpoint returns the hardcoded Trading API host for a region/env.
func defaultAPIEndpoint(region Region, env Environment) string {
	if env == EnvProd {
		return prodAPIEndpoints[region]
	}
	return uatAPIEndpoints[region]
}

// ResolveAPIEndpoint resolves the Trading API host for a region/env.
//
// Resolution rules:
//   - prod: always the hardcoded production host; the file override is ignored.
//   - uat:  the file override wins when non-empty, otherwise the hardcoded uat host.
//
// The override is the endpoints.api value persisted in profiles.json. It exists
// as an internal testing backdoor and is not configurable via CLI flags.
func ResolveAPIEndpoint(region Region, env Environment, override string) string {
	if env == EnvUAT && override != "" {
		return override
	}
	return defaultAPIEndpoint(region, env)
}
