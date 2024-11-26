// TO DO: Integrate Environment variables
// TO DO: Integrate User

package main

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"

	"application_profiling/cmdparser"
)

// logError checks for an error and logs it if present
func logError(err error, message string) {
	if err != nil {
		log.Fatalf("[ERROR] %s: %v\n", message, err)
	} else {
		log.Printf("[INFO] %s\n", message)
	}
}

// restartProcess handles restarting a process by PID
func restartProcess(pid int) {
	// Step 1: Extract command line arguments
	cmdlinePath := fmt.Sprintf("/proc/%d/cmdline", pid)
	cmdline, err := os.ReadFile(cmdlinePath)
	logError(err, "Read process cmdline")

	// Step 2: Get the executable path
	exePath := fmt.Sprintf("/proc/%d/exe", pid)
	executable, err := os.Readlink(exePath)
	logError(err, "Read process executable path")
	log.Printf("[DEBUG] Executable path: %s\n", executable)

	// Use the parser to reconstruct the command
	reconstructedCommand := cmdparser.ParseCmdline(executable, cmdline)
	log.Printf("[DEBUG] Reconstructed command: %s\n", reconstructedCommand)

	// Step 3: Get the working directory
	cwdPath := fmt.Sprintf("/proc/%d/cwd", pid)
	cwd, err := os.Readlink(cwdPath)
	logError(err, "Read process working directory")
	log.Printf("[DEBUG] Working directory: %s\n", cwd)

	// Step 4: Kill the process
	err = exec.Command("sudo", "kill", strconv.Itoa(pid)).Run()
	logError(err, fmt.Sprintf("Kill process with PID %d", pid))

	// Step 5: Restart the process
	// Use bash to execute the command string safely
	cmd := exec.Command("sudo", "bash", "-c", reconstructedCommand)
	cmd.Dir = cwd // Set the working directory
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	log.Printf("[INFO] Restarting process: %s\n", reconstructedCommand)
	err = cmd.Run()
	if err != nil {
		log.Fatalf("[ERROR] Failed to restart process: %s\n", stderr.String())
	} else {
		log.Println("[INFO] Process restarted successfully")
	}
}

func main() {
	// Example usage
	if len(os.Args) < 2 {
		log.Fatalf("[ERROR] Usage: %s <PID>\n", filepath.Base(os.Args[0]))
	}

	pid, err := strconv.Atoi(os.Args[1])
	if err != nil {
		log.Fatalf("[ERROR] Invalid PID: %v\n", err)
	}

	log.Printf("[INFO] Restarting process with PID: %d\n", pid)
	restartProcess(pid)
}
