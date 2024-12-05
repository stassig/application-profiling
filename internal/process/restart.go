// TO DO: Update filtering logic for chdir syscalls
// TO DO: Add more params to strace (e.g., mmap)
// TO DO: Integrate /etc/os-release info for accurate base image
// TO DO: Clean up: interfacing;

package process

import (
	"bytes"
	"fmt"
	"os/exec"
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

	// Restart process with monitoring
	terminateProcess(processInfo.PID)
	startProcessWithStrace(processInfo)
}

// terminateProcess stops the process with the given PID
func terminateProcess(processID int) {
	err := exec.Command("sudo", "kill", strconv.Itoa(processID)).Run()
	logger.Error(err, fmt.Sprintf("Terminating process with PID %d", processID))
	time.Sleep(5 * time.Second)
}

// startProcessWithStrace starts a process with strace monitoring
func startProcessWithStrace(info *ProcessInfo) {
	EnsureSocketDirectories(info.Sockets, info.ProcessOwner)

	// Prepare the strace command
	cmd := prepareStraceCommand(info)
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
	logger.Warning(fmt.Sprintf("Failed to kill strace process: %v", err))

	// Filter the strace log file to remove duplicates and invalid paths
	logfilePath := fmt.Sprintf("/home/stassig/go/application-profiling/strace_log_%d.log", info.PID)
	filteredLogfilePath := fmt.Sprintf("/home/stassig/go/application-profiling/filtered_strace_log_%d.log", info.PID)
	FilterStraceLog(logfilePath, filteredLogfilePath)
}

// prepareStraceCommand constructs the strace command to execute
func prepareStraceCommand(info *ProcessInfo) *exec.Cmd {
	logfilePath := fmt.Sprintf("/home/stassig/go/application-profiling/strace_log_%d.log", info.PID)

	// Prepare the strace command arguments
	cmdArgs := []string{
		"strace",
		"-f",
		"-e", "trace=open,openat,chdir",
		"-o", logfilePath,
		"bash", "-c", info.ReconstructedCommand,
	}

	cmd := exec.Command("sudo", cmdArgs...)
	cmd.Dir = info.WorkingDirectory
	cmd.Env = info.EnvironmentVariables

	return cmd
}
