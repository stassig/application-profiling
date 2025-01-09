package profiler

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/charmbracelet/log"
	"gopkg.in/yaml.v2"
)

// SaveAsYAML saves the ProcessInfo object to a YAML file
func (info *ProcessInfo) SaveAsYAML() {
	// Get the file path for the YAML file
	filePath := BuildFilePath(fmt.Sprintf("output/%d/profile", info.PID), "process_info.yaml")

	// Create or overwrite the specified file
	file, err := os.Create(filePath)
	if err != nil {
		log.Error("Failed to create YAML file", "filePath", filePath, "error", err)
	}
	defer file.Close()

	// Marshal the ProcessInfo object to YAML
	data, err := yaml.Marshal(info)
	if err != nil {
		log.Error("Failed to marshal ProcessInfo to YAML", "error", err)
	}

	// Write the YAML data to the file
	_, err = file.Write(data)
	if err != nil {
		log.Error("Failed to write YAML data to file", "filePath", filePath, "error", err)
	}
}

// LoadFromYAML loads process info from a YAML file.
func LoadFromYAML(path string) *ProcessInfo {
	data, err := os.ReadFile(path)
	if err != nil {
		log.Error("Failed to read file", "path", path, "error", err)
	}

	info := &ProcessInfo{}
	err = yaml.Unmarshal(data, info)
	if err != nil {
		log.Error("Failed to unmarshal YAML data", "error", err)
	}

	return info
}

// BuildFilePath constructs a full file path from a subdirectory and file name
func BuildFilePath(subDir, fileName string) string {
	currentDirectory, err := os.Getwd()
	if err != nil {
		log.Error("Failed to get current directory", "error", err)
	}

	// Build the full path to the desired subdirectory
	fullDir := filepath.Join(currentDirectory, subDir)

	// Ensure the directory exists
	err = os.MkdirAll(fullDir, os.ModePerm)
	if err != nil {
		log.Error("Failed to create tracing directory", "directory", fullDir, "error", err)
	}

	// Construct the full file path
	return filepath.Join(fullDir, fileName)
}
