package main

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"log"
	"os"
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
				if root.VER.Value != "16" {
					return nil, nil, fmt.Errorf("unsupported corpus file version: %s", root.VER.Value)
					// todo convert MakLink to Spoj
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
				if root.VER.Value != "16" {
					return nil, nil, fmt.Errorf("unsupported corpus file version: %s", root.VER.Value)
					// todo convert MakLink to Spoj
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
