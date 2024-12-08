package main

import (
	"archive/tar"
	"compress/gzip"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v2"
)

// ProcessInfo is an example structure similar to what you have.
// Adjust fields and tags as needed.
type ProcessInfo struct {
	PID                  int      `yaml:"pid" json:"pid"`
	ExecutablePath       string   `yaml:"executablepath" json:"executablepath"`
	CommandLineArgs      []string `yaml:"commandlineargs" json:"commandlineargs"`
	WorkingDirectory     string   `yaml:"workingdirectory" json:"workingdirectory"`
	EnvironmentVariables []string `yaml:"environmentvariables" json:"environmentvariables"`
	ProcessOwner         string   `yaml:"processowner" json:"processowner"`
	ReconstructedCommand string   `yaml:"reconstructedcommand" json:"reconstructedcommand"`
	UnixSockets          []string `yaml:"unixsockets" json:"unixsockets"`
	ListeningTCP         []int    `yaml:"listeningtcp" json:"listeningtcp"`
	ListeningUDP         []int    `yaml:"listeningudp" json:"listeningudp"`
}

func main() {
	processInfoPath := flag.String("process-info", "99261_process_info.yaml", "Path to YAML/JSON file containing process info")
	traceLogPath := flag.String("trace-log", "nginx_strace_log_filtered.log", "Path to filtered trace log with file paths")
	outputDockerfile := flag.String("dockerfile", "Dockerfile", "Output Dockerfile path")
	profileDir := flag.String("profile-dir", "./profile", "Directory for minimal filesystem")
	tarFile := flag.String("tar-file", "profile.tar.gz", "Tar file to create from the profile directory")
	flag.Parse()

	// 1. Load process info
	info, err := loadProcessInfo(*processInfoPath)
	if err != nil {
		log.Fatalf("Failed to load process info: %v", err)
	}

	// 2. Load file paths from trace log
	files, err := loadFilePaths(*traceLogPath)
	if err != nil {
		log.Fatalf("Failed to load file paths: %v", err)
	}

	// 3. Copy files to ./profile directory
	if err := os.RemoveAll(*profileDir); err != nil {
		log.Fatalf("Failed to clean profile directory: %v", err)
	}
	if err := copyFilesToProfile(files, *profileDir); err != nil {
		log.Fatalf("Failed to copy files to profile: %v", err)
	}

	// 4. Create tar archive from the profile directory
	if err := createTarArchive(*tarFile, *profileDir); err != nil {
		log.Fatalf("Failed to create tar archive: %v", err)
	}

	// 5. Generate Dockerfile
	if err := generateDockerfile(info, *outputDockerfile, filepath.Base(*tarFile)); err != nil {
		log.Fatalf("Failed to generate Dockerfile: %v", err)
	}

	log.Println("Done.")
}

func loadProcessInfo(path string) (*ProcessInfo, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	info := &ProcessInfo{}
	// Try YAML first
	if yaml.Unmarshal(data, info) == nil && info.PID != 0 {
		return info, nil
	}

	// Try JSON if YAML didn't yield a valid PID
	if json.Unmarshal(data, info) == nil && info.PID != 0 {
		return info, nil
	}

	return nil, errors.New("failed to unmarshal process info (check PID field)")
}

func loadFilePaths(path string) ([]string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	lines := strings.Split(strings.TrimSpace(string(data)), "\n")
	var files []string
	for _, l := range lines {
		l = strings.TrimSpace(l)
		if l == "" || strings.HasPrefix(l, "#") {
			continue
		}
		files = append(files, l)
	}
	return files, nil
}

func copyFilesToProfile(files []string, profileDir string) error {
	for _, f := range files {
		// Ensure leading slash
		if !strings.HasPrefix(f, "/") {
			continue
		}
		dst := filepath.Join(profileDir, f)
		if err := copyFileWithSymlinksPreserved(f, dst); err != nil {
			log.Printf("Warning: Failed to copy %s: %v", f, err)
		}
	}
	return nil
}

func copyFileWithSymlinksPreserved(src, dst string) error {
	stat, err := os.Lstat(src)
	if err != nil {
		return err
	}

	// Create parent dirs
	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return err
	}

	if stat.Mode()&os.ModeSymlink != 0 {
		// It's a symlink
		linkTarget, err := os.Readlink(src)
		if err != nil {
			return err
		}

		// Ensure that the link target is copied as well if it's an absolute path
		if filepath.IsAbs(linkTarget) {
			absTarget := linkTarget
			targetDst := filepath.Join(filepath.Dir(dst), linkTarget)
			if err := copyFileWithSymlinksPreserved(absTarget, targetDst); err != nil {
				log.Printf("Warning: Failed to copy symlink target %s: %v", absTarget, err)
			}
		}

		// Create the symlink in dst
		return os.Symlink(linkTarget, dst)
	}

	// If it's a dir
	if stat.IsDir() {
		return os.MkdirAll(dst, stat.Mode())
	}

	// Otherwise it's a regular file or something else
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.OpenFile(dst, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, stat.Mode())
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, in)
	return err
}

func createTarArchive(tarFile, srcDir string) error {
	f, err := os.Create(tarFile)
	if err != nil {
		return err
	}
	defer f.Close()

	gw := gzip.NewWriter(f)
	defer gw.Close()

	tw := tar.NewWriter(gw)
	defer tw.Close()

	return filepath.Walk(srcDir, func(path string, info os.FileInfo, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}

		relPath, err := filepath.Rel(srcDir, path)
		if err != nil {
			return err
		}
		relPath = filepath.ToSlash(relPath)
		if relPath == "." {
			return nil
		}

		link := ""
		if info.Mode()&os.ModeSymlink != 0 {
			link, err = os.Readlink(path)
			if err != nil {
				return err
			}
		}

		hdr, err := tar.FileInfoHeader(info, link)
		if err != nil {
			return err
		}
		hdr.Name = relPath

		if err := tw.WriteHeader(hdr); err != nil {
			return err
		}

		if info.Mode().IsRegular() {
			f, err := os.Open(path)
			if err != nil {
				return err
			}
			defer f.Close()
			if _, err := io.Copy(tw, f); err != nil {
				return err
			}
		}
		return nil
	})
}

func generateDockerfile(info *ProcessInfo, dockerfilePath, tarFile string) error {
	dockerfile := &strings.Builder{}
	fmt.Fprintln(dockerfile, "FROM ubuntu:latest")
	fmt.Fprintf(dockerfile, "COPY %s /\n", tarFile)
	fmt.Fprintf(dockerfile, "RUN tar --skip-old-files -xvf /%s -C / && rm /%s\n", tarFile, tarFile)

	// Set env vars
	for _, env := range info.EnvironmentVariables {
		// env assumed in form KEY=VAL
		if strings.Contains(env, "=") {
			fmt.Fprintf(dockerfile, "ENV %s\n", env)
		}
	}

	// If process owner is given, we assume it as "user:group". If group not known, same as user.
	userSpec := info.ProcessOwner
	if userSpec == "" {
		userSpec = "root:root"
	} else if !strings.Contains(userSpec, ":") {
		// If we only have a username, use that as group as well
		userSpec = userSpec + ":" + userSpec
	}
	fmt.Fprintf(dockerfile, "USER %s\n", userSpec)

	// Set working directory
	if info.WorkingDirectory != "" {
		fmt.Fprintf(dockerfile, "WORKDIR %s\n", info.WorkingDirectory)
	}

	// Expose ports derived from ListeningTCP or ListeningUDP
	for _, port := range info.ListeningTCP {
		fmt.Fprintf(dockerfile, "EXPOSE %d/tcp\n", port)
	}
	for _, port := range info.ListeningUDP {
		fmt.Fprintf(dockerfile, "EXPOSE %d/udp\n", port)
	}

	// Command
	cmdArgs := strings.Fields(info.ReconstructedCommand)
	if len(cmdArgs) > 0 {
		quoted := make([]string, 0, len(cmdArgs))
		for _, c := range cmdArgs {
			quoted = append(quoted, fmt.Sprintf("\"%s\"", c))
		}
		fmt.Fprintf(dockerfile, "CMD [%s]\n", strings.Join(quoted, ", "))
	}

	return os.WriteFile(dockerfilePath, []byte(dockerfile.String()), 0644)
}
