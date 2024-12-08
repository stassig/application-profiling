package main

import (
	"fmt"
	"os"

	"application_profiling/cmd/subcommands"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: application_profiling <subcommand> [flags]")
		fmt.Println("Available subcommands: dockerize, profile")
		os.Exit(1)
	}

	subcommand := os.Args[1]
	args := os.Args[2:] // the remaining arguments after the subcommand

	switch subcommand {
	case "dockerize":
		subcommands.RunDockerize(args)
	case "profile":
		subcommands.RunProfile(args)
	default:
		fmt.Printf("Unknown subcommand: %s\n", subcommand)
		os.Exit(1)
	}
}
