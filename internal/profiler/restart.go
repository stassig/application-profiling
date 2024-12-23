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
func RestartProcess(processInfo *ProcessInfo, sleepDuration time.Duration) {
	// Restart process with monitoring
	terminateProcess(processInfo.PID)
	startProcessWithStrace(processInfo, sleepDuration)
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
func startProcessWithStrace(info *ProcessInfo, sleepDuration time.Duration) {
	// Ensure the directories for the sockets exist
	EnsureSocketDirectories(info.UnixSockets, info.ProcessUser)

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

	// Trigger a curl request to the process to allow strace to capture html page
	// exec.Command("curl", "http://localhost").Run()

	// Sleep for the specified duration to allow strace to capture initial syscalls
	time.Sleep(sleepDuration)

	// Terminate the strace process after data collection
	err = cmd.Process.Kill()
	if err != nil {
		log.Error("Failed to kill strace process", "error", err)
	}
}

// prepareStraceCommand constructs the strace command to execute
func prepareStraceCommand(info *ProcessInfo, logfilePath string) *exec.Cmd {
	// Modify the reconstructed command to include sudo -u <process_user>
	// userPrefixedCommand := fmt.Sprintf("sudo -u %s %s", info.ProcessUser, info.ReconstructedCommand)

	// Prepare the strace command arguments
	cmdArgs := []string{
		"strace",
		"-f",
		"-e", "trace=file",
		"-o", logfilePath,
		"bash", "-c", info.ReconstructedCommand,
	}

	cmd := exec.Command("sudo", cmdArgs...)
	cmd.Dir = info.WorkingDirectory
	cmd.Env = info.EnvironmentVariables

	return cmd
}
