// Copyright 2024-2026 Webull Inc. All rights reserved.
// Licensed under the Apache License, Version 2.0.
// See LICENSE file in the project root for full license information.

package cmd

// silentError is returned when the error has already been printed to stderr.
// It prevents cobra from printing the error again while still setting a non-zero exit code.
type silentError struct{}

func (e silentError) Error() string { return "" }

// errSilent is the sentinel error used after API errors are already output.
var errSilent = silentError{}
