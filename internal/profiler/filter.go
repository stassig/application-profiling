package profiler

import (
	"application_profiling/internal/util/logger"
	"bufio"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

var (
	filePathRegex = regexp.MustCompile(`(?:\s|")((/|\.\/)[^" ]+)`)
	dirRegex      = regexp.MustCompile(`chdir\("([^"]+)"\)`)
)

// FilterStraceLog reads a raw strace log file, filters file paths, and writes to a new log file
func FilterStraceLog(info *ProcessInfo) {
	// Get the input and output file paths
	inputFilePath := BuildFilePath("bin/tracing", fmt.Sprintf("strace_log_%d.log", info.PID))
	outputFilePath := BuildFilePath("bin/tracing", fmt.Sprintf("strace_log_%d_filtered.log", info.PID))

	// Open input file
	inputFile, err := os.Open(inputFilePath)
	logger.Error(err, "Failed to open input file")
	defer inputFile.Close()

	// Open output file
	outputFile, err := os.Create(outputFilePath)
	logger.Error(err, "Failed to create output file")
	defer outputFile.Close()

	// Process the strace log
	err = processStraceLog(inputFile, outputFile, info.WorkingDirectory)
	logger.Error(err, "Failed to process strace log")
}

// processStraceLog scans the input file, filters file paths, and writes them to the output file
func processStraceLog(inputFile *os.File, outputFile *os.File, initialWorkingDirectory string) error {
	filePaths := []string{}
	seenPaths := make(map[string]bool)
	currentWorkingDirectory := initialWorkingDirectory

	scanner := bufio.NewScanner(inputFile)
	for scanner.Scan() {
		line := scanner.Text()

		// Skip lines with error indicators
		if containsErrorIndicators(line) {
			continue
		}

		// Update working directory if "chdir" syscall is encountered
		currentWorkingDirectory = updateWorkingDirectory(line, currentWorkingDirectory)

		// Extract and resolve file path
		filePath, err := extractFilePath(line, currentWorkingDirectory)
		if err != nil {
			continue
		}

		// Skip duplicates and invalid paths
		if seenPaths[filePath] || isGenericOrExcluded(filePath) {
			continue
		}

		// Mark as seen and append to list
		seenPaths[filePath] = true
		filePaths = append(filePaths, filePath)
	}

	if err := scanner.Err(); err != nil {
		return err
	}

	// Collapse application-specific directories
	filePaths = collapseApplicationSpecificDirs(filePaths)

	// Write the final paths
	for _, filePath := range filePaths {
		if _, err := outputFile.WriteString(filePath + "\n"); err != nil {
			return err
		}
	}

	return nil
}

// extractFilePath extracts and resolves the file path from a line
func extractFilePath(line, currentWorkingDirectory string) (string, error) {
	match := filePathRegex.FindStringSubmatch(line)
	if match == nil {
		return "", errors.New("No file path found")
	}
	filePath := match[1]

	// Resolve relative paths (e.g., "./file")
	if strings.HasPrefix(filePath, "./") {
		filePath = filepath.Join(currentWorkingDirectory, filePath[2:])
	}

	return filePath, nil
}

// updateWorkingDirectory updates the current directory if a chdir syscall is found
func updateWorkingDirectory(line, currentDir string) string {
	if !strings.Contains(line, "chdir(") {
		return currentDir
	}

	matches := dirRegex.FindStringSubmatch(line)
	if matches != nil && len(matches) > 1 {
		return matches[1]
	}
	return currentDir
}

// containsErrorIndicators checks if a line contains error-related indicators
func containsErrorIndicators(line string) bool {
	return strings.Contains(line, "(Invalid argument)") || strings.Contains(line, "(No such file or directory)")
}

// isGenericOrExcluded checks if a file path is generic or excluded
func isGenericOrExcluded(path string) bool {
	return isGenericPath(path) || hasExcludedPrefix(path)
}

// isGenericPath checks if a file path is system-generic
func isGenericPath(path string) bool {
	// Normalize by removing trailing slash
	clean := strings.TrimSuffix(path, "/")
	return GenericPathsSet[clean]
}

// hasExcludedPrefix checks if a file path starts with any excluded prefix
func hasExcludedPrefix(path string) bool {
	for prefix := range ExcludePrefixesSet {
		if strings.HasPrefix(path, prefix) {
			return true
		}
	}
	return false
}

// collapseApplicationSpecificDirs collapses application-specific directories
func collapseApplicationSpecificDirs(filePaths []string) []string {

	collapsed := make([]string, 0, len(filePaths))

	for _, path := range filePaths {
		// Split the path into components
		parts := strings.Split(strings.TrimPrefix(path, "/"), "/")
		if len(parts) <= 1 {
			// Nothing to collapse if it's just "/etc" or "/usr"
			collapsed = append(collapsed, path)
			continue
		}

		// Attempt to find the shortest prefix that is not generic.
		// We'll walk down the path components until we hit a directory
		// that isn't listed in GenericPaths.
		//
		// For example: /etc/nginx/conf.d
		//   - Check "/etc" (generic)
		//   - Next "/etc/nginx" (not generic?), if not generic and not excluded:
		//       we consider this an application-specific directory.
		//
		// If "/etc/nginx" is determined to be application-specific,
		// we collapse everything under it to "/etc/nginx".

		candidate := "/"
		collapsedPath := path // default to original if we can't collapse
		for i := 0; i < len(parts); i++ {
			candidate = filepath.Join(candidate, parts[i])
			// Once we pass the first component (the generic directory),
			// check if candidate is still generic or excluded.
			if i > 0 && !isGenericOrExcluded(candidate) {
				// We've found a directory that is not in GenericPaths and not excluded,
				// so treat it as the top-level application-specific directory.
				collapsedPath = candidate
				break
			}
		}

		collapsed = append(collapsed, collapsedPath)
	}

	// Deduplicate after collapsing
	seen := make(map[string]bool)
	finalPaths := []string{}
	for _, p := range collapsed {
		if !seen[p] {
			seen[p] = true
			finalPaths = append(finalPaths, p)
		}
	}

	sort.Strings(finalPaths) // Ensure sorted order
	return finalPaths
}
