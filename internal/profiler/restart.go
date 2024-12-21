// TO DO: Add executable to tar archive
// TO DO: Integrate /etc/os-release info for accurate base image
// TO DO: Add user groups based on process owner

// --- BACKLOG ---

// TO DO: More elegant solution than sleep for strace - commandline option for time
// TO DO: Inline error handling for better readability (charm bracelet log package)
// TO DO: Clean up: interfacing;
// TO DO: User groups, permissions (filter by user)
// TO DO: Performance profile (CPU & RAM)

package profiler

import (
	"bytes"
	"fmt"
	"os/exec"
	"strconv"
	"time"

	"github.com/charmbracelet/log"
)

// RestartProcess handles restarting a process using its ProcessInfo
func RestartProcess(processInfo *ProcessInfo) {
	// Restart process with monitoring
	terminateProcess(processInfo.PID)
	startProcessWithStrace(processInfo)
}

// terminateProcess stops the process with the given PID
func terminateProcess(processID int) {
	log.Info(fmt.Sprintf("Terminating process with PID %d", processID))
	err := exec.Command("sudo", "kill", strconv.Itoa(processID)).Run()
	if err != nil {
		log.Error("Failed to terminate process", "error", err)
	}
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
	log.Info(fmt.Sprintf("Starting process with strace: %s", info.ReconstructedCommand))
	err := cmd.Start()
	if err != nil {
		log.Error("Failed to start process with strace", "stderr", stderrBuffer.String(), "error", err)
	}

	// Sleep for a few seconds to allow strace to capture initial syscalls
	time.Sleep(5 * time.Second)

	// Terminate the strace process after data collection
	err = cmd.Process.Kill()
	if err != nil {
		log.Error("Failed to kill strace process", "error", err)
	}
}

// prepareStraceCommand constructs the strace command to execute
func prepareStraceCommand(info *ProcessInfo, logfilePath string) *exec.Cmd {
	// Modify the reconstructed command to include sudo -u <process_owner>
	// userPrefixedCommand := fmt.Sprintf("sudo -u %s %s", info.ProcessOwner, info.ReconstructedCommand)

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
