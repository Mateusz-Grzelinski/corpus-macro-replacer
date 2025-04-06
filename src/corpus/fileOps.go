package corpus

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
)

func FindFile(directory, filename string) (string, error) {
	var foundPath string
	err := filepath.Walk(directory, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && info.Name() == filename {
			foundPath = path
			return filepath.SkipDir // Stop searching once found
		}
		return nil
	})

	if err != nil {
		return "", err
	}

	if foundPath == "" {
		return "", fmt.Errorf("file '%s' not found in directory '%s'", filename, directory)
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
