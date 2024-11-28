package process

import (
	"application_profiling/internal/util/logger"
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

// FilterFatraceLog filters lines containing any of the monitored PIDs and replaces the log file
func FilterFatraceLog(logFilePath string, monitoredPIDs []int) {
	// Convert PIDs to strings for easier comparison
	pidStrings := make([]string, len(monitoredPIDs))
	for i, pid := range monitoredPIDs {
		pidStrings[i] = strconv.Itoa(pid)
	}

	// Open the fatrace log file for reading
	file, err := os.Open(logFilePath)
	if err != nil {
		fmt.Printf("[ERROR] Failed to open fatrace log file: %v\n", err)
		return
	}
	defer file.Close()

	// Parse the log file line by line
	var filteredLines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()

		// Check if the line contains any of the monitored PIDs
		for _, pid := range pidStrings {
			if strings.Contains(line, pid) {
				filteredLines = append(filteredLines, line)
				break // No need to check other PIDs for this line
			}
		}
	}

	// Check for scanning errors
	if err := scanner.Err(); err != nil {
		fmt.Printf("[ERROR] Error reading fatrace log file: %v\n", err)
		return
	}

	// Overwrite the log file with the filtered content
	err = os.WriteFile(logFilePath, []byte(strings.Join(filteredLines, "\n")+"\n"), 0644)
	logger.Error(err, "Failed to write filtered fatrace log file")

	fmt.Printf("[INFO] Successfully filtered: %s\n", logFilePath)
}
