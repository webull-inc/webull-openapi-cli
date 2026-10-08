// Copyright 2024-2026 Webull Inc. All rights reserved.
// Licensed under the Apache License, Version 2.0.
// See LICENSE file in the project root for full license information.

package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/webull-inc/webull-openapi-cli/internal/client"
	"github.com/webull-inc/webull-openapi-cli/internal/output"
)

var communityCmd = &cobra.Command{
	Use:   "community",
	Short: "Community (posts, comments, hot stocks)",
}

var communityPostCmd = &cobra.Command{Use: "post", Short: "Community posts"}
var communityCommentCmd = &cobra.Command{Use: "comment", Short: "Community comments"}

func init() {
	communityCmd.AddCommand(communityPostCmd)
	communityCmd.AddCommand(communityCommentCmd)
	communityCmd.AddCommand(communityHotStocksCmd)
	communityCmd.AddCommand(communityMyPostsCmd)

	communityPostCmd.AddCommand(communityPostCreateCmd)
	communityPostCmd.AddCommand(communityPostGetCmd)
	communityPostCmd.AddCommand(communityPostsBySymbolCmd)

	communityCommentCmd.AddCommand(communityCommentCreateCmd)
}

// ==================== Create Post ====================

var communityPostCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a plain-text community post",
	RunE: func(cmd *cobra.Command, args []string) error {
		text, _ := cmd.Flags().GetString("text")
		symbol, _ := cmd.Flags().GetString("symbol")
		category, _ := cmd.Flags().GetString("category")

		if text == "" {
			output.PrintError("MISSING_PARAMETER", "--text must not be empty", 0)
			return errSilent
		}
		// symbol and category must be provided together.
		if (symbol == "") != (category == "") {
			output.PrintError("INVALID_ARGUMENT", "--symbol and --category must be provided together", 0)
			return errSilent
		}

		body := map[string]interface{}{"text": text}
		if symbol != "" {
			body["symbol"] = symbol
			body["category"] = category
		}
		return doUSDataPost("/market-data/communities/posts/create", body)
	},
}

func init() {
	communityPostCreateCmd.Flags().String("text", "", "Post content (required)")
	communityPostCreateCmd.Flags().String("symbol", "", "Ticker symbol (optional, requires --category)")
	communityPostCreateCmd.Flags().String("category", "", "Market category (optional, requires --symbol): US_STOCK, HK_STOCK")
	communityPostCreateCmd.MarkFlagRequired("text")
}

// ==================== Create Comment ====================

var communityCommentCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a plain-text comment on a community post",
	RunE: func(cmd *cobra.Command, args []string) error {
		postID, _ := cmd.Flags().GetString("post-id")
		text, _ := cmd.Flags().GetString("text")
		if postID == "" {
			output.PrintError("MISSING_PARAMETER", "--post-id must not be empty", 0)
			return errSilent
		}
		if text == "" {
			output.PrintError("MISSING_PARAMETER", "--text must not be empty", 0)
			return errSilent
		}
		body := map[string]interface{}{
			"post_id": postID,
			"text":    text,
		}
		return doUSDataPost("/market-data/communities/comments/create", body)
	},
}

func init() {
	communityCommentCreateCmd.Flags().String("post-id", "", "Parent post identifier (required)")
	communityCommentCreateCmd.Flags().String("text", "", "Comment content (required)")
	communityCommentCreateCmd.MarkFlagRequired("post-id")
	communityCommentCreateCmd.MarkFlagRequired("text")
}

// ==================== Hot Stocks ====================

var communityHotStocksCmd = &cobra.Command{
	Use:   "hot-stocks",
	Short: "List stocks ranked by community activity",
	RunE: func(cmd *cobra.Command, args []string) error {
		return doUSDataGet("/market-data/communities/stock-rankings/list", map[string]string{})
	},
}

// ==================== List Posts by Symbol ====================

var communityPostsBySymbolCmd = &cobra.Command{
	Use:   "by-symbol",
	Short: "List community posts associated with a symbol",
	RunE: func(cmd *cobra.Command, args []string) error {
		symbol, _ := cmd.Flags().GetString("symbol")
		category, _ := cmd.Flags().GetString("category")
		keyword, _ := cmd.Flags().GetString("keyword")
		paginationKey, _ := cmd.Flags().GetString("pagination-key")
		if symbol == "" {
			output.PrintError("MISSING_PARAMETER", "--symbol must not be empty", 0)
			return errSilent
		}
		q := map[string]string{"symbol": symbol, "category": category}
		if keyword != "" {
			q["keyword"] = keyword
		}
		if paginationKey != "" {
			q["pagination_key"] = paginationKey
		}
		return doUSDataGet("/market-data/communities/posts/symbols/list", q)
	},
}

func init() {
	communityPostsBySymbolCmd.Flags().String("symbol", "", "Stock symbol (required)")
	communityPostsBySymbolCmd.Flags().String("category", "US_STOCK", "Market category: US_STOCK, HK_STOCK")
	communityPostsBySymbolCmd.Flags().String("keyword", "", "Optional search keyword")
	communityPostsBySymbolCmd.Flags().String("pagination-key", "", "Cursor from the previous response")
	communityPostsBySymbolCmd.MarkFlagRequired("symbol")
}

// ==================== Get Post Detail ====================

var communityPostGetCmd = &cobra.Command{
	Use:   "get",
	Short: "Get detailed information for a community post",
	RunE: func(cmd *cobra.Command, args []string) error {
		postID, _ := cmd.Flags().GetString("post-id")
		if postID == "" {
			output.PrintError("MISSING_PARAMETER", "--post-id must not be empty", 0)
			return errSilent
		}
		return doUSDataGet("/market-data/communities/posts/get", map[string]string{"post_id": postID})
	},
}

func init() {
	communityPostGetCmd.Flags().String("post-id", "", "Post identifier (required)")
	communityPostGetCmd.MarkFlagRequired("post-id")
}

// ==================== List My Posts ====================

var communityMyPostsCmd = &cobra.Command{
	Use:   "my-posts",
	Short: "List posts created by the current user",
	RunE: func(cmd *cobra.Command, args []string) error {
		paginationKey, _ := cmd.Flags().GetString("pagination-key")
		q := map[string]string{}
		if paginationKey != "" {
			q["pagination_key"] = paginationKey
		}
		return doUSDataGet("/market-data/communities/my-posts/list", q)
	},
}

func init() {
	communityMyPostsCmd.Flags().String("pagination-key", "", "Cursor from the previous response")
}

// ==================== Helper ====================

// doUSDataPost issues a US-only POST with a JSON body and prints the response.
func doUSDataPost(path string, body map[string]interface{}) error {
	cl, err := newUSDataClient()
	if err != nil {
		output.PrintError("CONFIG_ERROR", err.Error(), 0)
		return err
	}
	resp, err := cl.Do(context.Background(), "POST", path, &client.RequestOptions{
		Body: body,
	})
	if err != nil {
		output.PrintError("REQUEST_ERROR", err.Error(), 0)
		return err
	}
	if resp.StatusCode != 200 {
		fmt.Fprintln(os.Stderr, string(resp.Body))
		return errSilent
	}
	if flagCSV {
		return output.PrintCSV(resp.Body)
	}
	output.PrintRawJSON(resp.Body)
	return nil
}
