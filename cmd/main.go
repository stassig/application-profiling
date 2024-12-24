// TO DO: Update Dockerfile creation (separate user files copy)
// TO DO: Test with Nginx
// TO DO: Daimon off always (in cmd parser)
// TO DO: Kill only strace process (not application process)

// --- BACKLOG ---

// TO DO: Test production run
// TO DO: Better instructions (man page)
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
