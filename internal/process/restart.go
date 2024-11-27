package process

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"os/exec"
	"os/user"
	"strconv"
	"strings"
	"time"

	"application_profiling/internal/cmdparser"
	"application_profiling/internal/util"
)

// RestartProcess handles restarting a process by its Process ID (PID)
func RestartProcess(processID int) {
	executablePath := getProcessExecutablePath(processID)
	commandLineArgs := getProcessCommandLineArgs(processID)
	workingDirectory := getProcessWorkingDirectory(processID)
	environmentVariables := getProcessEnvironmentVariables(processID)
	processOwner := getProcessOwnerByPID(processID)
	// Parse the command-line string into a valid command
	reconstructedCommand := cmdparser.ParseCommandLine(executablePath, commandLineArgs)

	log.Printf("[DEBUG] Executable path: %s\n", executablePath)
	log.Printf("[DEBUG] Command-line arguments: %s\n", commandLineArgs)
	log.Printf("[DEBUG] Working directory: %s\n", workingDirectory)
	log.Printf("[DEBUG] Environment variables: %v\n", environmentVariables)
	log.Printf("[DEBUG] Process owner: %s\n", processOwner)
	log.Printf("[DEBUG] Reconstructed command: %s\n", reconstructedCommand)

	// Prepare log file for bpftrace monitoring
	logFilePath := fmt.Sprintf("file_access_log_%d.txt", processID)

	// Use a channel to synchronize
	started := make(chan bool)
	finished := make(chan bool)

	go monitorFileAccess(logFilePath, started, finished)

	// Wait until bpftrace starts
	<-started

	// Terminate the existing process and start a new one
	terminateProcess(processID)
	newPID := startProcess(reconstructedCommand, workingDirectory, environmentVariables, processOwner, executablePath)

	log.Printf("[INFO] New process started with PID: %d\n", newPID)

	// Wait for the bpftrace monitoring
	<-finished
}

// monitorFileAccess uses bpftrace to monitor file access events for a given PID
func monitorFileAccess(outputFile string, started chan bool, finished chan bool) {
	// Define the bpftrace script
    bpftraceScript := `
	tracepoint:syscalls:sys_enter_openat {
		printf("%s %s\n", comm, str(args->filename));
	}
	`

	// Redirect output to the log file
	output, err := os.Create(outputFile)
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
    time.Sleep(5 * time.Second)

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

    finished <- true
}

// getProcessExecutablePath retrieves the path to the executable of the process
func getProcessExecutablePath(processID int) string {
	executablePath := fmt.Sprintf("/proc/%d/exe", processID)
	resolvedPath, err := os.Readlink(executablePath)
	util.LogError(err, "Reading process executable path")
	return resolvedPath
}

// getProcessCommandLineArgs retrieves the command-line arguments of the process
func getProcessCommandLineArgs(processID int) []byte {
	commandLinePath := fmt.Sprintf("/proc/%d/cmdline", processID)
	commandLineArgs, err := os.ReadFile(commandLinePath)
	util.LogError(err, "Reading process command-line arguments")
	return commandLineArgs
}

// getProcessWorkingDirectory retrieves the working directory of the process
func getProcessWorkingDirectory(processID int) string {
	workingDirectoryPath := fmt.Sprintf("/proc/%d/cwd", processID)
	workingDirectory, err := os.Readlink(workingDirectoryPath)
	util.LogError(err, "Reading process working directory")
	return workingDirectory
}

// getProcessEnvironmentVariables retrieves and parses the environment variables of the process
func getProcessEnvironmentVariables(processID int) []string {
	environmentFilePath := fmt.Sprintf("/proc/%d/environ", processID)
	rawEnvironmentData, err := os.ReadFile(environmentFilePath)
	util.LogError(err, "Reading process environment variables")
	return parseEnvironmentVariables(rawEnvironmentData)
}

// getProcessOwnerByPID retrieves the user associated with the process ID
func getProcessOwnerByPID(processID int) string {
	statusFilePath := fmt.Sprintf("/proc/%d/status", processID)
	rawStatusData, err := os.ReadFile(statusFilePath)
	util.LogError(err, "Reading process status file")

	// Extract UID from the status file
	var userID string
	statusLines := strings.Split(string(rawStatusData), "\n")
	for _, line := range statusLines {
		if strings.HasPrefix(line, "Uid:") {
			fields := strings.Fields(line)
			if len(fields) > 1 {
				userID = fields[1] // The UID is the second field
			}
			break
		}
	}

	// Lookup username by UID
	userInfo, err := user.LookupId(userID)
	util.LogError(err, fmt.Sprintf("Looking up user by UID (%s)", userID))
	return userInfo.Username
}

// terminateProcess stops the process with the given PID
func terminateProcess(processID int) {
	err := exec.Command("sudo", "kill", strconv.Itoa(processID)).Run()
	util.LogError(err, fmt.Sprintf("Terminating process with PID %d", processID))
}

// startProcess starts a process with the given command, working directory, environment variables and user
func startProcess(command, workingDirectory string, environmentVariables []string, user string, executablePath string) int {
	cmd := exec.Command("sudo", "-u", user, "bash", "-c", command)
	cmd.Dir = workingDirectory
	cmd.Env = environmentVariables

	var stderrBuffer bytes.Buffer
	cmd.Stderr = &stderrBuffer

	log.Printf("[INFO] Starting process as user %s: %s\n", user, command)
	err := cmd.Run()

	util.LogError(err, fmt.Sprintf("Failed to start process: %s", stderrBuffer.String()))
	log.Println("[INFO] Process started successfully")

	newPID := GetProcessIDbyExecutable(executablePath)
	log.Printf("[INFO] New PID: %d\n", newPID)

	return newPID
}

// getProcessIDbyExecutable retrieves the PID of a process by its executable path
func GetProcessIDbyExecutable(executablePath string) int {
	output, err := exec.Command("pgrep", "-f", executablePath).Output()
	util.LogError(err, "Failed to retrieve PID for executable: "+executablePath)

	pid, err := strconv.Atoi(strings.TrimSpace(string(output)))
	util.LogError(err, "Failed to convert PID to integer for executable: "+executablePath)

	return pid
}

// parseEnvironmentVariables parses environment variables from a null-byte separated string
func parseEnvironmentVariables(rawData []byte) []string {
	rawVariables := strings.Split(string(rawData), "\x00")
	var environmentVariables []string

	for _, variable := range rawVariables {
		trimmedVariable := strings.TrimSpace(variable)
		if trimmedVariable == "" {
			continue // Skip empty entries
		}
		if strings.Contains(trimmedVariable, "=") { // Check for valid KEY=VALUE format
			environmentVariables = append(environmentVariables, trimmedVariable)
		} else {
			log.Printf("[WARNING] Ignoring invalid environment variable: %s\n", trimmedVariable)
		}
	}
	return environmentVariables
}

