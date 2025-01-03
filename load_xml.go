package main

import (
	"bufio"
	"encoding/xml"
	"fmt"
	"log"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"

	"golang.org/x/text/encoding/charmap"
	"golang.org/x/text/transform"
)

type ConvertLocalVariablesToGlobal string

const (
	Always ConvertLocalVariablesToGlobal = "always"
	OnlyIfValueIsTheSame
	/*
		{
			[VARIJABLE] // old
			grubosc=32

			[VARIJABLE] // new
			_grubosc=18

			[VARIJABLE] // output
			grubosc=32
		}
		note global value is ignored:
		{
			[VARIJABLE] // output
			_grubosc=32 // 32 is ignored
		}
	*/
	KeepLocal
	/*
		{
			[VARIJABLE] // old
			grubosc=if(sz>50;18;32)

			[VARIJABLE] // new
			_grubosc=18

			[VARIJABLE] // output
			grubosc=if(sz>50;18;32)
		}
	*/
	KeepLocalIfHasCondition
)

type Settings struct {
	minify                        bool
	convertLocalVariablesToGlobal ConvertLocalVariablesToGlobal
}

var settings Settings = Settings{minify: false, convertLocalVariablesToGlobal: Always}

func replaceMakroInCorpusE3DFile(inputFile string, outputFile string, makroFile string) {
	newMakro := LoadMakroFromCMKFile(makroFile)
	handleCorpusFile(inputFile, outputFile, func(decoder *xml.Decoder, start xml.StartElement) xml.Token {
		var oldMakro M1
		e := decoder.DecodeElement(&oldMakro, &start)
		if e != nil {
			log.Fatal(e)
		}
		UpdateMakro(&oldMakro, newMakro)
		return newMakro
	})
}

func handleCorpusFile(inputFile string, outputFile string, handleM1 func(decoder *xml.Decoder, start xml.StartElement) xml.Token) error {
	targetField := "M1"
	input, err := os.Open(inputFile)
	if err != nil {
		return fmt.Errorf("error opening input file: %w", err)
	}
	defer input.Close()

	output, err := os.Create(outputFile)
	if err != nil {
		return fmt.Errorf("error creating output file: %w", err)
	}
	defer output.Close()

	decoder := xml.NewDecoder(input)
	encoder := xml.NewEncoder(output)
	encoder.Indent("", "")

	for {
		token, err := decoder.Token()
		encoder.Flush()
		if err != nil {
			if err.Error() == "EOF" {
				break
			}
			return fmt.Errorf("error decoding XML: %w", err)
		}

		switch t := token.(type) {
		case xml.StartElement:
			if t.Name.Local == targetField {
				handleOut := handleM1(decoder, t)
				if handleOut != nil {
					if !settings.minify {
						encoder.Indent("", "  ")
						defer encoder.Indent("", "")
					}
					// indentation is actually considered xml.CharData, so pretty printing is actually modifying it
					err = encoder.Encode(handleOut)
					if err != nil {
						log.Fatal(err)
					}
				}
			} else {
				encoder.EncodeToken(t)
			}
		case xml.CharData:
			if settings.minify {
				charData := strings.TrimSpace(string(token.(xml.CharData))) // minify
				encoder.EncodeToken(xml.CharData(charData))
			} else {
				encoder.EncodeToken(t)
			}
		default:
			encoder.EncodeToken(t)
		}
	}
	return encoder.Flush()
}

type GenericAttribute struct { // Varijable
	DAT string `xml:"DAT,attr"`
}

type EmbeddedMakro struct {
	XMLName           xml.Name `xml:"MSMA"`
	EmbeddedMakroName string   `xml:"-"`
	DAT               string   `xml:"DAT,attr"`
	MAK               *M1      `xml:"MAK,omitempty"`
}

/*
Example value:
Note: names come from Croatian language
<M1 MN="">

	<MSVA DAT="WPUST_GLEBOKOSC=13,NUMER_NARZEDZIA=155"></MSVA>
	<MSFO DAT="pila_grubosc=4"></MSFO>
	<MSPI DAT="J=1,GB=if(typ_plecow=3;1;0),&#34;GN=frezowanie pila&#34;,GD=wpust_glebokosc,GX=pmaxx,GY=0,PX=pmaxx,PY=pmaxy,GS=obj2.autost,PSP=pila_grubosc,PO=0,PS=1,PP=1,PA=1,PMU=1,PMT=numer_narzedzia"></MSPI>
	<MSJO DAT="CONNECT=2345,mindistance=-16,maxdistance=10"></MSJO>

</M1>
*/
type M1 struct {
	MakroName string             `xml:"MN,attr"`
	Varijable GenericAttribute   `xml:"MSVA"`
	Formule   *GenericAttribute  `xml:"MSFO,omitempty"`
	Pila      []GenericAttribute `xml:"MSPI,omitempty"`
	Joint     *GenericAttribute  `xml:"MSJO,omitempty"`
	Grupa     []GenericAttribute `xml:"MSGR,omitempty"`
	Potrosni  []GenericAttribute `xml:"MSPO,omitempty"`
	Makro     []EmbeddedMakro    `xml:"MSMA,omitempty"`
}

const CMKLineSeparator = `,`

var SectionRegex = regexp.MustCompile(`\[((\w+?)\d*)\]`)

func appendM1Section(m *M1, currentSection string, currentSectionText strings.Builder) {
	switch strings.ToLower(currentSection) {
	case "":
	case "varijable":
		m.Varijable.DAT = currentSectionText.String()
	case "formule":
		m.Formule = new(GenericAttribute)
		m.Formule.DAT = currentSectionText.String()
	case "joint":
		m.Joint = new(GenericAttribute)
		m.Joint.DAT = currentSectionText.String()
	case "grupa":
		m.Grupa = append(m.Grupa, GenericAttribute{DAT: currentSectionText.String()})
	case "potrosni":
		m.Potrosni = append(m.Potrosni, GenericAttribute{DAT: currentSectionText.String()})
	case "makro":
		embeddedMakro := EmbeddedMakro{DAT: currentSectionText.String(), MAK: nil}
		scanner := bufio.NewScanner(strings.NewReader(embeddedMakro.DAT))
		for scanner.Scan() {
			line := scanner.Text()
			nameAndValue := strings.SplitN(line, "=", 2)
			if strings.ToLower(currentSection) == "makro" && strings.ToLower(nameAndValue[0]) == "name" {
				if len(nameAndValue) != 2 {
					log.Fatalf("Error, I do not know how to handle this case. Is there '=' in file name? %s", line)
				}
				embeddedMakro.EmbeddedMakroName = nameAndValue[1]
			}
		}
		m.Makro = append(m.Makro, embeddedMakro)
	default:
		log.Printf("unknown section name: %s", currentSection)
	}

}

const M1InitialMacroMarker string = ""

// create makro struct from CMK file
// assert that makro has the same name as file
// if there is a makro inside makro, it is assumed that it is placed in the same directory as original filename
func LoadMakroFromCMKFile(inputFile string) *M1 {
	var initialMakro *M1 = partialMakroFromFile(path.Join(path.Dir(inputFile), inputFile))

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

		var macro *M1 = partialMakroFromFile(path.Join(path.Dir(inputFile), macroToProcess+".CMK"))
		processedMacros[macroToProcess] = macro

		for _, subMacro := range macro.Makro {
			val, ok := unprocessedMakros[subMacro.EmbeddedMakroName]
			fmt.Print(val)
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

	for continueLoop := false; continueLoop; {
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

	return initialMakro
}

// same as MakroFromFile but might have unresolved data in m.makro
func partialMakroFromFile(filename string) *M1 {
	file, err := os.Open(filename)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	m := new(M1)
	m.MakroName = strings.SplitN(filepath.Base(filename), ".", 2)[0] // this name might be wrong, it can be redefined in software
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
				log.Fatalf("%s was parsed badly, I was looking for section name.\n", text)
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
	return m
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

/*
	Update old makro in smart way. Modifies newMacro in place

- do not touch old JOINT section
- do not touch old VARIJABLE, unless
-- there is new variable
-- maybe in future suport deleting unused variable
- discard other old sections (groupa, potrosni, makro, pila)
- todo handle case insensitive and global names: _VAR==VAR==var==vAr
*/
func UpdateMakro(oldMacro *M1, newMacro *M1) {
	// load everything related to old
	oldVariablesKeys := []string{} // not needed for now
	lastName := ""
	oldValues := map[string]string{}
	for _, line := range strings.Split(oldMacro.Varijable.DAT, CMKLineSeparator) {
		lineTrimmed := decodeCMKLine(line)
		isComment := strings.HasPrefix(lineTrimmed, "//")
		if isComment {
			// discard old comments
			continue
		}
		nameValue := strings.SplitN(line, "=", 2)
		lastName = nameValue[0]
		oldVariablesKeys = append(oldVariablesKeys, lastName)
		oldValues[nameValue[0]] = nameValue[1]
	}

	// load everything related to new
	newVariablesKeys := []string{}
	newValues := map[string]string{}
	newVariablesComments := map[string][]string{}
	lastName = ""
	for _, line := range strings.Split(oldMacro.Varijable.DAT, CMKLineSeparator) {
		lineTrimmed := decodeCMKLine(line)
		isComment := strings.HasPrefix(lineTrimmed, "//")
		if isComment {
			newVariablesComments[lastName] = append(newVariablesComments[lastName], lineTrimmed)
		} else {
			nameValue := strings.SplitN(lineTrimmed, "=", 2)
			lastName = nameValue[0]
			newVariablesKeys = append(newVariablesKeys, nameValue[0])
			newValues[nameValue[0]] = nameValue[1]
		}
	}

	// combine old and new in "smart way"
	var outputVarijable strings.Builder
	// write initial comment
	for _, line := range newVariablesComments[M1InitialMacroMarker] {
		outputVarijable.WriteString(encodeCMKLine(line))
	}
	delete(newVariablesComments, M1InitialMacroMarker)
	for _, newName := range newVariablesKeys {
		oldName, _ := CMKFindOldName(oldVariablesKeys, newName)
		// setting: convert local values to global: always; if value stays the same; convert to evar expression; Keep as is
		// todo if oldValue is not integer, do not allow to convert it to global value
		// one=4
		// one=evar.one+20
		// _one=4//4 is ignored
		oldValue, ok := oldValues[oldName]
		if ok {
			outputVarijable.WriteString(encodeCMKLine(newName + "=" + oldValue))
		} else {
			outputVarijable.WriteString(encodeCMKLine(newName + "=" + newValues[newName]))
		}
		delete(newValues, newName)
		for _, comment := range newVariablesComments[newName] {
			outputVarijable.WriteString(encodeCMKLine(comment))
		}
	}

	// append missing new variables
	for name, newValue := range newValues {
		outputVarijable.WriteString(encodeCMKLine(name + "=" + newValue))
		for _, comment := range newVariablesComments[name] {
			outputVarijable.WriteString(encodeCMKLine(comment))
		}
	}

	// old variables that no longer exist are discarded

	newMacro.Varijable.DAT = outputVarijable.String()
	newMacro.Joint = oldMacro.Joint
}

func CMKFindOldName(oldVariablesNames []string, name string) (string, bool) {
	cleanupName, _ := strings.CutPrefix(name, "_")
	cleanupName = strings.ToLower(cleanupName)
	for index, possibleMatch := range oldVariablesNames {
		cleanupPossibleMatch, _ := strings.CutPrefix(possibleMatch, "_")
		cleanupPossibleMatch = strings.ToLower(cleanupPossibleMatch)
		if cleanupName == possibleMatch {
			return oldVariablesNames[index], true
		}
	}
	return name, false
}

func main() {
	// openCorpusE3D("Szablon biurowy prosty.E3D")
	makroFile := `C:\Tri D Corpus\Corpus 5.0\Makro\custom.CMK`
	inputFile := `C:\Tri D Corpus\Corpus 5.0\elmsav\_modifications\simple_original_custom.E3D`
	outputFile := `C:\Tri D Corpus\Corpus 5.0\elmsav\_modifications\simple_original_custom_output.E3D`
	replaceMakroInCorpusE3DFile(inputFile, outputFile, makroFile)
	// m := LoadMakroFromCMKFile(makroFile)
	// fmt.Println("writing output!")
	// out, err := os.Create("out.xml")
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// // enc := transform.NewWriter(out, charmap.Windows1250.NewDecoder())
	// mybytes, err := xml.MarshalIndent(&m, "", "  ")
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// out.Write(mybytes)
}
