package main

import (
	"bufio"
	"bytes"
	"encoding/xml"
	"fmt"
	"os"
	"strings"
)

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

type GenericNode struct {
	XMLName  xml.Name
	Attr     []xml.Attr    `xml:",any,attr"`
	Content  []GenericNode `xml:",any"`
	CharData string        `xml:",chardata"`
	// Comments string        `xml:",comment"`
}

// func (g *GenericNode) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
// 	type Alias GenericNode // Create an alias to avoid recursion
// 	var alias Alias
// 	if err := d.DecodeElement(&alias, &start); err != nil {
// 		return err
// 	}
// 	*g = GenericNode(alias)

// 	g.CharData = strings.TrimSpace(g.CharData)
// 	return nil
// }

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

func main() {
	// Open the XML file
	xmlFile, err := os.Open("lewy_gorny.E3D.xml")
	if err != nil {
		fmt.Println("Error opening file:", err)
		return
	}
	defer xmlFile.Close()

	// Read the file content
	byteValue := bufio.NewReader(xmlFile)

	// Unmarshal into the GenericNode struct
	var root ElementFile
	raw := xml.NewDecoder(byteValue)
	decoder := xml.NewTokenDecoder(TrimmerDecoder{raw})

	for {
		token, err := decoder.Token()
		if err != nil {
			if err.Error() == "EOF" {
				break
			}
			fmt.Errorf("error decoding XML: %w", err)
		}

		switch t := token.(type) {
		case xml.StartElement:
			if t.Name.Local == "ELEMENTFILE" {
				err := decoder.DecodeElement(&root, &t)
				if err != nil {
					fmt.Println(err)
				}
			}
			// encoder.EncodeToken(t)
		case xml.CharData:
			// encoder.EncodeToken(t)
		default:
			// encoder.EncodeToken(t)
		}
	}
	// err = xml.Unmarshal(byteValue, &root)
	// iterateXMLTrimChardataSpaces(&root.GenericNode)
	fmt.Println(root.Element[0].Elinks[0].Spoj)
	if err != nil {
		fmt.Println("Error unmarshalling XML:", err)
		return
	}

	// Traverse and access attributes
	// printNodeAttributes(root, 0)

	// Marshal back to XML
	outputXML, err := xml.MarshalIndent(root, "", "  ")
	if err != nil {
		fmt.Println("Error marshalling XML:", err)
		return
	}

	// Save to output.xml
	err = os.WriteFile("output.xml", outputXML, 0644)
	if err != nil {
		fmt.Println("Error writing to file:", err)
		return
	}

	fmt.Println("XML file processed and saved as output.xml")
}

func printNodeAttributes(node GenericNode, depth int) {
	indent := strings.Repeat("  ", depth)
	fmt.Printf("%sNode: %s\n", indent, node.XMLName.Local)
	for _, attr := range node.Attr {
		fmt.Printf("%s  Attribute: %s=%s\n", indent, attr.Name.Local, attr.Value)
	}
	for _, child := range node.Content {
		printNodeAttributes(child, depth+1)
	}
}
