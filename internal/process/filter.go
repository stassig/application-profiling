package process

import (
	"application_profiling/internal/util/logger"
	"bufio"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

var (
	excludePrefixes = []string{
		"/dev/", "/proc/", "/sys/", "/run/", "/tmp/", "/usr/lib/locale/", "/usr/share/locale/",
	}
)

// FilterStraceLog reads a raw strace log file, filters file paths, and writes to a new log file
func FilterStraceLog(inputFilePath, outputFilePath string) {
	// Compile a regex to match valid file paths
	filePathRegex := regexp.MustCompile(`(?:\s|")((/|\.\/)[^" ]+)`)

	// Open input file
	inputFile, err := os.Open(inputFilePath)
	logger.Error(err, "Failed to open input file")
	defer inputFile.Close()

	// Open output file
	outputFile, err := os.Create(outputFilePath)
	logger.Error(err, "Failed to create output file")
	defer outputFile.Close()

	// Process the file paths
	err = processStraceLog(inputFile, outputFile, filePathRegex)
	logger.Error(err, "Failed to process strace log")
}

// processStraceLog scans the input file, filters file paths, and writes them to the output file
func processStraceLog(inputFile *os.File, outputFile *os.File, filePathRegex *regexp.Regexp) error {
	// Track seen file paths to remove duplicates
	seenPaths := make(map[string]bool)
	currentWorkingDirectory := "/"

	scanner := bufio.NewScanner(inputFile)
	for scanner.Scan() {
		line := scanner.Text()

		// Update working directory if "chdir" syscall is encountered
		currentWorkingDirectory = updateWorkingDirectory(line, currentWorkingDirectory)

		// Extract file paths from the line
		match := filePathRegex.FindStringSubmatch(line)
		if match == nil {
			continue
		}

		filePath := match[1]

		// Resolve relative paths (e.g., "./file")
		if strings.HasPrefix(filePath, "./") {
			filePath = filepath.Join(currentWorkingDirectory, filePath[2:])
		}

		// Skip duplicates and invalid paths
		if seenPaths[filePath] || isShortGenericPath(filePath) || hasExcludedPrefix(filePath) {
			continue
		}

		// Add to seen paths and write to output file
		seenPaths[filePath] = true
		if _, err := outputFile.WriteString(filePath + "\n"); err != nil {
			return err
		}
	}

	// Handle scanner errors
	if err := scanner.Err(); err != nil {
		return err
	}

	return nil
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

// updateWorkingDirectory checks for chdir syscall and updates the current directory
func updateWorkingDirectory(line, currentDir string) string {
	if !strings.Contains(line, "chdir(") {
		return currentDir
	}

	dirRegex := regexp.MustCompile(`chdir\("([^"]+)"\)`)
	matches := dirRegex.FindStringSubmatch(line)
	if matches != nil && len(matches) > 1 {
		return matches[1]
	}

	return currentDir
}
