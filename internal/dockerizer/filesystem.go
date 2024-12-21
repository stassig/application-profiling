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
			log.Printf("Warning: Failed to copy %s: %v", filePath, err)
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
		return copySymlink(sourcePath, destinationPath, profileDirectory)
	}

	// Handle directories.
	if sourceFileInfo.IsDir() {
		return copyDirectory(sourcePath, destinationPath, profileDirectory)
	}

	// Handle regular files.
	return copyRegularFile(sourcePath, destinationPath, sourceFileInfo.Mode())
}

// copySymlink handles copying symlinks into the profile directory.
func copySymlink(sourcePath, destinationPath, profileDirectory string) error {
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
		log.Printf("Warning: Failed to copy symlink target %s: %v", linkTarget, err)
	}

	// Ensure the parent folder exists before making a symlink.
	if err := os.MkdirAll(filepath.Dir(destinationPath), 0o755); err != nil {
		return err
	}

	return os.Symlink(linkTarget, destinationPath)
}

// copyDirectory handles copying a directory and its contents recursively.
func copyDirectory(sourcePath, destinationPath, profileDirectory string) error {
	// Create the destination directory with the same permissions.
	if err := os.MkdirAll(destinationPath, 0o755); err != nil {
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
			log.Printf("Warning: Failed to copy entry %s: %v", entryPath, err)
		}
	}

	return nil
}

// copyRegularFile copies a regular file from sourcePath to destinationPath.
func copyRegularFile(sourcePath, destinationPath string, fileMode os.FileMode) error {
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
	destinationFile, err := os.OpenFile(destinationPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, fileMode)
	if err != nil {
		return err
	}
	defer destinationFile.Close()

	// Copy the file contents.
	_, err = io.Copy(destinationFile, sourceFile)
	return err
}

// CreateTarArchive creates a tar.gz archive from the specified source directory.
func CreateTarArchive(tarFilePath, sourceDirectory string) error {
	// Create the tar file on disk.
	tarFile, err := os.Create(tarFilePath)
	if err != nil {
		return err
	}
	defer tarFile.Close()

	// Wrap the file in a gzip writer and tar writer.
	gzipWriter := gzip.NewWriter(tarFile)
	defer gzipWriter.Close()

	tarWriter := tar.NewWriter(gzipWriter)
	defer tarWriter.Close()

	// Walk through the source directory and add files to the tar archive.
	return filepath.Walk(sourceDirectory, func(currentPath string, fileInfo os.FileInfo, walkError error) error {
		if walkError != nil {
			return walkError
		}

		return addToTarArchive(tarWriter, sourceDirectory, currentPath, fileInfo)
	})
}

// addToTarArchive adds a file or directory to the tar archive.
func addToTarArchive(tarWriter *tar.Writer, sourceDirectory, currentPath string, fileInfo os.FileInfo) error {
	// Compute the file’s path relative to sourceDirectory.
	relativePath, err := filepath.Rel(sourceDirectory, currentPath)
	if err != nil {
		return err
	}
	relativePath = filepath.ToSlash(relativePath)
	if relativePath == "." {
		return nil
	}

	// Handle symlinks.
	var linkTarget string
	if fileInfo.Mode()&os.ModeSymlink != 0 {
		linkTarget, err = os.Readlink(currentPath)
		if err != nil {
			return err
		}
	}

	// Create a tar header for the file or directory.
	tarHeader, err := tar.FileInfoHeader(fileInfo, linkTarget)
	if err != nil {
		return err
	}
	tarHeader.Name = relativePath // Ensure relative paths are used in the tar archive.

	if err := tarWriter.WriteHeader(tarHeader); err != nil {
		return err
	}

	// If it’s a regular file, write its contents to the tar archive.
	if fileInfo.Mode().IsRegular() {
		return writeFileToTar(tarWriter, currentPath)
	}

	return nil
}

// writeFileToTar writes the contents of a regular file to the tar archive.
func writeFileToTar(tarWriter *tar.Writer, filePath string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = io.Copy(tarWriter, file)
	return err
}
