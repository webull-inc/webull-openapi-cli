// Copyright 2024-2026 Webull Inc. All rights reserved.
// Licensed under the Apache License, Version 2.0.
// See LICENSE file in the project root for full license information.

package config

// OAuthEndpoints holds the OAuth-specific endpoints for a given environment.
type OAuthEndpoints struct {
	// Token is the authorization server host for device code and token operations.
	Token string
	// H5 is the authorization page base URL (scheme + host + path).
	H5 string
}

// prodOAuthEndpoints / sandboxOAuthEndpoints hold the hardcoded OAuth endpoints.
// Only prod and sandbox values live in the open-source code; internal test
// environments are reached through the profiles.json override (uat only).
var (
	prodOAuthEndpoints = OAuthEndpoints{
		Token: "u1suserauth.webullfintech.com",
		H5:    "passport.webull.com/oauth2/ai-cli/login",
	}
	sandboxOAuthEndpoints = OAuthEndpoints{
		Token: "u1suserauth.sandbox.webullfintech.com",
		H5:    "passport.webull.com/sandbox/oauth2/ai-cli/login",
	}
)

// defaultOAuthEndpoints returns the hardcoded OAuth endpoints for an environment.
func defaultOAuthEndpoints(env Environment) OAuthEndpoints {
	if env == EnvProd {
		return prodOAuthEndpoints
	}
	return sandboxOAuthEndpoints
}

// ResolveOAuthEndpoints resolves OAuth token/h5 endpoints for an environment.
//
// Resolution rules:
//   - prod: always the hardcoded production endpoints; overrides are ignored.
//   - uat:  each override field wins when non-empty, otherwise the hardcoded
//     sandbox value.
//
// The override carries the token/h5 values persisted in profiles.json. It exists
// as an internal testing backdoor and is not configurable via CLI flags.
func ResolveOAuthEndpoints(env Environment, override *OAuthEndpoints) OAuthEndpoints {
	ep := defaultOAuthEndpoints(env)
	if env == EnvUAT && override != nil {
		if override.Token != "" {
			ep.Token = override.Token
		}
		if override.H5 != "" {
			ep.H5 = override.H5
		}
	}
	return ep
}
