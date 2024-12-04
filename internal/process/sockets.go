package process

import (
	"application_profiling/internal/util/logger"
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// GetSockets retrieves a list of socket file paths used by the process and its child processes
func GetSockets(processID int) []string {
	// Collect all relevant process IDs
	processIDs := append([]int{processID}, GetChildProcessIDs(processID)...)

	// Get inodes associated with the processes
	inodeSet := getProcessSocketInodes(processIDs)

	// Map inodes to socket paths
	socketPaths := mapInodesToPaths(inodeSet)

	return socketPaths
}

// getProcessSocketInodes returns a set of socket inodes used by the given process IDs
func getProcessSocketInodes(processIDs []int) map[string]struct{} {
	inodeSet := make(map[string]struct{})

	for _, pid := range processIDs {
		fdPath := fmt.Sprintf("/proc/%d/fd", pid)
		fds, err := os.ReadDir(fdPath)
		if err != nil {
			logger.Error(err, fmt.Sprintf("Reading file descriptors for PID %d", pid))
			continue
		}

		for _, fd := range fds {
			fdFullPath := filepath.Join(fdPath, fd.Name())
			linkTarget, err := os.Readlink(fdFullPath)
			if err != nil {
				logger.Error(err, fmt.Sprintf("Reading link for fd %s", fdFullPath))
				continue
			}

			// Check if the link target is a socket (e.g., "socket:[12345]")
			if strings.HasPrefix(linkTarget, "socket:[") && strings.HasSuffix(linkTarget, "]") {
				// Extract the inode number from "socket:[inode]"
				inode := linkTarget[len("socket:[") : len(linkTarget)-1]
				inodeSet[inode] = struct{}{}
			}
		}
	}

	return inodeSet
}

// mapInodesToPaths maps inodes to their corresponding socket paths by parsing /proc/net/unix
func mapInodesToPaths(inodeSet map[string]struct{}) []string {
	var socketPaths []string
	inodeToPath, err := parseProcNetUnix()
	if err != nil {
		logger.Error(err, "Parsing /proc/net/unix")
		return socketPaths
	}

	for inode := range inodeSet {
		if path, exists := inodeToPath[inode]; exists && path != "" {
			socketPaths = append(socketPaths, path)
		}
	}

	return socketPaths
}

// parseProcNetUnix parses /proc/net/unix to build a map of inode to socket path
func parseProcNetUnix() (map[string]string, error) {
	inodeToPath := make(map[string]string)

	file, err := os.Open("/proc/net/unix")
	if err != nil {
		return nil, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	// Skip the header line
	if scanner.Scan() {
		// Header line skipped
	}

	for scanner.Scan() {
		fields := strings.Fields(scanner.Text())
		if len(fields) >= 7 {
			inode := fields[6]
			var path string
			if len(fields) >= 8 {
				path = fields[7]
			}
			inodeToPath[inode] = path
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return inodeToPath, nil
}
