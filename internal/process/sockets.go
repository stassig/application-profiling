package process

import (
	"application_profiling/internal/util/logger"
	"bufio"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

// GetSockets retrieves a list of socket file paths used by the process and its child processes
func GetSockets(processID int) []string {
    var socketPaths []string
    // Get all process IDs: the main process and its children
    processIDs := []int{processID}
    childPIDs := GetChildProcessIDs(processID) 
    processIDs = append(processIDs, childPIDs...)

    inodeSet := make(map[string]struct{})
    for _, pid := range processIDs {
        fdPath := fmt.Sprintf("/proc/%d/fd", pid)
        fds, err := ioutil.ReadDir(fdPath)
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

            // Check if the link target is a socket
            if strings.HasPrefix(linkTarget, "socket") {
                // Extract the inode number
                inode := linkTarget[8:len(linkTarget)-1]
                inodeSet[inode] = struct{}{}
            }
        }
    }

    // Now read /proc/net/unix to map inodes to socket paths
    socketInodes := make(map[string]string)
    file, err := os.Open("/proc/net/unix")
    if err != nil {
        logger.Error(err, "Opening /proc/net/unix")
        return socketPaths
    }
    defer file.Close()

    scanner := bufio.NewScanner(file)
    // Skip the header line
    if scanner.Scan() {
        // Header line skipped
    }

    for scanner.Scan() {
        line := scanner.Text()
        fields := strings.Fields(line)
        if len(fields) >= 7 {
            inode := fields[6]
            path := ""
            if len(fields) >= 8 {
                path = fields[7]
            }
            socketInodes[inode] = path
        }
    }
    if err := scanner.Err(); err != nil {
        logger.Error(err, "Scanning /proc/net/unix")
    }

    // Collect the socket paths
    for inode := range inodeSet {
        if path, ok := socketInodes[inode]; ok && path != "" {
            socketPaths = append(socketPaths, path)
        }
    }

    return socketPaths
}
