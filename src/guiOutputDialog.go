package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

func NewPopupListOfFiles(w fyne.Window, files []string) *widget.List {
	return widget.NewList(
		func() int {
			return len(files)
		},
		func() fyne.CanvasObject {
			label := widget.NewLabel("")
			label.Truncation = fyne.TextTruncateEllipsis
			label.Alignment = fyne.TextAlignLeading
			return container.NewBorder(nil, nil,
				widget.NewIcon(theme.FileIcon()),
				widget.NewButtonWithIcon("", theme.ContentCopyIcon(), func() {
					w.Clipboard().SetContent(label.Text)
				}),
				label,
			)
		},
		func(i widget.ListItemID, o fyne.CanvasObject) {
			hbox := o.(*fyne.Container).Objects
			l := hbox[0].(*widget.Label)
			l.SetText(files[i])
			// b := hbox[2].(*widget.Button)
			// b.OnTapped = func() {
			// 	files = RemoveIndex(files, i)
			// 	ListOfLoadedFilesContainer.Refresh()
			// }
		})
}

func removeRelativePrefix(path string) string {
	for strings.HasPrefix(path, "..\\") {
		path = strings.TrimPrefix(path, "..\\")
	}
	return path
}

func WriteOutputTask(inputFile string, outputFile string, makrosToReplace map[string]*M1, err *string) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("panic occured: ", r)
			*err = fmt.Sprintf("⚠ ERROR: %s", inputFile)
		}
	}()
	ReplaceMakroInCorpusFile(inputFile, outputFile, makrosToReplace)
}

const LogFile = "Corpus_Macro_Replacer_log.txt"

func WriteOutput(logData binding.StringList, foundCorpusFiles []string, outputDir string, makroFiles []string) {
	log.Printf("Generating output to: %s", outputDir)
	// originalStderr := os.Stderr
	// // Create a pipe to capture stderr
	// r, w, err := os.Pipe()
	// if err != nil {
	// 	log.Printf("error creating pipe: %s", err)
	// }
	// os.Stderr = w

	// todo make panic handler to write to stderr

	// var buf bytes.Buffer
	makrosToReplace := ReadMakrosFromCMK(makroFiles)
	errored := []string{}
	for i, inputFile := range foundCorpusFiles {
		currentLog, _ := logData.Get()
		currentLog = append(currentLog, fmt.Sprintf("%d/%d: %s", i+1, len(foundCorpusFiles), inputFile))
		relInputFile, _ := filepath.Rel(outputDir, inputFile)
		cleanedRelInputFile := removeRelativePrefix(relInputFile)
		outputFile := filepath.Join(outputDir, cleanedRelInputFile)
		var err *string = new(string)
		WriteOutputTask(inputFile, outputFile, makrosToReplace, err)
		if *err != "" {
			errored = append(errored, *err)
		} else {
			currentLog = append(currentLog, "Zapinano: "+outputFile)
		}
		// _, err := io.Copy(&buf, r)
		// fmt.Print(err)
		// buf.ReadString(r)
		// logWindow.Set(append(current, buf.String()))
		logData.Set(currentLog)
	}
	currentLog, _ := logData.Get()
	if len(errored) != 0 {
		message := fmt.Sprintf("⚠ W %d plikach wystąpiły błędy", len(errored))
		log.Println(message)
		log.Println(errored)

		currentLog = append(currentLog, message)
		currentLog = append(currentLog, errored...)
	}
	// os.Stderr = originalStderr
	// w.Close()
	// r.Close()
	curDir, _ := os.Getwd()
	currentLog = append(currentLog, fmt.Sprintf("Zapisano log: %s", filepath.Join(curDir, LogFile)))
	logData.Set(currentLog)
	os.WriteFile(LogFile, []byte(strings.Join(currentLog, "\n")), 0644)
}

func onTappedOutputPopup(self widget.ToolbarItem, w fyne.Window) func() {
	var popup *widget.PopUp
	// Define an action when the "Show Pop-up" button is tapped

	return func() {
		foundCorpusFiles := []string{}
		for _, dirOrFile := range loadedFiles {
			foundCorpusFiles = append(foundCorpusFiles, FindCorpusFiles(dirOrFile.path)...)
		}
		outputPath := widget.NewEntry()
		outputPath.PlaceHolder = CorpusMacroReplacerDefaultPath
		outputPath.SetText(CorpusMacroReplacerDefaultPath + time.Now().Format("2006-01-02"))
		logData := binding.NewStringList()
		// logList :=
		logData.Set([]string{`Wciśnij Wykonaj aby uruchomić. Pliki zostaną nadpisane`})
		listWidget := widget.NewListWithData(logData,
			func() fyne.CanvasObject {
				label := widget.NewLabel("") // Template for list items
				label.Wrapping = fyne.TextWrapBreak
				return container.NewBorder(nil, nil, nil, widget.NewButtonWithIcon("", theme.ContentCopyIcon(), func() {
					w.Clipboard().SetContent(label.Text)
				}), label)
			},
			func(data binding.DataItem, item fyne.CanvasObject) {
				con := item.(*fyne.Container)
				str := data.(binding.String) // Get the bound string
				label := con.Objects[0].(*widget.Label)
				label.Bind(str)
			},
		)

		logWindow := container.NewBorder(nil, nil, nil, nil, container.NewHScroll(listWidget))
		// logWindow := widget.NewLabelWithData(logData)
		content := container.NewVSplit(
			container.NewBorder(
				widget.NewLabel(fmt.Sprintf("Znaleziono pliki: %d", len(foundCorpusFiles))),
				nil, nil, nil,
				NewPopupListOfFiles(w, foundCorpusFiles),
			),
			container.NewBorder(
				container.NewBorder(nil, nil,
					widget.NewLabel("Zapisz wybrane pliki do:"),
					container.NewHBox(
						widget.NewButtonWithIcon("", theme.FolderOpenIcon(), func() {
							fileOpenDialog := dialog.NewFolderOpen(func(lu fyne.ListableURI, err error) {
								if err != nil {
									dialog.ShowError(err, w)
									return
								}
								if lu == nil {
									return
								}
								path := lu.Path()
								outputPath.SetText(path)
							}, w)
							fileOpenDialog.Resize(DialogSizeDefault)
							fileOpenDialog.Show()
						}),
						widget.NewButtonWithIcon("Wykonaj", theme.MediaPlayIcon(), func() {
							macrosTochange := []string{}
							for _, e := range MacrosToChangeEntries {
								macrosTochange = append(macrosTochange, e.Text)
							}
							WriteOutput(logData, foundCorpusFiles, outputPath.Text, macrosTochange)
							logWindow.Refresh()
						}),
					),
					outputPath,
				), nil, nil, nil,
				container.NewScroll(logWindow),
			),
		)

		// Create a container for the pop-up content, including a "Close" button
		popUpContent := container.NewBorder(nil,
			widget.NewButton("Close", func() {
				popup.Hide() // Function to hide the pop-up
			}), nil, nil,
			content,
		)

		popup = widget.NewModalPopUp(popUpContent, w.Canvas())
		popup.Resize(fyne.NewSize(w.Canvas().Size().Width*0.9, w.Canvas().Size().Height*0.9))
		popup.Show()
	}
}
