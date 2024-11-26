// TO DO: Integrate Environment variables
// TO DO: Integrate User

package main

import (
	"log"
	"os"
	"path/filepath"
	"strconv"

	"application_profiling/processmanager"
)

func main() {
	// Validate arguments
	if len(os.Args) < 2 {
		log.Fatalf("[ERROR] Usage: %s <PID>\n", filepath.Base(os.Args[0]))
	}

	// Parse the PID from command-line arguments
	pid, err := strconv.Atoi(os.Args[1])
	if err != nil {
		log.Fatalf("[ERROR] Invalid PID: %v\n", err)
	}

	log.Printf("[INFO] Restarting process with PID: %d\n", pid)

	// Call the restart process logic
	processmanager.RestartProcess(pid)
}
