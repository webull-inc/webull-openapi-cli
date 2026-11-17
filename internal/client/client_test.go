// Copyright 2024-2026 Webull Inc. All rights reserved.
// Licensed under the Apache License, Version 2.0.
// See LICENSE file in the project root for full license information.

package client

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

// noopProvider authenticates nothing; these tests only look at the request path.
type noopProvider struct{}

func (noopProvider) Authenticate(req *http.Request, body []byte) error { return nil }

func (noopProvider) HandleAuthError(resp *http.Response) (bool, error) { return false, nil }

func TestDoPathPrefixByAuthMethod(t *testing.T) {
	tests := []struct {
		name       string
		authMethod string
		wantPath   string
	}{
		{"oauth adds the gateway prefix", "oauth", "/openapi-oauth-cli/trading/accounts/list"},
		{"ak_sk keeps the plain path", "ak_sk", "/trading/accounts/list"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var gotPath string
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				gotPath = r.URL.Path
				w.Header().Set("Content-Type", "application/json")
				_, _ = w.Write([]byte(`{}`))
			}))
			defer srv.Close()

			c := NewAPIClient(noopProvider{}, srv.URL, 5*time.Second, tt.authMethod)
			if _, err := c.Do(context.Background(), "GET", "/trading/accounts/list", nil); err != nil {
				t.Fatalf("Do returned error: %v", err)
			}
			if gotPath != tt.wantPath {
				t.Fatalf("request path = %q, want %q", gotPath, tt.wantPath)
			}
		})
	}
}
