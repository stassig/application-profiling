// TO DO: Clean up the code (use struct for process info)
// TO DO: Use waitgroups to synchronize monitoring
// TO DO: Ensure bpftrace starts without explicit sleep

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
	restartWithMonitoring(
		processID, reconstructedCommand, workingDirectory, environmentVariables, processOwner, executablePath,
	)
}

// restartWithMonitoring handles monitoring, terminating, and restarting the process
func restartWithMonitoring( processID int, reconstructedCommand string, workingDirectory string,
	environmentVariables []string, processOwner string, executablePath string) {
	// Use channels to synchronize monitoring
	started := make(chan bool)
	finished := make(chan bool)
	logFilePath := fmt.Sprintf("file_access_log_fatrace_%d.log", processID)

	// Start monitoring in a separate goroutine
	go StartFatrace(logFilePath, started, finished)

	// Wait until monitoring starts
	<-started

	// Terminate the existing process and start a new one
	terminateProcess(processID)
	newPID := startProcess(reconstructedCommand, workingDirectory, environmentVariables, processOwner, executablePath)
	childProcesses := GetChildProcessIDs(newPID)
	bootStrapPID := newPID - 1 // Init process that exits after process setup

	// Wait for monitoring to finish
	<-finished

	log.Printf("[INFO] New process started with PID: %d\n", newPID)
	log.Printf("[INFO] Child processes: %v\n", childProcesses)

	// Add the new process and its children to the list of monitored PIDs
	monitoredPIDs := append([]int{newPID, bootStrapPID}, childProcesses...)
	// Filter the log file for monitored PIDs
	FilterFatraceLog(logFilePath, monitoredPIDs)
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
	log.Printf("[INFO] Init PID: %d\n", cmd.Process.Pid)

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