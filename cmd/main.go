// --- BACKLOG ---

// TODO: Clean up codebase: interfacing?

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
	case "-h", "--help":
		printUsageAndExit()
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
	fmt.Println(`
Usage: vm2container <command> [flags]

Commands:
  profile     Analyze Unix processes to collect runtime application dependencies.
              Accepts comma-separated process IDs (PIDs). The last PID is treated
              as the main application process.

  dockerize   Generate container artifacts for the profiled application.
              Requires the main application PID of the profiled processes.

Flags:
  -trace-wait <seconds>    (profile only) Duration to wait while capturing
                           runtime data. Default: 5 seconds.

  -h, --help               Display this help message.

Examples:
  vm2container profile -trace-wait 10 1234,5678
  vm2container dockerize 5678

For detailed documentation, see the README.
    `)
	os.Exit(0)
}
