package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"
)

type TokenData struct {
	Header    map[string]interface{}
	Claims    map[string]interface{}
	Signature string
}

type TokenParseResult struct {
	Data    *TokenData
	Errored bool
	Error   string
}

func loadToken() (bool, string) {
	file, err := os.Open("userConfig.json")

	if err != nil {
		return false, ""
	}
	defer file.Close()

	var data map[string]string
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&data); err != nil {
		return false, ""
	}

	authToken, ok := data["authToken"]
	if !ok {
		fmt.Printf("\n\033[91mError: missing auth token\033[0m\n")
		return false, ""
	}

	if !checkToken(authToken) {
		return false, ""
	}

	return true, authToken
}

func checkToken(token string) bool {
	if token == "" {
		return false
	}

	result := parseToken(token)
	if result.Errored {
		fmt.Printf("\n\033[91mError: %s\033[0m\n", result.Error)
		return false
	}

	expiration, ok := result.Data.Claims["exp"].(float64)
	if !ok {
		fmt.Printf("\n\033[91mError: 'exp' claim is missing or not a float64\033[0m\n")
		return false
	}

	expTime := time.Unix(int64(expiration), 0)

	if time.Now().After(expTime) {
		fmt.Printf("\n\033[91mError: Token Expired!\033[0m\n")
		return false
	}

	verifyResult := verifyToken(token)

	if !verifyResult {
		fmt.Printf("\n\033[91mError: Token Invalid!\033[0m\n")
		return false
	}

	return true
}

func parseToken(token string) TokenParseResult {
	var result TokenParseResult
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		result.Errored = true
		result.Error = "Invalid JWT token format"
		return result
	}

	tkn := &TokenData{}

	// Parse Header
	headerJSON, err := decodeBase64URL(parts[0])
	if err != nil {
		result.Errored = true
		result.Error = "Failed to decode JWT header: " + err.Error()
		return result
	}
	if err := json.Unmarshal([]byte(headerJSON), &tkn.Header); err != nil {
		result.Errored = true
		result.Error = "Failed to parse JWT header: " + err.Error()
		return result
	}

	// Parse Claims
	claimsJSON, err := decodeBase64URL(parts[1])
	if err != nil {
		result.Errored = true
		result.Error = "Failed to decode JWT claims: " + err.Error()
		return result
	}
	if err := json.Unmarshal([]byte(claimsJSON), &tkn.Claims); err != nil {
		result.Errored = true
		result.Error = "Failed to parse JWT claims: " + err.Error()
		return result
	}

	// Parse Signature
	sig, err := decodeBase64URL(parts[2])
	if err != nil || len(sig) == 0 {
		result.Errored = true
		result.Error = "Failed to decode JWT signature: " + err.Error()
		return result
	}
	tkn.Signature = parts[2]

	result.Data = tkn
	return result
}

func decodeBase64URL(data string) (string, error) {
	decoded, err := base64.RawURLEncoding.DecodeString(data)
	if err != nil {
		return "", err
	}
	return string(decoded), nil
}

func verifyToken(authToken string) bool {
	var apiUrl = "https://api.jamlaunch.com/projects"

	data, success := fetch(apiUrl, authToken)

	if success == nil {
		if _, exists := data["projects"]; exists {
			return true
		} else {
			return false
		}
	} else {
		fmt.Printf("\033[91m%s\033[0m\n", success)
		return false
	}
}
