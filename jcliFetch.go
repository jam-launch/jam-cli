package main

import (
	"encoding/json"
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
