package main

import (
	"log"
	"os"
	"path/filepath"
	"strconv"

	"application_profiling/internal/process"
)

func main() {
	// Ensure the correct number of arguments are provided
	if len(os.Args) < 2 {
		log.Fatalf("[ERROR] Usage: %s <ProcessID>\n", filepath.Base(os.Args[0]))
	}

	// Parse the Process ID (PID) from the command-line arguments
	processID, err := strconv.Atoi(os.Args[1])
	if err != nil {
		log.Fatalf("[ERROR] Invalid Process ID (PID): %v\n", err)
	}

	log.Printf("[INFO] Attempting to restart process with PID: %d\n", processID)

	// Invoke the restart process functionality from the process manager
	process.RestartProcess(processID)
}
