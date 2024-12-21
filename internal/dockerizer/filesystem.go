package dockerizer

import (
	"archive/tar"
	"compress/gzip"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
)

var visitedFiles = make(map[string]bool)

// LoadFilePaths loads file paths from a trace log.
func LoadFilePaths(path string) ([]string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	lines := strings.Split(strings.TrimSpace(string(data)), "\n")
	var files []string
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		if strings.HasPrefix(line, "/") {
			files = append(files, line)
		}
	}
	return files, nil
}

// CopyFilesToProfile copies files to a profile directory.
func CopyFilesToProfile(files []string, profileDir string) error {
	for _, file := range files {
		if err := copyFileRecursively(file, profileDir); err != nil {
			log.Printf("Warning: Failed to copy %s: %v", file, err)
		}
	}
	return nil
}

func copyFileRecursively(src, profileDir string) error {
	if src == "" || visitedFiles[src] {
		return nil
	}
	visitedFiles[src] = true

	stat, err := os.Lstat(src)
	if err != nil {
		return err
	}

	dst := filepath.Join(profileDir, src)
	if stat.Mode()&os.ModeSymlink != 0 {
		linkTarget, err := os.Readlink(src)
		if err != nil {
			return err
		}
		if !filepath.IsAbs(linkTarget) {
			linkTarget = filepath.Join(filepath.Dir(src), linkTarget)
		}
		if err := copyFileRecursively(linkTarget, profileDir); err != nil {
			log.Printf("Warning: Failed to copy symlink target %s: %v", linkTarget, err)
		}
		if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
			return err
		}
		return os.Symlink(linkTarget, dst)
	}

	if stat.IsDir() {
		if err := os.MkdirAll(dst, stat.Mode()); err != nil {
			return err
		}
		entries, err := os.ReadDir(src)
		if err != nil {
			return err
		}
		for _, entry := range entries {
			entryPath := filepath.Join(src, entry.Name())
			if err := copyFileRecursively(entryPath, profileDir); err != nil {
				log.Printf("Warning: Failed to copy entry %s: %v", entryPath, err)
			}
		}
		return nil
	}

	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return err
	}

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

// CreateTarArchive creates a tar archive from a directory.
func CreateTarArchive(tarFile, srcDir string) error {
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
