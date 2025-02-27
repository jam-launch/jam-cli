package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

func main() {
	// Step 1: Request Device Code
	fmt.Println("Welcome to the JamLaunch CLI!")

	fmt.Print("Checking token...")
	result, token := loadToken()

	if !result {
		fmt.Println("\033[91mToken not found or invalid! User must authenticate again.\033[0m")

		deviceCodeResp, err := requestUserCode()
		if err != nil {
			fmt.Printf("\033[91mError requesting user code: %v\033[0m\n", err)
			return
		}

		// Step 2: Display User Instructions
		fmt.Printf("\033[93mVisit:\033[0m %s?user_code=%s\n", userAuthEndpoint, deviceCodeResp.UserCode)
		fmt.Printf("\033[93mEnter the code:\033[0m %s\n", deviceCodeResp.UserCode)

		// Step 3: Poll for Access Token
		authResponse, err := checkAuth(deviceCodeResp)
		if err != nil {
			fmt.Printf("\033[91mError polling for token: %v\033[0m\n", err)
			return
		}

		if err := saveToken(authResponse.AccessToken); err != nil {
			fmt.Printf("\033[91mError saving token: %v\033[0m\n", err)
			return
		}
	} else {
		fmt.Printf("\033[92mLogin successful!\033[0m\n")
	}

	reader := bufio.NewReader(os.Stdin)

	fmt.Println("Type your message below. Type 'exit' to quit.")

	for {
		// Display a prompt
		fmt.Print("> ")

		// Read user input
		input, err := reader.ReadString('\n')
		if err != nil {
			fmt.Printf("\033[91mError reading input: %s\033[0m\n", err)
			continue
		}

		// Trim whitespace
		input = strings.TrimSpace(input)

		if strings.ToLower(input) == "login" {
			login()
		} else if len(input) >= 8 && strings.ToLower(input[:8]) == "projects" {
			parts := strings.Fields(input)
			if len(parts) == 1 && strings.ToLower(parts[0]) == "projects" {
				projects(token)
			} else if len(parts) == 2 {
				projects_id(token, parts[1])
			} else if len(parts) == 3 {
				projects_sessions(token, parts[1])
			} else {
				projects_sessions_with_id(token, parts[1], parts[3])
			}
		} else if len(input) >= 4 && strings.ToLower(input[:4]) == "help" {
			help(input)
		} else if strings.ToLower(input) == "exit" {
			fmt.Println("Goodbye!")
			break
		} else {
			fmt.Printf("\033[31m%s is not a valid command!\033[0m\n", input)
		}
	}
}

// https://admin-api.jamlaunch.com/account/transactions
