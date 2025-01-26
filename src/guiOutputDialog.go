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

const LogFile = "Corpus_Macro_Replacer_log.txt"

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

func WriteOutputTask(inputFile string, outputFile string, makrosToReplace map[string]*M1, err *string, alwaysConvertLocalToGlobal bool, verbose bool, minify bool) error {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("panic occured: ", r)
			*err = fmt.Sprintf("⚠ ERROR (panic): %s", inputFile)
		}
	}()
	return ReplaceMakroInCorpusFile(inputFile, outputFile, makrosToReplace, alwaysConvertLocalToGlobal, verbose, minify)
}

func WriteOutput(logData binding.StringList, foundCorpusFiles []string, outputDir string, makroFiles []string, alwaysConvertLocalToGlobal bool, verbose bool, minify bool) {
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
	makrosToReplace, err := ReadMakrosFromCMK(makroFiles)
	currentLog, _ := logData.Get()
	if err != nil {
		log.Println(err)
		currentLog = append(currentLog, fmt.Sprintf("⚠ ERROR: Przerwano, bo: %s", err))
		logData.Set(currentLog)
		return
	}
	if len(makrosToReplace) == 0 {
		log.Println(err)
		currentLog = append(currentLog, fmt.Sprintf("⚠ ERROR: Nie znaleziono żadnych makr: %s", makroFiles))
		logData.Set(currentLog)
		return
	}
	panicErrors := []string{}
	normalErrors := []error{}
	for i, inputFile := range foundCorpusFiles {
		currentLog = append(currentLog, fmt.Sprintf("%d/%d: %s", i+1, len(foundCorpusFiles), inputFile))
		relInputFile, _ := filepath.Rel(outputDir, inputFile)
		cleanedRelInputFile := removeRelativePrefix(relInputFile)
		outputFile := filepath.Join(outputDir, cleanedRelInputFile)
		var specialError *string = new(string)
		err := WriteOutputTask(inputFile, outputFile, makrosToReplace, specialError, alwaysConvertLocalToGlobal, verbose, minify)
		if *specialError != "" {
			panicErrors = append(panicErrors, *specialError)
			currentLog = append(currentLog, fmt.Sprint("⚠ FATAL: '%s'", outputFile))
		}
		if err != nil {
			normalErrors = append(normalErrors, err)
			currentLog = append(currentLog, fmt.Sprintf("⚠ ERROR: '%s' %s", outputFile, err))
		}
		// _, err := io.Copy(&buf, r)
		// fmt.Print(err)
		// buf.ReadString(r)
		// logWindow.Set(append(current, buf.String()))
		logData.Set(currentLog)
	}
	if len(normalErrors) != 0 {
		message := fmt.Sprintf("⚠ W %d plikach wystąpiły błędy", len(normalErrors))
		log.Println(message)
		log.Println(normalErrors)

		currentLog = append(currentLog, message)
		currentLog = append(currentLog, panicErrors...)
	}
	if len(panicErrors) != 0 {
		message := fmt.Sprintf("⚠ W %d plikach wystąpiły nietypowe błędy", len(panicErrors))
		log.Println(message)
		log.Println(panicErrors)

		currentLog = append(currentLog, message)
		currentLog = append(currentLog, panicErrors...)
	}
	// os.Stderr = originalStderr
	// w.Close()
	// r.Close()
	curDir, _ := os.Getwd()
	currentLog = append(currentLog, fmt.Sprintf("Zapisano log: %s", filepath.Join(curDir, LogFile)))
	logData.Set(currentLog)
	os.WriteFile(LogFile, []byte(strings.Join(currentLog, "\n")), 0644)
}

func onTappedOutputPopup(a fyne.App, self widget.ToolbarItem, w fyne.Window) func() {
	var popup *widget.PopUp
	// Define an action when the "Show Pop-up" button is tapped

	return func() {
		foundCorpusFiles := []string{}
		for _, dirOrFile := range loadedFiles {
			foundCorpusFiles = append(foundCorpusFiles, FindCorpusFiles(dirOrFile.path)...)
		}
		outputPath := widget.NewEntry()
		// outputPath.PlaceHolder = CorpusMacroReplacerDefaultPath
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
							alwaysConvertLocalToGlobal := a.Preferences().Bool("alwaysConvertLocalToGlobal")
							verbose := a.Preferences().Bool("minify")
							minify := a.Preferences().Bool("verbose")
							WriteOutput(logData, foundCorpusFiles, outputPath.Text, macrosTochange, alwaysConvertLocalToGlobal, verbose, minify)
							logWindow.Refresh()
						}),
						widget.NewButtonWithIcon("", theme.MoreVerticalIcon(), func() {
							checkMinify := widget.NewCheck("Zmniejsz rozmiar plików (eksperymentalne)", func(b bool) {
								a.Preferences().SetBool("minfy", b)
							})
							checkMinify.Checked = a.Preferences().Bool("minify")

							checkVerbose := widget.NewCheck("Więcej logów (widoczne w terminalu)", func(b bool) {
								a.Preferences().SetBool("verbose", b)
							})
							checkVerbose.Checked = a.Preferences().Bool("verbose")
							popup := dialog.NewCustom("Ustawienia wynikowych plików", "Ok", container.NewVBox(checkMinify, checkVerbose), w)
							popup.Show()
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
