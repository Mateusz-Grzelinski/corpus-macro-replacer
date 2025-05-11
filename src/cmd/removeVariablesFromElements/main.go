package main

import (
	"bytes"
	"corpus_macro_replacer/corpus"
	"encoding/xml"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

const Version = "0.1"

var variablesToRemove []string = []string{
	"ilosc_zawiasow=.*",
	"producent_zawiasu=.*",
	"typ_prowadnika_rodzaj=.*",
	"typ_zawiasu_s=.*",
	"zmiana_pozycji_dolnego=.*",
	"zmiana_pozycji_gornego=.*",
	"zmiana_pozycji_3=.*",
	"zmiana_pozycji_4=.*",
	"zmiana_pozycji_5=.*",
	"zmiana_pozycji_6=.*",
	"przesuniecie_prowadnik=.*",
	"przesuniecie_zawias=.*",
}

func main() {
	flag.Usage = func() {
		w := flag.CommandLine.Output()
		fmt.Fprint(w, `This program is used to remove certain variables from elements and subelements`)
		fmt.Fprintf(w, "Usage of %s -input <PATH> -output <PATH> -makro <PATH>:\n", os.Args[0])
		flag.PrintDefaults()
	}
	var version *bool = flag.Bool("v", false, "print version and exit")
	var input *string = flag.String("input", "", "required. File or dir, must exist. If dir then changes macro recursively for all .E3D files.")
	var output *string = flag.String("output", "", `required. output directory, does not need to exist.`)
	flag.Parse()

	if *version {
		fmt.Printf("CorpusVariableRemover v%s\n", Version)
		os.Exit(0)
	}
	if *input == "" {
		log.Fatalln("-input can not be empty")
	}
	if *output == "" {
		log.Fatalln("-output can not be empty")
	}
	err := parseAllfiles(*output, *input, variablesToRemove)
	if err != nil {
		log.Fatal(err)
	}
	log.Print("Done")
}

func parseAllfiles(output string, input string, variablesToRemove []string) error {
	foundCorpusFiles := []string{}
	foundCorpusFiles = append(foundCorpusFiles, corpus.FindCorpusFiles(input)...)

	regexes := make([]*regexp.Regexp, len(variablesToRemove))
	for i, pattern := range variablesToRemove {
		r := regexp.MustCompile(pattern)
		regexes[i] = r
	}

	errors_occured := []string{}
	for _, inputFile := range foundCorpusFiles {
		result, err := parseSingleFile(inputFile, regexes)
		if err != nil {
			errors_occured = append(errors_occured, inputFile)
			log.Printf("error processing file %s: %s", inputFile, err)
			continue
		}
		outputFile := corpus.GetCleanOutputpath(output, inputFile)
		log.Printf("parsing: %s", outputFile)

		err = os.MkdirAll(filepath.Dir(outputFile), os.ModePerm)
		if err != nil {
			return fmt.Errorf("can not create path: '%s': %w", outputFile, err)
		}

		output, err := os.Create(outputFile)
		if err != nil {
			return fmt.Errorf("error creating output file: %w", err)
		}
		defer output.Close()

		var encodedData bytes.Buffer
		encoder := xml.NewEncoder(&encodedData)
		encoder.Indent("", "  ")
		if err = encoder.Encode(result); err != nil {
			log.Printf("Error during encode: %s", err)
			return err
		}
		_, err = output.Write(encodedData.Bytes())
		if err != nil {
			return err
		}
		log.Printf("Done writing file: '%s'", outputFile)
	}
	if len(errors_occured) > 0 {
		log.Printf("%d errors occured in files (see log above)", len(errors_occured))
		for i := range errors_occured {
			log.Printf("  %s", errors_occured[i])
		}
		return fmt.Errorf("some errors occured, see above")
	}
	return nil
}

func parseSingleFile(input string, regexes []*regexp.Regexp) (any, error) {
	projectFile, elementFile, err := corpus.NewCorpusFile(input)
	if err != nil {
		return nil, err
	}
	if projectFile != nil {
		projectFile, err := RemoveVariablesFromFile(projectFile, regexes)
		return projectFile, err
	} else if elementFile != nil {
		return RemoveVariablesFromFile(elementFile, regexes)
	} else {
		log.Fatal("should not happen")
	}
	return nil, fmt.Errorf("Unknown error occured (should not happen)")
}

type CanWalkElements interface {
	VisitElementsAndSubelements(f func(*corpus.Element))
}

func RemoveVariablesFromFile[T CanWalkElements](projectFile T, removePatterns []*regexp.Regexp) (T, error) {
	RemoveVariablesCallback := func(e *corpus.Element) {
		newAttributes := []xml.Attr{}
	to_next_attribute:
		for _, attr := range e.Evar.Attr {
			if !strings.HasPrefix(attr.Name.Local, "VAR") {
				// some unknown attribute, leave it alone
				newAttributes = append(newAttributes, attr)
				continue
			}
			for _, pattern := range removePatterns {
				// all attributes here are VAR0, VAR1, ...
				if pattern.MatchString(attr.Value) {
					// matches the pattern, remove it
					continue to_next_attribute
				}
			}
			// does not match pattern, leave it alone
			newName := fmt.Sprintf("VAR%d", len(newAttributes))
			newAttributes = append(newAttributes, xml.Attr{Name: xml.Name{Local: newName}, Value: attr.Value})
		}
		if len(newAttributes) != len(e.Evar.Attr) {
			log.Printf("Element: %s", e.EName.Value)
			log.Printf("Removed %d/%d variables", len(e.Evar.Attr)-len(newAttributes), len(e.Evar.Attr))
		}
		e.Evar.Attr = newAttributes
	}
	projectFile.VisitElementsAndSubelements(RemoveVariablesCallback)
	return projectFile, nil
}
