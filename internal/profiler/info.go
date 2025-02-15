package profiler

import (
	"fmt"
	"os"
	"os/exec"
	"os/user"
	"strconv"
	"strings"

	"github.com/charmbracelet/log"
)

// ProcessInfo represents the process metadata.
type ProcessInfo struct {
	PID                  int            `yaml:"pid"`                  // Process ID
	ChildPIDs            []int          `yaml:"childpids"`            // Child process IDs
	ProcessUser          string         `yaml:"processuser"`          // User running the process
	ProcessGroup         string         `yaml:"processgroup"`         // Group running the process
	ExecutablePath       string         `yaml:"executablepath"`       // Path to the executable
	CommandLineArguments []FlagArgument `yaml:"commandlinearguments"` // Command-line arguments
	ReconstructedCommand string         `yaml:"reconstructedcommand"` // Reconstructed command string
	WorkingDirectory     string         `yaml:"workingdirectory"`     // Current working directory
	EnvironmentVariables []string       `yaml:"environmentvariables"` // Environment variables
	UnixSockets          []string       `yaml:"unixsockets"`          // Unix domain sockets in use
	ListeningTCP         []int          `yaml:"listeningtcp"`         // TCP ports in use
	ListeningUDP         []int          `yaml:"listeningudp"`         // UDP ports in use
	OSImage              string         `yaml:"osimage"`              // Operating system information
	ResourceUsage        *ProcessUsage  `yaml:"resourceusage"`        // Resource usage information
}

// FlagArgument represents a cmdline flag and its associated value.
type FlagArgument struct {
	Flag  string // e.g., "-g"
	Value string // e.g., "daemon on;"
}

// ResourceUsageInfo holds resource usage information for a process.
type ProcessUsage struct {
	CPUCores    float64 // Fraction of CPU cores used
	MemoryMB    float64 // Memory usage in MB
	DiskReadMB  float64 // Disk read in MB
	DiskWriteMB float64 // Disk write in MB
}

// GetProcessInfo retrieves key information about a process by its Process ID (PID)
func GetProcessInfo(processID int) *ProcessInfo {
	// Initialize ProcessInfo object
	info := &ProcessInfo{
		PID: processID,
	}

	// Get process information
	info.ChildPIDs = GetChildProcessIDs(processID)
	info.ExecutablePath = GetExecutablePath(processID)
	info.WorkingDirectory = GetWorkingDirectory(processID)
	info.EnvironmentVariables = GetEnvironmentVariables(processID)
	info.ProcessUser, info.ProcessGroup = GetProcessUserAndGroup(processID)
	info.OSImage = GetOSRelease()

	// Reconstruct command line
	rawCommandLineArguments := GetCommandLineArgs(processID)
	info.ReconstructedCommand, info.CommandLineArguments = ParseCommandLine(info.ExecutablePath, rawCommandLineArguments)

	// Get resource usage and network/socket details
	processIDs := append([]int{info.PID}, info.ChildPIDs...)
	info.ResourceUsage = GetTotalResourceUsage(processIDs)
	inodeSet := GetProcessInodeSet(processIDs)
	info.UnixSockets = GetUnixDomainSockets(inodeSet)
	info.ListeningTCP = GetListeningTCPPorts(inodeSet)
	info.ListeningUDP = GetListeningUDPPorts(inodeSet)

	return info
}

// GetExecutablePath retrieves the path to the executable of the process
func GetExecutablePath(processID int) string {
	// Read the symbolic link to the executable from /proc/<PID>/exe
	executablePath := fmt.Sprintf("/proc/%d/exe", processID)
	resolvedPath, err := os.Readlink(executablePath)
	if err != nil {
		log.Error("Failed to read process executable path", "error", err)
	}
	return resolvedPath
}

// GetCommandLineArgs retrieves the command-line arguments of the process
func GetCommandLineArgs(processID int) []string {
	// Read the command-line arguments from /proc/<PID>/cmdline
	commandLinePath := fmt.Sprintf("/proc/%d/cmdline", processID)
	commandLineData, err := os.ReadFile(commandLinePath)
	if err != nil {
		log.Error("Failed to read process command-line arguments", "error", err)
	}
	// Replace null bytes with spaces and split the string into fields
	commandLineArgs := strings.Fields(strings.ReplaceAll(string(commandLineData), "\x00", " "))
	return commandLineArgs
}

// GetWorkingDirectory retrieves the working directory of the process
func GetWorkingDirectory(processID int) string {
	// Read the symbolic link to the working directory from /proc/<PID>/cwd
	workingDirectoryPath := fmt.Sprintf("/proc/%d/cwd", processID)
	workingDirectory, err := os.Readlink(workingDirectoryPath)
	if err != nil {
		log.Error("Failed to read process working directory", "error", err)
	}
	return workingDirectory
}

// GetEnvironmentVariables retrieves and parses the environment variables of the process
func GetEnvironmentVariables(processID int) []string {
	// Read the environment variables from /proc/<PID>/environ
	environmentFilePath := fmt.Sprintf("/proc/%d/environ", processID)
	rawEnvironmentData, err := os.ReadFile(environmentFilePath)
	if err != nil {
		log.Error("Failed to read process environment variables", "error", err)
	}
	return parseEnvironmentVariables(rawEnvironmentData)
}

// GetProcessUserAndGroup retrieves the user and group associated with the process ID.
func GetProcessUserAndGroup(processID int) (string, string) {
	// Read the status file to get the UID of the process
	statusFilePath := fmt.Sprintf("/proc/%d/status", processID)
	rawStatusData, err := os.ReadFile(statusFilePath)
	if err != nil {
		log.Error("Failed to read process status file", "error", err)
	}

	var userID string
	statusLines := strings.Split(string(rawStatusData), "\n")
	for _, line := range statusLines {
		if strings.HasPrefix(line, "Uid:") {
			fields := strings.Fields(line)
			if len(fields) > 1 {
				userID = fields[1]
			}
			break
		}
	}

	userInfo, err := user.LookupId(userID)
	if err != nil {
		log.Error(fmt.Sprintf("Failed to look up user by UID (%s)", userID), "error", err)
	}

	groupInfo, err := user.LookupGroupId(userInfo.Gid)
	if err != nil {
		log.Error(fmt.Sprintf("Failed to look up group by GID (%s)", userInfo.Gid), "error", err)
	}

	return userInfo.Username, groupInfo.Name
}

// GetChildProcessIDs retrieves a list of child process IDs for a given parent process ID
func GetChildProcessIDs(parentPID int) []int {
	// Execute pgrep -P <parentPID>
	output, err := exec.Command("pgrep", "-P", strconv.Itoa(parentPID)).Output()
	if err != nil {
		log.Info("No child processes found for parent PID: " + strconv.Itoa(parentPID))
		return []int{} // Return an empty slice if there are no child processes
	}

	// Split the output into lines and parse each line into an integer
	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	var childProcessIDs []int
	for _, line := range lines {
		if line == "" {
			continue
		}
		pid, err := strconv.Atoi(line)
		if err != nil {
			log.Error("Failed to convert PID to integer: "+line, "error", err)
		}
		childProcessIDs = append(childProcessIDs, pid)
	}
	return childProcessIDs
}

// GetProcessIDbyExecutable retrieves the PID of a process by its executable path
func GetProcessIDbyExecutable(executablePath string) int {
	// Execute pgrep -f <executablePath>
	output, err := exec.Command("pgrep", "-f", executablePath).Output()
	if err != nil {
		log.Error("Failed to retrieve PID for executable: "+executablePath, "error", err)
	}

	pid, err := strconv.Atoi(strings.TrimSpace(string(output)))
	if err != nil {
		log.Error("Failed to convert PID to integer for executable: "+executablePath, "error", err)
	}

	return pid
}

func GetOSRelease() string {
	// Read the /etc/os-release file
	data, err := os.ReadFile("/etc/os-release")
	if err != nil {
		log.Error("Failed to read /etc/os-release", "error", err)
		return "ubuntu:latest" // Default fallback
	}

	// Parse the file to extract NAME and VERSION_ID
	var name, versionID string
	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "NAME=") {
			name = strings.Trim(strings.Split(line, "=")[1], "\"")
		}
		if strings.HasPrefix(line, "VERSION_ID=") {
			versionID = strings.Trim(strings.Split(line, "=")[1], "\"")
		}
	}

	// Default to ubuntu:latest if parsing fails
	if name == "" || versionID == "" {
		return "ubuntu:latest"
	}

	// Format the base image name
	return fmt.Sprintf("%s:%s", strings.ToLower(name), versionID)
}

// parseEnvironmentVariables parses environment variables from a null-byte separated string
func parseEnvironmentVariables(rawData []byte) []string {
	rawVariables := strings.Split(string(rawData), "\x00")
	var environmentVariables []string

	for _, variable := range rawVariables {
		trimmedVariable := strings.TrimSpace(variable)
		if trimmedVariable == "" {
			continue
		}
		if strings.Contains(trimmedVariable, "=") {
			environmentVariables = append(environmentVariables, trimmedVariable)
		}
	}
	return environmentVariables
}
