package corpus

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

func FindFile(directory, filenameWithSeparators string) (string, error) {
	dir := filepath.Dir(filenameWithSeparators)
	basename := filepath.Base(filenameWithSeparators)
	searchDir := filepath.Join(directory, dir)
	var foundPath string
	err := filepath.Walk(searchDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && info.Name() == basename {
			foundPath = path
			return filepath.SkipDir // Stop searching once found
		}
		return nil
	})

	if err != nil {
		return "", err
	}

	if foundPath == "" {
		return "", fmt.Errorf("file '%s' not found in directory '%s'", filenameWithSeparators, directory)
	}

	return foundPath, nil
}

// CopyFile copies a file from src to dst.
func CopyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, in)
	if err != nil {
		return err
	}

	return out.Close()
}

func GetCleanOutputpath(outputDir string, inputFile string) string {
	relInputFile, _ := filepath.Rel(outputDir, inputFile)
	cleanedRelInputFile := removeRelativePrefix(relInputFile)
	outputFile := filepath.Join(outputDir, cleanedRelInputFile)
	return outputFile
}

func removeRelativePrefix(path string) string {
	for strings.HasPrefix(path, "..\\") {
		path = strings.TrimPrefix(path, "..\\")
	}
	return path
}
