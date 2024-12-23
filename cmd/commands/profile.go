package commands

import (
	"flag"
	"strconv"
	"strings"
	"time"

	"application_profiling/internal/profiler"
	"application_profiling/internal/util"

	"github.com/charmbracelet/log"
)

// ProfileOptions represents the options for the Profile command
type ProfileOptions struct {
	TraceWaitDuration time.Duration
	ProcessIDs        []int
}

// RunProfile handles the "profile" command logic
func RunProfile(arguments []string) {
	// Parse command line arguments
	options := parseProfileArguments(arguments)

	// Profile each process
	for _, processID := range options.ProcessIDs {
		profileProcess(processID, options.TraceWaitDuration)
	}

	// Merge filtered logs from all processes
	util.MergeFilteredLogs(options.ProcessIDs)
}

// parseProfileArguments parses command line arguments for the profile command
func parseProfileArguments(arguments []string) ProfileOptions {
	// Initialize a flag set and define the trace-wait flag
	flagSet := flag.NewFlagSet("profile", flag.ExitOnError)
	traceWait := flagSet.Int("trace-wait", 5, "Duration (in seconds) to wait while the tracer captures data")
	flagSet.Parse(arguments)

	// Convert traceWait to a duration
	traceWaitDuration := time.Duration(*traceWait) * time.Second

	// Retrieve PIDs from arguments
	processIDs := getProcessIDs(flagSet.Args())

	return ProfileOptions{
		TraceWaitDuration: traceWaitDuration,
		ProcessIDs:        processIDs,
	}
}

// getProcessIDs retrieves process IDs from arguments
func getProcessIDs(arguments []string) []int {
	// Parse comma-separated PIDs from the first argument
	processIDStrings := strings.Split(arguments[0], ",")
	processIDs := []int{}

	// Convert each PID to an integer
	for _, processIDString := range processIDStrings {
		processID, err := strconv.Atoi(strings.TrimSpace(processIDString))
		if err != nil {
			log.Errorf("Invalid Process ID: %s", processIDString)
			continue
		}
		processIDs = append(processIDs, processID)
	}
	// If no valid PIDs are found, exit the program
	if len(processIDs) == 0 {
		log.Fatalf("No valid PIDs found for profiling.")
	}

	return processIDs
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
