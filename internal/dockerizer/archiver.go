package dockerizer

import (
	"archive/tar"
	"compress/gzip"
	"io"
	"os"
	"path/filepath"
)

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
