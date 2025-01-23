package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
)

func login() {
	fmt.Println("Requesting new token...")

	deviceCodeResp, err := requestUserCode()
	if err != nil {
		fmt.Printf("\033[91mError requesting user code: %v\033[0m\n", err)
		return
	}

	fmt.Printf("\033[93mVisit:\033[0m %s?user_code=%s\n", userAuthEndpoint, deviceCodeResp.UserCode)
	fmt.Printf("\033[93mEnter the code:\033[0m %s\n", deviceCodeResp.UserCode)

	authResponse, err := checkAuth(deviceCodeResp)
	if err != nil {
		fmt.Printf("\033[91mError polling for token: %v\033[0m\n", err)
		return
	}

	if err := saveToken(authResponse.AccessToken); err != nil {
		fmt.Printf("\033[91mError saving token: %v\033[0m\n", err)
		return
	}
}

func projects(authToken string) bool {
	req, err := http.NewRequest("GET", "https://api.jamlaunch.com/projects", nil)
	if err != nil {
		log.Fatalf("\033[91mError creating request: %v\033[0m", err)
		return false
	}

	req.Header.Add("Authorization", "Bearer "+authToken)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Fatalf("\033[91mError making request: %v\033[0m", err)
		return false
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("\033[91mError reading response: %v\033[0m", err)
		return false
	}

	var data map[string]interface{}
	if err := json.Unmarshal(body, &data); err != nil {
		log.Fatalf("\033[91mError unmarshaling JSON: %v\033[0m", err)
		return false
	}

	if projects, ok := data["projects"].([]interface{}); ok {
		if len(projects) == 0 {
			fmt.Println("You currently do not have any projects!")
		} else {
			fmt.Println("List of current projects:")

			for _, project := range projects {
				if projMap, ok := project.(map[string]interface{}); ok {
					fmt.Printf("ID: %v, Name: %v\n",
						projMap["id"], projMap["project_name"])
				}
			}
		}
	} else {
		log.Printf("\033[91mError: projects is not an array!\033[0m\n")
		return false
	}

	return true
}
