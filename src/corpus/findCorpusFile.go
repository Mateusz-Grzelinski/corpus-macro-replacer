package corpus

import (
	"io/fs"
	"path/filepath"
	"slices"
	"strings"
)

var corpustExtensions = []string{".s3d", ".e3d"} // use with strings.Fold

func IsCorpusExtension(path string) bool {
	// todo use strings.EqualFold()?
	ext := strings.ToLower(filepath.Ext(path))
	return slices.Contains(corpustExtensions, ext)
}

func FindCorpusFiles(inputFolder string) []string {
	foundCorpusFiles := []string{}
	filepath.Walk(inputFolder, func(path string, info fs.FileInfo, err error) error {
		if info != nil && !info.IsDir() && IsCorpusExtension(info.Name()) {
			foundCorpusFiles = append(foundCorpusFiles, path)
		}
		return nil
	})
	return foundCorpusFiles
}
