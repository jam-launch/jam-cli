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

		var err error
		token, err = getDevToken()
		if err != nil {
			fmt.Printf("\033[31mFailed to get tokens: %v\n - exiting...\033[0m\n", err)
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
		} else if len(input) >= 3 && strings.ToLower(input[:3]) == "get" {
			parts := strings.Fields(input)
			apiGet(parts[1], token)
		} else if len(input) >= 8 && strings.ToLower(input[:8]) == "game-get" {
			parts := strings.Fields(input)
			gameToken, err := getGameUserToken(parts[1], token)
			if err != nil {
				fmt.Printf("\033[31mFailed: %v\033[0m\n", err)
				continue
			}
			apiGet(parts[2], gameToken)
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
