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

// GetProcessInodeSet retrieves a set of socket inodes used by the specified process and its children
func GetProcessInodeSet(processID int) map[string]struct{} {
	processIDs := append([]int{processID}, GetChildProcessIDs(processID)...)
	return getProcessSocketInodes(processIDs)
}

// GetUnixDomainSockets retrieves all Unix domain sockets for the specified process
func GetUnixDomainSockets(inodeSet map[string]struct{}) []string {
	return mapInodesToPaths(inodeSet)
}

// GetListeningTCPPorts retrieves all TCP listening ports for the specified process
func GetListeningTCPPorts(inodeSet map[string]struct{}) []int {
	tcpPorts := parseListeningPortsFromNet(inodeSet, "/proc/net/tcp")
	tcp6Ports := parseListeningPortsFromNet(inodeSet, "/proc/net/tcp6")
	return removeDuplicatePorts(append(tcpPorts, tcp6Ports...))
}

// GetListeningUDPPorts retrieves all UDP listening ports for the specified process
func GetListeningUDPPorts(inodeSet map[string]struct{}) []int {
	udpPorts := parseListeningPortsFromNet(inodeSet, "/proc/net/udp")
	udp6Ports := parseListeningPortsFromNet(inodeSet, "/proc/net/udp6")
	return removeDuplicatePorts(append(udpPorts, udp6Ports...))
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
	logger.Error(err, "Parsing /proc/net/unix")

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

// parseListeningPortsFromNet parses a /proc/net/* file for sockets in LISTEN state
func parseListeningPortsFromNet(inodeSet map[string]struct{}, netFilePath string) []int {
	file, err := os.Open(netFilePath)
	logger.Error(err, fmt.Sprintf("Failed to open %s", netFilePath))
	defer file.Close()

	var listeningPorts []int
	scanner := bufio.NewScanner(file)

	// Skip the header line
	if scanner.Scan() {
		// Skip the header
	}

	for scanner.Scan() {
		fields := strings.Fields(scanner.Text())
		if len(fields) < 10 {
			continue
		}

		// Extract inode and state
		localAddress := fields[1] // Format: "0100007F:1F90" (IP:PORT in hex)
		state := fields[3]        // Connection state (e.g., "0A" for LISTEN)
		inode := fields[9]        // Inode number

		// Check if this socket is in LISTEN state and belongs to the target process
		if state == "0A" { // 0A == LISTEN
			if _, exists := inodeSet[inode]; exists {
				port := extractPortFromHex(localAddress)
				if port > 0 {
					listeningPorts = append(listeningPorts, port)
				}
			}
		}
	}

	return listeningPorts
}

// extractPortFromHex extracts the port from a hex-formatted "IP:PORT" string
func extractPortFromHex(addrPort string) int {
	parts := strings.Split(addrPort, ":")
	if len(parts) != 2 {
		return 0
	}
	portHex := parts[1]
	port, err := strconv.ParseInt(portHex, 16, 32)
	if err != nil {
		return 0
	}
	return int(port)
}

// removeDuplicatePorts removes duplicates from a slice of ports
func removeDuplicatePorts(ports []int) []int {
	seen := make(map[int]struct{})
	var uniquePorts []int

	for _, port := range ports {
		if _, exists := seen[port]; !exists {
			seen[port] = struct{}{}
			uniquePorts = append(uniquePorts, port)
		}
	}

	return uniquePorts
}
