package profiler

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/charmbracelet/log"
)

var (
	filePathRegex = regexp.MustCompile(`(?:\s|")((/|\.\/)[^" ]+)`)
	dirRegex      = regexp.MustCompile(`chdir\("([^"]+)"\)`)
)

// FilterStraceLog reads a raw strace log file, filters file paths, and writes to a new log file
func FilterStraceLog(info *ProcessInfo) {
	// Get the input and output file paths

	inputFilePath := BuildFilePath(fmt.Sprintf("vm2container/%d/profile", info.PID), "strace_raw.log")
	outputFilePath := BuildFilePath(fmt.Sprintf("vm2container/%d/profile", info.PID), "strace_filtered.log")

	// Open input file
	inputFile, err := os.Open(inputFilePath)
	if err != nil {
		log.Error("Failed to open input file", "error", err)
	}
	defer inputFile.Close()

	// Open output file
	outputFile, err := os.Create(outputFilePath)
	if err != nil {
		log.Error("Failed to create output file", "error", err)
	}
	defer outputFile.Close()

	// Process the strace log
	err = processStraceLog(inputFile, outputFile, info.WorkingDirectory, info.ExecutablePath)
	if err != nil {
		log.Error("Failed to process strace log", "error", err)
	}
}

// processStraceLog scans the input file, filters file paths, and writes them to the output file
func processStraceLog(inputFile *os.File, outputFile *os.File, initialWorkingDirectory, executablePath string) error {
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

	// Ensure the executable path is included
	if !seenPaths[executablePath] {
		filePaths = append(filePaths, executablePath)
		seenPaths[executablePath] = true
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
		return "", errors.New("no file path found")
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
	if len(matches) > 1 {
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

// collapseApplicationSpecificDirs reduces file paths by identifying the shortest
// application-specific directory that is not system-generic or excluded.
// Example: For "/etc/nginx/conf.d", it checks:
//   - "/etc" (generic)
//   - "/etc/nginx" (application-specific)
//
// If "/etc/nginx" is valid, all subpaths collapse to it.
func collapseApplicationSpecificDirs(filePaths []string) []string {
	collapsed := make([]string, 0, len(filePaths))

	for _, path := range filePaths {
		// Split path into components
		parts := strings.Split(strings.TrimPrefix(path, "/"), "/")
		if len(parts) <= 1 {
			// Keep short paths like "/etc" or "/usr" as-is
			collapsed = append(collapsed, path)
			continue
		}

		// Find the first non-generic, non-excluded directory
		candidate := "/"
		collapsedPath := path // Default to original if no collapse possible
		for i := 0; i < len(parts); i++ {
			candidate = filepath.Join(candidate, parts[i])
			if i > 0 && !isGenericOrExcluded(candidate) {
				// Use this directory as the top-level application-specific path
				collapsedPath = candidate
				break
			}
		}

		collapsed = append(collapsed, collapsedPath)
	}

	// Deduplicate and sort results
	seen := make(map[string]bool)
	finalPaths := []string{}
	for _, path := range collapsed {
		if !seen[path] {
			seen[path] = true
			finalPaths = append(finalPaths, path)
		}
	}

	sort.Strings(finalPaths)
	return finalPaths
}
