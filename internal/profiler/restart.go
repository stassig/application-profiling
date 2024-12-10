// TO DO: Proper mapping for cmdline arguments
// TO DO: Solve /usr/lib/mysql/plugin/ → /usr/lib/mysql/plugin/auth_socket.so (if parent directory exists in the list -> skip?)
// TO DO: Add rules for /etc/nginx, /var/lib/mysql
// TO DO: User groups, permissions, etc.
// TO DO: Move filter call to main

// --- BACKLOG ---

// TO DO: Add user in cmdparser
// TO DO: Integrate /etc/os-release info for accurate base image
// TO DO: More elegant solution than sleep for strace
// TO DO: Clean up: interfacing?; PortInfo struct?

package profiler

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"time"

	"application_profiling/internal/util/logger"
)

// RestartProcess handles restarting a process using its ProcessInfo
func RestartProcess(processInfo *ProcessInfo) {
	// Restart process with monitoring
	terminateProcess(processInfo.PID)
	startProcessWithStrace(processInfo)
}

// terminateProcess stops the process with the given PID
func terminateProcess(processID int) {
	logger.Info(fmt.Sprintf("Terminating process with PID %d", processID))
	err := exec.Command("sudo", "kill", strconv.Itoa(processID)).Run()
	logger.Error(err, fmt.Sprintf("Failed to terminate process with PID %d", processID))
	// Sleep for a few seconds to allow the process to terminate
	time.Sleep(5 * time.Second)
}

// startProcessWithStrace starts a process with strace monitoring
func startProcessWithStrace(info *ProcessInfo) {
	// Ensure the directories for the sockets exist
	EnsureSocketDirectories(info.UnixSockets, info.ProcessOwner)

	// Get the log file paths
	logfilePath := getLogFilePath(info.PID, "")
	filteredLogfilePath := getLogFilePath(info.PID, "_filtered")

	// Prepare the strace command
	cmd := prepareStraceCommand(info, logfilePath)
	var stderrBuffer bytes.Buffer
	cmd.Stderr = &stderrBuffer

	// Start the process with strace
	logger.Info(fmt.Sprintf("Starting process with strace: %s", info.ReconstructedCommand))
	err := cmd.Start()
	logger.Error(err, fmt.Sprintf("Failed to start process: %s", stderrBuffer.String()))

	// Sleep for a few seconds to allow strace to capture initial syscalls
	time.Sleep(5 * time.Second)

	// Terminate the strace process after data collection
	err = cmd.Process.Kill()
	logger.Error(err, fmt.Sprintf("Failed to kill strace process"))

	// Filter the strace log file to remove duplicates and invalid paths
	FilterStraceLog(logfilePath, filteredLogfilePath, info.WorkingDirectory)
}

// prepareStraceCommand constructs the strace command to execute
func prepareStraceCommand(info *ProcessInfo, logfilePath string) *exec.Cmd {
	// Prepare the strace command arguments
	cmdArgs := []string{
		"strace",
		"-f",
		"-e", "trace=openat,chdir,mkdir",
		"-o", logfilePath,
		"bash", "-c", info.ReconstructedCommand,
	}

	cmd := exec.Command("sudo", cmdArgs...)
	cmd.Dir = info.WorkingDirectory
	cmd.Env = info.EnvironmentVariables

	return cmd
}

// getLogFilePath generates the path for the strace log file
func getLogFilePath(pid int, suffix string) string {
	currentDirectory, err := os.Getwd()
	logger.Error(err, "Failed to get current directory")

	// Append the desired subdirectory to the current directory
	tracingDir := filepath.Join(currentDirectory, "bin", "tracing")

	// Ensure the directory exists
	err = os.MkdirAll(tracingDir, os.ModePerm)
	logger.Error(err, fmt.Sprintf("Failed to create tracing directory: %s", tracingDir))

	// Return the full path for the log file
	return filepath.Join(tracingDir, fmt.Sprintf("strace_log_%d%s.log", pid, suffix))
}
