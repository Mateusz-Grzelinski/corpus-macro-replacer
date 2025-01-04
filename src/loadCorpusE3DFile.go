package main

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"log"
	"os"
	"strings"
)

type GenericNode struct {
	XMLName  xml.Name
	Attr     []xml.Attr    `xml:",any,attr"`
	Content  []GenericNode `xml:",any"`
	Chardata string        `xml:",chardata"`
	Comment  xml.Comment   `xml:",comment"`
}

type ElementFile struct {
	GenericNode
	FILE    xml.Attr  `xml:"FILE,attr"`
	VER     xml.Attr  `xml:"VER,attr"`
	Element []Element `xml:"ELEMENT"`
}
type Element struct {
	GenericNode
	Daske  []Daske  `xml:"DASKE"`
	Elinks []Elinks `xml:"ELINKS"`
}
type Daske struct {
	GenericNode
	DCount xml.Attr `xml:"DCOUNT,attr"`
	AD     []AD     `xml:"AD"`
}
type AD struct {
	GenericNode
	DName xml.Attr `xml:"DName,attr"`
}
type Elinks struct {
	GenericNode
	Spoj []Spoj `xml:"SPOJ"`
}
type Spoj struct {
	GenericNode
	O1DaskeIndex xml.Attr `xml:"O1,attr"`
	/* unknown */
	O2 xml.Attr `xml:"O2,attr"`
	/* unknown */
	SP xml.Attr `xml:"SP,attr"`
	// M1
	/* unknown */
	// M2
}

type GenericAttribute struct { // Varijable
	DAT  string `xml:"DAT,attr"`
	Attr string `xml:",any,attr"`
}

type EmbeddedMakro struct {
	XMLName           xml.Name `xml:"MSMA"`
	EmbeddedMakroName string   `xml:"-"`
	DAT               string   `xml:"DAT,attr"`
	MAK               *M1      `xml:"MAK,omitempty"`
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
	/* MN is not obligatory. Empty names means that makro is not save in any file. */
	MakroName string             `xml:"MN,attr"`
	Varijable GenericAttribute   `xml:"MSVA"`
	Formule   *GenericAttribute  `xml:"MSFO,omitempty"`
	Pila      []GenericAttribute `xml:"MSPI,omitempty"`
	Joint     *GenericAttribute  `xml:"MSJO,omitempty"`
	Grupa     []GenericAttribute `xml:"MSGR,omitempty"`
	Potrosni  []GenericAttribute `xml:"MSPO,omitempty"`
	Raster    []GenericAttribute `xml:"MSRA,omitempty"`
	Makro     []EmbeddedMakro    `xml:"MSMA,omitempty"`
}

func handleCorpusFile(inputFile string, outputFile string, minify bool, handleM1 func(decoder *xml.Decoder, start xml.StartElement) xml.Token) error {
	targetField := "M1"
	log.Printf("Reading Corpus file: '%s'", inputFile)
	input, err := os.Open(inputFile)
	if err != nil {
		return fmt.Errorf("error opening input file: %w", err)
	}
	defer input.Close()

	var encodedData bytes.Buffer
	decoder := xml.NewDecoder(input)
	encoder := xml.NewEncoder(&encodedData)
	// indentation is actually considered xml.CharData, so pretty printing is actually modifying it
	encoder.Indent("", "")

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
			if t.Name.Local == targetField {
				handleOut := handleM1(decoder, t)
				if handleOut != nil {
					if !minify {
						encoder.Indent("", "  ")
					}
					err = encoder.Encode(handleOut)
					if err != nil {
						log.Fatal(err)
					}
					if !minify {
						encoder.Indent("", "")
					}
				}
			} else {
				encoder.EncodeToken(t)
			}
		case xml.CharData:
			if minify {
				charData := strings.TrimSpace(string(token.(xml.CharData)))
				encoder.EncodeToken(xml.CharData(charData))
			} else {
				encoder.EncodeToken(t)
			}
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
