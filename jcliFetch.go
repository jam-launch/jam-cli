package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
)

func fetch(apiUrl string, authToken string) (map[string]interface{}, bool) {
	req, err := http.NewRequest("GET", apiUrl, nil)
	if err != nil {
		log.Fatalf("\033[91mError creating request: %v\033[0m", err)
		return map[string]interface{}{}, false
	}

	req.Header.Add("Authorization", "Bearer "+authToken)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Fatalf("\033[91mError making request: %v\033[0m", err)
		return map[string]interface{}{}, false
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("\033[91mError reading response: %v\033[0m", err)
		return map[string]interface{}{}, false
	}

	var data map[string]interface{}
	if err := json.Unmarshal(body, &data); err != nil {
		log.Fatalf("\033[91mError unmarshaling JSON: %v\033[0m", err)
		return map[string]interface{}{}, false
	}

	return data, true
}

func apiPost(apiUrl string, authToken string, body map[string]interface{}) (map[string]interface{}, error) {
	jsonData, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal JSON: %w", err)
	}

	req, err := http.NewRequest("POST", apiUrl, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+authToken)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("\033[91mError reading response: %v\033[0m", err)
		return map[string]interface{}{}, fmt.Errorf("\033[91mError reading response: %v\033[0m", err)
	}

	var data map[string]interface{}
	if err := json.Unmarshal(respBody, &data); err != nil {
		return map[string]interface{}{}, fmt.Errorf("\033[91mError unmarshaling JSON: %v\033[0m", err)
	}

	return data, nil
}
