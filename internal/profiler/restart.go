// TO DO: Don't add entire directories from trace file
// TO DO: Add rules for /etc/nginx, /var/lib/mysql
// TO DO: Proper mapping for cmdline arguments
// ТО DO: Indentation for Dockerfile
// TO DO: Refactor dockerize action

// --- BACKLOG ---

// TO DO: Add user in cmdparser
// TO DO: Integrate /etc/os-release info for accurate base image
// TO DO: More elegant solution than sleep for strace
// TO DO: Clean up: interfacing; PortInfo struct?
// TO DO: User groups, permissions, etc.

package profiler

import (
	"bytes"
	"fmt"
	"os/exec"
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

	// Get the output file path for strace
	logfilePath := BuildFilePath("bin/tracing", fmt.Sprintf("strace_log_%d.log", info.PID))

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
