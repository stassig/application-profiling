package util

import (
	"application_profiling/internal/profiler"
	"os"

	"github.com/charmbracelet/log"
)

// LogProcessDetails logs key details about a process in one method.
func LogProcessDetails(processInfo *profiler.ProcessInfo) {
	// Initialize logger
	logger := log.New(os.Stderr)
	logger.SetLevel(log.DebugLevel)

	// Log process details
	logger.Debugf("Parent Process ID: %d", processInfo.PID)
	logger.Debugf("Child process IDs: %v", processInfo.ChildPIDs)
	logger.Debugf("Executable path: %s", processInfo.ExecutablePath)
	logger.Debugf("Command-line arguments: %s", processInfo.CommandLineArguments)
	logger.Debugf("Working directory: %s", processInfo.WorkingDirectory)
	logger.Debugf("Environment variables: %v", processInfo.EnvironmentVariables)
	logger.Debugf("Process user: %s", processInfo.ProcessUser)
	logger.Debugf("Process group: %s", processInfo.ProcessGroup)
	logger.Debugf("Reconstructed command: %s", processInfo.ReconstructedCommand)
	logger.Debugf("Sockets: %v", processInfo.UnixSockets)
	logger.Debugf("Listening TCP ports: %v", processInfo.ListeningTCP)
	logger.Debugf("Listening UDP ports: %v", processInfo.ListeningUDP)
	logger.Debugf("OS Version: %s", processInfo.OSImage)
	logger.Debugf("Memory usage: %.2f MB", processInfo.ResourceUsage.MemoryMB)
	logger.Debugf("CPU cores used: %.2f", processInfo.ResourceUsage.CPUCores)
	logger.Debugf("Disk Read: %.2f MB", processInfo.ResourceUsage.DiskReadMB)
	logger.Debugf("Disk Write: %.2f MB", processInfo.ResourceUsage.DiskWriteMB)
}
