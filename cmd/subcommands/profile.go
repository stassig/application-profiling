package subcommands

import (
	"flag"
	"log"
	"os"
	"path/filepath"
	"strconv"

	"application_profiling/internal/profiler"
)

func RunProfile(args []string) {
	fs := flag.NewFlagSet("profile", flag.ExitOnError)
	useExecutable := fs.Bool("use-executable", false, "Use a hardcoded executable to determine the PID")
	fs.Parse(args)

	processID := getProcessID(fs.Args(), *useExecutable)

	// 1. Retrieve process information
	processInfo := profiler.GetProcessInfo(processID)

	// 2. Log debug information
	processInfo.LogProcessDetails()

	// 3. Save process information to a YAML file
	processInfo.SaveAsYAML()

	// 4. Restart the process with strace monitoring
	profiler.RestartProcess(processInfo)
}

// getProcessID retrieves the process ID from the arguments or by using the executable path
func getProcessID(args []string, useExecutable bool) int {
	if useExecutable {
		executablePath := "/usr/sbin/nginx"
		processID := profiler.GetProcessIDbyExecutable(executablePath)
		if processID == 0 {
			log.Fatalf("[ERROR] Failed to retrieve PID for executable: %s\n", executablePath)
		}
		log.Printf("[INFO] Using PID %d for executable: %s\n", processID, executablePath)
		return processID
	}

	if len(args) < 1 {
		log.Fatalf("[ERROR] Usage: %s profile [-use-executable] <ProcessID>\n", filepath.Base(os.Args[0]))
	}

	processID, err := strconv.Atoi(args[0])
	if err != nil {
		log.Fatalf("[ERROR] Invalid Process ID (PID): %v\n", err)
	}
	log.Printf("[INFO] Using PID from arguments: %d\n", processID)
	return processID
}
