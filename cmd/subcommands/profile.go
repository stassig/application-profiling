// TO DO: Refactor cmd & subcommands
// TO DO: Investigate Dockerfile: "RUN groupadd -r mysql" & file permissions

// --- BACKLOG ---

// TO DO: Clean up codebase: interfacing?

package subcommands

import (
	"flag"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"application_profiling/internal/profiler"
	"application_profiling/internal/util"

	"github.com/charmbracelet/log"
)

func RunProfile(args []string) {
	fs := flag.NewFlagSet("profile", flag.ExitOnError)
	useExecutable := fs.Bool("use-executable", false, "Use a hardcoded executable to determine the PID")
	traceWait := fs.Int("trace-wait", 5, "Duration (in seconds) to wait while the tracer captures data")
	fs.Parse(args)

	traceWaitDuration := time.Duration(*traceWait) * time.Second

	// Retrieve PIDs from arguments or executable
	processIDs := getProcessIDs(fs.Args(), *useExecutable)
	if len(processIDs) == 0 {
		log.Fatalf("[ERROR] No valid PIDs found for profiling.")
	}

	for _, processID := range processIDs {
		profileProcess(processID, traceWaitDuration)
	}

	// Merge filtered logs from all processes
	util.MergeFilteredLogs(processIDs)
}

// profileProcess profiles a single process by ID
func profileProcess(processID int, traceWaitDuration time.Duration) {
	// 1. Retrieve process information
	processInfo := profiler.GetProcessInfo(processID)

	// 2. Log debug information
	util.LogProcessDetails(processInfo)

	// 3. Save process information to a YAML file
	processInfo.SaveAsYAML()

	// 4. Restart the process with strace monitoring
	profiler.RestartProcess(processInfo, traceWaitDuration)

	// 5. Filter the strace log file to remove duplicates and invalid paths
	profiler.FilterStraceLog(processInfo)
}

// getProcessIDs retrieves process IDs from arguments or executable
func getProcessIDs(args []string, useExecutable bool) []int {
	if useExecutable {
		executablePath := "/usr/sbin/nginx"
		processID := profiler.GetProcessIDbyExecutable(executablePath)
		if processID == 0 {
			log.Fatalf("[ERROR] Failed to retrieve PID for executable: %s\n", executablePath)
		}
		log.Infof("Using PID %d for executable: %s", processID, executablePath)
		return []int{processID}
	}

	if len(args) < 1 {
		log.Fatalf("[ERROR] Usage: %s profile [-use-executable] <ProcessID>\n", filepath.Base(os.Args[0]))
	}

	// Parse comma-separated PIDs from arguments
	processIDStrings := strings.Split(args[0], ",")
	processIDs := []int{}

	for _, processIDString := range processIDStrings {
		processID, err := strconv.Atoi(strings.TrimSpace(processIDString))
		if err != nil {
			log.Errorf("[ERROR] Invalid Process ID: %s", processIDString)
			continue
		}
		processIDs = append(processIDs, processID)
	}
	return processIDs
}
