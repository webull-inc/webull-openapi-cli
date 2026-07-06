// Copyright 2024-2026 Webull Inc. All rights reserved.
// Licensed under the Apache License, Version 2.0.
// See LICENSE file in the project root for full license information.

package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/spf13/cobra"
)

// readJSONInput reads JSON input from --body flag, --file flag, or stdin.
// Priority: --body > --file > stdin (if pipe detected)
func readJSONInput(cmd *cobra.Command) (map[string]interface{}, error) {
	// Try --body
	body, _ := cmd.Flags().GetString("body")
	if body != "" {
		var result map[string]interface{}
		if err := json.Unmarshal([]byte(body), &result); err != nil {
			return nil, fmt.Errorf("invalid JSON in --body: %w", err)
		}
		return result, nil
	}

	// Try --file
	file, _ := cmd.Flags().GetString("file")
	if file != "" {
		data, err := os.ReadFile(file)
		if err != nil {
			return nil, fmt.Errorf("failed to read file %q: %w", file, err)
		}
		var result map[string]interface{}
		if err := json.Unmarshal(data, &result); err != nil {
			return nil, fmt.Errorf("invalid JSON in file %q: %w", file, err)
		}
		return result, nil
	}

	// Try stdin (only if piped, not interactive terminal)
	stat, _ := os.Stdin.Stat()
	if (stat.Mode() & os.ModeCharDevice) == 0 {
		data, err := io.ReadAll(os.Stdin)
		if err != nil {
			return nil, fmt.Errorf("failed to read stdin: %w", err)
		}
		if len(data) > 0 {
			var result map[string]interface{}
			if err := json.Unmarshal(data, &result); err != nil {
				return nil, fmt.Errorf("invalid JSON from stdin: %w", err)
			}
			return result, nil
		}
	}

	return nil, fmt.Errorf("no input provided. Use --body, --file, or pipe JSON via stdin")
}
