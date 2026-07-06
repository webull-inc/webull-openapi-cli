// Copyright 2024-2026 Webull Inc. All rights reserved.
// Licensed under the Apache License, Version 2.0.
// See LICENSE file in the project root for full license information.

package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/webull-inc/webull-openapi-cli/internal/auth"
	"github.com/webull-inc/webull-openapi-cli/internal/config"
)

// APIClient is the HTTP client that handles signing, retries, and requests.
type APIClient struct {
	httpClient *http.Client
	provider   auth.CredentialProvider
	endpoints  config.RegionEndpoints
	verbose    bool
	debug      bool
}

const (
	// defaultClientSource identifies this client to the backend.
	defaultClientSource = "cli"
	// clientSourceEnv allows overriding the client source via environment variable.
	clientSourceEnv = "WEBULL_CLIENT_SOURCE"
)

// NewAPIClient creates a new API client.
func NewAPIClient(provider auth.CredentialProvider, endpoints config.RegionEndpoints, timeout time.Duration) *APIClient {
	return &APIClient{
		httpClient: &http.Client{Timeout: timeout},
		provider:   provider,
		endpoints:  endpoints,
	}
}

// SetVerbose enables verbose logging to stderr.
func (c *APIClient) SetVerbose(v bool) { c.verbose = v }

// SetDebug enables debug logging to stderr.
func (c *APIClient) SetDebug(d bool) { c.debug = d }

// RequestOptions holds options for an API request.
type RequestOptions struct {
	Module config.ApiModule
	Query  map[string]string
	Body   interface{}
}

// Response holds the API response.
type Response struct {
	StatusCode int
	Headers    http.Header
	Body       json.RawMessage
}

// Do executes an API request.
func (c *APIClient) Do(ctx context.Context, method, path string, opts *RequestOptions) (*Response, error) {
	if opts == nil {
		opts = &RequestOptions{}
	}

	// Resolve host based on module
	host := c.resolveHost(opts.Module)

	// Build URL
	url := fmt.Sprintf("https://%s%s", host, path)

	// Serialize body
	var bodyBytes []byte
	var bodyReader io.Reader
	if opts.Body != nil {
		var err error
		bodyBytes, err = json.Marshal(opts.Body)
		if err != nil {
			return nil, fmt.Errorf("failed to serialize request body: %w", err)
		}
		bodyReader = bytes.NewReader(bodyBytes)
	}

	// Create request
	req, err := http.NewRequestWithContext(ctx, method, url, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set query params
	if len(opts.Query) > 0 {
		q := req.URL.Query()
		for k, v := range opts.Query {
			q.Set(k, v)
		}
		req.URL.RawQuery = q.Encode()
	}

	// Content-Type
	if bodyBytes != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	// Authenticate
	if err := c.provider.Authenticate(req, bodyBytes); err != nil {
		return nil, fmt.Errorf("authentication failed: %w", err)
	}

	// Client source header
	req.Header.Set("x-webull-client-source", getClientSource())

	// Log verbose
	if c.verbose {
		fmt.Fprintf(os.Stderr, "%s https://%s%s\n", method, host, path)
	}

	// Execute with retry for GET
	resp, err := c.doWithRetry(ctx, req, method)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// Handle auth errors
	if resp.StatusCode == http.StatusUnauthorized {
		fakeResp := &http.Response{StatusCode: resp.StatusCode, Body: io.NopCloser(bytes.NewReader(respBody))}
		retry, authErr := c.provider.HandleAuthError(fakeResp)
		if retry {
			// Re-execute after token refresh
			req2, _ := http.NewRequestWithContext(ctx, method, url, bytes.NewReader(bodyBytes))
			req2.Header = req.Header.Clone()
			if err := c.provider.Authenticate(req2, bodyBytes); err != nil {
				return nil, err
			}
			resp2, err := c.httpClient.Do(req2)
			if err != nil {
				return nil, err
			}
			defer resp2.Body.Close()
			respBody, _ = io.ReadAll(resp2.Body)
			return &Response{StatusCode: resp2.StatusCode, Headers: resp2.Header, Body: respBody}, nil
		}
		if authErr != nil {
			return nil, authErr
		}
	}

	return &Response{
		StatusCode: resp.StatusCode,
		Headers:    resp.Header,
		Body:       respBody,
	}, nil
}

// doWithRetry executes with retry for GET requests only.
func (c *APIClient) doWithRetry(ctx context.Context, req *http.Request, method string) (*http.Response, error) {
	maxRetries := 0
	if method == http.MethodGet {
		maxRetries = 3
	}

	var lastErr error
	for attempt := 0; attempt <= maxRetries; attempt++ {
		resp, err := c.httpClient.Do(req)
		if err != nil {
			lastErr = err
			if c.verbose {
				fmt.Fprintf(os.Stderr, "[error] %s %s: %v\n", method, req.URL.String(), err)
			}
			if attempt < maxRetries {
				if c.verbose {
					fmt.Fprintf(os.Stderr, "[retry] attempt %d/%d, backing off...\n", attempt+1, maxRetries)
				}
				sleepWithBackoff(ctx, attempt)
				continue
			}
			return nil, fmt.Errorf("request failed after %d attempts: %w", attempt+1, lastErr)
		}

		if c.verbose {
			fmt.Fprintf(os.Stderr, "%s %s -> %d\n", method, req.URL.String(), resp.StatusCode)
		}

		// Retry on 5xx for GET only
		if method == http.MethodGet && resp.StatusCode >= 500 && attempt < maxRetries {
			resp.Body.Close()
			if c.verbose {
				fmt.Fprintf(os.Stderr, "[retry] server error %d, attempt %d/%d\n", resp.StatusCode, attempt+1, maxRetries)
			}
			sleepWithBackoff(ctx, attempt)
			continue
		}

		return resp, nil
	}
	return nil, lastErr
}

func (c *APIClient) resolveHost(module config.ApiModule) string {
	switch module {
	case config.ModuleQuotes:
		return c.endpoints.Quotes
	case config.ModuleEvents:
		return c.endpoints.Events
	default:
		return c.endpoints.API
	}
}

func sleepWithBackoff(ctx context.Context, attempt int) {
	delay := time.Duration(1<<uint(attempt)) * time.Second
	select {
	case <-ctx.Done():
	case <-time.After(delay):
	}
}

func getClientSource() string {
	if v := os.Getenv(clientSourceEnv); v != "" {
		return v
	}
	return defaultClientSource
}
