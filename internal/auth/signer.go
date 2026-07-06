// Copyright 2024-2026 Webull Inc. All rights reserved.
// Licensed under the Apache License, Version 2.0.
// See LICENSE file in the project root for full license information.

package auth

import (
	"crypto/hmac"
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"net/url"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"
)

// SignAlgorithm represents the HMAC signing algorithm.
type SignAlgorithm string

const (
	AlgoHMACSHA1   SignAlgorithm = "HMAC-SHA1"
	AlgoHMACSHA256 SignAlgorithm = "HMAC-SHA256"
)

const signatureVersion = "1.0"

// Header key constants.
const (
	headerAppKey    = "x-app-key"
	headerTimestamp = "x-timestamp"
	headerSigVer    = "x-signature-version"
	headerSigAlgo   = "x-signature-algorithm"
	headerSigNonce  = "x-signature-nonce"
	headerSig       = "x-signature"
	headerHost      = "Host"
)

// SignRequest contains all parameters needed to sign a request.
type SignRequest struct {
	AppKey    string
	AppSecret string
	Host      string
	Path      string
	Query     map[string]string
	Body      []byte
	Algorithm SignAlgorithm
}

// SignResult contains the computed signature and headers to attach.
type SignResult struct {
	Signature string
	Headers   map[string]string
}

// Sign computes the HMAC signature for a Webull API request.
func Sign(req SignRequest) SignResult {
	timestamp := time.Now().UTC().Format("2006-01-02T15:04:05Z")
	nonce := uuid.New().String()

	algo := req.Algorithm
	if algo == "" {
		algo = AlgoHMACSHA1
	}

	// Step 1: Build sign headers (will also be used as output headers)
	signHeaders := map[string]string{
		headerAppKey:    req.AppKey,
		headerTimestamp: timestamp,
		headerSigVer:    signatureVersion,
		headerSigAlgo:   string(algo),
		headerSigNonce:  nonce,
		headerHost:      req.Host,
	}

	// Step 2: Merge with query params for signing
	signParams := make(map[string]string)
	// Use lowercase "host" for signing (per protocol spec)
	for k, v := range signHeaders {
		key := k
		if k == headerHost {
			key = "host"
		}
		signParams[key] = v
	}
	for k, v := range req.Query {
		if existing, ok := signParams[k]; ok {
			signParams[k] = existing + "&" + v
		} else {
			signParams[k] = v
		}
	}

	// Step 3: Sort by key
	keys := make([]string, 0, len(signParams))
	for k := range signParams {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	// Step 4: Build sign string
	pairs := make([]string, 0, len(keys))
	for _, k := range keys {
		pairs = append(pairs, k+"="+signParams[k])
	}
	sortedPairs := strings.Join(pairs, "&")

	var signStr string
	if req.Path != "" {
		signStr = req.Path + "&" + sortedPairs
	} else {
		signStr = sortedPairs
	}

	// Step 5: Body hash
	if len(req.Body) > 0 {
		var bodyHash string
		if algo == AlgoHMACSHA256 {
			h := sha256.Sum256(req.Body)
			bodyHash = fmt.Sprintf("%X", h)
		} else {
			h := md5.Sum(req.Body)
			bodyHash = fmt.Sprintf("%X", h)
		}
		signStr = signStr + "&" + bodyHash
	}

	// Step 6: URL encode
	encoded := url.QueryEscape(signStr)
	encoded = strings.ReplaceAll(encoded, "+", "%20")

	// Step 7: HMAC
	signingKey := req.AppSecret + "&"
	var signature string
	if algo == AlgoHMACSHA256 {
		mac := hmac.New(sha256.New, []byte(signingKey))
		mac.Write([]byte(encoded))
		signature = base64.StdEncoding.EncodeToString(mac.Sum(nil))
	} else {
		mac := hmac.New(sha1.New, []byte(signingKey))
		mac.Write([]byte(encoded))
		signature = base64.StdEncoding.EncodeToString(mac.Sum(nil))
	}

	// Add signature to the same headers map
	signHeaders[headerSig] = signature

	return SignResult{
		Signature: signature,
		Headers:   signHeaders,
	}
}
