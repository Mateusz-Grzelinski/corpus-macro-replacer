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

func replaceMakroInCorpusE3DFile(inputFile string, outputFile string, makroFile string) {
	// newMakro := LoadMakroFromCMKFile(makroFile)
	handleCorpusFile(inputFile, outputFile, handleM1Makro)
}

func handleM1Makro(decoder *xml.Decoder, t xml.StartElement) xml.Token {
	var oldMakro M1
	e := decoder.DecodeElement(&oldMakro, &t)
	if e != nil {
		log.Fatal(e)
	}
	return oldMakro
}

func handleCorpusFile(inputFile string, outputFile string, handler func(decoder *xml.Decoder, start xml.StartElement) xml.Token) error {
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
				handleOut := handler(decoder, t)
				if handleOut != nil {
					encoder.Indent("", "  ") // each, weird.
					// indentation is actually considered xml.CharData, so I do not know how to handle it
					err = encoder.Encode(handleOut)
					if err != nil {
						log.Fatal(err)
					}
					encoder.Indent("", "")
				}
			} else {
				encoder.EncodeToken(t)
			}
		// case xml.CharData:
		// 	charData := strings.TrimSpace(string(token.(xml.CharData)))
		// 	encoder.EncodeToken(xml.CharData(charData))
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
	resolveSubMakros := map[string]*M1{"": initialMakro}
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
			currentSectionText.WriteString(`"` + text + `"`)
			currentSectionText.WriteString(CMKLineSeparator)
		}
	}
	if currentSectionText.Len() != 0 {
		appendM1Section(m, currentSection, currentSectionText)
	}
	// m.Varijable = append(m.MSFO)
	return m
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
	// load everythin related to old
	oldVariablesKeys := []string{}
	oldVariables := map[string]string{}
	lastName := ""
	for _, line := range strings.Split(oldMacro.Varijable.DAT, "\n") {
		leadingWhiteCharTrimmed := strings.TrimSpace(line)
		isComment := strings.HasPrefix(leadingWhiteCharTrimmed, "//")
		if isComment {
			// discard old comments
			continue
		}
		nameValue := strings.SplitN(line, "=", 2)
		lastName = nameValue[0]
		oldVariablesKeys = append(oldVariablesKeys, lastName)
		oldVariables[nameValue[0]] = nameValue[1]
	}

	// load everything related to new
	newVariablesKeys := []string{}
	newVariables := map[string]string{}
	newVariablesComments := make(map[string][]string)
	lastName = ""
	for _, line := range strings.Split(oldMacro.Varijable.DAT, "\n") {
		leadingWhiteCharTrimmed := strings.TrimSpace(line)
		isComment := strings.HasPrefix(leadingWhiteCharTrimmed, "//")
		if isComment {
			newVariablesComments[lastName] = append(newVariablesComments[lastName], line)
		} else {
			nameValue := strings.SplitN(line, "=", 2)
			newVariablesKeys = append(newVariablesKeys, nameValue[0])
			newVariables[nameValue[0]] = nameValue[1]
		}
	}

	// combine old and new in "smart way"
	var outputVarijable strings.Builder
	for _, name := range newVariablesKeys {
		oldValue, ok := oldVariables[name]
		if ok {
			outputVarijable.WriteString(name + "=" + oldValue + CMKLineSeparator)
		} else {
			outputVarijable.WriteString(name + "=" + newVariables[name] + CMKLineSeparator)
		}
		delete(newVariables, name)
		for _, comment := range newVariablesComments[name] {
			outputVarijable.WriteString(comment + CMKLineSeparator)
		}
	}

	// append missing new variables
	for name, newValue := range newVariables {
		outputVarijable.WriteString(name + "=" + newValue + CMKLineSeparator)
		for _, comment := range newVariablesComments[name] {
			outputVarijable.WriteString(comment + CMKLineSeparator)
		}
	}

	newMacro.Varijable.DAT = outputVarijable.String()
	newMacro.Joint = oldMacro.Joint
	// outM := new(M1)
	// copy(outM, newMacro)
	// return outM
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
