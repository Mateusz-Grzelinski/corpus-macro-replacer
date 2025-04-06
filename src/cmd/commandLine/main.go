package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"corpus_macro_replacer/corpus"
)

//go:generate fyne bundle -o src/bundled.go assets/

const Version = "0.7"

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
	var verbose *bool = flag.Bool("verbose", false, "print more output")
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
		fmt.Printf("Corpus_Macro_Replacer v%s\n", Version)
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

	statInput, errInput := os.Stat(*input)
	if errInput != nil {
		log.Fatalf("input '%s' is invalid: %s", *input, errInput)
	}
	statOutput, errOutput := os.Stat(*output)
	if errOutput == nil && !statOutput.IsDir() && !*force {
		log.Fatalf("output %s already exists. Add --force to override", *output)
	}
	if statInput.IsDir() {
		corpus.ReplaceMakroInCorpusFolder(*input, *output, makroFiles, *alwaysConvertLocalToGlobal, *verbose, *minify) // todo support from cmd line all options
	} else {
		if errOutput == nil && statOutput.IsDir() || !strings.HasSuffix(strings.ToLower(*output), ".e3d") {
			var newOutput string = filepath.Join(*output, filepath.Base(*input))
			output = &newOutput
		}
		_, errNewOutput := os.Stat(*output)
		if errNewOutput == nil && !*force {
			log.Fatalf("output %s already exists. Add --force to override", *output)
		}

		makrosToReplace, err := corpus.ReadMakrosFromCMK(makroFiles, nil, nil)
		if err != nil {
			log.Fatalf("can not read makros: %s", err)
		}
		corpus.ReplaceMakroInCorpusFile(*input, *output, makrosToReplace, map[string]string{}, *alwaysConvertLocalToGlobal, *verbose, *minify)
		// todo support from cmd line all options
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
