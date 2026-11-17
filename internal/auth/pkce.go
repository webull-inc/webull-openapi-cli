// Copyright 2024-2026 Webull Inc. All rights reserved.
// Licensed under the Apache License, Version 2.0.
// See LICENSE file in the project root for full license information.

package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
)

// PKCE holds a generated PKCE code_verifier / code_challenge pair (S256).
type PKCE struct {
	// Verifier is the high-entropy random string kept secret by the client
	// and sent when exchanging the device code for a token.
	Verifier string
	// Challenge is base64url(SHA256(Verifier)), sent when requesting the device code.
	Challenge string
	// Method is the challenge method, always "S256".
	Method string
}

// GeneratePKCE creates a new PKCE pair using the S256 method.
// The verifier is 32 random bytes base64url-encoded (43 chars, RFC 7636 compliant).
func GeneratePKCE() (*PKCE, error) {
	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		return nil, fmt.Errorf("failed to generate PKCE verifier: %w", err)
	}
	verifier := base64.RawURLEncoding.EncodeToString(buf)

	sum := sha256.Sum256([]byte(verifier))
	challenge := base64.RawURLEncoding.EncodeToString(sum[:])

	return &PKCE{
		Verifier:  verifier,
		Challenge: challenge,
		Method:    "S256",
	}, nil
}
