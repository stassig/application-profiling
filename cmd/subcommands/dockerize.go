package subcommands

import (
	"flag"
	"log"
	"os"
	"path/filepath"

	"application_profiling/internal/docker"
	"application_profiling/internal/profiler"
)

func RunDockerize(args []string) {
	fs := flag.NewFlagSet("dockerize", flag.ExitOnError)
	processInfoPath := fs.String("process-info", "samples/process_info.yaml", "Path to YAML file containing process info")
	traceLogPath := fs.String("trace-log", "samples/strace.log", "Path to filtered trace log with file paths")
	outputDockerfile := fs.String("dockerfile", "bin/containerization/Dockerfile", "Output Dockerfile path")
	profileDir := fs.String("profile-dir", "bin/containerization/profile", "Directory for minimal filesystem")
	tarFile := fs.String("tar-file", "bin/containerization/profile.tar.gz", "Tar file to create from the profile directory")
	fs.Parse(args)

	// 1. Load process info
	info := profiler.LoadFromYAML(*processInfoPath)

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
