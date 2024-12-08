package main

import (
	"application_profiling/internal/process"
	"flag"
	"log"
	"os"
	"path/filepath"
	"strconv"
)

func main() {
	// Define a flag to use a hardcoded executable for testing
	useExecutable := flag.Bool("use-executable", false, "Use a hardcoded executable to determine the PID")
	flag.Parse()

	var processID int
	var err error

	if *useExecutable {
		// Hardcoded executable path for testing
		executablePath := "/usr/sbin/nginx"
		processID = process.GetProcessIDbyExecutable(executablePath)
		if processID == 0 {
			log.Fatalf("[ERROR] Failed to retrieve PID for executable: %s\n", executablePath)
		}
		log.Printf("[INFO] Using PID %d for executable: %s\n", processID, executablePath)
	} else {
		// Ensure the correct number of arguments are provided
		if len(flag.Args()) < 1 {
			log.Fatalf("[ERROR] Usage: %s [-use-executable] <ProcessID>\n", filepath.Base(os.Args[0]))
		}

		// Parse the Process ID (PID) from the command-line arguments
		processID, err = strconv.Atoi(flag.Args()[0])
		if err != nil {
			log.Fatalf("[ERROR] Invalid Process ID (PID): %v\n", err)
		}
		log.Printf("[INFO] Using PID from arguments: %d\n", processID)
	}

	// Invoke the restart process functionality
	process.RestartProcess(processID)
}