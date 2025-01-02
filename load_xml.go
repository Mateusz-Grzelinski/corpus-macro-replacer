package main

import (
	"bufio"
	"encoding/xml"
	"fmt"
	"io"
	"log"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"

	"golang.org/x/text/encoding/charmap"
	"golang.org/x/text/transform"
)

func openDemo() {
	// Open our xmlFile
	filename := "demo.xml"
	xmlFile, err := os.Open(filename)
	// if we os.Open returns an error then handle it
	if err != nil {
		fmt.Println(err)
	}

	fmt.Println("Successfully Opened {filename}")

	// defer the closing of our xmlFile so that we can parse it later on
	defer xmlFile.Close()
}

func openCorpus(filename string) {
	xmlFile, err := os.Open(filename)
	if err != nil {
		fmt.Println(err)
	}

	fmt.Printf("Successfully Opened %s\n", filename)
	// defer the closing of our xmlFile so that we can parse it later on
	defer xmlFile.Close()

	xml.EscapeText(os.Stdout, []byte("\"Mateusz 'Grzeliński excaped!:)\n"))

	fmt.Println()
	s := `test '123'`
	test := xml.StartElement{Name: xml.Name{Local: `test`}}
	xml.NewEncoder(os.Stdout).EncodeElement(s, test)

	byteValue, _ := io.ReadAll(xmlFile)
	output, err := xml.Marshal(byteValue)
	println(output)

}

type GenericElement struct {
	XMLName  xml.Name
	Attrs    []xml.Attr       `xml:",any,attr"`
	Content  string           `xml:",chardata"`
	Children []GenericElement `xml:",any"`
}

func modifyXMLTokenByToken(inputFile, outputFile, targetField, newValue string) error {
	// Open the input file
	input, err := os.Open(inputFile)
	if err != nil {
		return fmt.Errorf("error opening input file: %w", err)
	}
	defer input.Close()

	// Create the output file
	output, err := os.Create(outputFile)
	if err != nil {
		return fmt.Errorf("error creating output file: %w", err)
	}
	defer output.Close()

	// Create an XML decoder and encoder
	decoder := xml.NewDecoder(input)
	encoder := xml.NewEncoder(output)
	encoder.Indent("", "  ")

	// Walk through the XML
	var inTargetElement, charDataVisited bool
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
			inTargetElement = t.Name.Local == targetField
			if inTargetElement {
				var m M1
				e := decoder.DecodeElement(&m, &t)
				if e != nil {
					log.Fatal(e)
				}
				charDataVisited = false
			}
			encoder.EncodeToken(t)
		case xml.EndElement:
			if inTargetElement && !charDataVisited {
				encoder.EncodeToken(xml.CharData(newValue))
			}
			if t.Name.Local == targetField {
				inTargetElement = false
			}
			encoder.EncodeToken(t)
		case xml.CharData:
			if inTargetElement {
				t = xml.CharData(newValue)
				charDataVisited = true
			}
			encoder.EncodeToken(t)
		default:
			// Pass through other tokens
			encoder.EncodeToken(t)
		}
	}

	// Flush the encoder
	return encoder.Flush()
}

type GenericAttribute struct { // Varijable
	// Text string `xml:",chardata"`
	DAT string `xml:"DAT,attr"`
}

type EmbeddedMakro struct {
	XMLName           xml.Name `xml:"MSMA"`
	EmbeddedMakroName string   `xml:"-"`
	// Text    string   `xml:",chardata"`
	DAT string `xml:"DAT,attr"`
	MAK *M1    `xml:"MAK,omitempty"`
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
	// XMLName   xml.Name           `xml:"M1"`
	// Text      string             `xml:",chardata"`
	MakroName string             `xml:"MN,attr"`
	Varijable GenericAttribute   `xml:"MSVA"`
	Formule   *GenericAttribute  `xml:"MSFO,omitempty"`
	Pila      []GenericAttribute `xml:"MSPI,omitempty"`
	Joint     *GenericAttribute  `xml:"MSJO,omitempty"`
	Grupa     []GenericAttribute `xml:"MSGR,omitempty"`
	Potrosni  []GenericAttribute `xml:"MSPO,omitempty"`
	Makro     []EmbeddedMakro    `xml:"MSMA,omitempty"`
	// Children  []GenericElement   `xml:",any,-"` // anything that we missed, usually empty
}

const CMKSeparator = `,`

func loadMakroFromXML(inputFile string) error {
	input, err := os.Open(inputFile)
	if err != nil {
		return fmt.Errorf("error opening input file: %w", err)
	}
	byteValue, _ := io.ReadAll(input)
	var m M1
	err = xml.Unmarshal(byteValue, &m)
	// if err != nil {
	// 	log.Fatal(err)
	// 	return fmt.Errorf("error opening input file: %w", err)
	// }
	fmt.Println(err)
	return nil
}

var SectionRegex = regexp.MustCompile(`\[((\w+?)\d*)\]`)

func appendSection(m *M1, currentSection string, currentSectionText strings.Builder) {
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
			nameAndValue := strings.Split(line, "=")
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
func MakroFromFile(filename string) *M1 {
	var initialMakro *M1 = partialMakroFromFile(path.Join(path.Dir(filename), filename))

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

		var macro *M1 = partialMakroFromFile(path.Join(path.Dir(filename), macroToProcess+".CMK"))
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
			appendSection(m, currentSection, currentSectionText)
			currentSectionText.Reset()
			currentSection = sectionName
		} else {
			currentSectionText.WriteString(`"` + text + `"`)
			currentSectionText.WriteString(CMKSeparator)
		}
	}
	if currentSectionText.Len() != 0 {
		appendSection(m, currentSection, currentSectionText)
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

	// load everythin related to new
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
			outputVarijable.WriteString(name + "=" + oldValue + CMKSeparator)
		} else {
			outputVarijable.WriteString(name + "=" + newVariables[name] + CMKSeparator)
		}
		delete(newVariables, name)
		for _, comment := range newVariablesComments[name] {
			outputVarijable.WriteString(`"` + comment + `"` + CMKSeparator)
		}
	}

	// append missing new variables
	for name, newValue := range newVariables {
		outputVarijable.WriteString(name + "=" + newValue + CMKSeparator)
		for _, comment := range newVariablesComments[name] {
			outputVarijable.WriteString(`"` + comment + `"` + CMKSeparator)
		}
	}

	newMacro.Varijable.DAT = outputVarijable.String()
	newMacro.Joint = oldMacro.Joint
	// outM := new(M1)
	// copy(outM, newMacro)
	// return outM
}

func main() {
	// openCorpus("Szablon biurowy prosty.E3D")
	// modifyXMLTokenByToken(
	// 	"Szablon biurowy prosty.E3D", "Szablon biurowy prosty.E3D.out.xml",
	// 	"M1", "ala!!")
	// loadMakroFromXML("M1.xml")
	m := MakroFromFile(`C:\Tri D Corpus\Corpus 5.0\Makro\custom.CMK`)
	// fmt.Println(m)
	fmt.Println("writing output!")
	out, err := os.Create("out.xml")
	if err != nil {
		log.Fatal(err)
	}
	// enc := transform.NewWriter(out, charmap.Windows1250.NewDecoder())
	mybytes, err := xml.MarshalIndent(&m, "", "  ")
	if err != nil {
		log.Fatal(err)
	}
	out.Write(mybytes)
}
