package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"encoding/json"
	"net/http"
	"time"
	"bytes"
)

const (
	deviceCodeEndpoint = "https://api.jamlaunch.com/device-auth/request"
	userAuthEndpoint = "https://app.jamlaunch.com/device-auth"
	client_id = "jamlaunch-addon"
)

type DeviceCodeRequest struct {
	ClientId    string `json:"clientId"`
	Scope       string `json:"scope"`
}

type DeviceCodeResponse struct {
	DeviceCode  string `json:"deviceCode"`
	UserCode    string `json:"userCode"`
}

type CheckAuthResponse struct {
	AccessToken string `json:"accessKey,omitempty"`
	AccessState string `json:"state"`
}

func requestUserCode() (*DeviceCodeResponse, error) {
    payload := DeviceCodeRequest {
		ClientId : client_id,
		Scope : "developer",
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
			fmt.Println("Access granted!")
			return &authResponse, nil
		}

		// Delay before the next request
		time.Sleep(time.Second)
    }
}

func main() {
	// Step 1: Request Device Code
	fmt.Println("Welcome to the JamLaunch CLI!")
	fmt.Println("Requesting device code...")
	deviceCodeResp, err := requestUserCode()
	if err != nil {
		fmt.Printf("Error requesting user code: %v\n", err)
		return
	}

	// Step 2: Display User Instructions
	fmt.Printf("Visit: %s?user_code=%s\n", userAuthEndpoint, deviceCodeResp.UserCode)
	fmt.Printf("Enter the code: %s\n", deviceCodeResp.UserCode)

	// Step 3: Poll for Access Token
	fmt.Println("Waiting for user authentication...")
	authResponse, err := checkAuth(deviceCodeResp)
	if err != nil {
		fmt.Printf("Error polling for token: %v\n", err)
		return
	}

	fmt.Printf("Access Token: %s\n", authResponse.AccessToken)

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

		// Exit condition
		if strings.ToLower(input) == "exit" {
			fmt.Println("Goodbye!")
			break
		}

		// Display the user's input
		fmt.Println("You said:", input)
	}
}