package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

func printError(errStr error) {
	if errStr != nil {
		fmt.Printf("\033[91m%s\033[0m\n", errStr)
	}
}

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
			err = login()

			printError(err)
		} else if len(input) >= 8 && strings.ToLower(input[:8]) == "projects" {
			parts := strings.Fields(input)
			if len(parts) == 1 && strings.ToLower(parts[0]) == "projects" {
				err = projects(token)

				printError(err)
			} else if len(parts) == 2 {
				err = projectsName(token, parts[1])

				printError(err)
			} else if len(parts) == 3 {
				err = projectSessions(token, parts[1])

				printError(err)
			} else {
				err = projectSessionId(token, parts[1], parts[3])

				printError(err)
			}
		} else if len(input) >= 3 && strings.ToLower(input[:3]) == "get" {
			parts := strings.Fields(input)

			err = apiGet(parts[1], token)

			printError(err)
		} else if len(input) >= 8 && strings.ToLower(input[:8]) == "game-get" {
			parts := strings.Fields(input)
			gameToken, err := getGameUserToken(parts[1], token)
			if err != nil {
				fmt.Printf("\033[31mFailed: %v\033[0m\n", err)
				continue
			}
			err = apiGet(parts[2], gameToken)

			printError(err)
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
