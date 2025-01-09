package util

import (
	"application_profiling/internal/profiler"
	"bufio"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/charmbracelet/log"
)

// MergeFilteredLogs merges the filtered logs of the given PIDs into a single file.
func MergeFilteredLogs(processIDs []int) {
	// Create a map to store unique paths
	mergedPaths := make(map[string]bool)

	// Read filtered logs for each PID
	for _, pid := range processIDs {
		filteredFilePath := profiler.BuildFilePath(fmt.Sprintf("output/%d/profile", pid), "strace_filtered.log")

		file, err := os.Open(filteredFilePath)
		if err != nil {
			log.Errorf("Failed to open filtered log for PID %d: %v", pid, err)
			continue
		}

		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			if line != "" {
				mergedPaths[line] = true
			}
		}
		file.Close()
	}

	// Convert map to sorted slice
	finalLines := make([]string, 0, len(mergedPaths))
	for path := range mergedPaths {
		finalLines = append(finalLines, path)
	}
	sort.Strings(finalLines)

	// Write to a new merged file
	lastPID := processIDs[len(processIDs)-1]
	mergedFilePath := profiler.BuildFilePath(fmt.Sprintf("output/%d/profile", lastPID), "strace_merged.log")

	mergedFile, err := os.Create(mergedFilePath)
	if err != nil {
		log.Errorf("Failed to create merged log file: %v", err)
		return
	}
	defer mergedFile.Close()

	for _, line := range finalLines {
		_, _ = mergedFile.WriteString(line + "\n")
	}

	log.Infof("Merged strace logs have been written to: %s", mergedFilePath)
}
