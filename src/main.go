package main

import (
	"encoding/xml"
	"flag"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"
)

var Settings ProgramSettings = ProgramSettings{minify: false, alwaysConvertLocalToGlobal: false}

func replaceMakroInCorpusE3DFile(inputFile string, outputFile string, makroFile string) {
	newMakro := LoadMakroFromCMKFile(makroFile)

	err := os.MkdirAll(filepath.Dir(outputFile), os.ModePerm)
	if err != nil {
		log.Fatalf("Can not create path: %s", outputFile)
	}

	macrosUpdated := 0
	macrosSkipped := 0
	handleCorpusFile(inputFile, outputFile, Settings.minify, func(decoder *xml.Decoder, start xml.StartElement) xml.Token {
		var oldMakro M1
		decoder.Strict = true
		e := decoder.DecodeElement(&oldMakro, &start)
		decoder.Strict = false
		if e != nil {
			log.Fatal(e)
		}
		if oldMakro.MakroName == newMakro.MakroName {
			macrosUpdated++
			UpdateMakro(&oldMakro, newMakro, Settings.alwaysConvertLocalToGlobal)
			return newMakro
		} else {
			macrosSkipped++
		}
		return oldMakro
	})
	log.Printf("  Updated %d macros, %d skipped\n", macrosUpdated, macrosSkipped)
}

func replaceMakroInCorpusE3DFolder(inputFolder string, outputFolder string, makroFile string) {
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

	foundCorpusFiles := []string{}
	filepath.Walk(inputFolder, func(path string, info fs.FileInfo, err error) error {
		if !info.IsDir() && strings.HasSuffix(strings.ToLower(info.Name()), ".e3d") {
			foundCorpusFiles = append(foundCorpusFiles, path)
		}
		return nil
	})
	log.Printf("Found %d files in %s", len(foundCorpusFiles), inputFolder)

	for _, inputFile := range foundCorpusFiles {
		relInputFile, _ := filepath.Rel(inputFolder, inputFile)
		outputFile := filepath.Join(outputFolder, relInputFile)
		replaceMakroInCorpusE3DFile(inputFile, outputFile, makroFile)
		// err = os.MkdirAll(filepath.Dir(outputFile), os.ModePerm)
		// if err != nil {
		// 	log.Printf("error when creating dir: %s: %s", filepath.Dir(outputFile), err)
		// }
	}
}

func main() {
	flag.Usage = func() {
		w := flag.CommandLine.Output()
		fmt.Fprintln(w, `This program is used to update makro in Copus (.E3D) files. 
It is alternative to doule ticks in macro editor that actually works: 
- it does not edit [JOINT] section
- does a smart merge on [VARIJABLE] section, see README
`)
		fmt.Fprintf(w, "Usage of %s:\n", os.Args[0])
		flag.PrintDefaults()
	}
	var version *bool = flag.Bool("v", false, "print version")
	var input *string = flag.String("input", "", "required. File or dir, must exist. If dir then changes macro recursively for all .E3D files.")
	var output *string = flag.String("output", "", `required. File or dir, does not need to exist. 
If input is dir then output must be dir, but will be created if does not exist. Directory structure of input is mirrored.
If input is file the output can be file (must end with .E3D) or directory.`)
	var makroFile *string = flag.String("makro", "", `required. Path to macro that should be replaced. Usually one of files in "C:\Tri D Corpus\Corpus 5.0\Makro"`)
	var force *bool = flag.Bool("force", false, `default: false. Specify to override file specified in -output`)
	var minify *bool = flag.Bool("minify", false, `default: false. Reduce file size by deleting spaces, (~7% size reduction)`)
	var alwaysConvertLocalToGlobal *bool = flag.Bool("alwaysConvertLocalToGlobal", false, `default: false. Global variable start with "_" prefix - it takes value from "evar". 
Default logic allows adding "_" prefix to variables that consists only from integers (no if statements, no +-* operations). It prevents from erasing your custom logic.`)

	flag.Parse()

	if *version {
		fmt.Println("Corpus_Macro_Replacer v0.2")
		os.Exit(0)
	}
	if *input == "" {
		log.Fatalln("-input can not be empty")
	}
	if *output == "" {
		log.Fatalln("-output can not be empty")
	}
	if *makroFile == "" {
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
		replaceMakroInCorpusE3DFolder(*input, *output, *makroFile)
	} else {
		if errOutput == nil && statOutput.IsDir() || !strings.HasSuffix(strings.ToLower(*output), ".e3d") {
			var newOutput string = filepath.Join(*output, filepath.Base(*input))
			output = &newOutput
		}
		_, errNewOutput := os.Stat(*output)
		if errNewOutput == nil && !*force {
			log.Fatalf("output %s already exists. Add --force to override", *output)
		}
		replaceMakroInCorpusE3DFile(*input, *output, *makroFile)
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
