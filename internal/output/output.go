// Copyright 2024-2026 Webull Inc. All rights reserved.
// Licensed under the Apache License, Version 2.0.
// See LICENSE file in the project root for full license information.

package output

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

// Format represents the output format.
type Format string

const (
	FormatJSON Format = "json"
	FormatCSV  Format = "csv"
)

// PrintJSON outputs data as formatted JSON to stdout.
func PrintJSON(data interface{}) error {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(data)
}

// PrintRawJSON outputs raw JSON bytes to stdout.
func PrintRawJSON(data json.RawMessage) {
	var buf bytes.Buffer
	if err := json.Indent(&buf, data, "", "  "); err == nil {
		fmt.Fprintln(os.Stdout, buf.String())
	} else {
		fmt.Fprintln(os.Stdout, string(data))
	}
}

// PrintError outputs an error to stderr in JSON format.
func PrintError(errorCode string, message string, status int) {
	errObj := map[string]interface{}{
		"error_code": errorCode,
		"message":    message,
	}
	if status > 0 {
		errObj["status"] = status
	}
	data, _ := json.Marshal(errObj)
	fmt.Fprintln(os.Stderr, string(data))
}

// PrintCSV outputs data as CSV. Expects a slice of maps or a single map.
func PrintCSV(data json.RawMessage) error {
	// Try to parse as array
	var rows []map[string]interface{}
	if err := json.Unmarshal(data, &rows); err != nil {
		// Try single object
		var single map[string]interface{}
		if err := json.Unmarshal(data, &single); err != nil {
			return fmt.Errorf("CSV output requires JSON object or array")
		}
		rows = []map[string]interface{}{single}
	}

	if len(rows) == 0 {
		return nil
	}

	// Extract headers from first row
	headers := make([]string, 0)
	for k := range rows[0] {
		headers = append(headers, k)
	}

	w := csv.NewWriter(os.Stdout)
	defer w.Flush()

	// Write header
	w.Write(headers)

	// Write rows
	for _, row := range rows {
		record := make([]string, len(headers))
		for i, h := range headers {
			v := row[h]
			record[i] = formatValue(v)
		}
		w.Write(record)
	}

	return nil
}

// Mask masks a sensitive string (show first 6 and last 6).
func Mask(s string) string {
	if len(s) <= 12 {
		return "******"
	}
	return s[:6] + "******" + s[len(s)-6:]
}

func formatValue(v interface{}) string {
	if v == nil {
		return ""
	}
	switch val := v.(type) {
	case string:
		return val
	case float64:
		return strings.TrimRight(strings.TrimRight(fmt.Sprintf("%f", val), "0"), ".")
	default:
		b, _ := json.Marshal(val)
		return string(b)
	}
}
