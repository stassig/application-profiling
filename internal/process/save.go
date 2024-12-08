package process

import (
	"encoding/json"
	"fmt"
	"os"

	"application_profiling/internal/util/logger"

	"gopkg.in/yaml.v2"
)

// SaveAsYAML saves the ProcessInfo object to a YAML file
func (info *ProcessInfo) SaveAsYAML() {
	// Create the file path
	filePath := fmt.Sprintf("%d_process_info.yaml", info.PID)

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

// SaveAsJSON saves the ProcessInfo object to a JSON file
func (info *ProcessInfo) SaveAsJSON() {
	// Create the file path
	filePath := fmt.Sprintf("%d_process_info.json", info.PID)

	// Create or overwrite the specified file
	file, err := os.Create(filePath)
	logger.Error(err, fmt.Sprintf("Failed to create JSON file: %s", filePath))
	defer file.Close()

	// Marshal the ProcessInfo object to JSON
	data, err := json.MarshalIndent(info, "", "  ") // Pretty-print JSON
	logger.Error(err, "Failed to marshal ProcessInfo to JSON")

	// Write the JSON data to the file
	_, err = file.Write(data)
	logger.Error(err, fmt.Sprintf("Failed to write JSON data to file: %s", filePath))
}
