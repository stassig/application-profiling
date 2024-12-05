package process

import (
	"application_profiling/internal/util/logger"
	"bufio"
	"fmt"
	"log"
	"os"
	"os/user"
	"path/filepath"
	"strconv"
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

// EnsureSocketDirectories ensures that the directories for the given socket paths exist
// and sets their ownership to the specified user.
func EnsureSocketDirectories(sockets []string, username string) {
	// Get user information
	uid, gid := getUIDGID(username)

	for _, socketPath := range sockets {
		dirPath := filepath.Dir(socketPath)
		log.Printf("[DEBUG] Ensuring directory for socket: %s", dirPath)

		// Create directory if it doesn't exist
		err := os.MkdirAll(dirPath, 0755)
		logger.Error(err, fmt.Sprintf("Failed to create directory %s", dirPath))

		// Change ownership of the directory
		err = os.Chown(dirPath, uid, gid)
		logger.Error(err, fmt.Sprintf("Failed to change ownership of directory %s", dirPath))
	}
}

// getUIDGID retrieves the UID and GID for a given username.
func getUIDGID(username string) (int, int) {
	usr, err := user.Lookup(username)
	logger.Error(err, fmt.Sprintf("Failed to look up user %s", username))

	uid, err := strconv.Atoi(usr.Uid)
	logger.Error(err, fmt.Sprintf("Failed to convert UID %s", usr.Uid))

	gid, err := strconv.Atoi(usr.Gid)
	logger.Error(err, fmt.Sprintf("Failed to convert GID %s", usr.Gid))

	return uid, gid
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
