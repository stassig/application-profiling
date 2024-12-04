// TO DO: Integrate socket files (mkdir -p /var/run/mysqld/) & user permissions (test mysql)
// TO DO: Add more params to strace (e.g., mmap)
// TO DO: Integrate /etc/os-release info for accurate base image
// TO DO: Clean up: struct for process info; interfacing

package process

import (
	"bytes"
	"fmt"
	"log"
	"os/exec"
	"strconv"
	"time"

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
	terminateProcess(processID)
	startProcessWithStrace(processID, reconstructedCommand, workingDirectory, environmentVariables, processOwner)
}

// terminateProcess stops the process with the given PID
func terminateProcess(processID int) {
	err := exec.Command("sudo", "kill", strconv.Itoa(processID)).Run()
	logger.Error(err, fmt.Sprintf("Terminating process with PID %d", processID))
}

// startProcessWithStrace starts a process with `strace` monitoring
func startProcessWithStrace(processID int, command, workingDirectory string, environmentVariables []string, user string)  {
	// Hardcoded path for strace output log file (for now)
	logfilePath := fmt.Sprintf("/home/stassig/go/application-profiling/strace_log_%d.log", processID)
	filteredLogfilePath := fmt.Sprintf("/home/stassig/go/application-profiling/filtered_strace_log_%d.log", processID)
    // Prepare the strace command
    cmd := exec.Command("sudo", "-u", user, "strace", "-f", "-e", "trace=open,openat", "-o", logfilePath, "bash", "-c", command)
    cmd.Dir = workingDirectory
    cmd.Env = environmentVariables
    var stderrBuffer bytes.Buffer
    cmd.Stderr = &stderrBuffer

    log.Printf("[INFO] Starting process as user %s with strace: %s\n", user, command)
    err := cmd.Start()
    logger.Error(err, fmt.Sprintf("Failed to start process: %s", stderrBuffer.String()))

    // Sleep for 5 seconds to allow strace to gather data
    time.Sleep(5 * time.Second)

    log.Println("[INFO] Process started successfully")
    log.Printf("[INFO] strace PID: %d\n", cmd.Process.Pid)

    // Terminate the strace process after data collection
    err = cmd.Process.Kill()
	logger.Warning(fmt.Sprintf("Failed to kill strace process: %v", err))

	// Filter the strace log file to remove duplicates and invalid paths
	FilterStraceLog(logfilePath, filteredLogfilePath)
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