package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/jedib0t/go-pretty/table"
)

var (
	colProjectIndex = "#"
	colProjectName  = "Project Name"
	rowHeader       = table.Row{colProjectIndex, colProjectName}
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
			t := table.NewWriter()
			tTemp := table.Table{}
			tTemp.Render()
			t.AppendHeader(rowHeader)
			t.SetTitle("Current Projects")
			t.SetStyle(table.StyleColoredDark)

			for _, project := range projects {
				if projMap, ok := project.(map[string]interface{}); ok {
					t.AppendRow(table.Row{projMap["id"], projMap["project_name"]})
				}
			}

			fmt.Println(t.Render())
		}
	} else {
		log.Printf("\033[91mError: projects is not an array!\033[0m\n")
		log.Printf("\033[91mPlease visit https://app.jamlaunch.com/projects and try again!\033[0m\n")
		return false
	}

	return true
}

func help(input string) {
	parts := strings.Fields(input)

	if len(parts) == 1 && strings.ToLower(parts[0]) == "help" {
		fmt.Println("For more information on a specific command, type HELP command-name")
		fmt.Println("HELP        Provides Help information for Jam Launch CLI commands.")
		fmt.Println("LOGIN       Prompts the user to log in again.")
		fmt.Println("PROJECTS    Displays a list of the users current projects.")
	} else if len(parts) == 2 && strings.ToLower(parts[0]) == "help" && strings.ToLower(parts[1]) == "projects" {
		fmt.Println("PROJECTS command details:")
		fmt.Println("Displays a list of the user's current projects.")
		fmt.Println("")
		fmt.Println("PROJECTS (No Parameters)")
		fmt.Println("")
		fmt.Println("This command will display the id and name of each project in a table format.")
	} else if len(parts) == 2 && strings.ToLower(parts[0]) == "help" && strings.ToLower(parts[1]) == "help" {
		fmt.Println("HELP command details:")
		fmt.Println("Provides Help information for Jam Launch CLI commands.")
		fmt.Println("")
		fmt.Println("HELP (No Parameters)")
		fmt.Println("HELP [Command Name]")
		fmt.Println("")
		fmt.Println("Running help with parameters will display detailed help information for the command specified by the parameter.")
	} else if len(parts) == 2 && strings.ToLower(parts[0]) == "help" && strings.ToLower(parts[1]) == "login" {
		fmt.Println("LOGIN command details:")
		fmt.Println("Prompts the user to log in again.")
		fmt.Println("")
		fmt.Println("LOGIN (No Parameters)")
		fmt.Println("")
		fmt.Println("Running this command will prompt the user to generate a new authentication token and replace the old one regardles if it is valid or not.")
	} else {
		fmt.Printf("\033[31mCommand formatted incorrectly. Use 'HELP' or 'HELP command-name'.\033[0m\n")
	}
}
