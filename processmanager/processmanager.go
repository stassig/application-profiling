// TO DO: Integrate Environment variables
// TO DO: Integrate User

package processmanager

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"os/exec"
	"os/user"
	"strconv"
	"strings"

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

// RestartProcess handles restarting a process by PID
func RestartProcess(pid int) {
	// Step 1: Extract command line arguments
	cmdlinePath := fmt.Sprintf("/proc/%d/cmdline", pid)
	cmdline, err := os.ReadFile(cmdlinePath)
	logError(err, "Read process cmdline")

	// Step 2: Get the executable path
	exePath := fmt.Sprintf("/proc/%d/exe", pid)
	executable, err := os.Readlink(exePath)
	logError(err, "Read process executable path")
	log.Printf("[DEBUG] Executable path: %s\n", executable)

	// Step 3: Get the working directory
	cwdPath := fmt.Sprintf("/proc/%d/cwd", pid)
	cwd, err := os.Readlink(cwdPath)
	logError(err, "Read process working directory")
	log.Printf("[DEBUG] Working directory: %s\n", cwd)

	// Step 4: Get environment variables
	environPath := fmt.Sprintf("/proc/%d/environ", pid)
	environData, err := os.ReadFile(environPath)
	logError(err, "Read process environment variables")
	environment := parseEnvironment(environData)
	log.Printf("[DEBUG] Environment Variables: %v\n", environment)

	// Step 5: Get the user associated with the process
	userName := getUserByPID(pid)
	log.Printf("[DEBUG] User: %s\n", userName)

	// Use the parser to reconstruct the command
	reconstructedCommand := cmdparser.ParseCmdline(executable, cmdline)
	log.Printf("[DEBUG] Reconstructed command: %s\n", reconstructedCommand)

	// Step 6: Kill the process
	err = exec.Command("sudo", "kill", strconv.Itoa(pid)).Run()
	logError(err, fmt.Sprintf("Kill process with PID %d", pid))

	// Step 7: Restart the process
	cmd := exec.Command("sudo", "bash", "-c", reconstructedCommand)
	cmd.Dir = cwd              // Set the working directory
	cmd.Env = environment      // Set the environment variables
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

// parseEnvironment parses environment variables from a null-byte separated string
func parseEnvironment(data []byte) []string {
	// Split by null byte
	envVars := strings.Split(string(data), "\x00")
	
	// Filter out empty and invalid entries
	var validEnvVars []string
	for _, env := range envVars {
		env = strings.TrimSpace(env) // Remove any leading/trailing whitespace

		if env == "" {
			continue // Skip empty entries
		}
		if strings.Contains(env, "=") { // Check for valid KEY=VALUE format
			validEnvVars = append(validEnvVars, env)
		} else {
			log.Printf("[WARNING] Ignoring invalid environment variable: %s\n", env)
		}
	}

	return validEnvVars
}

// getUserByPID retrieves the user associated with a process by PID
func getUserByPID(pid int) string {
	statusPath := fmt.Sprintf("/proc/%d/status", pid)
	data, err := os.ReadFile(statusPath)
	if err != nil {
		log.Fatalf("[ERROR] Failed to read process status: %v", err)
	}

	// Extract UID from the status file
	var uid string
	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "Uid:") {
			fields := strings.Fields(line)
			if len(fields) > 1 {
				uid = fields[1] // The UID is the second field
			}
			break
		}
	}

	// Lookup username by UID
	userInfo, err := user.LookupId(uid)
	if err != nil {
		log.Fatalf("[ERROR] Failed to lookup user by UID (%s): %v", uid, err)
	}
	return userInfo.Username
}
