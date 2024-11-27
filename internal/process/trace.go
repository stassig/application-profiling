package process

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Printf("Usage: %s <PID>\n", os.Args[0])
		os.Exit(1)
	}

	oldPid := os.Args[1]

	// Get the service name from the PID
	serviceName, err := getServiceName(oldPid)
	if err != nil {
		fmt.Printf("Error getting service name: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Service Name: %s\n", serviceName)

	// Prepare to run fatrace and restart the service
	var wg sync.WaitGroup
	wg.Add(1)

	// Channel to capture the new PID after restart
	pidChan := make(chan string, 1)

	// Start fatrace in a goroutine
	go func() {
		defer wg.Done()
		err := runFatrace(pidChan)
		if err != nil {
			fmt.Printf("Error running fatrace: %v\n", err)
			os.Exit(1)
		}
	}()

	// Give fatrace a moment to start
	time.Sleep(1 * time.Second)

	// Restart the service and get the new PID
	fmt.Printf("Restarting service: %s\n", serviceName)
	newPid, err := restartServiceAndGetNewPid(serviceName)
	if err != nil {
		fmt.Printf("Error restarting service: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("New PID: %s\n", newPid)

	// Send the new PID to the fatrace goroutine for filtering
	pidChan <- newPid
	close(pidChan)

	// Wait for fatrace to finish
	wg.Wait()
}

func getServiceName(pid string) (string, error) {
	data, err := os.ReadFile(fmt.Sprintf("/proc/%s/cgroup", pid))
	if err != nil {
		return "", err
	}

	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		fields := strings.Split(line, ":")
		if len(fields) == 3 && strings.Contains(fields[2], ".service") {
			parts := strings.Split(fields[2], "/")
			for _, part := range parts {
				if strings.HasSuffix(part, ".service") {
					return part, nil
				}
			}
		}
	}
	return "", fmt.Errorf("service not found for PID %s", pid)
}

func restartServiceAndGetNewPid(serviceName string) (string, error) {
	// Restart the service
	cmd := exec.Command("systemctl", "restart", serviceName)
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("failed to restart service: %v", err)
	}

	// Get the new PID of the service
	output, err := exec.Command("systemctl", "show", serviceName, "-p", "MainPID", "--value").Output()
	if err != nil {
		return "", fmt.Errorf("failed to get new PID: %v", err)
	}
	newPid := strings.TrimSpace(string(output))
	if newPid == "0" || newPid == "" {
		return "", fmt.Errorf("could not retrieve new PID for service: %s", serviceName)
	}
	return newPid, nil
}

func runFatrace(pidChan chan string) error {
    cmd := exec.Command("fatrace")

    // Capture the output of fatrace
    stdout, err := cmd.StdoutPipe()
    if err != nil {
        return fmt.Errorf("failed to capture stdout: %v", err)
    }

    // Open the log file for writing
    logFile, err := os.Create("fatrace.log")
    if err != nil {
        return fmt.Errorf("failed to create log file: %v", err)
    }
    defer logFile.Close()

    // Start the command
    if err := cmd.Start(); err != nil {
        return fmt.Errorf("failed to start fatrace: %v", err)
    }

    // Wait for the new PID from the main goroutine
    newPid := <-pidChan
    fmt.Printf("Filtering fatrace output for PID: %s\n", newPid)

    // Prepare the format to match, e.g., "(<PID>):"
    filter := fmt.Sprintf("(%s):", newPid)

    // Read and log output line by line
    scanner := bufio.NewScanner(stdout)
    fmt.Println("Filtered fatrace output:")
    for scanner.Scan() {
        line := scanner.Text()

        // Write unfiltered line to the log file
        _, err := logFile.WriteString(line + "\n")
        if err != nil {
            return fmt.Errorf("failed to write to log file: %v", err)
        }

        // Print filtered lines to the console
        if strings.Contains(line, filter) {
            fmt.Println(line)
        }
    }

    // Wait for the command to finish
    if err := cmd.Wait(); err != nil {
        return fmt.Errorf("fatrace exited with error: %v", err)
    }

    return nil
}
