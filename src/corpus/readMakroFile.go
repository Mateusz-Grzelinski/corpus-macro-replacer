package corpus

import (
	"log"
	"path/filepath"
	"strings"
)

func ReadMakrosFromCMK(makroFiles []string, makroNamesOverrides []*string, makroRootPath *string, makroMapping MakroMappings) (map[string]*M1, error) {
	makrosToReplace := map[string]*M1{}
	for i, makroFile := range makroFiles {
		absPathMakroFile := strings.SplitN(filepath.Base(makroFile), ".", 2)[0] // this name might be wrong, it can be redefined in software
		_, exists := makrosToReplace[absPathMakroFile]
		if exists {
			log.Printf("Warning: Makro path seems to be duplicated: '%s' (all paths: %s)", absPathMakroFile, makroFiles)
		}
		// todo this call requires new flags to work properly
		makro, err := NewMakroFromCMKFile(nil, makroFile, makroRootPath, makroMapping)
		if err != nil {
			return nil, err
		}
		if len(makroNamesOverrides) > i && makroNamesOverrides[i] != nil {
			makrosToReplace[*makroNamesOverrides[i]] = makro
		} else {
			makrosToReplace[absPathMakroFile] = makro
		}
	}
	return makrosToReplace, nil
}
