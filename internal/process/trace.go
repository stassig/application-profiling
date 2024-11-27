package process

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"os/exec"
	"time"

	"application_profiling/internal/util"
)

// StartMonitoring uses bpftrace to monitor file access events for a given PID
func StartMonitoring(processID int, started chan bool, finished chan bool) {
	defer func() { finished <- true }()

	logFilePath := fmt.Sprintf("file_access_log_%d.txt", processID)
	// Define the bpftrace script
	bpftraceScript := `
	tracepoint:syscalls:sys_enter_openat {
		printf("%s %s\n", comm, str(args->filename));
	}
	`

	// Redirect output to the log file
	output, err := os.Create(logFilePath)
	util.LogError(err, "Failed to create log file")
	defer output.Close()

	// Prepare the bpftrace command
	cmd := exec.Command("sudo", "bpftrace", "-e", bpftraceScript)
	var stderr bytes.Buffer
	cmd.Stdout = output
	cmd.Stderr = &stderr

	log.Println("[INFO] Starting bpftrace monitoring for file accesses.")

	// Start the bpftrace process
	err = cmd.Start()
	util.LogError(err, "Failed to start bpftrace")

	// Allow bpftrace to initialize and signal readiness
	time.Sleep(1 * time.Second)
	started <- true

	// Monitor for a fixed duration
	time.Sleep(3 * time.Second)

	// Terminate the bpftrace process
	log.Printf("[INFO] Stopping bpftrace monitoring\n")
	err = cmd.Process.Kill()
	if err != nil {
		log.Printf("[WARNING] Failed to kill bpftrace process: %v\n", err)
	}

	// Wait for the bpftrace process to exit and capture any errors
	err = cmd.Wait()
	if err != nil {
		log.Printf("[ERROR] bpftrace process exited with error: %v\n", err)
	}

	// Log any bpftrace errors
	if stderr.Len() > 0 {
		log.Printf("[ERROR] bpftrace error: %s\n", stderr.String())
	} else {
		log.Println("[INFO] bpftrace monitoring stopped successfully.")
	}
}
