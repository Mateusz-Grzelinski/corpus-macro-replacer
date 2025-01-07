package main

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

const CMKLineSeparator = `,`

var SectionRegex = regexp.MustCompile(`\[((\w+?)\d*)\]`)

func appendM1Section(m *M1, currentSection string, currentSectionTextBuilder strings.Builder) {
	currentSectionText, _ := strings.CutSuffix(currentSectionTextBuilder.String(), CMKLineSeparator)
	switch strings.ToLower(currentSection) {
	case "":
	case "raster":
		m.Raster = append(m.Raster, GenericAttribute{DAT: currentSectionText})
	case "varijable":
		m.Varijable.DAT = currentSectionText
	case "formule":
		m.Formule = new(GenericAttribute)
		m.Formule.DAT = currentSectionText
	case "joint":
		m.Joint = new(GenericAttribute)
		m.Joint.DAT = currentSectionText
	case "grupa":
		m.Grupa = append(m.Grupa, GenericAttribute{DAT: currentSectionText})
	case "potrosni":
		m.Potrosni = append(m.Potrosni, GenericAttribute{DAT: currentSectionText})
	case "makro":
		embeddedMakro := EmbeddedMakro{DAT: currentSectionText, MAK: nil}
		for _, line := range strings.Split(currentSectionText, CMKLineSeparator) { // todo iterate
			line = decodeCMKLine(line)
			nameAndValue := strings.SplitN(line, "=", 2)
			if strings.ToLower(currentSection) == "makro" && strings.ToLower(nameAndValue[0]) == "name" {
				if len(nameAndValue) != 2 {
					log.Printf("Error, I do not know how to handle this case. Is there '=' in file name? %s", line)
				}
				embeddedMakro.EmbeddedMakroName = nameAndValue[1]
			}
		}
		m.Makro = append(m.Makro, embeddedMakro)
	default:
		log.Printf("ERROR: unknown section name: '%s', sectionText: %s\n", currentSection, currentSectionText)
	}

}

const M1InitialMacroMarker string = ""

// create makro struct from CMK file
// assert that makro has the same name as file
// if there is a makro inside makro, it is assumed that it is placed in the same directory as original filename
func LoadMakroFromCMKFile(makroFile string) (*M1, error) {
	makroFile, _ = filepath.Abs(makroFile)
	initialMakro, err := partialLoadMakroFromCMKFile(makroFile)
	if err != nil {
		return nil, err
	}

	unprocessedMakros := map[string]bool{}
	processedMacros := map[string]*M1{}
	for _, subMacro := range initialMakro.Makro {
		unprocessedMakros[subMacro.EmbeddedMakroName] = false
	}
	for {
		var macroToProcess string
		// pick next not processed macro
		for k, v := range unprocessedMakros {
			if !v {
				macroToProcess = k
				break
			}
		}
		unprocessedMakros[macroToProcess] = true
		// end condition
		if macroToProcess == "" {
			break
		}

		submacroName := filepath.Join(filepath.Dir(makroFile), macroToProcess+".CMK")
		macro, err := partialLoadMakroFromCMKFile(submacroName)
		if err != nil {
			return nil, err
		}
		processedMacros[macroToProcess] = macro

		for _, subMacro := range macro.Makro {
			_, ok := unprocessedMakros[subMacro.EmbeddedMakroName]
			if !ok {
				unprocessedMakros[subMacro.EmbeddedMakroName] = false
			}
		}
	}

	// make sure all macros are resolved. Follow topological sort
	resolveSubMakros := map[string]*M1{M1InitialMacroMarker: initialMakro}
	for name, macro := range processedMacros {
		resolveSubMakros[name] = macro
	}

	for continueLoop := true; continueLoop; {
		continueLoop = false
		for _, macro := range resolveSubMakros {
			for i, submacro := range macro.Makro {
				if submacro.EmbeddedMakroName != "" && submacro.MAK == nil {
					macro.Makro[i].MAK = resolveSubMakros[submacro.EmbeddedMakroName]
					continueLoop = true
				}
			}
		}
	}

	// sanity checks
	for _, macro := range resolveSubMakros {
		for _, submacro := range macro.Makro {
			if submacro.MAK == nil && submacro.EmbeddedMakroName != "" {
				log.Printf("ERROR: Makro '%s' not loaded properly! This might be a bug.", submacro.EmbeddedMakroName)
			}
		}
	}

	return initialMakro, nil
}

// same as MakroFromFile but might have unresolved data in m.makro
func partialLoadMakroFromCMKFile(makroFile string) (*M1, error) {
	log.Printf("Reading makro: '%s'", makroFile)
	file, err := os.Open(makroFile)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	m := new(M1)
	m.MakroName = strings.SplitN(filepath.Base(makroFile), ".", 2)[0] // this name might be wrong, it can be redefined in software
	// true makro name is in MakroCollection.dat (binary file)

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
				return nil, fmt.Errorf("%s was parsed badly, I was looking for section name.\n", text)
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

func decodeAllCMKLines(DAT string) []string {
	lines := []string{}
	for _, line := range strings.Split(DAT, CMKLineSeparator) {
		lines = append(lines, decodeCMKLine(line))
	}
	return lines
}
