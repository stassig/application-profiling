package profiler

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
	genericPaths = []string{
		"/", "/bin", "/boot", "/boot/efi", "/dev", "/dev/pts", "/dev/shm", "/etc",
		"/etc/network", "/etc/opt", "/etc/ssl", "/home", "/lib", "/lib32", "/lib64",
		"/lib/firmware", "/lib/x86_64-linux-gnu", "/media", "/mnt", "/opt", "/proc",
		"/root", "/run", "/run/lock", "/run/shm", "/sbin", "/srv", "/sys", "/tmp",
		"/usr", "/usr/bin", "/usr/games", "/usr/include", "/usr/lib", "/usr/lib64",
		"/usr/libexec", "/usr/lib/locale", "/usr/local", "/usr/local/bin",
		"/usr/local/games", "/usr/local/lib", "/usr/local/lib64", "/usr/local/sbin",
		"/usr/sbin", "/usr/share", "/usr/share/doc", "/usr/share/fonts",
		"/usr/share/icons", "/usr/share/locale", "/usr/share/man", "/usr/share/themes",
		"/var", "/var/backups", "/var/cache", "/var/lib", "/var/lib/apt",
		"/var/lib/dhcp", "/var/lib/dpkg", "/var/lib/snapd", "/var/lib/systemd",
		"/var/lock", "/var/log", "/var/mail", "/var/opt", "/var/run", "/var/spool",
		"/var/tmp", "/var/www", "/usr/local/bin/bash", "/usr/local/sbin/bash",
		"/usr/sbin/bash", "/usr/bin/bash",
	}

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
		if seenPaths[filePath] || isGenericPath(filePath) || hasExcludedPrefix(filePath) {
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

// containsErrorIndicators checks if a line contains error-related indicators
func containsErrorIndicators(line string) bool {
	return strings.Contains(line, "(Invalid argument)") || strings.Contains(line, "(No such file or directory)")
}

// isGenericPath checks if a file path is generic and can be excluded
func isGenericPath(path string) bool {
	for _, generic := range genericPaths {
		if path == generic || path == generic+"/" {
			return true
		}
	}
	return false
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
