package main

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"log"
	"os"
	"strings"
)

func handleCorpusFile(inputFile string, outputFile string, minify bool, handleM1 func(decoder *xml.Decoder, start xml.StartElement) xml.Token) error {
	targetField := "M1"
	log.Printf("Reading Corpus file: '%s'", inputFile)
	input, err := os.Open(inputFile)
	if err != nil {
		return fmt.Errorf("error opening input file: %w", err)
	}
	defer input.Close()

	// writer := bufio.NewReadWriter()
	var encodedData bytes.Buffer
	decoder := xml.NewDecoder(input)
	encoder := xml.NewEncoder(&encodedData)
	// indentation is actually considered xml.CharData, so pretty printing is actually modifying it
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
	log.Printf("Done: '%s'", outputFile)
	return nil
}
