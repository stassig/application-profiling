package subcommands

import (
	"flag"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"application_profiling/internal/profiler"

	"github.com/charmbracelet/log"
)

func RunProfile(args []string) {
	fs := flag.NewFlagSet("profile", flag.ExitOnError)
	useExecutable := fs.Bool("use-executable", false, "Use a hardcoded executable to determine the PID")
	traceWait := fs.Int("trace-wait", 5, "Duration (in seconds) to wait while the tracer captures data")
	fs.Parse(args)

	traceWaitDuration := time.Duration(*traceWait) * time.Second

	processID := getProcessID(fs.Args(), *useExecutable)

	// 1. Retrieve process information
	processInfo := profiler.GetProcessInfo(processID)

	// 2. Log debug information
	processInfo.LogProcessDetails()

	// 3. Save process information to a YAML file
	processInfo.SaveAsYAML()

	// 4. Restart the process with strace monitoring
	profiler.RestartProcess(processInfo, traceWaitDuration)

	// 5. Filter the strace log file to remove duplicates and invalid paths
	profiler.FilterStraceLog(processInfo)
}

// getProcessID retrieves the process ID from the arguments or by using the executable path
func getProcessID(args []string, useExecutable bool) int {
	if useExecutable {
		executablePath := "/usr/sbin/nginx"
		processID := profiler.GetProcessIDbyExecutable(executablePath)
		if processID == 0 {
			log.Fatalf("[ERROR] Failed to retrieve PID for executable: %s\n", executablePath)
		}
		log.Infof("Using PID %d for executable: %s", processID, executablePath)
		return processID
	}

	if len(args) < 1 {
		log.Fatalf("[ERROR] Usage: %s profile [-use-executable] <ProcessID>\n", filepath.Base(os.Args[0]))
	}

	processID, err := strconv.Atoi(args[0])
	if err != nil {
		log.Fatalf("[ERROR] Invalid Process ID (PID): %v\n", err)
	}
	log.Info("Retrieved process ID from arguments", "PID", processID)
	return processID
}
