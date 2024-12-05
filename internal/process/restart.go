// TO DO: Update filtering logic for chdir syscalls
// TO DO: Add more params to strace (e.g., mmap)
// TO DO: Integrate /etc/os-release info for accurate base image
// TO DO: Clean up: struct for process info; interfacing; refactor startProcessWithStrace

package process

import (
	"bytes"
	"fmt"
	"log"
	"os/exec"
	"strconv"
	"time"

	"application_profiling/internal/parser"
	"application_profiling/internal/util/logger"
)

// RestartProcess handles restarting a process by its Process ID (PID)
func RestartProcess(processID int) {
	executablePath := GetExecutablePath(processID)
	commandLineArgs := GetCommandLineArgs(processID)
	workingDirectory := GetWorkingDirectory(processID)
	environmentVariables := GetEnvironmentVariables(processID)
	processOwner := GetProcessOwner(processID)
	sockets := GetSockets(processID)
	// Parse the command-line string into a valid command
	reconstructedCommand := parser.ParseCommandLine(executablePath, commandLineArgs)

	// Log debug information
	logProcessDetails(processID, executablePath, commandLineArgs, workingDirectory, environmentVariables, processOwner, reconstructedCommand, sockets)

	// Restart process with monitoring
	terminateProcess(processID)
	startProcessWithStrace(processID, reconstructedCommand, workingDirectory, environmentVariables, processOwner, sockets)
}

// terminateProcess stops the process with the given PID
func terminateProcess(processID int) {
	err := exec.Command("sudo", "kill", strconv.Itoa(processID)).Run()
	logger.Error(err, fmt.Sprintf("Terminating process with PID %d", processID))
	time.Sleep(5 * time.Second)
}

// startProcessWithStrace starts a process with strace monitoring
func startProcessWithStrace(processID int, command, workingDirectory string, environmentVariables []string, user string, sockets []string) {
	// Ensure the directories for the sockets exist
	EnsureSocketDirectories(sockets, user)

	// Prepare the strace command
	cmd := prepareStraceCommand(processID, command, workingDirectory, environmentVariables)
	var stderrBuffer bytes.Buffer
	cmd.Stderr = &stderrBuffer

	// Start the process with strace
	logger.Info(fmt.Sprintf("Starting process with strace: %s", command))
	err := cmd.Start()
	logger.Error(err, fmt.Sprintf("Failed to start process: %s", stderrBuffer.String()))

	// Sleep for a few seconds to allow strace to capture
	time.Sleep(5 * time.Second)

	// Terminate the strace process after data collection
	err = cmd.Process.Kill()
	logger.Warning(fmt.Sprintf("Failed to kill strace process: %v", err))

	// Filter the strace log file to remove duplicates and invalid paths
	logfilePath := fmt.Sprintf("/home/stassig/go/application-profiling/strace_log_%d.log", processID)
	filteredLogfilePath := fmt.Sprintf("/home/stassig/go/application-profiling/filtered_strace_log_%d.log", processID)
	FilterStraceLog(logfilePath, filteredLogfilePath)
}

// prepareStraceCommand constructs the strace command to execute
func prepareStraceCommand(processID int, command, workingDirectory string, environmentVariables []string) *exec.Cmd {
	logfilePath := fmt.Sprintf("/home/stassig/go/application-profiling/strace_log_%d.log", processID)

	// Prepare the strace command arguments
	cmdArgs := []string{
		"strace",
		"-f",
		"-e", "trace=open,openat,chdir",
		"-o", logfilePath,
		"bash", "-c", command,
	}

	cmd := exec.Command("sudo", cmdArgs...)
	cmd.Dir = workingDirectory
	cmd.Env = environmentVariables

	return cmd
}

// logProcessDetails logs key details about a process in one method.
func logProcessDetails(processID int, executablePath string, commandLineArgs []byte, workingDirectory string, environmentVariables []string, processOwner, reconstructedCommand string, sockets []string) {
	log.Printf("[DEBUG] Process ID: %d", processID)
	log.Printf("[DEBUG] Executable path: %s", executablePath)
	log.Printf("[DEBUG] Command-line arguments: %s", commandLineArgs)
	log.Printf("[DEBUG] Working directory: %s", workingDirectory)
	log.Printf("[DEBUG] Environment variables: %v", environmentVariables)
	log.Printf("[DEBUG] Process owner: %s", processOwner)
	log.Printf("[DEBUG] Reconstructed command: %s", reconstructedCommand)
	log.Printf("[DEBUG] Sockets: %v", sockets)
}
