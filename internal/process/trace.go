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
func StartFatrace(processID int, started chan bool, finished chan bool) {
	defer func() { finished <- true }()

	logFilePath := fmt.Sprintf("file_access_log_fatrace_%d.log", processID)

	// Redirect output to the log file
	output, err := os.Create(logFilePath)
	logger.Error(err, "Failed to create log file for fatrace")
	defer output.Close()

	// Prepare a single pipeline command: fatrace | grep "nginx"
	cmd := exec.Command("sh", "-c", "sudo fatrace | grep nginx")

	// Redirect the pipeline output to the log file
	cmd.Stdout = output
	cmd.Stderr = os.Stderr // Optional: to capture any errors directly to stderr

	logger.Info("Starting fatrace monitoring for file accesses.")

	// Start the pipeline process
	err = cmd.Start()
	logger.Error(err, "Failed to start fatrace pipeline")

	// Allow the pipeline to initialize and signal readiness
	time.Sleep(1 * time.Second)
	started <- true

	// Monitor for a fixed duration
	time.Sleep(5 * time.Second)

	// Terminate the pipeline
	logger.Info("Stopping fatrace monitoring.")
	err = cmd.Process.Kill()
	if err != nil {
		logger.Warning(fmt.Sprintf("Failed to kill fatrace process: %v", err))
	}

	// Wait for the pipeline to exit and capture any errors
	err = cmd.Wait()
	if err != nil {
		logger.Warning(fmt.Sprintf("fatrace pipeline exited with error: %v", err))
	}

	logger.Info("fatrace monitoring stopped successfully.")
}