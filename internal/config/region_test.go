// Copyright 2024-2026 Webull Inc. All rights reserved.
// Licensed under the Apache License, Version 2.0.
// See LICENSE file in the project root for full license information.
package config

import "testing"

func TestRegionAU(t *testing.T) {
	if !RegionAU.IsValid() {
		t.Fatal("au should be a valid region")
	}
	if got := ResolveAPIEndpoint(RegionAU, EnvProd, ""); got != "api.webull.com.au" {
		t.Errorf("au prod endpoint = %q", got)
	}
	if got := ResolveAPIEndpoint(RegionAU, EnvUAT, ""); got != "au-api.uat.webullbroker.com" {
		t.Errorf("au uat endpoint = %q", got)
	}
	if Region("eu").IsValid() {
		t.Error("eu should remain unsupported")
	}
}
