package main

import (
	"fmt"
	"log"
	"strings"

	"github.com/jedib0t/go-pretty/table"
)

var (
	colProjectIndex = "#"
	colProjectName  = "Project Name"
	projectHeader   = table.Row{colProjectIndex, colProjectName}
)

var (
	colUsername   = "Username"
	colLevel      = "Level"
	membersHeader = table.Row{colUsername, colLevel}
)

var (
	colId             = "id"
	colCreatedAt      = "Created At"
	colDefaultRelease = "Default Release"
	colPublic         = "Public"
	colNetworkMode    = "Network Mode"
	colServerBuild    = "Server Build"
	colAllowGuests    = "Allow Guests"
	releasesHeader    = table.Row{colId, colCreatedAt, colDefaultRelease, colPublic, colNetworkMode, colServerBuild, colAllowGuests}
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
	var apiUrl = "https://api.jamlaunch.com/projects"

	data, success := fetch(apiUrl, authToken)

	if success {
		if projects, ok := data["projects"].([]interface{}); ok {
			if len(projects) == 0 {
				fmt.Println("You currently do not have any projects!")
			} else {
				t := table.NewWriter()
				tTemp := table.Table{}
				tTemp.Render()
				t.AppendHeader(projectHeader)
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
	}

	return true
}

func projects_id(authToken string, name string) bool {
	var apiUrlName = "https://api.jamlaunch.com/projects"

	nameData, successName := fetch(apiUrlName, authToken)

	if successName && nameData != nil {
		var projectId string

		if projects, ok := nameData["projects"].([]interface{}); ok {
			for _, p := range projects {
				project := p.(map[string]interface{})

				if project["project_name"].(string) == name {
					projectId = project["id"].(string)
					break
				}
			}
		}

		var apiUrlId = "https://api.jamlaunch.com/projects/" + projectId

		data, successId := fetch(apiUrlId, authToken)

		if successId {
			if data != nil && data["project_name"] != nil {
				fmt.Printf("\033[93mProject Name:\033[0m %s\n", data["project_name"].(string))
				fmt.Printf("\033[93mCreated At:\033[0m %s\n", data["created_at"].(string)[:10])
				fmt.Printf("\033[93mProject Id:\033[0m %s\n", data["id"].(string))
				fmt.Printf("\033[93mActive:\033[0m %t\n", data["active"].(bool))
				fmt.Println("")

				if members, ok := data["members"].([]interface{}); ok && len(members) > 0 {
					t := table.NewWriter()
					tTemp := table.Table{}
					tTemp.Render()
					t.AppendHeader(membersHeader)
					t.SetTitle("Current Members")
					t.SetStyle(table.StyleColoredDark)

					for _, member := range members {
						if memMap, ok := member.(map[string]interface{}); ok {
							t.AppendRow(table.Row{memMap["username"], memMap["level"]})
						}
					}

					fmt.Println(t.Render())
				}

				if releases, ok := data["releases"].([]interface{}); ok && len(releases) > 0 {
					fmt.Println("")

					t := table.NewWriter()
					tTemp := table.Table{}
					tTemp.Render()
					t.AppendHeader(releasesHeader)
					t.SetTitle("Current Releases")
					t.SetStyle(table.StyleColoredDark)

					for _, release := range releases {
						if relMap, ok := release.(map[string]interface{}); ok {
							t.AppendRow(table.Row{
								relMap["id"],
								relMap["created_at"],
								relMap["is_default"],
								relMap["public"],
								relMap["network_mode"],
								relMap["server_build"],
								relMap["allow_guests"],
							})
						}
					}

					fmt.Println(t.Render())
				}
			} else {
				fmt.Printf("\033[91mError: Project not found!\033[0m\n")
			}
		} else {
			fmt.Printf("\033[91mError: Unable to retrieve project data!\033[0m\n")
		}
	} else {
		fmt.Printf("\033[91mError: Unable to retrieve projects!\033[0m\n")
	}

	return true
}

func projects_sessions(authToken string, name string) bool {
	var apiUrlName = "https://api.jamlaunch.com/projects"

	nameData, successName := fetch(apiUrlName, authToken)

	if successName && nameData != nil {
		var projectId string

		if projects, ok := nameData["projects"].([]interface{}); ok {
			for _, p := range projects {
				project := p.(map[string]interface{})

				if project["project_name"].(string) == name {
					projectId = project["id"].(string)
					break
				}
			}
		}

		var apiUrlSessions = "https://api.jamlaunch.com/projects/" + projectId + "/sessions"

		fmt.Printf("Command: %s\n", apiUrlSessions)
	} else {
		fmt.Printf("\033[91mError: Unable to retrieve projects!\033[0m\n")
	}

	return true
}

func projects_sessions_with_id(authToken string, name string, sessionId string) bool {
	var apiUrlName = "https://api.jamlaunch.com/projects"

	nameData, successName := fetch(apiUrlName, authToken)

	if successName && nameData != nil {
		var projectId string

		if projects, ok := nameData["projects"].([]interface{}); ok {
			for _, p := range projects {
				project := p.(map[string]interface{})

				if project["project_name"].(string) == name {
					projectId = project["id"].(string)
					break
				}
			}
		}

		var apiUrlSessionsWithId = "https://api.jamlaunch.com/projects/" + projectId + "/sessions/" + sessionId

		fmt.Printf("Command: %s\n", apiUrlSessionsWithId)
	} else {
		fmt.Printf("\033[91mError: Unable to retrieve projects!\033[0m\n")
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
		fmt.Println("PROJECTS (Project Name)")
		fmt.Println("PROJECTS (Project Name) SESSIONS")
		fmt.Println("PROJECTS (Project Name) SESSIONS (Session ID)")
		fmt.Println("")
		fmt.Println("This command will display the id and name of each project in a table format.")
		fmt.Println("Running projects with parameters will display more specific details about a specific project.")
		fmt.Println("Running projects with parameters and the \"sessions\" keyword will display session information about the project")
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
