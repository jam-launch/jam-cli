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
		fmt.Println("\033[91mToken not found or invalid! User must authenticate again.")

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
		} else if strings.ToLower(input) == "projects" {
			projects(token)
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
