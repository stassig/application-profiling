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
func GetTotalResourceUsage(parentPID int, childPIDs []int) *ResourceUsageInfo {
	totalCPUPercent := 0.0
	totalMemoryMB := 0.0
	totalCPUCores := 0.0
	totalDiskReadMB := 0.0
	totalDiskWriteMB := 0.0

	// Aggregate usage for parent process
	parentUsage := GetResourceUsageForPID(parentPID)
	if parentUsage != nil {
		totalCPUPercent += parentUsage.CPUPercent
		totalMemoryMB += parentUsage.MemoryMB
		totalCPUCores += parentUsage.CPUCores
		readMB, writeMB := GetDiskIOStatsForPID(parentPID)
		totalDiskReadMB += readMB
		totalDiskWriteMB += writeMB
	}

	// Aggregate usage for each child process
	for _, childPID := range childPIDs {
		childUsage := GetResourceUsageForPID(childPID)
		if childUsage != nil {
			totalCPUPercent += childUsage.CPUPercent
			totalMemoryMB += childUsage.MemoryMB
			totalCPUCores += childUsage.CPUCores
			readMB, writeMB := GetDiskIOStatsForPID(childPID)
			totalDiskReadMB += readMB
			totalDiskWriteMB += writeMB
		}
	}

	// Round all values to 2 decimal places before returning
	return &ResourceUsageInfo{
		CPUPercent:  roundToTwoDecimalPlaces(totalCPUPercent),
		CPUCores:    roundToTwoDecimalPlaces(totalCPUCores),
		MemoryMB:    roundToTwoDecimalPlaces(totalMemoryMB),
		DiskReadMB:  roundToTwoDecimalPlaces(totalDiskReadMB),
		DiskWriteMB: roundToTwoDecimalPlaces(totalDiskWriteMB),
	}
}

// GetResourceUsageForPID retrieves approximate CPU usage and memory usage for the given PID.
func GetResourceUsageForPID(pid int) *ResourceUsageInfo {
	// 1) Read /proc/uptime to find total system uptime
	uptime, err := getSystemUptime()
	if err != nil {
		log.Error("Failed to read system uptime", "error", err)
	}

	// 2) Fetch clock ticks per second
	clockTicks, err := getClockTicks()
	if err != nil {
		log.Error("Failed to retrieve clock ticks per second (SC_CLK_TCK)", "error", err)
		clockTicks = 100 // Default fallback
	}

	// 3) Read /proc/<pid>/stat for process CPU times and start time
	procStat, err := os.ReadFile(fmt.Sprintf("/proc/%d/stat", pid))
	if err != nil {
		log.Errorf("Failed to read /proc/%d/stat", pid)
	}

	// Parse CPU times from /proc/<pid>/stat
	fields := strings.Fields(string(procStat))
	if len(fields) < 22 {
		log.Errorf("Unexpected format in /proc/%d/stat", pid)
	}

	// utime = fields[13], stime = fields[14], starttime = fields[21] (in clock ticks)
	utimeTicks, _ := strconv.ParseFloat(fields[13], 64)
	stimeTicks, _ := strconv.ParseFloat(fields[14], 64)
	startTimeTicks, _ := strconv.ParseFloat(fields[21], 64)

	// Convert from clock ticks to seconds
	totalTimeSec := (utimeTicks + stimeTicks) / clockTicks

	// Process start time in seconds since boot
	startTimeSec := startTimeTicks / clockTicks

	// 4) Calculate approximate CPU usage in cores
	processUptime := uptime - startTimeSec
	numCores := float64(runtime.NumCPU()) // Get number of CPU cores
	cpuCoresUsed := (totalTimeSec / processUptime) * numCores

	// 5) Read /proc/<pid>/statm for memory usage
	statmData, err := os.ReadFile(fmt.Sprintf("/proc/%d/statm", pid))
	if err != nil {
		log.Errorf("Failed to read /proc/%d/statm: %v", pid, err)
	}

	statmFields := strings.Fields(string(statmData))
	if len(statmFields) < 2 {
		log.Errorf("Unexpected format in /proc/%d/statm", pid)
	}

	// resident set size in pages = statmFields[1]
	rssPages, _ := strconv.ParseInt(statmFields[1], 10, 64)
	pageSize := int64(4096)
	memoryBytes := rssPages * pageSize

	// Convert memory to MB
	memoryMB := float64(memoryBytes) / (1024 * 1024)

	// Return resource usage information
	return &ResourceUsageInfo{
		CPUPercent: cpuCoresUsed / numCores * 100,
		CPUCores:   cpuCoresUsed,
		MemoryMB:   memoryMB,
	}
}

// GetDiskIOStatsForPID retrieves disk I/O stats for the given PID.
func GetDiskIOStatsForPID(pid int) (float64, float64) {
	ioFilePath := fmt.Sprintf("/proc/%d/io", pid)
	data, err := os.ReadFile(ioFilePath)
	if err != nil {
		log.Warnf("Failed to read disk I/O stats for PID %d: %v", pid, err)
		return 0, 0
	}

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

// getClockTicks retrieves the SC_CLK_TCK value (clock ticks per second).
func getClockTicks() (float64, error) {
	output, err := exec.Command("getconf", "CLK_TCK").Output()
	if err != nil {
		return 0, fmt.Errorf("failed to retrieve clock ticks: %w", err)
	}
	clockTicks, err := strconv.ParseFloat(strings.TrimSpace(string(output)), 64)
	if err != nil {
		return 0, fmt.Errorf("failed to parse clock ticks: %w", err)
	}
	return clockTicks, nil
}

// getSystemUptime returns the system uptime in seconds from /proc/uptime
func getSystemUptime() (float64, error) {
	data, err := os.ReadFile("/proc/uptime")
	if err != nil {
		return 0, err
	}
	fields := strings.Fields(string(data))
	if len(fields) < 1 {
		return 0, fmt.Errorf("unexpected format in /proc/uptime")
	}
	return strconv.ParseFloat(fields[0], 64)
}

func roundToTwoDecimalPlaces(value float64) float64 {
	return math.Round(value*100) / 100
}
