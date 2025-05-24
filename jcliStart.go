package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/jessevdk/go-flags"
)

func printError(errStr error) {
	if errStr != nil {
		fmt.Printf("\033[91m%s\033[0m\n", errStr)
	}
}

var token string

type LoginCommand struct{}

func (c *LoginCommand) Execute(args []string) error {
	return login()
}

type ProjectsCommand struct{}

func (c *ProjectsCommand) Execute(args []string) error {
	switch len(args) {
	case 0:
		return projects(token)
	case 1:
		return projectsName(token, args[0])
	case 2:
		if strings.ToLower(args[1]) == "sessions" {
			return projectSessions(token, args[0])
		}
		return fmt.Errorf("usage: projects <project-name> sessions")
	case 3:
		return projectSessionId(token, args[0], args[2])
	default:
		return fmt.Errorf("invalid usage of 'projects'")
	}
}

type GetCommand struct{}

func (c *GetCommand) Execute(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("missing endpoint for get command")
	}
	return apiGet(args[0], token)
}

type GameGetCommand struct{}

func (c *GameGetCommand) Execute(args []string) error {
	if len(args) < 2 {
		return fmt.Errorf("usage: game-get <user-id> <endpoint>")
	}
	gameToken, err := getGameUserToken(args[0], token)
	if err != nil {
		return fmt.Errorf("failed to get game token: %v", err)
	}
	return apiGet(args[1], gameToken)
}

type HelpCommand struct{}

func (c *HelpCommand) Execute(args []string) error {
	input := "help"
	if len(args) > 0 {
		input += " " + strings.Join(args, " ")
	}
	help(input)
	return nil
}

type Command struct {
	Login    LoginCommand    `command:"login" description:"Log in to the system"`
	Projects ProjectsCommand `command:"projects" description:"Manage projects"`
	Get      GetCommand      `command:"get" description:"Call an API GET endpoint"`
	GameGet  GameGetCommand  `command:"game-get" description:"Call an API GET endpoint as a game user"`
	Help     HelpCommand     `command:"help" description:"Show help"`
}

func main() {
	// Step 1: Request Device Code
	fmt.Println("Welcome to the JamLaunch CLI!")

	fmt.Print("Checking token...")
	result, tok := loadToken()

	if !result {
		fmt.Println("\033[91mToken not found or invalid! User must authenticate again.\033[0m")

		var err error
		tok, err = getDevToken()
		if err != nil {
			fmt.Printf("\033[31mFailed to get tokens: %v\n - exiting...\033[0m\n", err)
		}
	} else {
		fmt.Printf("\033[92mLogin successful!\033[0m\n")
	}

	token = tok

	reader := bufio.NewReader(os.Stdin)

	fmt.Println("Type your message below. Type 'exit' to quit.")

	var cmd Command
	parser := flags.NewParser(&cmd, flags.None)

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

		if strings.ToLower(input) == "exit" {
			fmt.Println("Goodbye!")
			break
		}

		args := strings.Fields(input)
		_, err = parser.ParseArgs(args)
		printError(err)
	}
}

// https://admin-api.jamlaunch.com/account/transactions
