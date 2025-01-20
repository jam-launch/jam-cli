package main

import (
	"bufio"
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

const (
	deviceCodeEndpoint = "https://api.jamlaunch.com/device-auth/request"
	userAuthEndpoint   = "https://app.jamlaunch.com/device-auth"
	client_id          = "jamlaunch-addon"
)

type DeviceCodeRequest struct {
	ClientId string `json:"clientId"`
	Scope    string `json:"scope"`
}

type DeviceCodeResponse struct {
	DeviceCode string `json:"deviceCode"`
	UserCode   string `json:"userCode"`
}

type CheckAuthResponse struct {
	AccessToken string `json:"accessKey,omitempty"`
	AccessState string `json:"state"`
}

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

func requestUserCode() (*DeviceCodeResponse, error) {
	payload := DeviceCodeRequest{
		ClientId: client_id,
		Scope:    "developer",
	}
	body, _ := json.Marshal(payload)

	resp, err := http.Post(deviceCodeEndpoint, "application/json", bytes.NewBuffer(body))
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("received non-OK HTTP status: %s", resp.Status)
	}

	var deviceCodeResp DeviceCodeResponse
	err = json.NewDecoder(resp.Body).Decode(&deviceCodeResp)
	if err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &deviceCodeResp, nil
}

func checkAuth(deviceCodeResp *DeviceCodeResponse) (*CheckAuthResponse, error) {
	checkURL := fmt.Sprintf("%s/%s/%s", deviceCodeEndpoint, deviceCodeResp.UserCode, deviceCodeResp.DeviceCode)

	for {
		// Send GET request
		resp, err := http.Get(checkURL)
		if err != nil {
			return nil, fmt.Errorf("failed to make request: %w", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("received non-OK HTTP status: %s", resp.Status)
		}

		// Decode JSON response
		var authResponse CheckAuthResponse
		if err := json.NewDecoder(resp.Body).Decode(&authResponse); err != nil {
			return nil, fmt.Errorf("failed to decode response: %w", err)
		}

		// Check the state
		if authResponse.AccessState == "allowed" {
			fmt.Println("Login Successful!")
			return &authResponse, nil
		}

		// Delay before the next request
		time.Sleep(time.Second)
	}
}

func saveToken(token string) error {
	data := map[string]string{"authToken": token}

	file, err := os.Create("userConfig.json")
	if err != nil {
		return fmt.Errorf("Failed to create file: %w", err)
	}

	defer file.Close()

	encoder := json.NewEncoder(file)
	if err := encoder.Encode(data); err != nil {
		return fmt.Errorf("Failed to write token to file: %w", err)
	}

	return nil
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

	authToken := data["authToken"]

	if authToken == "" {
		return false, ""
	}

	result := parseToken(authToken)
	if result.Errored {
		fmt.Println("Error:", result.Error)
		return false, ""
	}

	expiration, ok := result.Data.Claims["exp"].(float64)
	if !ok {
		fmt.Println("Error: 'exp' claim is missing or not a float64")
		return false, ""
	}

	expTime := time.Unix(int64(expiration), 0)

	if time.Now().After(expTime) {
		fmt.Println("Error: Token Expired!")
		return false, ""
	}

	verifyResult := verifyToken(authToken)

	if verifyResult == false {
		fmt.Println("Error: Token Invalid!")
		return false, ""
	}

	return true, authToken
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
	req, err := http.NewRequest("GET", "https://api.jamlaunch.com/projects", nil)
	if err != nil {
		log.Fatalf("Error creating request: %v", err)
		return false
	}

	req.Header.Add("Authorization", "Bearer "+authToken)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Fatalf("Error making request: %v", err)
		return false
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("Error reading response: %v", err)
		return false
	}

	var data map[string]interface{}
	if err := json.Unmarshal(body, &data); err != nil {
		log.Fatalf("Error unmarshaling JSON: %v", err)
	}

	if _, exists := data["projects"]; exists {
		return true
	} else {
		return false
	}
}

func login() {
	fmt.Println("User has requested a new login token...")

	deviceCodeResp, err := requestUserCode()
	if err != nil {
		fmt.Printf("Error requesting user code: %v\n", err)
		return
	}

	fmt.Printf("Visit: %s?user_code=%s\n", userAuthEndpoint, deviceCodeResp.UserCode)
	fmt.Printf("Enter the code: %s\n", deviceCodeResp.UserCode)

	authResponse, err := checkAuth(deviceCodeResp)
	if err != nil {
		fmt.Printf("Error polling for token: %v\n", err)
		return
	}

	if err := saveToken(authResponse.AccessToken); err != nil {
		fmt.Printf("Error saving token: %v\n", err)
		return
	}
}

func main() {
	// Step 1: Request Device Code
	fmt.Println("Welcome to the JamLaunch CLI!")

	fmt.Println("Checking Token...")
	result, token := loadToken()

	if result == false {
		fmt.Println("Token not found or invalid! User must authenticate again.")

		deviceCodeResp, err := requestUserCode()
		if err != nil {
			fmt.Printf("Error requesting user code: %v\n", err)
			return
		}

		// Step 2: Display User Instructions
		fmt.Printf("Visit: %s?user_code=%s\n", userAuthEndpoint, deviceCodeResp.UserCode)
		fmt.Printf("Enter the code: %s\n", deviceCodeResp.UserCode)

		// Step 3: Poll for Access Token
		authResponse, err := checkAuth(deviceCodeResp)
		if err != nil {
			fmt.Printf("Error polling for token: %v\n", err)
			return
		}

		if err := saveToken(authResponse.AccessToken); err != nil {
			fmt.Printf("Error saving token: %v\n", err)
			return
		}
	} else {
		fmt.Printf("Valid token found: %s\n", token)
	}

	reader := bufio.NewReader(os.Stdin)

	fmt.Println("Type your message below. Type 'exit' to quit.")

	for {
		// Display a prompt
		fmt.Print("> ")

		// Read user input
		input, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println("Error reading input:", err)
			continue
		}

		// Trim whitespace
		input = strings.TrimSpace(input)

		if input == "login" {
			login()
		}

		// Exit condition
		if strings.ToLower(input) == "exit" {
			fmt.Println("Goodbye!")
			break
		}

		// Display the user's input
		fmt.Println("You said:", input)
	}
}
