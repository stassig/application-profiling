package profiler

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/charmbracelet/log"
)

// ResourceUsageInfo holds approximate CPU and memory usage stats for a process.
type ResourceUsageInfo struct {
	CPUPercent float64 // Approximate % of CPU usage
	MemoryMB   float64 // Memory usage in MB
}

// GetTotalResourceUsage calculates the total CPU and memory usage for a parent process and its children.
func GetTotalResourceUsage(parentPID int, childPIDs []int) *ResourceUsageInfo {
	totalCPU := 0.0
	totalMemory := 0.0

	// Calculate usage for the parent process
	parentUsage := GetResourceUsageForPID(parentPID)
	if parentUsage != nil {
		totalCPU += parentUsage.CPUPercent
		totalMemory += parentUsage.MemoryMB
	}

	// Calculate usage for each child process
	for _, childPID := range childPIDs {
		childUsage := GetResourceUsageForPID(childPID)
		if childUsage != nil {
			totalCPU += childUsage.CPUPercent
			totalMemory += childUsage.MemoryMB
		}
	}

	return &ResourceUsageInfo{
		CPUPercent: totalCPU,
		MemoryMB:   totalMemory,
	}
}

// GetResourceUsageForPID retrieves approximate CPU usage and memory usage for the given PID.
func GetResourceUsageForPID(pid int) *ResourceUsageInfo {
	// 1) Read /proc/uptime to find total system uptime
	uptime, err := getSystemUptime()
	if err != nil {
		log.Error("Failed to read system uptime", "error", err)
	}

	// 2) Read /proc/<pid>/stat for process CPU times and start time
	procStat, err := os.ReadFile(fmt.Sprintf("/proc/%d/stat", pid))
	if err != nil {
		log.Errorf("Failed to read /proc/%d/stat", pid)
	}

	// 3) Parse CPU times from /proc/<pid>/stat
	// stat format reference: http://man7.org/linux/man-pages/man5/proc.5.html
	fields := strings.Fields(string(procStat))
	if len(fields) < 22 {
		log.Errorf("Unexpected format in /proc/%d/stat", pid)
	}

	// utime = fields[13], stime = fields[14], starttime = fields[21] (in clock ticks)
	utimeTicks, _ := strconv.ParseFloat(fields[13], 64)
	stimeTicks, _ := strconv.ParseFloat(fields[14], 64)
	startTimeTicks, _ := strconv.ParseFloat(fields[21], 64)

	// Convert from clock ticks to seconds. Typically sysconf(_SC_CLK_TCK) = 100 on most Linux systems,
	// but you may want to dynamically fetch the clock ticks (e.g. sysconf SC_CLK_TCK).
	clockTicks := float64(100)
	totalTimeSec := (utimeTicks + stimeTicks) / clockTicks

	// Process start time in seconds since boot
	startTimeSec := startTimeTicks / clockTicks

	// 4) Calculate approximate CPU usage as a percentage
	// Process uptime in seconds = system uptime - (start time of process in seconds)
	processUptime := uptime - startTimeSec
	var cpuPercent float64
	if processUptime > 0 {
		cpuPercent = (totalTimeSec / processUptime) * 100.0
	}

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
	pageSize := int64(4096) // Typically 4KB on x86_64, but can vary
	memoryBytes := rssPages * pageSize

	// Convert memory to MB
	memoryMB := float64(memoryBytes) / (1024 * 1024)

	usage := &ResourceUsageInfo{
		CPUPercent: cpuPercent,
		MemoryMB:   memoryMB,
	}

	return usage
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
