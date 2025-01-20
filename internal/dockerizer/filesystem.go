package dockerizer

import (
	"io"
	"os"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/charmbracelet/log"
)

var visitedFiles = make(map[string]bool)

// LoadFilePaths loads file paths from a trace log.
func LoadFilePaths(traceLogPath string) ([]string, error) {
	// Read the trace log file
	traceLogData, err := os.ReadFile(traceLogPath)
	if err != nil {
		return nil, err
	}

	// Extract and return file paths directly
	filePathLines := strings.Split(strings.TrimSpace(string(traceLogData)), "\n")
	return filePathLines, nil
}

// CopyFilesToProfile copies a list of files (and directories) into the specified profile directory.
func CopyFilesToProfile(filePaths []string, profileDirectory string) error {
	for _, filePath := range filePaths {
		if err := copyFileRecursively(filePath, profileDirectory); err != nil {
			log.Warnf("Failed to copy %s: %v", filePath, err)
		}
	}
	return nil
}

// copyFileRecursively copies a single file or directory into the profile directory.
// Handles symlinks, directories, and regular files.
func copyFileRecursively(sourcePath, profileDirectory string) error {
	// Skip empty paths or already-visited files.
	if sourcePath == "" || visitedFiles[sourcePath] {
		return nil
	}
	visitedFiles[sourcePath] = true

	// Gather file metadata.
	sourceFileInfo, err := os.Lstat(sourcePath)
	if err != nil {
		return err
	}

	// Determine the destination path inside profileDirectory.
	destinationPath := filepath.Join(profileDirectory, sourcePath)

	// Handle symlinks.
	if sourceFileInfo.Mode()&os.ModeSymlink != 0 {
		return copySymlink(sourcePath, destinationPath, profileDirectory, sourceFileInfo)
	}

	// Handle directories.
	if sourceFileInfo.IsDir() {
		return copyDirectory(sourcePath, destinationPath, profileDirectory, sourceFileInfo)
	}

	// Handle regular files.
	return copyRegularFile(sourcePath, destinationPath, sourceFileInfo)
}

// copySymlink handles copying symlinks into the profile directory.
func copySymlink(sourcePath, destinationPath, profileDirectory string, sourceFileInfo os.FileInfo) error {
	// Read the symlink target.
	linkTarget, err := os.Readlink(sourcePath)
	if err != nil {
		return err
	}

	// Convert relative symlinks to absolute, if needed.
	if !filepath.IsAbs(linkTarget) {
		linkTarget = filepath.Join(filepath.Dir(sourcePath), linkTarget)
	}

	// Recursively copy the symlink target.
	if err := copyFileRecursively(linkTarget, profileDirectory); err != nil {
		log.Warnf("Failed to copy symlink target %s: %v", linkTarget, err)
	}

	// Ensure the parent folder exists before making a symlink.
	if err := os.MkdirAll(filepath.Dir(destinationPath), 0o755); err != nil {
		return err
	}

	// Create the symlink.
	if err := os.Symlink(linkTarget, destinationPath); err != nil {
		return err
	}

	// Preserve symlink ownership with Lchown.
	uid, gid, err := getUIDGIDFromFileInfo(sourceFileInfo)
	if err != nil {
		return err
	}
	if err := os.Lchown(destinationPath, uid, gid); err != nil {
		return err
	}

	return nil
}

// copyDirectory handles copying a directory and its contents recursively.
func copyDirectory(sourcePath, destinationPath, profileDirectory string, sourceFileInfo os.FileInfo) error {
	// Create the destination directory with the same permissions.
	if err := os.MkdirAll(destinationPath, sourceFileInfo.Mode()); err != nil {
		return err
	}

	// Preserve ownership (UID/GID).
	uid, gid, err := getUIDGIDFromFileInfo(sourceFileInfo)
	if err != nil {
		return err
	}
	if err := os.Chown(destinationPath, uid, gid); err != nil {
		return err
	}

	// Read directory entries and copy each recursively.
	directoryEntries, err := os.ReadDir(sourcePath)
	if err != nil {
		return err
	}

	for _, directoryEntry := range directoryEntries {
		entryPath := filepath.Join(sourcePath, directoryEntry.Name())
		if err := copyFileRecursively(entryPath, profileDirectory); err != nil {
			log.Warnf("Failed to copy entry %s: %v", entryPath, err)
		}
	}

	return nil
}

// copyRegularFile copies a regular file from sourcePath to destinationPath.
func copyRegularFile(sourcePath, destinationPath string, sourceFileInfo os.FileInfo) error {
	// Ensure the destination directory exists.
	if err := os.MkdirAll(filepath.Dir(destinationPath), 0o755); err != nil {
		return err
	}

	// Open source file for reading.
	sourceFile, err := os.Open(sourcePath)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	// Open destination file for writing.
	destinationFile, err := os.OpenFile(destinationPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, sourceFileInfo.Mode())
	if err != nil {
		return err
	}
	defer destinationFile.Close()

	// Preserve ownership (UID/GID).
	uid, gid, err := getUIDGIDFromFileInfo(sourceFileInfo)
	if err != nil {
		return err
	}
	// For regular files, use os.Chown.
	if err := os.Chown(destinationPath, uid, gid); err != nil {
		return err
	}

	// Copy the file contents.
	_, err = io.Copy(destinationFile, sourceFile)
	return err
}

// getUIDGIDFromFileInfo extracts the user and group ID from the given FileInfo.
func getUIDGIDFromFileInfo(fileInfo os.FileInfo) (int, int, error) {
	statT, ok := fileInfo.Sys().(*syscall.Stat_t)
	if !ok {
		// Not on a Unix-based system or something unexpected.
		return 0, 0, nil
	}
	return int(statT.Uid), int(statT.Gid), nil
}
