package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"
)

const (
	deviceCodeEndpoint = "https://api.jamlaunch.com/device-auth/request"
	userAuthEndpoint   = "https://app.jamlaunch.com/device-auth"
	DevClientId        = "jamlaunch-addon"
	UserClientId       = "jam-play"
	ApiBaseUrl         = "https://api.jamlaunch.com"
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

func requestUserCode(clientId string, scope string) (*DeviceCodeResponse, error) {
	payload := DeviceCodeRequest{
		ClientId: clientId,
		Scope:    scope,
	}
	body, _ := json.Marshal(payload)

	resp, err := http.Post(deviceCodeEndpoint, "application/json", bytes.NewBuffer(body))
	if err != nil {
		return nil, fmt.Errorf("\033[91mfailed to send request: %w\033[0m", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("\033[91mreceived non-OK HTTP status: %s\033[0m", resp.Status)
	}

	var deviceCodeResp DeviceCodeResponse
	err = json.NewDecoder(resp.Body).Decode(&deviceCodeResp)
	if err != nil {
		return nil, fmt.Errorf("\033[91mfailed to decode response: %w\033[0m", err)
	}

	return &deviceCodeResp, nil
}

func checkAuth(deviceCodeResp *DeviceCodeResponse) (*CheckAuthResponse, error) {
	checkURL := fmt.Sprintf("%s/%s/%s", deviceCodeEndpoint, deviceCodeResp.UserCode, deviceCodeResp.DeviceCode)

	for {
		// Send GET request
		resp, err := http.Get(checkURL)
		if err != nil {
			return nil, fmt.Errorf("\033[91mfailed to make request: %w\033[0m", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("\033[91mreceived non-OK HTTP status: %s\033[0m", resp.Status)
		}

		// Decode JSON response
		var authResponse CheckAuthResponse
		if err := json.NewDecoder(resp.Body).Decode(&authResponse); err != nil {
			return nil, fmt.Errorf("\033[91mfailed to decode response: %w\033[0m", err)
		}

		// Check the state
		if authResponse.AccessState == "allowed" {
			fmt.Printf("\033[92mLogin successful!\033[0m\n")
			return &authResponse, nil
		}

		// Delay before the next request
		time.Sleep(time.Second)
	}
}

func saveToken(authToken string) error {
	data := map[string]string{"authToken": authToken}

	file, err := os.Create("userConfig.json")
	if err != nil {
		return fmt.Errorf("\033[91mFailed to create file: %w\033[0m", err)
	}

	defer file.Close()

	encoder := json.NewEncoder(file)
	if err := encoder.Encode(data); err != nil {
		return fmt.Errorf("\033[91mFailed to write token to file: %w\033[0m", err)
	}

	return nil
}

func getDevToken() (string, error) {
	devResp, err := deviceAuthFlow(DevClientId, "developer")
	if err != nil {
		return "", fmt.Errorf("failed to get developer token: %v", err)
	}

	if err = saveToken(devResp.AccessToken); err != nil {
		return "", fmt.Errorf("error saving tokens: %v", err)
	}

	return devResp.AccessToken, nil
}

func deviceAuthFlow(clientId string, scope string) (*CheckAuthResponse, error) {
	deviceCodeResp, err := requestUserCode(clientId, scope)
	if err != nil {
		return nil, fmt.Errorf("error requesting user code: %v", err)
	}

	// Step 2: Display User Instructions
	fmt.Printf("\033[93mVisit:\033[0m %s?user_code=%s\n", userAuthEndpoint, deviceCodeResp.UserCode)
	fmt.Printf("\033[93mEnter the code:\033[0m %s\n", deviceCodeResp.UserCode)

	// Step 3: Poll for Access Token
	authResponse, err := checkAuth(deviceCodeResp)
	if err != nil {
		return nil, fmt.Errorf("error polling for token: %v", err)
	}

	return authResponse, nil
}

func getGameUserToken(gameId string, token string) (string, error) {
	parts := strings.SplitN(gameId, "-", 2)
	body := map[string]interface{}{
		"release":  parts[1],
		"test_num": 99,
	}

	res, err := apiPost(fmt.Sprintf("%s/projects/%s/testkey", ApiBaseUrl, parts[0]), token, body)
	if err != nil {
		return "", fmt.Errorf("error getting test token for game: %v", err)
	}

	token, ok := res["test_jwt"].(string)
	if !ok {
		return "", fmt.Errorf("game test token missing from API response")
	}

	return token, nil
}
