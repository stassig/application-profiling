package profiler

import (
	"fmt"
	"os"
	"path/filepath"

	"application_profiling/internal/util/logger"

	"gopkg.in/yaml.v2"
)

// SaveAsYAML saves the ProcessInfo object to a YAML file
func (info *ProcessInfo) SaveAsYAML() {
	// Get the current working directory
	currentDirectory, err := os.Getwd()
	logger.Error(err, "Failed to get current directory")

	// Create the target directory path
	targetDir := filepath.Join(currentDirectory, "bin", "process_info")

	// Ensure the directory exists
	err = os.MkdirAll(targetDir, os.ModePerm)
	logger.Error(err, fmt.Sprintf("Failed to create directory: %s", targetDir))

	// Create the file path
	filePath := filepath.Join(targetDir, fmt.Sprintf("%d_process_info.yaml", info.PID))

	// Create or overwrite the specified file
	file, err := os.Create(filePath)
	logger.Error(err, fmt.Sprintf("Failed to create YAML file: %s", filePath))
	defer file.Close()

	// Marshal the ProcessInfo object to YAML
	data, err := yaml.Marshal(info)
	logger.Error(err, "Failed to marshal ProcessInfo to YAML")

	// Write the YAML data to the file
	_, err = file.Write(data)
	logger.Error(err, fmt.Sprintf("Failed to write YAML data to file: %s", filePath))
}

// LoadFromYAML loads process info from a YAML file.
func LoadFromYAML(path string) *ProcessInfo {
	data, err := os.ReadFile(path)
	logger.Error(err, fmt.Sprintf("Failed to read file: %s", path))

	info := &ProcessInfo{}
	err = yaml.Unmarshal(data, info)
	logger.Error(err, "Failed to unmarshal YAML data")

	return info
}
