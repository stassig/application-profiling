package process

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"time"

	"application_profiling/internal/util/logger"
)

// StartBpftrace uses bpftrace to monitor file access events for a given PID
func StartBpftrace(processID int, started chan bool, finished chan bool) {
	defer func() { finished <- true }()

	logFilePath := fmt.Sprintf("file_access_log_bpftrace_%d.log", processID)
	// Define the bpftrace script
	bpftraceScript := `
	tracepoint:syscalls:sys_enter_openat {
		printf("%d %s %s\n", tid, comm, str(args->filename));
	}
	`

	// Redirect output to the log file
	output, err := os.Create(logFilePath)
	logger.Error(err, "Failed to create log file for bpftrace")
	defer output.Close()

	// Prepare the bpftrace command
	cmd := exec.Command("sudo", "bpftrace", "-e", bpftraceScript)
	var stderr bytes.Buffer
	cmd.Stdout = output
	cmd.Stderr = &stderr

	logger.Info("Starting bpftrace monitoring for file accesses.")

	// Start the bpftrace process
	err = cmd.Start()
	logger.Error(err, "Failed to start bpftrace")

	// Allow bpftrace to initialize and signal readiness
	time.Sleep(1 * time.Second)
	started <- true

	// Monitor for a fixed duration
	time.Sleep(3 * time.Second)

	// Terminate the bpftrace process
	logger.Info("Stopping bpftrace monitoring.")
	err = cmd.Process.Kill()
	if err != nil {
		logger.Warning(fmt.Sprintf("Failed to kill bpftrace process: %v", err))
	}

	// Wait for the bpftrace process to exit and capture any errors
	err = cmd.Wait()
	if err != nil {
		logger.Warning(fmt.Sprintf("bpftrace process exited with error: %v", err))
	}

	// Log any bpftrace errors
	if stderr.Len() > 0 {
		logger.Error(fmt.Errorf(stderr.String()), "bpftrace error")
	} else {
		logger.Info("bpftrace monitoring stopped successfully.")
	}
}

// StartFatrace uses fatrace to monitor file access events and filter by a process name (e.g., nginx)
func StartFatrace(logFilePath string, started chan bool, finished chan bool) {
	defer func() { finished <- true }()

	// Open the log file for writing
	output, err := os.Create(logFilePath)
	if err != nil {
		logger.Error(err, "Failed to create log file for fatrace")
		started <- false
		return
	}
	defer output.Close()

	// Prepare the fatrace command
	cmd := exec.Command("sudo", "fatrace")
	cmd.Stdout = output   // Directly write fatrace output to the log file
	cmd.Stderr = os.Stderr // Optional: log errors directly to stderr

	logger.Info("Starting fatrace monitoring for file accesses.")

	// Start the fatrace process
	err = cmd.Start()
	if err != nil {
		logger.Error(err, "Failed to start fatrace")
		started <- false
		return
	}

	time.Sleep(1 * time.Second)

	// Signal that monitoring has started
	started <- true

	// Monitor for a fixed duration
	time.Sleep(4 * time.Second)

	// Stop monitoring
	logger.Info("Stopping fatrace monitoring.")
	err = cmd.Process.Kill()
	if err != nil {
		logger.Warning(fmt.Sprintf("Failed to kill fatrace process: %v", err))
	}

	// Ensure the process has exited
	err = cmd.Wait()
	if err != nil {
		logger.Warning(fmt.Sprintf("fatrace process exited with error: %v", err))
	}

	logger.Info(fmt.Sprintf("fatrace log saved to: %s", logFilePath))
}
