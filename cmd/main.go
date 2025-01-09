// TODO: - Instructions for test environment setup
//	- VM: Ubuntu 24.04 server (4 CPU cores, 8GB RAM, 50GB disk)
//  - Tools: Docker, MySQL, NGINX, strace

// TO DO: Bundle the application as a binary?
// TO DO: -h flag (man page) for help
// TO DO: Readme: instructions for running the application

// --- BACKLOG ---

// TO DO: Go Doc?
// TO DO: Clean up codebase: interfacing?

package main

import (
	"fmt"
	"os"

	"application_profiling/cmd/commands"
)

func main() {
	// At least one argument is required (profile, dockerize, etc.)
	if len(os.Args) < 2 {
		printUsageAndExit()
	}

	// Parse command and arguments
	command := os.Args[1]
	arguments := os.Args[2:]

	// Run the appropriate command
	switch command {
	case "dockerize":
		commands.RunDockerize(arguments)
	case "profile":
		commands.RunProfile(arguments)
	default:
		printUsageAndExit()
	}
}

// printUsageAndExit prints the usage message and exits with status code 1
func printUsageAndExit() {
	fmt.Println("Usage: application_profiling <command> [flags]")
	fmt.Println("Available commands: dockerize, profile")
	os.Exit(1)
}
