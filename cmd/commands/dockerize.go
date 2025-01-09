package commands

import (
	"fmt"
	"os"
	"path/filepath"

	"application_profiling/internal/dockerizer"
	"application_profiling/internal/profiler"

	"github.com/charmbracelet/log"
)

// DockerizeOptions represents the options for the Dockerize command
type DockerizeOptions struct {
	ProcessInfoFile  string
	TraceLogFile     string
	DockerfilePath   string
	ProfileDirectory string
	TarArchivePath   string
}

// RunDockerize handles the "dockerize" command logic
func RunDockerize(arguments []string) {
	// Parse command-line arguments
	options := parseDockerizeArguments(arguments[0])

	// Execute the Dockerization process
	executeDockerization(options)
}

// parseDockerizeArguments generates DockerizeOptions using the provided PID
func parseDockerizeArguments(pid string) DockerizeOptions {
	// Define file paths
	processInfoFile := fmt.Sprintf("output/%s/profile/process_info.yaml", pid)
	traceLogFile := fmt.Sprintf("output/%s/profile/strace_merged.log", pid)
	dockerfilePath := fmt.Sprintf("output/%s/dockerize/Dockerfile", pid)
	profileDirectory := fmt.Sprintf("output/%s/dockerize/profile", pid)
	tarArchivePath := fmt.Sprintf("output/%s/dockerize/profile.tar.gz", pid)

	return DockerizeOptions{
		ProcessInfoFile:  processInfoFile,
		TraceLogFile:     traceLogFile,
		DockerfilePath:   dockerfilePath,
		ProfileDirectory: profileDirectory,
		TarArchivePath:   tarArchivePath,
	}
}

// executeDockerization executes the Dockerization process
func executeDockerization(options DockerizeOptions) {
	// 1. Load process information
	log.Info("Loading static process information...")
	processInfo := profiler.LoadFromYAML(options.ProcessInfoFile)

	// 2. Load file paths from trace log
	log.Info("Loading runtime data from trace log...")
	filePaths, err := dockerizer.LoadFilePaths(options.TraceLogFile)
	if err != nil {
		log.Fatalf("Failed to load file paths from trace log: %v", err)
	}

	// 3. Prepare the profile directory
	log.Info("Copying files to minimal profile filesystem...")
	if err := os.RemoveAll(options.ProfileDirectory); err != nil {
		log.Fatalf("Failed to clean up profile directory: %v", err)
	}
	if err := dockerizer.CopyFilesToProfile(filePaths, options.ProfileDirectory); err != nil {
		log.Fatalf("Failed to copy files to profile directory: %v", err)
	}

	// 4. Create a tar archive of the profile directory
	log.Info("Creating tar archive of profile directory...")
	if err := dockerizer.CreateTarArchive(options.TarArchivePath, options.ProfileDirectory); err != nil {
		log.Fatalf("Failed to create tar archive: %v", err)
	}

	// 5. Generate the Dockerfile
	log.Info("Generating Dockerfile...")
	if err := dockerizer.GenerateDockerfile(processInfo, options.DockerfilePath, filepath.Base(options.TarArchivePath), filepath.Base(options.ProfileDirectory)); err != nil {
		log.Fatalf("Failed to generate Dockerfile: %v", err)
	}

	log.Info("Dockerization complete.")
}
