package main

import (
	"encoding/xml"
	"flag"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

//go:generate fyne bundle -o src/bundled.go assets

var Settings ProgramSettings = ProgramSettings{minify: false, alwaysConvertLocalToGlobal: false, verbose: true, hideElementsWithZeroMacros: true}

func ReadMakrosFromCMK(makroFiles []string) map[string]*M1 {
	makrosToReplace := map[string]*M1{}
	for _, makroFile := range makroFiles {
		absPathMakroFile := strings.SplitN(filepath.Base(makroFile), ".", 2)[0] // this name might be wrong, it can be redefined in software
		_, exists := makrosToReplace[absPathMakroFile]
		if exists {
			log.Fatalf("Makro path seems to be duplicated: '%s' (all paths: %s)", absPathMakroFile, makroFiles)
		}
		makro, err := LoadMakroFromCMKFile(makroFile)
		if err != nil {
			log.Fatal(err)
		}
		makrosToReplace[absPathMakroFile] = makro
	}
	return makrosToReplace
}

// func updateElements()

func ReplaceMakroInCorpusFile(inputFile string, outputFile string, makrosToReplace map[string]*M1) {
	err := os.MkdirAll(filepath.Dir(outputFile), os.ModePerm)
	if err != nil {
		log.Fatalf("Can not create path: %s", outputFile)
	}

	macrosUpdated := 0
	macrosSkipped := 0

	// todo how to handle 2 very similar types??? ElementFile vs ProjectFile
	visitCorpusE3DFile := func(decoder *xml.Decoder, start xml.StartElement) xml.Token {
		var root ElementFile
		decoder.Strict = true
		err := decoder.DecodeElement(&root, &start)
		decoder.Strict = false
		if err != nil {
			log.Printf("%s: %s", inputFile, err)
		}
		for _, element := range root.Element {
			visitedDaske := []string{}
			updatedDaske := map[string]int{}
			skippedDaske := map[string]int{}
			for _, spoj := range element.Elinks.Spoj {
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
				UpdateMakro(&oldMakro, newMakro, Settings.alwaysConvertLocalToGlobal)

				// todo reorder variables so that ones with the same name are next to each other
				// oldVariablesKeys, oldValues, _ := loadValuesFromSection(oldMacro.Varijable.DAT)
				// newVariablesKeys, newValues, newVariablesComments := loadValuesFromSection(newMacro.Varijable.DAT)
				macrosUpdated++
				updatedDaske[daskeName]++
			}
			if Settings.verbose {
				log.Printf("  Cabinet '%s'\n", element.EName.Value)
				for _, name := range visitedDaske {
					log.Printf("    Updated %d macros, %d skipped in plate '%s'\n", updatedDaske[name], skippedDaske[name], name)
				}
			}
		}
		log.Printf("  Summary: updated %d macros, %d skipped\n", macrosUpdated, macrosSkipped)

		return root
	}

	visitCorpusS3DFile := func(decoder *xml.Decoder, start xml.StartElement) xml.Token {
		var root ProjectFile
		decoder.Strict = true
		err := decoder.DecodeElement(&root, &start)
		decoder.Strict = false
		if err != nil {
			log.Printf("%s: %s", inputFile, err)
		}
		for _, element := range root.Element {
			visitedDaske := []string{}
			updatedDaske := map[string]int{}
			skippedDaske := map[string]int{}
			for _, spoj := range element.Elinks.Spoj {
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
				UpdateMakro(&oldMakro, newMakro, Settings.alwaysConvertLocalToGlobal)
				macrosUpdated++
				updatedDaske[daskeName]++
			}
			if Settings.verbose {
				log.Printf("  Cabinet '%s'\n", element.EName.Value)
				for _, name := range visitedDaske {
					log.Printf("    Updated %d macros, %d skipped in plate '%s'\n", updatedDaske[name], skippedDaske[name], name)
				}
			}
		}
		log.Printf("  Summary: updated %d macros, %d skipped\n", macrosUpdated, macrosSkipped)

		return root
	}

	ReadWriteCorpusFile(inputFile, outputFile, Settings.minify, visitCorpusE3DFile, visitCorpusS3DFile)
}

func FindCorpusFiles(inputFolder string) []string {
	foundCorpusFiles := []string{}
	filepath.Walk(inputFolder, func(path string, info fs.FileInfo, err error) error {
		if info != nil && !info.IsDir() && isCorpusExtension(info.Name()) {
			foundCorpusFiles = append(foundCorpusFiles, path)
		}
		return nil
	})
	return foundCorpusFiles
}

func ReplaceMakroInCorpusFolder(inputFolder string, outputFolder string, makroFiles []string) {
	inputFolderStat, err := os.Stat(inputFolder)
	if err != nil {
		log.Fatal(err)
	}
	if !inputFolderStat.IsDir() {
		log.Fatalf("%s must be a directory", inputFolder)
	}
	err = os.MkdirAll(outputFolder, os.ModePerm)
	if err != nil {
		log.Fatal(err)
	}

	foundCorpusFiles := FindCorpusFiles(inputFolder)

	log.Printf("Found %d files in %s", len(foundCorpusFiles), inputFolder)

	makrosToReplace := ReadMakrosFromCMK(makroFiles)
	for _, inputFile := range foundCorpusFiles {
		relInputFile, _ := filepath.Rel(inputFolder, inputFile)
		outputFile := filepath.Join(outputFolder, relInputFile)
		ReplaceMakroInCorpusFile(inputFile, outputFile, makrosToReplace)
	}
}

type arrayFlags []string

// String is an implementation of the flag.Value interface
func (i *arrayFlags) String() string {
	return fmt.Sprintf("%v", *i)
}

// Set is an implementation of the flag.Value interface
func (i *arrayFlags) Set(value string) error {
	*i = append(*i, value)
	return nil
}

func main() {
	flag.Usage = func() {
		w := flag.CommandLine.Output()
		fmt.Fprint(w, `This program is used to update makro in Copus (.E3D) files. 
It is alternative to doule ticks in macro editor that actually works: 
- it does not edit [JOINT] section
- does a smart merge on [VARIJABLE] section, see README: https://github.com/Mateusz-Grzelinski/corpus-macro-replacer
`)
		fmt.Fprintf(w, "Usage of %s -input <PATH> -output <PATH> -makro <PATH>:\n", os.Args[0])
		flag.PrintDefaults()
	}
	var version *bool = flag.Bool("v", false, "print version")
	var input *string = flag.String("input", "", "required. File or dir, must exist. If dir then changes macro recursively for all .E3D files.")
	var output *string = flag.String("output", "", `required. File or dir, does not need to exist. 
If input is dir then output must be dir, but will be created if does not exist. Directory structure of input is mirrored.
If input is file the output can be file (must end with .E3D) or directory.`)
	var makroFiles arrayFlags
	flag.Var(&makroFiles, "makro", `required. Path to macro that should be replaced. Can be specified multiple times. Usually one of files in "C:\Tri D Corpus\Corpus 5.0\Makro"`)
	var force *bool = flag.Bool("force", false, `default: false. Specify to override file specified in -output`)
	var minify *bool = flag.Bool("minify", false, `default: false. Reduce file size by deleting spaces, (~7% size reduction)`)
	var alwaysConvertLocalToGlobal *bool = flag.Bool("alwaysConvertLocalToGlobal", false, `default: false. Global variable start with "_" prefix - it takes value from "evar". 
Default logic allows adding "_" prefix to variables that consists only from integers (no if statements, no +-* operations). It prevents from erasing your custom logic.`)

	flag.Parse()

	if *version {
		fmt.Println("Corpus_Macro_Replacer v0.2")
		os.Exit(0)
	}
	if *input == "" && *output == "" && len(makroFiles) == 0 {
		RunGui()
		os.Exit(0)
	}
	if *input == "" {
		log.Fatalln("-input can not be empty")
	}
	if *output == "" {
		log.Fatalln("-output can not be empty")
	}
	if len(makroFiles) == 0 {
		log.Fatalln("-makroFile can not be empty")
	}

	Settings.minify = *minify
	Settings.alwaysConvertLocalToGlobal = *alwaysConvertLocalToGlobal

	statInput, errInput := os.Stat(*input)
	if errInput != nil {
		log.Fatalf("input '%s' is invalid: %s", *input, errInput)
	}
	statOutput, errOutput := os.Stat(*output)
	if errOutput == nil && !statOutput.IsDir() && !*force {
		log.Fatalf("output %s already exists. Add --force to override", *output)
	}
	if statInput.IsDir() {
		ReplaceMakroInCorpusFolder(*input, *output, makroFiles)
	} else {
		if errOutput == nil && statOutput.IsDir() || !strings.HasSuffix(strings.ToLower(*output), ".e3d") {
			var newOutput string = filepath.Join(*output, filepath.Base(*input))
			output = &newOutput
		}
		_, errNewOutput := os.Stat(*output)
		if errNewOutput == nil && !*force {
			log.Fatalf("output %s already exists. Add --force to override", *output)
		}

		makrosToReplace := ReadMakrosFromCMK(makroFiles)
		ReplaceMakroInCorpusFile(*input, *output, makrosToReplace)
	}

	// makroFile := `C:\Tri D Corpus\Corpus 5.0\Makro\custom.CMK`
	// inputFile := `C:\Tri D Corpus\Corpus 5.0\elmsav\_modifications\simple_original_custom_v1.E3D`
	// outputFile := `C:\Tri D Corpus\Corpus 5.0\elmsav\_modifications\simple_original_custom_v1_output.E3D`
	// replaceMakroInCorpusE3DFile(inputFile, outputFile, makroFile)

	// inputFolder := `C:\Tri D Corpus\Corpus 5.0\elmsav\_modifications`
	// outputFolder := `C:\Tri D Corpus\Corpus 5.0\elmsav\_modifications_output`
	// replaceMakroInCorpusE3DFolder(inputFolder, outputFolder, makroFile)

	// m := LoadMakroFromCMKFile(makroFile)
	// fmt.Println("writing output!")
	// out, err := os.Create("out.xml")
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// // enc := transform.NewWriter(out, charmap.Windows1250.NewDecoder())
	// mybytes, err := xml.MarshalIndent(&m, "", "  ")
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// out.Write(mybytes)
}
