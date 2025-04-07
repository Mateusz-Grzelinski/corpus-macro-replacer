package main

import (
	"corpus_macro_replacer/corpus"
	"encoding/xml"
	"flag"
	"fmt"
	"log"
	"os"
	"regexp"
	"strings"
)

const Version = "0.1"

var variablesToRemove []string = []string{
	"(?i)_*most_outer_global_var.*=.*",
	"(?i)unused_parent_variable",
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
	var output *string = flag.String("output", "", `required. File or dir, does not need to exist.`)
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

	regexes := make([]*regexp.Regexp, len(variablesToRemove))
	for _, pattern := range variablesToRemove {
		r := regexp.MustCompile(pattern)
		regexes = append(regexes, r)
	}
	projectFile, elementFile, err := corpus.NewCorpusFile(*input)
	if err != nil {
		return // fmt.Errorf("")
	}
	if projectFile != nil {
		RemoveVariablesFromFile(projectFile, regexes)
	} else if elementFile != nil {
		RemoveVariablesFromFile(elementFile, regexes)
	} else {
		log.Fatal("should not happen")
	}
}

type CanWalkElements interface {
	VisitElementsAndSubelements(f func(*corpus.Element))
}

func RemoveVariablesFromFile[T CanWalkElements](projectFile T, removePatterns []*regexp.Regexp) (T, error) {
	RemoveVariablesCallback := func(e *corpus.Element) {
		log.Printf("Element: %s", e.EName.Value)
		newAttributes := []xml.Attr{}
		for _, attr := range e.Evar.Attr {
			for _, pattern := range removePatterns {
				if !strings.HasPrefix(attr.Name.Local, "VAR") {
					newAttributes = append(newAttributes, attr)
					continue
				}
				if pattern.MatchString(attr.Value) {
					continue
				}
				newName := fmt.Sprintf("VAR%d", len(newAttributes))
				newAttributes = append(newAttributes, xml.Attr{Name: xml.Name{Local: newName}, Value: attr.Value})
			}
		}
		log.Printf("Removed %d/%d variables", len(e.Evar.Attr)-len(newAttributes), len(e.Evar.Attr))
		e.Evar.Attr = newAttributes
	}
	projectFile.VisitElementsAndSubelements(RemoveVariablesCallback)
	return projectFile, nil
}
