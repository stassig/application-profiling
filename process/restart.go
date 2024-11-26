package process

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

// RestartProcess handles restarting a process by its Process ID (PID)
func RestartProcess(processID int) {
	executablePath := getProcessExecutablePath(processID)
	commandLineArgs := getProcessCommandLineArgs(processID)
	workingDirectory := getProcessWorkingDirectory(processID)
	environmentVariables := getProcessEnvironmentVariables(processID)
	processOwner := getProcessOwnerByPID(processID)

	log.Printf("[DEBUG] Executable path: %s\n", executablePath)
	log.Printf("[DEBUG] Command-line arguments: %s\n", commandLineArgs)
	log.Printf("[DEBUG] Working directory: %s\n", workingDirectory)
	log.Printf("[DEBUG] Environment variables: %v\n", environmentVariables)
	log.Printf("[DEBUG] Process owner: %s\n", processOwner)

	reconstructedCommand := cmdparser.ParseCommandLine(executablePath, commandLineArgs)
	log.Printf("[DEBUG] Reconstructed command: %s\n", reconstructedCommand)

	terminateProcess(processID)
	startProcess(reconstructedCommand, workingDirectory, environmentVariables)
}

// getProcessExecutablePath retrieves the path to the executable of the process
func getProcessExecutablePath(processID int) string {
	executablePath := fmt.Sprintf("/proc/%d/exe", processID)
	resolvedPath, err := os.Readlink(executablePath)
	logError(err, "Reading process executable path")
	return resolvedPath
}

// getProcessCommandLineArgs retrieves the command-line arguments of the process
func getProcessCommandLineArgs(processID int) []byte {
	commandLinePath := fmt.Sprintf("/proc/%d/cmdline", processID)
	commandLineArgs, err := os.ReadFile(commandLinePath)
	logError(err, "Reading process command-line arguments")
	return commandLineArgs
}

// getProcessWorkingDirectory retrieves the working directory of the process
func getProcessWorkingDirectory(processID int) string {
	workingDirectoryPath := fmt.Sprintf("/proc/%d/cwd", processID)
	workingDirectory, err := os.Readlink(workingDirectoryPath)
	logError(err, "Reading process working directory")
	return workingDirectory
}

// getProcessEnvironmentVariables retrieves and parses the environment variables of the process
func getProcessEnvironmentVariables(processID int) []string {
	environmentFilePath := fmt.Sprintf("/proc/%d/environ", processID)
	rawEnvironmentData, err := os.ReadFile(environmentFilePath)
	logError(err, "Reading process environment variables")
	return parseEnvironmentVariables(rawEnvironmentData)
}

// getProcessOwnerByPID retrieves the user associated with the process ID
func getProcessOwnerByPID(processID int) string {
	statusFilePath := fmt.Sprintf("/proc/%d/status", processID)
	rawStatusData, err := os.ReadFile(statusFilePath)
	logError(err, "Reading process status file")

	// Extract UID from the status file
	var userID string
	statusLines := strings.Split(string(rawStatusData), "\n")
	for _, line := range statusLines {
		if strings.HasPrefix(line, "Uid:") {
			fields := strings.Fields(line)
			if len(fields) > 1 {
				userID = fields[1] // The UID is the second field
			}
			break
		}
	}

	// Lookup username by UID
	userInfo, err := user.LookupId(userID)
	logError(err, fmt.Sprintf("Looking up user by UID (%s)", userID))
	return userInfo.Username
}

// terminateProcess stops the process with the given PID
func terminateProcess(processID int) {
	err := exec.Command("sudo", "kill", strconv.Itoa(processID)).Run()
	logError(err, fmt.Sprintf("Terminating process with PID %d", processID))
}

// startProcess starts a process with the given command, working directory, and environment variables
func startProcess(command, workingDirectory string, environmentVariables []string) {
	cmd := exec.Command("sudo", "bash", "-c", command)
	cmd.Dir = workingDirectory
	cmd.Env = environmentVariables

	var stderrBuffer bytes.Buffer
	cmd.Stderr = &stderrBuffer

	log.Printf("[INFO] Starting process: %s\n", command)
	err := cmd.Run()

	logError(err, fmt.Sprintf("Failed to start process: %s", stderrBuffer.String()))
	log.Println("[INFO] Process started successfully")
}

// parseEnvironmentVariables parses environment variables from a null-byte separated string
func parseEnvironmentVariables(rawData []byte) []string {
	rawVariables := strings.Split(string(rawData), "\x00")
	var environmentVariables []string

	for _, variable := range rawVariables {
		trimmedVariable := strings.TrimSpace(variable)
		if trimmedVariable == "" {
			continue // Skip empty entries
		}
		if strings.Contains(trimmedVariable, "=") { // Check for valid KEY=VALUE format
			environmentVariables = append(environmentVariables, trimmedVariable)
		} else {
			log.Printf("[WARNING] Ignoring invalid environment variable: %s\n", trimmedVariable)
		}
	}
	return environmentVariables
}

// logError checks for an error and logs it if present
func logError(err error, contextMessage string) {
	if err != nil {
		log.Fatalf("[ERROR] %s: %v\n", contextMessage, err)
	}
}
