package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Printf("Usage: %s <PID>\n", os.Args[0])
		os.Exit(1)
	}

	pid := os.Args[1]

	// Get the service name from the PID
	serviceName, err := getServiceName(pid)
	if err != nil {
		fmt.Printf("Error getting service name: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Service Name: %s\n", serviceName)


	// Restart the service
	fmt.Printf("Restarting service: %s\n", serviceName)
	if err := restartService(serviceName); err != nil {
		fmt.Printf("Error restarting service: %v\n", err)
		os.Exit(1)
	}
}

func getServiceName(pid string) (string, error) {
	data, err := ioutil.ReadFile(fmt.Sprintf("/proc/%s/cgroup", pid))
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

func restartService(serviceName string) error {
	cmd := exec.Command("systemctl", "restart", serviceName)
	return cmd.Run()
}

