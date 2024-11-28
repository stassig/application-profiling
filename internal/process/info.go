package process

import (
	"fmt"
	"os"
	"os/exec"
	"os/user"
	"strconv"
	"strings"

	"application_profiling/internal/util/logger"
)

// GetExecutablePath retrieves the path to the executable of the process
func GetExecutablePath(processID int) string {
	executablePath := fmt.Sprintf("/proc/%d/exe", processID)
	resolvedPath, err := os.Readlink(executablePath)
	logger.Error(err, "Reading process executable path")
	return resolvedPath
}

// GetCommandLineArgs retrieves the command-line arguments of the process
func GetCommandLineArgs(processID int) []byte {
	commandLinePath := fmt.Sprintf("/proc/%d/cmdline", processID)
	commandLineArgs, err := os.ReadFile(commandLinePath)
	logger.Error(err, "Reading process command-line arguments")
	return commandLineArgs
}

// GetWorkingDirectory retrieves the working directory of the process
func GetWorkingDirectory(processID int) string {
	workingDirectoryPath := fmt.Sprintf("/proc/%d/cwd", processID)
	workingDirectory, err := os.Readlink(workingDirectoryPath)
	logger.Error(err, "Reading process working directory")
	return workingDirectory
}

// GetEnvironmentVariables retrieves and parses the environment variables of the process
func GetEnvironmentVariables(processID int) []string {
	environmentFilePath := fmt.Sprintf("/proc/%d/environ", processID)
	rawEnvironmentData, err := os.ReadFile(environmentFilePath)
	logger.Error(err, "Reading process environment variables")
	return parseEnvironmentVariables(rawEnvironmentData)
}

// GetOwnerByPID retrieves the user associated with the process ID
func GetProcessOwner(processID int) string {
	statusFilePath := fmt.Sprintf("/proc/%d/status", processID)
	rawStatusData, err := os.ReadFile(statusFilePath)
	logger.Error(err, "Reading process status file")

	var userID string
	statusLines := strings.Split(string(rawStatusData), "\n")
	for _, line := range statusLines {
		if strings.HasPrefix(line, "Uid:") {
			fields := strings.Fields(line)
			if len(fields) > 1 {
				userID = fields[1]
			}
			break
		}
	}

	userInfo, err := user.LookupId(userID)
	logger.Error(err, fmt.Sprintf("Looking up user by UID (%s)", userID))
	return userInfo.Username
}

// GetChildProcessIDs retrieves a list of child process IDs for a given parent process ID
func GetChildProcessIDs(parentPID int) []int {
	// Execute pgrep -P <parentPID>
	output, err := exec.Command("pgrep", "-P", strconv.Itoa(parentPID)).Output()
	logger.Error(err, "Failed to retrieve child process IDs for parent PID: "+strconv.Itoa(parentPID))

	// Split the output into lines and parse each line into an integer
	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	var childProcessIDs []int
	for _, line := range lines {
		if line == "" {
			continue
		}
		pid, err := strconv.Atoi(line)
		logger.Error(err, "Failed to convert PID to integer: "+line)
		childProcessIDs = append(childProcessIDs, pid)
	}
	return childProcessIDs
}

// GetProcessIDbyExecutable retrieves the PID of a process by its executable path
func GetProcessIDbyExecutable(executablePath string) int {
	// Execute pgrep -f <executablePath>
	output, err := exec.Command("pgrep", "-f", executablePath).Output()
	logger.Error(err, "Failed to retrieve PID for executable: "+executablePath)

	pid, err := strconv.Atoi(strings.TrimSpace(string(output)))
	logger.Error(err, "Failed to convert PID to integer for executable: "+executablePath)

	return pid
}

// parseEnvironmentVariables parses environment variables from a null-byte separated string
func parseEnvironmentVariables(rawData []byte) []string {
	rawVariables := strings.Split(string(rawData), "\x00")
	var environmentVariables []string

	for _, variable := range rawVariables {
		trimmedVariable := strings.TrimSpace(variable)
		if trimmedVariable == "" {
			continue
		}
		if strings.Contains(trimmedVariable, "=") {
			environmentVariables = append(environmentVariables, trimmedVariable)
		}
	}
	return environmentVariables
}
