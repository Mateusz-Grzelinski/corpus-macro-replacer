package corpus

import (
	"encoding/xml"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
)

func ReplaceMakroInCorpusFile(inputFile string, outputFile string, makrosToReplace map[string]*M1, makroRename map[string]string, alwaysConvertLocalToGlobal bool, verbose bool, minify bool) error {
	macrosUpdated := 0
	macrosSkipped := 0

	handleVisitElement := func(element *Element) {
		visitedDaske := []string{}
		updatedDaske := map[string]int{}
		skippedDaske := map[string]int{}
		for i, spoj := range element.Elinks.Spoj {
			adIndex, _ := strconv.Atoi(spoj.O1.Value)
			daske := element.Daske.AD[adIndex]
			daskeName := daske.DName.Value
			visitedDaske = append(visitedDaske, daskeName)
			oldMakro := spoj.Makro1
			newMakro, newMakroExists := makrosToReplace[oldMakro.MakroName]
			if !newMakroExists {
				macrosSkipped++
				skippedDaske[daskeName]++
				continue
			}
			renameMakro, found := makroRename[oldMakro.MakroName]
			var renameTo *string
			if !found {
				renameTo = nil
			} else {
				renameTo = &renameMakro
			}
			newMakroCopyUntilIFixTheUpdateMakro := *newMakro
			UpdateMakro(&oldMakro, &newMakroCopyUntilIFixTheUpdateMakro, renameTo, alwaysConvertLocalToGlobal)
			element.Elinks.Spoj[i].Makro1 = newMakroCopyUntilIFixTheUpdateMakro
			// spoj.Makro1 = *newMakro

			// todo reorder variables so that ones with the same name are next to each other
			// oldVariablesKeys, oldValues, _ := loadValuesFromSection(oldMacro.Varijable.DAT)
			// newVariablesKeys, newValues, newVariablesComments := loadValuesFromSection(newMacro.Varijable.DAT)
			macrosUpdated++
			updatedDaske[daskeName]++
		}
		if verbose {
			log.Printf("  Cabinet '%s'\n", element.EName.Value)
			for _, name := range visitedDaske {
				log.Printf("    Updated %d macros, %d skipped in plate '%s'\n", updatedDaske[name], skippedDaske[name], name)
			}
		}
	}

	// todo how to handle 2 very similar types??? ElementFile vs ProjectFile
	visitCorpusE3DFile := func(decoder *xml.Decoder, start xml.StartElement) xml.Token {
		var rootCorpusFile ElementFile
		decoder.Strict = true
		err := decoder.DecodeElement(&rootCorpusFile, &start)
		decoder.Strict = false
		if err != nil {
			log.Printf("%s: %s", inputFile, err)
		}
		// todo visit all elements including groups
		rootCorpusFile.VisitElementsAndSubelements(handleVisitElement)
		log.Printf("  Summary: updated %d macros, %d skipped\n", macrosUpdated, macrosSkipped)

		return rootCorpusFile
	}

	visitCorpusS3DFile := func(decoder *xml.Decoder, start xml.StartElement) xml.Token {
		var rootCorpusFile ProjectFile
		decoder.Strict = true
		err := decoder.DecodeElement(&rootCorpusFile, &start)
		decoder.Strict = false
		if err != nil {
			log.Printf("%s: %s", inputFile, err)
		}
		rootCorpusFile.VisitElementsAndSubelements(handleVisitElement)
		log.Printf("  Summary: updated %d macros, %d skipped\n", macrosUpdated, macrosSkipped)

		return rootCorpusFile
	}

	err := ReadWriteCorpusFile(inputFile, outputFile, minify, visitCorpusE3DFile, visitCorpusS3DFile)
	if err != nil {
		log.Printf("error when operating on corpus file: %s", err)
	}
	return err
}

func ReplaceMakroInCorpusFolder(inputFolder string, outputFolder string, makroFiles []string, macroNamesOverrides []*string, alwaysConvertLocalToGlobal bool, verbose bool, minify bool) error {
	inputFolderStat, err := os.Stat(inputFolder)
	if err != nil {
		return fmt.Errorf("error reading input folder: %w", err)
	}
	if !inputFolderStat.IsDir() {
		return fmt.Errorf("%s must be a directory", inputFolder)
	}
	err = os.MkdirAll(outputFolder, os.ModePerm)
	if err != nil {
		return fmt.Errorf("can not create dir: %w", err)
	}

	foundCorpusFiles := FindCorpusFiles(inputFolder)

	log.Printf("Found %d files in %s", len(foundCorpusFiles), inputFolder)

	// todo support the rest of parameters
	makrosToReplace, err := ReadMakrosFromCMK(makroFiles, macroNamesOverrides, nil, nil)
	if err != nil {
		return fmt.Errorf("error reading CMK macros: %w", err)
	}
	var errOut error
	for _, inputFile := range foundCorpusFiles {
		relInputFile, _ := filepath.Rel(inputFolder, inputFile)
		outputFile := filepath.Join(outputFolder, relInputFile)
		err := ReplaceMakroInCorpusFile(inputFile, outputFile, makrosToReplace, map[string]string{}, alwaysConvertLocalToGlobal, verbose, minify)
		if err != nil {
			if errOut != nil {
				errOut = fmt.Errorf("%w\n%w", errOut, err)
			} else {
				errOut = fmt.Errorf("%w", err)
			}
		}
	}
	if errOut != nil {
		return errOut
	}
	return nil
}
