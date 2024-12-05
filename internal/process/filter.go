package process

import (
	"application_profiling/internal/util/logger"
	"bufio"
	"errors"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

var (
	excludePrefixes = []string{
		"/dev/", "/proc/", "/sys/", "/run/", "/tmp/", "/usr/lib/locale/", "/usr/share/locale/",
	}
	filePathRegex = regexp.MustCompile(`(?:\s|")((/|\.\/)[^" ]+)`)
	dirRegex      = regexp.MustCompile(`chdir\("([^"]+)"\)`)
)

// FilterStraceLog reads a raw strace log file, filters file paths, and writes to a new log file
func FilterStraceLog(inputFilePath, outputFilePath, initialWorkingDirectory string) {
	// Open input file
	inputFile, err := os.Open(inputFilePath)
	logger.Error(err, "Failed to open input file")
	defer inputFile.Close()

	// Open output file
	outputFile, err := os.Create(outputFilePath)
	logger.Error(err, "Failed to create output file")
	defer outputFile.Close()

	// Process the strace log
	err = processStraceLog(inputFile, outputFile, initialWorkingDirectory)
	logger.Error(err, "Failed to process strace log")
}

// processStraceLog scans the input file, filters file paths, and writes them to the output file
func processStraceLog(inputFile *os.File, outputFile *os.File, initialWorkingDirectory string) error {
	filePaths := []string{}
	seenPaths := make(map[string]bool)
	currentWorkingDirectory := initialWorkingDirectory

	scanner := bufio.NewScanner(inputFile)
	for scanner.Scan() {
		// Read the next line
		line := scanner.Text()

		// Update working directory if "chdir" syscall is encountered
		currentWorkingDirectory = updateWorkingDirectory(line, currentWorkingDirectory)

		// Extract and resolve file path
		filePath, err := extractFilePath(line, currentWorkingDirectory)
		if err != nil {
			continue
		}

		// Skip duplicates and invalid paths
		if seenPaths[filePath] || isShortGenericPath(filePath) || hasExcludedPrefix(filePath) {
			continue
		}

		// Add to seen paths and collect for sorting
		seenPaths[filePath] = true
		filePaths = append(filePaths, filePath)
	}

	// Handle scanner errors
	if err := scanner.Err(); err != nil {
		return err
	}

	// Sort the file paths
	sort.Strings(filePaths)

	// Write the sorted paths to the output file
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

// isShortGenericPath checks if a file path is too short to be meaningful
func isShortGenericPath(path string) bool {
	return len(path) < 6
}

// hasExcludedPrefix checks if a file path starts with an excluded prefix
func hasExcludedPrefix(path string) bool {
	for _, prefix := range excludePrefixes {
		if strings.HasPrefix(path, prefix) {
			return true
		}
	}
	return false
}
