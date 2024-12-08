package main

import (
	"flag"
	"log"
	"os"
	"path/filepath"

	"application_profiling/internal/docker"
	"application_profiling/internal/process"
)

// main is the entry point of the application.
func main() {
	processInfoPath := flag.String("process-info", "process_info.yaml", "Path to YAML file containing process info")
	traceLogPath := flag.String("trace-log", "nginx_strace_log_filtered.log", "Path to filtered trace log with file paths")
	outputDockerfile := flag.String("dockerfile", "Dockerfile", "Output Dockerfile path")
	profileDir := flag.String("profile-dir", "./profile", "Directory for minimal filesystem")
	tarFile := flag.String("tar-file", "profile.tar.gz", "Tar file to create from the profile directory")
	flag.Parse()

	// 1. Load process info
	info := process.LoadFromYAML(*processInfoPath)

	// 2. Load file paths from trace log
	files, err := docker.LoadFilePaths(*traceLogPath)
	if err != nil {
		log.Fatalf("Failed to load file paths: %v", err)
	}

	// 3. Copy files to the profile directory
	if err := os.RemoveAll(*profileDir); err != nil {
		log.Fatalf("Failed to clean profile directory: %v", err)
	}
	if err := docker.CopyFilesToProfile(files, *profileDir); err != nil {
		log.Fatalf("Failed to copy files to profile: %v", err)
	}

	// 4. Create tar archive from the profile directory
	if err := docker.CreateTarArchive(*tarFile, *profileDir); err != nil {
		log.Fatalf("Failed to create tar archive: %v", err)
	}

	// 5. Generate Dockerfile
	if err := docker.GenerateDockerfile(info, *outputDockerfile, filepath.Base(*tarFile)); err != nil {
		log.Fatalf("Failed to generate Dockerfile: %v", err)
	}

	log.Println("Done.")
}
