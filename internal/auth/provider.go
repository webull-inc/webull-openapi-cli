// Copyright 2024-2026 Webull Inc. All rights reserved.
// Licensed under the Apache License, Version 2.0.
// See LICENSE file in the project root for full license information.

package auth

import (
	"net/http"
)

// CredentialProvider is the interface for all authentication methods.
// Phase 1: AkSkProvider
// Phase 2: OAuthProvider (to be implemented)
type CredentialProvider interface {
	// Authenticate adds authentication headers to the request.
	Authenticate(req *http.Request, body []byte) error

	// HandleAuthError handles auth errors (e.g., INVALID_TOKEN).
	// Returns true if the error was handled and the request should be retried.
	HandleAuthError(resp *http.Response) (retry bool, err error)
}
