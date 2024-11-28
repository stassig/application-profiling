// TO DO: Explicitly typed variables
// TO DO: Clean up the code (move proccess info to a struct and separate file)
// TO DO: Use waitgroups to synchronize monitoring
// TO DO: Get child processes IDs based on new process ID
// TO DO: Filter the trace log based on the PIDs (& add the PID to the log output)
// TO DO: Ensure bpftrace has started without sleep

// PHASE 1: Dependency Gathering

// Step 1: Get process information (executable path, command-line arguments, working directory, environment variables, process owner, sockets, user permissions, etc.)
// Step 2: Save process information to a file
// Step 3: Restart the process
// Step 4: Get new PID and child processes
// Step 5: Monitor the new process and log file access
// Step 6: Filter the trace log based on the PIDs & clean up duplicate logs

// PHASE 2: Dockerization

// Step 1: Copy files from trace log to "profiling" directory (ensure working symlinks)
// Step 2: Map the "profiling" directory to the Dockerfile
// Step 3: Map process info file to the Dockerfile

package process

import (
	"bytes"
	"fmt"
	"log"
	"os/exec"
	"strconv"

	"application_profiling/internal/cmdparser"
	"application_profiling/internal/util/logger"
)

// RestartProcess handles restarting a process by its Process ID (PID)
func RestartProcess(processID int) {
	executablePath := GetExecutablePath(processID)
	commandLineArgs := GetCommandLineArgs(processID)
	workingDirectory := GetWorkingDirectory(processID)
	environmentVariables := GetEnvironmentVariables(processID)
	processOwner := GetProcessOwner(processID)
	// Parse the command-line string into a valid command
	reconstructedCommand := cmdparser.ParseCommandLine(executablePath, commandLineArgs)

	// Log debug information
	logProcessDetails(processID, executablePath, commandLineArgs, workingDirectory, environmentVariables, processOwner, reconstructedCommand)

	// Restart process with monitoring
	newPID := restartWithMonitoring(
		processID, reconstructedCommand, workingDirectory, environmentVariables, processOwner, executablePath,
	)

	log.Printf("[INFO] New process started with PID: %d\n", newPID)
}

// restartWithMonitoring handles monitoring, terminating, and restarting the process
func restartWithMonitoring( processID int, reconstructedCommand string, workingDirectory string,
	environmentVariables []string, processOwner string, executablePath string) int {
	// Use channels to synchronize monitoring
	started := make(chan bool)
	finished := make(chan bool)

	// Start monitoring in a separate goroutine
	go StartMonitoring(processID, started, finished)

	// Wait until monitoring starts
	<-started

	// Terminate the existing process and start a new one
	terminateProcess(processID)
	newPID := startProcess(reconstructedCommand, workingDirectory, environmentVariables, processOwner, executablePath)

	// Wait for monitoring to finish
	<-finished

	return newPID
}

// terminateProcess stops the process with the given PID
func terminateProcess(processID int) {
	err := exec.Command("sudo", "kill", strconv.Itoa(processID)).Run()
	logger.Error(err, fmt.Sprintf("Terminating process with PID %d", processID))
}

// startProcess starts a process with the given command, working directory, environment variables and user
func startProcess(command, workingDirectory string, environmentVariables []string, user string, executablePath string) int {
	cmd := exec.Command("sudo", "-u", user, "bash", "-c", command)
	cmd.Dir = workingDirectory
	cmd.Env = environmentVariables

	var stderrBuffer bytes.Buffer
	cmd.Stderr = &stderrBuffer

	log.Printf("[INFO] Starting process as user %s: %s\n", user, command)
	err := cmd.Run()

	logger.Error(err, fmt.Sprintf("Failed to start process: %s", stderrBuffer.String()))
	log.Println("[INFO] Process started successfully")

	newPID := GetProcessIDbyExecutable(executablePath)
	log.Printf("[INFO] New PID: %d\n", newPID)

	return newPID
}

// logProcessDetails logs key details about a process in one method.
func logProcessDetails(processID int, executablePath string, commandLineArgs []byte, workingDirectory string, 
	environmentVariables []string, processOwner, reconstructedCommand string) {
	log.Printf("[DEBUG] Process ID: %d", processID)
	log.Printf("[DEBUG] Executable path: %s", executablePath)
	log.Printf("[DEBUG] Command-line arguments: %s", commandLineArgs)
	log.Printf("[DEBUG] Working directory: %s", workingDirectory)
	log.Printf("[DEBUG] Environment variables: %v", environmentVariables)
	log.Printf("[DEBUG] Process owner: %s", processOwner)
	log.Printf("[DEBUG] Reconstructed command: %s", reconstructedCommand)
}