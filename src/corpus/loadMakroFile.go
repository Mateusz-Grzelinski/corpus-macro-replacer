package corpus

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"golang.org/x/text/encoding/charmap"
	"golang.org/x/text/transform"
)

type CMKUnknownMakroError struct {
	Name string
}

func (e *CMKUnknownMakroError) Error() string {
	return fmt.Sprintf("can not find makro \"%s\"   ", e.Name)
}

const CMKLineSeparator = `,`

var SectionRegex = regexp.MustCompile(`\[((\w+?)\d*)\]`)

func appendM1Section(m *M1, currentSection string, currentSectionTextBuilder strings.Builder) {
	currentSectionText, _ := strings.CutSuffix(currentSectionTextBuilder.String(), CMKLineSeparator)
	switch strings.ToLower(currentSection) {
	case "":
	case "raster":
		m.Raster = append(m.Raster, GenericNodeWithDat{DAT: currentSectionText})
	case "varijable":
		m.Varijable.DAT = currentSectionText
	case "formule":
		m.Formule = new(GenericNodeWithDat)
		m.Formule.DAT = currentSectionText
	case "joint":
		m.Joint = new(GenericNodeWithDat)
		m.Joint.DAT = currentSectionText
	case "grupa":
		m.Grupa = append(m.Grupa, GenericNodeWithDat{DAT: currentSectionText})
	case "potrosni":
		m.Potrosni = append(m.Potrosni, GenericNodeWithDat{DAT: currentSectionText})
	case "pocket":
		m.Pocket = append(m.Pocket, GenericNodeWithDat{DAT: currentSectionText})
	case "pila":
		m.Pila = append(m.Pila, GenericNodeWithDat{DAT: currentSectionText})
	case "makro":
		embeddedMakro := M1EmbeddedMakro{GenericNodeWithDat: GenericNodeWithDat{DAT: currentSectionText}, MAK: nil}
		embeddedMakro.EmbeddedMakroName = embeddedMakro.CalledWith()
		m.Makro = append(m.Makro, embeddedMakro)
	default:
		log.Printf("ERROR: unknown section name: '%s', sectionText: %s\n", currentSection, currentSectionText)
	}
}

// dict["<makro name>"] = "<path to file>"
type MakroMappings map[string]string

const InitialMacroKey string = ""

// create makro struct from CMK file
// assert that makro has the same name as file
// if there is a makro inside makro, it is assumed that it is placed in the same directory as original filename
// use MakroMappings to resolve names in [MAKRO] section. Usually contains paths relative to makroRootPath
// makroRootPath + "<makro name>" is used as fallback path when makro name is not in makroMapping. When nil makroFile base path is treated as MakroRootPath, what might be false!
// usually makroRootPath="C:\Tri D Corpus\Corpus 5.0\Makro\"
func NewMakroFromCMKFile(makroName *string, makroFile string, makroRootPath *string, makroNameToPathRelative MakroMappings) (*M1, error) {
	if makroFile == "" {
		return nil, fmt.Errorf("missing input makro file")
	}
	if makroRootPath == nil {
		tmp := filepath.Dir(makroFile)
		makroRootPath = &tmp
	}
	makroFile, _ = filepath.Abs(makroFile)
	if makroName == nil {
		tmp := GetMacroNameByFileName(makroFile, makroFile, &MakroCollectionCache) // ugh refering to global var
		makroName = &tmp
	}
	initialMakro, err := partialNewMakroFromCMKFile(*makroName, makroFile)
	if err != nil {
		return nil, err
	}

	allMakros := map[string]bool{}
	for _, subMacro := range initialMakro.Makro {
		allMakros[subMacro.EmbeddedMakroName] = false
	}
	processedMakros := map[string]*M1{}
	for {
		var makroToProcessName *string
		// pick next not processed macro
		for name, isProcessed := range allMakros {
			if !isProcessed {
				makroToProcessName = &name
				break
			}
		}
		// end condition
		if makroToProcessName == nil {
			break
		}
		if *makroToProcessName == "" {
			return nil, fmt.Errorf("[MAKRO] specifies submakro with empty name")
		}
		allMakros[*makroToProcessName] = true

		submacroPathAbs, found := makroNameToPathRelative[*makroToProcessName]
		if found {
			submacroPathAbs = filepath.Join(*makroRootPath, submacroPathAbs)
		} else {
			// best effort search for file in makroRootPath
			submakroFoundPath, found := FindFile(*makroRootPath, *makroToProcessName+".CMK")
			submacroPathAbs = submakroFoundPath
			if found != nil {
				return nil, &CMKUnknownMakroError{Name: *makroToProcessName}
			} else {
				log.Printf("Warning: makro \"%s\" was found by searching \"%s\": \"%s\"", *makroToProcessName, *makroRootPath, submakroFoundPath)
			}
		}
		makro, err := partialNewMakroFromCMKFile(*makroToProcessName, submacroPathAbs)
		if err != nil {
			return nil, err
		}
		processedMakros[*makroToProcessName] = makro

		for _, subMakro := range makro.Makro {
			_, found := allMakros[subMakro.EmbeddedMakroName]
			// avoid reading again makro with same name (infinite recursion)
			if !found {
				allMakros[subMakro.EmbeddedMakroName] = false
			}
		}
	}

	// make sure all macros are resolved
	// can contain multiple pointers to same macro
	resolveSubMakros := map[string]*M1{InitialMacroKey: initialMakro}
	for name, makro := range processedMakros {
		resolveSubMakros[name] = makro
	}
	for continueLoop := true; continueLoop; {
		continueLoop = false
		for _, macro := range resolveSubMakros {
			for i, submacro := range macro.Makro {
				if submacro.EmbeddedMakroName != "" && submacro.MAK == nil {
					// make a copy just to be sure, probably can be optimized
					var macroCopy = *resolveSubMakros[submacro.EmbeddedMakroName]
					// embedded makro should get variables from parent, leaving any data here causes bugs
					macroCopy.Varijable.DAT = ""
					macro.Makro[i].MAK = &macroCopy
					continueLoop = true
				}
			}
		}
	}

	// sanity checks
	for _, macro := range resolveSubMakros {
		for _, submacro := range macro.Makro {
			if submacro.MAK == nil && submacro.EmbeddedMakroName != "" {
				return nil, fmt.Errorf("ERROR: Makro '%s' not loaded properly! This might be a bug", submacro.EmbeddedMakroName)
			}
			if submacro.MAK.Varijable.DAT != "" {
				return nil, fmt.Errorf("ERROR: Makro '%s' not loaded properly! This might be a bug", submacro.EmbeddedMakroName)
			}
		}
	}

	return initialMakro, nil
}

// same as MakroFromFile but might have unresolved data in m.makro
// makroMapping can come from Corpus settings (MakroCollection.dat)
func partialNewMakroFromCMKFile(makroName string, makroFile string) (*M1, error) {
	log.Printf("Reading makro: '%s'", makroFile)
	file, err := os.Open(makroFile)
	if err != nil {
		return nil, err
	}
	stat, err := file.Stat()
	if err != nil {
		return nil, err
	}
	if !stat.Mode().IsRegular() {
		abs, err := filepath.Abs(makroFile)
		if err == nil {
			return nil, fmt.Errorf("path is not a file %s (%s)", abs, makroFile)
		} else {
			return nil, fmt.Errorf("path is not a file %s", makroFile)
		}
	}
	defer file.Close()

	m := new(M1)
	m.MakroName = makroName

	dec := transform.NewReader(file, charmap.Windows1250.NewDecoder())
	scanner := bufio.NewScanner(dec)
	scanner.Split(bufio.ScanLines)
	var currentSection string
	var allSections map[string]int = make(map[string]int)
	var currentSectionText strings.Builder

	for scanner.Scan() {
		text := scanner.Text()
		if text == "" {
			continue
		}
		if strings.HasPrefix(text, "[") {
			matched := SectionRegex.FindStringSubmatch(text)
			if matched == nil {
				return nil, fmt.Errorf("%s was parsed badly, was looking start of section name, got nil", text)
			}
			fullSectionName, sectionName := matched[1], matched[2]
			allSections[fullSectionName]++
			appendM1Section(m, currentSection, currentSectionText)
			currentSectionText.Reset()
			currentSection = sectionName
		} else {
			currentSectionText.WriteString(encodeCMKLine(text))
		}
	}
	if currentSectionText.Len() != 0 {
		appendM1Section(m, currentSection, currentSectionText)
	}
	// m.Varijable = append(m.MSFO)
	return m, nil
}

func encodeCMKLine(text string) string {
	if strings.Contains(text, " ") || strings.Contains(text, "\t") {
		return `"` + text + `"` + CMKLineSeparator
	}
	return text + CMKLineSeparator
}

func decodeCMKLine(line string) string {
	lineTrimmed := strings.TrimSpace(line)
	lineTrimmed, _ = strings.CutPrefix(lineTrimmed, `"`)
	lineTrimmed, _ = strings.CutSuffix(lineTrimmed, `"`)
	return lineTrimmed
}

func DecodeAllCMKLines(DAT string) []string {
	lines := []string{}
	for _, line := range strings.Split(DAT, CMKLineSeparator) {
		lines = append(lines, decodeCMKLine(line))
	}
	return lines
}
