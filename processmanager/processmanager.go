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

// RestartProcess handles restarting a process by PID
func RestartProcess(pid int) {
	executable := getExecutablePath(pid)
	cmdline := getCmdline(pid)
	cwd := getWorkingDirectory(pid)
	environment := getEnvironment(pid)
	userName := getUserByPID(pid)

	log.Printf("[DEBUG] Executable path: %s\n", executable)
	log.Printf("[DEBUG] Command-line: %s\n", cmdline)
	log.Printf("[DEBUG] Working directory: %s\n", cwd)
	log.Printf("[DEBUG] Environment Variables: %v\n", environment)
	log.Printf("[DEBUG] User: %s\n", userName)

	reconstructedCommand := cmdparser.ParseCmdline(executable, cmdline)
	log.Printf("[DEBUG] Reconstructed command: %s\n", reconstructedCommand)

	killProcess(pid)
	restartProcess(reconstructedCommand, cwd, environment)
}

// getExecutablePath retrieves the executable path of the process
func getExecutablePath(pid int) string {
	exePath := fmt.Sprintf("/proc/%d/exe", pid)
	executable, err := os.Readlink(exePath)
	logError(err, "Read process executable path")
	return executable
}

// getCmdline retrieves the command-line arguments of the process
func getCmdline(pid int) []byte {
	cmdlinePath := fmt.Sprintf("/proc/%d/cmdline", pid)
	cmdline, err := os.ReadFile(cmdlinePath)
	logError(err, "Read process cmdline")
	return cmdline
}

// getWorkingDirectory retrieves the working directory of the process
func getWorkingDirectory(pid int) string {
	cwdPath := fmt.Sprintf("/proc/%d/cwd", pid)
	cwd, err := os.Readlink(cwdPath)
	logError(err, "Read process working directory")
	return cwd
}

// getEnvironment retrieves and parses the environment variables of the process
func getEnvironment(pid int) []string {
	environPath := fmt.Sprintf("/proc/%d/environ", pid)
	environData, err := os.ReadFile(environPath)
	logError(err, "Read process environment variables")
	return parseEnvironment(environData)
}

// getUserByPID retrieves the user associated with the process by PID
func getUserByPID(pid int) string {
	statusPath := fmt.Sprintf("/proc/%d/status", pid)
	data, err := os.ReadFile(statusPath)
	logError(err, "Read process status")

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
	logError(err, fmt.Sprintf("Lookup user by UID (%s)", uid))
	return userInfo.Username
}

// killProcess kills the process with the given PID
func killProcess(pid int) {
	err := exec.Command("sudo", "kill", strconv.Itoa(pid)).Run()
	logError(err, fmt.Sprintf("Kill process with PID %d", pid))
}

// restartProcess restarts the process with the given command, working directory, and environment
func restartProcess(command, cwd string, environment []string) {
	cmd := exec.Command("sudo", "bash", "-c", command)
	cmd.Dir = cwd              // Set the working directory
	cmd.Env = environment      // Set the environment variables

	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	log.Printf("[INFO] Restarting process: %s\n", command)
	err := cmd.Run()

	logError(err, fmt.Sprintf("Failed to restart process: %s", stderr.String()))
	log.Println("[INFO] Process restarted successfully")
}

// parseEnvironment parses environment variables from a null-byte separated string
func parseEnvironment(data []byte) []string {
	envVars := strings.Split(string(data), "\x00")
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

// logError checks for an error and logs it if present
func logError(err error, message string) {
	if err != nil {
		log.Fatalf("[ERROR] %s: %v\n", message, err)
	}
}
