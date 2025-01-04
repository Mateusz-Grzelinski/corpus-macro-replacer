package main

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"log"
	"os"
)

type GenericNode struct {
	XMLName  xml.Name
	Attr     []xml.Attr    `xml:",any,attr"`
	Content  []GenericNode `xml:",any"`
	Chardata string        `xml:",chardata"`
	// Comment  xml.Comment   `xml:",comment"`
}

type ElementFile struct {
	XMLName xml.Name `xml:"ELEMENTFILE"`
	GenericNode
	FILE    xml.Attr  `xml:"FILE,attr"`
	VER     xml.Attr  `xml:"VER,attr"`
	Element []Element `xml:"ELEMENT"`
}
type Element struct {
	GenericNode
	EName  xml.Attr `xml:"ENAME,attr"`
	Daske  Daske    `xml:"DASKE"`
	Elinks Elinks   `xml:"ELINKS"`
}
type Daske struct {
	GenericNode
	DCount xml.Attr `xml:"DCOUNT,attr"`
	AD     []AD     `xml:"AD"`
}
type AD struct {
	GenericNode
	DName xml.Attr `xml:"DNAME,attr"`
}
type Elinks struct {
	GenericNode
	Spoj []Spoj `xml:"SPOJ"`
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

type GenericAttribute struct { // Varijable
	DAT  string `xml:"DAT,attr"`
	Attr string `xml:",any,attr"`
}

type EmbeddedMakro struct {
	GenericNode
	EmbeddedMakroName string `xml:"-"`
	DAT               string `xml:"DAT,attr"`
	MAK               *M1    `xml:"MAK,omitempty"`
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

type TrimmerDecoder struct {
	decoder *xml.Decoder
}

func (tr TrimmerDecoder) Token() (xml.Token, error) {
	t, err := tr.decoder.Token()
	if cd, ok := t.(xml.CharData); ok {
		t = xml.CharData(bytes.TrimSpace(cd))
	}
	return t, err
}

func handleCorpusFile(inputFile string, outputFile string, minify bool, handleRootElement func(decoder *xml.Decoder, start xml.StartElement) xml.Token) error {
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
			if t.Name.Local == "ELEMENTFILE" {
				handleOut := handleRootElement(decoder, t)
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
