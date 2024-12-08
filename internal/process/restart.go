// TO DO: Copy entire directories if present (e.g. /var/lib/mysql)
// TO DO: Proper mapping for cmdline arguments
// TO DO: Solve /usr/lib/mysql/plugin/ → /usr/lib/mysql/plugin/auth_socket.so (if parent directory exists in the list -> skip?)
// TO DO: Add rules for /etc/nginx, /var/lib/mysql
// TO DO: User groups, permissions, etc.
// ├── cmd/
// │   ├── main.go
// │   ├── dockerize.go      # CLI entry point for the "dockerize" command
// │   ├── profile.go        # CLI entry point for the "profile" command

// --- BACKLOG ---

// TO DO: Add user in cmdparser
// TO DO: Integrate /etc/os-release info for accurate base image
// TO DO: More elegant solution than sleep for strace
// TO DO: Clean up: interfacing?; PortInfo struct?

package process

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

// RestartProcess handles restarting a process by its Process ID (PID)
func RestartProcess(processID int) {
	// Retrieve process information
	processInfo := GetProcessInfo(processID)

	// Log debug information
	processInfo.LogProcessDetails()

	// Save process information to a YAML file
	processInfo.SaveAsYAML()

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
	return filepath.Join(currentDirectory, fmt.Sprintf("strace_log_%d%s.log", pid, suffix))
}
