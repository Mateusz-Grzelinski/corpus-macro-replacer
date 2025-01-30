package main

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
)

type GenericNode struct {
	XMLName  xml.Name
	Attr     []xml.Attr    `xml:",any,attr"`
	Content  []GenericNode `xml:",any"`
	Chardata string        `xml:",chardata"`
	// Comment  xml.Comment   `xml:",comment"`
}

// E3D file
type ElementFile struct {
	XMLName xml.Name `xml:"ELEMENTFILE"`
	GenericNode
	FILE    xml.Attr  `xml:"FILE,attr"`
	VER     xml.Attr  `xml:"VER,attr"`
	Element []Element `xml:"ELEMENT"`
}

// S3D file
type ProjectFile struct {
	XMLName xml.Name `xml:"PROJECTFILE"`
	ElementFile
}

// single cabinet
type Element struct {
	GenericNode
	EName   xml.Attr `xml:"ENAME,attr"`
	Daske   Daske    `xml:"DASKE"`
	ElmList ElmList  `xml:"ELMLIST"`
	Elinks  Elinks   `xml:"ELINKS"`
}

// list of nested elements
type ElmList struct {
	GenericNode
	ECount xml.Attr  `xml:"ECOUNT,attr"`
	Elm    []Element `xml:"ELM"`
}
type Daske struct {
	GenericNode
	DCount xml.Attr `xml:"DCOUNT,attr"`
	AD     []AD     `xml:"AD"`
}

// plank
type AD struct {
	GenericNode
	DName xml.Attr `xml:"DNAME,attr"`
}
type Elinks struct {
	GenericNode
	COUNT xml.Attr `xml:"COUNT,attr"`
	Spoj  []Spoj   `xml:"SPOJ"`
}
type Spoj struct {
	GenericNode
	/* index of: element.Daske.AD[O1] */
	O1 xml.Attr `xml:"O1,attr"`
	/* unknown */
	O2 xml.Attr `xml:"O2,attr"`
	/* unknown */
	SP     xml.Attr `xml:"SP,attr"`
	Makro1 M1       `xml:"M1"`
	/* unknown, M2 seems to be unused */
	Makro2 GenericNode `xml:"M2"`
}

type GenericNodeWithDat struct { // Varijable
	GenericNode
	DAT string `xml:"DAT,attr"`
	// Attr string `xml:",any,attr"`
}

type EmbeddedMakro struct {
	// GenericNode
	GenericNodeWithDat
	EmbeddedMakroName string `xml:"-"`
	// DAT               string `xml:"DAT,attr"`
	MAK *M1 `xml:"MAK,omitempty"`
}

/*
Represents makro as defined in Corpus 5.0 (reverse engineered).
Help is available only in Corpus makro editor.

Note: names come from Croatian language

Example value:

	{
		<M1 MN="">
			<MSVA DAT="WPUST_GLEBOKOSC=13,NUMER_NARZEDZIA=155"></MSVA>
			<MSFO DAT="pila_grubosc=4"></MSFO>
			<MSPI DAT="J=1,GB=if(typ_plecow=3;1;0),&#34;GN=frezowanie pila&#34;,GD=wpust_glebokosc,..."></MSPI>
			<MSJO DAT="CONNECT=2345,mindistance=-16,maxdistance=10"></MSJO>
		</M1>
	}
*/
type M1 struct {
	GenericNode
	/* MN is not obligatory. Empty names means that makro is not save in any file. */
	MakroName string               `xml:"MN,attr"`
	Varijable GenericNodeWithDat   `xml:"MSVA"`
	Formule   *GenericNodeWithDat  `xml:"MSFO,omitempty"`
	Pila      []GenericNodeWithDat `xml:"MSPI,omitempty"`
	Joint     *GenericNodeWithDat  `xml:"MSJO,omitempty"`
	Grupa     []GenericNodeWithDat `xml:"MSGR,omitempty"`
	Potrosni  []GenericNodeWithDat `xml:"MSPO,omitempty"`
	Pocket    []GenericNodeWithDat `xml:"MSPOCK,omitempty"`
	Raster    []GenericNodeWithDat `xml:"MSRA,omitempty"`
	Makro     []EmbeddedMakro      `xml:"MSMA,omitempty"`
}

type TrimmerDecoder struct {
	decoder *xml.Decoder
}

// save makro as CMK file
func (m *M1) Save(w io.Writer) error {
	if _, err := w.Write([]byte("[VARIJABLE]\n")); err != nil {
		return err
	}
	{
		text := strings.Join(decodeAllCMKLines(m.Varijable.DAT), "\n")
		if _, err := w.Write([]byte(text)); err != nil {
			return err
		}
	}
	if m.Joint != nil {
		if _, err := w.Write([]byte("\n[JOINT]\n")); err != nil {
			return err
		}
		text := strings.Join(decodeAllCMKLines(m.Joint.DAT), "\n")
		if _, err := w.Write([]byte(text)); err != nil {
			return err
		}
	}
	if m.Formule != nil {
		if _, err := w.Write([]byte("\n[FORMULE]\n")); err != nil {
			return err
		}
		text := strings.Join(decodeAllCMKLines(m.Formule.DAT), "\n")
		if _, err := w.Write([]byte(text)); err != nil {
			return err
		}
	}

	for i, item := range m.Pocket {
		section := fmt.Sprintf("\n[POCKET%d]\n", i)
		if _, err := w.Write([]byte(section)); err != nil {
			return err
		}
		text := strings.Join(decodeAllCMKLines(item.DAT), "\n")
		if _, err := w.Write([]byte(text)); err != nil {
			return err
		}
	}
	for i, item := range m.Potrosni {
		section := fmt.Sprintf("\n[POTROSNI%d]\n", i)
		if _, err := w.Write([]byte(section)); err != nil {
			return err
		}
		text := strings.Join(decodeAllCMKLines(item.DAT), "\n")
		if _, err := w.Write([]byte(text)); err != nil {
			return err
		}
	}
	for i, item := range m.Grupa {
		section := fmt.Sprintf("\n[GRUPA%d]\n", i)
		if _, err := w.Write([]byte(section)); err != nil {
			return err
		}
		text := strings.Join(decodeAllCMKLines(item.DAT), "\n")
		if _, err := w.Write([]byte(text)); err != nil {
			return err
		}
	}
	for i, item := range m.Raster {
		section := fmt.Sprintf("\n[RASTER%d]\n", i)
		if _, err := w.Write([]byte(section)); err != nil {
			return err
		}
		text := strings.Join(decodeAllCMKLines(item.DAT), "\n")
		if _, err := w.Write([]byte(text)); err != nil {
			return err
		}
	}
	for i, item := range m.Makro {
		section := fmt.Sprintf("\n[MAKRO%d]\n", i)
		if _, err := w.Write([]byte(section)); err != nil {
			return err
		}
		text := strings.Join(decodeAllCMKLines(item.DAT), "\n")
		if _, err := w.Write([]byte(text)); err != nil {
			return err
		}
	}

	return nil
}

func (tr TrimmerDecoder) Token() (xml.Token, error) {
	t, err := tr.decoder.Token()
	if cd, ok := t.(xml.CharData); ok {
		t = xml.CharData(bytes.TrimSpace(cd))
	}
	return t, err
}

// normal idiomatic way of reading corpus file
func NewCorpusFile(inputFile string) (*ProjectFile, *ElementFile, error) {
	log.Printf("Reading Corpus file: '%s'", inputFile)
	input, err := os.Open(inputFile)
	if err != nil {
		return nil, nil, fmt.Errorf("error opening input file: %w", err)
	}
	defer input.Close()

	rawDecoder := xml.NewDecoder(input)
	decoder := xml.NewTokenDecoder(TrimmerDecoder{rawDecoder})
	for {
		token, err := decoder.Token()
		if err != nil {
			if err.Error() == "EOF" {
				break
			}
			return nil, nil, fmt.Errorf("error decoding XML: %w", err)
		}

		switch t := token.(type) {
		case xml.StartElement:
			if t.Name.Local == "PROJECTFILE" {
				var root *ProjectFile = new(ProjectFile)
				decoder.Strict = true
				err := decoder.DecodeElement(&root, &t)
				decoder.Strict = false
				if err != nil {
					return nil, nil, err
				}
				return root, nil, nil
			} else if t.Name.Local == "ELEMENTFILE" {
				var root *ElementFile = new(ElementFile)
				decoder.Strict = true
				err := decoder.DecodeElement(&root, &t)
				decoder.Strict = false
				if err != nil {
					return nil, nil, err
				}
				return nil, root, nil
			}
		default:
		}
	}
	return nil, nil, fmt.Errorf("something went wrong when reading corpus file. Wrong file format?")
}

// goes via file token by token thus has better chance of being correct
func ReadWriteCorpusFile(inputFile string, outputFile string, minify bool,
	handleE3DFile func(decoder *xml.Decoder, start xml.StartElement) xml.Token,
	handleS3DFile func(decoder *xml.Decoder, start xml.StartElement) xml.Token,
) error {
	log.Printf("Reading Corpus file: '%s'", inputFile)
	input, err := os.Open(inputFile)
	if err != nil {
		return fmt.Errorf("error opening input file: %w", err)
	}
	defer input.Close()

	var encodedData bytes.Buffer
	rawDecoder := xml.NewDecoder(input)
	decoder := xml.NewTokenDecoder(TrimmerDecoder{rawDecoder})
	encoder := xml.NewEncoder(&encodedData)
	// indentation is actually considered xml.CharData, so pretty printing is actually modifying it
	if !minify {
		encoder.Indent("", "  ")
	} else {
		encoder.Indent("", "")
	}

	for {
		token, err := decoder.Token()
		if err != nil {
			if err.Error() == "EOF" {
				break
			}
			return fmt.Errorf("error decoding XML: %w", err)
		}

		switch t := token.(type) {
		case xml.StartElement:
			if t.Name.Local == "PROJECTFILE" {
				handleOut := handleS3DFile(decoder, t)
				if handleOut != nil {
					if err = encoder.Encode(handleOut); err != nil {
						log.Fatal(err)
					}
				}
			} else if t.Name.Local == "ELEMENTFILE" {
				handleOut := handleE3DFile(decoder, t)
				if handleOut != nil {
					if err = encoder.Encode(handleOut); err != nil {
						log.Fatal(err)
					}
				}
			} else {
				encoder.EncodeToken(t)
			}
		case xml.CharData:
			encoder.EncodeToken(t)
		case xml.Comment:
			encoder.EncodeToken(t)
		default:
			encoder.EncodeToken(t)
		}
	}
	err = encoder.Flush()
	if err != nil {
		return err
	}

	output, err := os.Create(outputFile)
	if err != nil {
		return fmt.Errorf("error creating output file: %w", err)
	}
	defer output.Close()

	output.Write(encodedData.Bytes())
	log.Printf("Done writing file  : '%s'", outputFile)
	return nil
}
