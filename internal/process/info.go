package process

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"os/user"
	"strconv"
	"strings"

	"application_profiling/internal/parser"
	"application_profiling/internal/util/logger"
)

// ProcessInfo contains key information about a process
type ProcessInfo struct {
	PID                  int
	ExecutablePath       string
	CommandLineArgs      []string
	WorkingDirectory     string
	EnvironmentVariables []string
	ProcessOwner         string
	ReconstructedCommand string
	Sockets              []string
}

// GetProcessInfo retrieves key information about a process by its Process ID (PID)
func GetProcessInfo(processID int) *ProcessInfo {
	info := &ProcessInfo{
		PID: processID,
	}

	info.ExecutablePath = GetExecutablePath(processID)
	info.CommandLineArgs = GetCommandLineArgs(processID)
	info.WorkingDirectory = GetWorkingDirectory(processID)
	info.EnvironmentVariables = GetEnvironmentVariables(processID)
	info.ProcessOwner = GetProcessOwner(processID)
	info.Sockets = GetSockets(processID)
	info.ReconstructedCommand = parser.ParseCommandLine(info.ExecutablePath, info.CommandLineArgs)

	return info
}

// LogProcessDetails logs key details about a process in one method.
func (info *ProcessInfo) LogProcessDetails() {
	log.Printf("[DEBUG] Process ID: %d", info.PID)
	log.Printf("[DEBUG] Executable path: %s", info.ExecutablePath)
	log.Printf("[DEBUG] Command-line arguments: %s", info.CommandLineArgs)
	log.Printf("[DEBUG] Working directory: %s", info.WorkingDirectory)
	log.Printf("[DEBUG] Environment variables: %v", info.EnvironmentVariables)
	log.Printf("[DEBUG] Process owner: %s", info.ProcessOwner)
	log.Printf("[DEBUG] Reconstructed command: %s", info.ReconstructedCommand)
	log.Printf("[DEBUG] Sockets: %v", info.Sockets)
}

// GetExecutablePath retrieves the path to the executable of the process
func GetExecutablePath(processID int) string {
	// Read the symbolic link to the executable from /proc/<PID>/exe
	executablePath := fmt.Sprintf("/proc/%d/exe", processID)
	resolvedPath, err := os.Readlink(executablePath)
	logger.Error(err, "Reading process executable path")
	return resolvedPath
}

// GetCommandLineArgs retrieves the command-line arguments of the process
func GetCommandLineArgs(processID int) []string {
	// Read the command-line arguments from /proc/<PID>/cmdline
	commandLinePath := fmt.Sprintf("/proc/%d/cmdline", processID)
	commandLineData, err := os.ReadFile(commandLinePath)
	logger.Error(err, "Reading process command-line arguments")

	// Replace null bytes with spaces and split the string into fields
	commandLineArgs := strings.Fields(strings.ReplaceAll(string(commandLineData), "\x00", " "))
	return commandLineArgs
}

// GetWorkingDirectory retrieves the working directory of the process
func GetWorkingDirectory(processID int) string {
	// Read the symbolic link to the working directory from /proc/<PID>/cwd
	workingDirectoryPath := fmt.Sprintf("/proc/%d/cwd", processID)
	workingDirectory, err := os.Readlink(workingDirectoryPath)
	logger.Error(err, "Reading process working directory")
	return workingDirectory
}

// GetEnvironmentVariables retrieves and parses the environment variables of the process
func GetEnvironmentVariables(processID int) []string {
	// Read the environment variables from /proc/<PID>/environ
	environmentFilePath := fmt.Sprintf("/proc/%d/environ", processID)
	rawEnvironmentData, err := os.ReadFile(environmentFilePath)
	logger.Error(err, "Reading process environment variables")
	return parseEnvironmentVariables(rawEnvironmentData)
}

// GetOwnerByPID retrieves the user associated with the process ID
func GetProcessOwner(processID int) string {
	// Read the status file to get the UID of the process
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
	if err != nil {
		logger.Warning("No child processes found or failed to retrieve child processes for parent PID: " + strconv.Itoa(parentPID))
		return []int{} // Return an empty slice if there are no child processes
	}

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
