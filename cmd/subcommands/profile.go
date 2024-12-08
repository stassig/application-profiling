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

	var processID int
	var err error

	if *useExecutable {
		executablePath := "/usr/sbin/nginx"
		processID = profiler.GetProcessIDbyExecutable(executablePath)
		if processID == 0 {
			log.Fatalf("[ERROR] Failed to retrieve PID for executable: %s\n", executablePath)
		}
		log.Printf("[INFO] Using PID %d for executable: %s\n", processID, executablePath)
	} else {
		if fs.NArg() < 1 {
			log.Fatalf("[ERROR] Usage: %s profile [-use-executable] <ProcessID>\n", filepath.Base(os.Args[0]))
		}

		processID, err = strconv.Atoi(fs.Arg(0))
		if err != nil {
			log.Fatalf("[ERROR] Invalid Process ID (PID): %v\n", err)
		}
		log.Printf("[INFO] Using PID from arguments: %d\n", processID)
	}

	// Invoke the restart process functionality
	profiler.RestartProcess(processID)
}
