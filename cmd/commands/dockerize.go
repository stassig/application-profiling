package commands

import (
	"flag"
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
func RunDockerize(args []string) {
	// Parse command-line arguments
	options := parseDockerizeArguments(args)

	// Execute the Dockerization process
	executeDockerization(options)
}

// parseDockerizeArguments parses command-line arguments for the Dockerize command
func parseDockerizeArguments(args []string) DockerizeOptions {
	// Create a new flag set for the Dockerize command
	flagSet := flag.NewFlagSet("dockerize", flag.ExitOnError)

	// Define command-line flags
	processInfoFile := flagSet.String("process-info", "samples/process_info.yaml", "Path to the YAML file containing process information")
	traceLogFile := flagSet.String("trace-log", "samples/strace.log", "Path to the filtered trace log containing file paths")
	dockerfilePath := flagSet.String("dockerfile", "bin/containerization/Dockerfile", "Path to save the generated Dockerfile")
	profileDirectory := flagSet.String("profile-dir", "bin/containerization/profile", "Directory to build the minimal filesystem")
	tarArchivePath := flagSet.String("tar-file", "bin/containerization/profile.tar.gz", "Path to save the tar archive of the profile directory")

	// Parse flags from the arguments
	flagSet.Parse(args)

	return DockerizeOptions{
		ProcessInfoFile:  *processInfoFile,
		TraceLogFile:     *traceLogFile,
		DockerfilePath:   *dockerfilePath,
		ProfileDirectory: *profileDirectory,
		TarArchivePath:   *tarArchivePath,
	}
}

// executeDockerization executes the Dockerization process
func executeDockerization(options DockerizeOptions) {
	// 1. Load process information
	processInfo := profiler.LoadFromYAML(options.ProcessInfoFile)

	// 2. Load file paths from trace log
	filePaths, err := dockerizer.LoadFilePaths(options.TraceLogFile)
	if err != nil {
		log.Fatalf("Failed to load file paths from trace log: %v", err)
	}

	// 3. Prepare the profile directory
	if err := os.RemoveAll(options.ProfileDirectory); err != nil {
		log.Fatalf("Failed to clean up profile directory: %v", err)
	}
	if err := dockerizer.CopyFilesToProfile(filePaths, options.ProfileDirectory); err != nil {
		log.Fatalf("Failed to copy files to profile directory: %v", err)
	}

	// 4. Create a tar archive of the profile directory
	if err := dockerizer.CreateTarArchive(options.TarArchivePath, options.ProfileDirectory); err != nil {
		log.Fatalf("Failed to create tar archive: %v", err)
	}

	// 5. Generate the Dockerfile
	if err := dockerizer.GenerateDockerfile(processInfo, options.DockerfilePath, filepath.Base(options.TarArchivePath)); err != nil {
		log.Fatalf("Failed to generate Dockerfile: %v", err)
	}

	log.Printf("Dockerization process completed successfully.")
}
