package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
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

func requestUserCode() (*DeviceCodeResponse, error) {
	payload := DeviceCodeRequest{
		ClientId: client_id,
		Scope:    "developer",
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

func saveToken(token string) error {
	data := map[string]string{"authToken": token}

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
