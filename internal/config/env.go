// Copyright 2024-2026 Webull Inc. All rights reserved.
// Licensed under the Apache License, Version 2.0.
// See LICENSE file in the project root for full license information.

package config

import "os"

const (
	EnvAppKey         = "WEBULL_APP_KEY"
	EnvAppSecret      = "WEBULL_APP_SECRET"
	EnvAPIEndpoint    = "WEBULL_API_ENDPOINT"
	EnvQuotesEndpoint = "WEBULL_QUOTES_ENDPOINT"
	EnvEventsEndpoint = "WEBULL_EVENTS_ENDPOINT"
	EnvProfile        = "WEBULL_PROFILE"
)

// EnvCredentials attempts to load AK/SK from environment variables.
// Returns app_key, app_secret, ok.
func EnvCredentials() (string, string, bool) {
	key := os.Getenv(EnvAppKey)
	secret := os.Getenv(EnvAppSecret)
	if key != "" && secret != "" {
		return key, secret, true
	}
	return "", "", false
}

// EnvEndpointOverrides returns endpoint overrides from environment variables.
func EnvEndpointOverrides() *RegionEndpoints {
	api := os.Getenv(EnvAPIEndpoint)
	quotes := os.Getenv(EnvQuotesEndpoint)
	events := os.Getenv(EnvEventsEndpoint)

	if api == "" && quotes == "" && events == "" {
		return nil
	}
	return &RegionEndpoints{
		API:    api,
		Quotes: quotes,
		Events: events,
	}
}
