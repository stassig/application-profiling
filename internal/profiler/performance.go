package profiler

import (
	"fmt"
	"math"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"

	"github.com/charmbracelet/log"
)

// GetTotalResourceUsage aggregates resource usage for a parent process and its child processes.
func GetTotalResourceUsage(parentPID int, childPIDs []int) *ProcessUsage {
	totalUsage := &ProcessUsage{}

	// 1) Aggregate usage for parent process
	aggregateProcessUsage(totalUsage, parentPID)

	// 2) Aggregate usage for child processes
	for _, childPID := range childPIDs {
		aggregateProcessUsage(totalUsage, childPID)
	}

	// 3) Round all values to 2 decimal places
	totalUsage = roundProcessUsage(totalUsage)

	return totalUsage
}

// GetResourceUsageForPID calculates CPU and memory usage for a process.
func GetResourceUsageForPID(pid int) *ProcessUsage {
	// Get CPU usage as both percentage and cores
	cpuCoresUsed := calculateCPUUsage(pid)

	// Get memory usage in MB
	memoryMB := getMemoryUsage(pid)

	// Get disk I/O stats
	diskReadMB, diskWriteMB := getDiskIOStatsForPID(pid)

	// Return ProcessUsage struct
	return &ProcessUsage{
		CPUCores:    cpuCoresUsed,
		MemoryMB:    memoryMB,
		DiskReadMB:  diskReadMB,
		DiskWriteMB: diskWriteMB,
	}
}

// GetDiskIOStatsForPID retrieves disk I/O stats for the given PID.
func getDiskIOStatsForPID(pid int) (float64, float64) {
	// Read disk I/O stats from /proc/<pid>/io
	ioFilePath := fmt.Sprintf("/proc/%d/io", pid)
	data, err := os.ReadFile(ioFilePath)
	if err != nil {
		log.Warnf("Failed to read disk I/O stats for PID %d: %v", pid, err)
		return 0, 0
	}
	// Parse read_bytes and write_bytes
	var readBytes, writeBytes float64

	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "read_bytes:") {
			fields := strings.Fields(line)
			readBytes, _ = strconv.ParseFloat(fields[1], 64)
		}
		if strings.HasPrefix(line, "write_bytes:") {
			fields := strings.Fields(line)
			writeBytes, _ = strconv.ParseFloat(fields[1], 64)
		}
	}

	// Convert bytes to MB
	return readBytes / (1024 * 1024), writeBytes / (1024 * 1024)
}

// getMemoryUsage retrieves the memory usage in MB for a process.
func getMemoryUsage(pid int) float64 {
	data, err := os.ReadFile(fmt.Sprintf("/proc/%d/statm", pid))
	if err != nil {
		log.Errorf("Failed to read /proc/%d/statm: %v", pid, err)
		return 0
	}

	fields := strings.Fields(string(data))
	if len(fields) < 2 {
		log.Errorf("Unexpected format in /proc/%d/statm", pid)
		return 0
	}

	rssPages, _ := strconv.ParseInt(fields[1], 10, 64)
	return convertBytesToMB(float64(rssPages) * 4096)
}

// calculateCPUUsage calculates the CPU usage in cores for a process.
func calculateCPUUsage(pid int) float64 {
	// Get number of CPU cores
	numCores := float64(runtime.NumCPU())

	// Get system uptime
	systemUptimeSeconds := getSystemUptime()

	// Get clock ticks per second
	clockTicksPerSecond := getClockTicks()

	// Get process CPU stats
	userTimeTicks, systemTimeTicks, startTimeTicks := getProcessStatFields(pid)

	// Calculate total CPU time in seconds
	totalCPUTimeSeconds := (userTimeTicks + systemTimeTicks) / clockTicksPerSecond

	// Calculate process start time in seconds
	processStartTimeSeconds := startTimeTicks / clockTicksPerSecond

	// Calculate process uptime
	processUptimeSeconds := systemUptimeSeconds - processStartTimeSeconds

	// Calculate CPU cores used
	cpuCoresUsed := (totalCPUTimeSeconds / processUptimeSeconds) * numCores

	return cpuCoresUsed
}

// getClockTicks retrieves the SC_CLK_TCK value (clock ticks per second).
func getClockTicks() float64 {
	output, err := exec.Command("getconf", "CLK_TCK").Output()
	if err != nil {
		log.Error("Failed to retrieve clock ticks", "error", err)
		return 100 // Default value
	}
	clockTicks, err := strconv.ParseFloat(strings.TrimSpace(string(output)), 64)
	if err != nil {
		log.Error("Failed to parse clock ticks", "error", err)
		return 100 // Default value
	}
	return clockTicks
}

// getSystemUptime returns the system uptime in seconds from /proc/uptime
func getSystemUptime() float64 {
	data, err := os.ReadFile("/proc/uptime")
	if err != nil {
		log.Error("Failed to read /proc/uptime", "error", err)
		return 0
	}

	fields := strings.Fields(string(data))
	if len(fields) < 1 {
		log.Error("Unexpected format in /proc/uptime")
		return 0
	}

	uptime, _ := strconv.ParseFloat(fields[0], 64)
	return uptime
}

// getProcessStatFields retrieves user, system, and start time ticks for a process.
func getProcessStatFields(processID int) (float64, float64, float64) {
	// Parse CPU times from /proc/<pid>/stat
	data, err := os.ReadFile(fmt.Sprintf("/proc/%d/stat", processID))
	if err != nil {
		log.Errorf("Failed to read /proc/%d/stat: %v", processID, err)
		return 0, 0, 0
	}

	fields := strings.Fields(string(data))
	if len(fields) < 22 {
		log.Errorf("Unexpected format in /proc/%d/stat", processID)
		return 0, 0, 0
	}

	// Extract user time, system time, and start time ticks
	userTimeTicks, _ := strconv.ParseFloat(fields[13], 64)
	systemTimeTicks, _ := strconv.ParseFloat(fields[14], 64)
	startTimeTicks, _ := strconv.ParseFloat(fields[21], 64)

	return userTimeTicks, systemTimeTicks, startTimeTicks
}

// aggregateProcessUsage aggregates the resource usage of a single process into the provided total.
func aggregateProcessUsage(totalUsage *ProcessUsage, pid int) {
	processUsage := GetResourceUsageForPID(pid)

	if processUsage != nil {
		// Add CPU and memory stats
		totalUsage.CPUCores += processUsage.CPUCores
		totalUsage.MemoryMB += processUsage.MemoryMB
		// Add disk I/O stats
		totalUsage.DiskReadMB += processUsage.DiskReadMB
		totalUsage.DiskWriteMB += processUsage.DiskWriteMB
	}
}

// roundProcessUsage rounds all values in a ProcessUsage struct to two decimal places.
func roundProcessUsage(usage *ProcessUsage) *ProcessUsage {
	usage.CPUCores = roundToTwoDecimalPlaces(usage.CPUCores)
	usage.MemoryMB = roundToTwoDecimalPlaces(usage.MemoryMB)
	usage.DiskReadMB = roundToTwoDecimalPlaces(usage.DiskReadMB)
	usage.DiskWriteMB = roundToTwoDecimalPlaces(usage.DiskWriteMB)
	return usage
}

// convertBytesToMB converts bytes to megabytes.
func convertBytesToMB(bytes float64) float64 {
	return bytes / (1024 * 1024)
}

// roundToTwoDecimalPlaces rounds a float64 value to two decimal places.
func roundToTwoDecimalPlaces(value float64) float64 {
	return math.Round(value*100) / 100
}
