package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"corpus_macro_replacer/corpus"

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

func WriteOutputTask(inputFile string, outputFile string, makrosToReplace map[string]*corpus.M1, makroRename map[string]string, err *string, alwaysConvertLocalToGlobal bool, verbose bool, minify bool) error {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("panic occured: ", r)
			*err = fmt.Sprintf("💀 FATAL: %s: %s", inputFile, r)
		}
	}()
	return corpus.ReplaceMakroInCorpusFile(inputFile, outputFile, makrosToReplace, makroRename, alwaysConvertLocalToGlobal, verbose, minify)
}

func WriteOutput(
	logData binding.StringList,
	foundCorpusFiles []string,
	outputDir string,
	makroFiles []string,
	makroNamesOverrides []*string,
	makroOldNameToNewName map[string]string,
	alwaysConvertLocalToGlobal bool,
	verbose bool,
	minify bool,
	makroRootPath *string,
	makroMapping corpus.MakroMappings,
) {
	// todo someday: capture stdout and print to log window?
	log.Printf("Generating output to: %s", outputDir)
	// todo makro names are lost...
	makrosToReplace, err := corpus.ReadMakrosFromCMK(makroFiles, makroNamesOverrides, makroRootPath, makroMapping)

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
	normalErrors := []string{}
	for i, inputFile := range foundCorpusFiles {
		currentLog = append(currentLog, fmt.Sprintf("%d/%d: %s", i+1, len(foundCorpusFiles), inputFile))
		outputFile := corpus.GetCleanOutputpath(outputDir, inputFile)
		var panicErrorToReport *string = new(string)
		err := WriteOutputTask(inputFile, outputFile, makrosToReplace, makroOldNameToNewName, panicErrorToReport, alwaysConvertLocalToGlobal, verbose, minify)
		if *panicErrorToReport != "" {
			panicErrors = append(panicErrors, *panicErrorToReport)
			currentLog = append(currentLog, *panicErrorToReport)
		}
		if err != nil {
			normalErrorMessage := fmt.Sprintf("⚠ ERROR: '%s' %s", outputFile, err)
			normalErrors = append(normalErrors, normalErrorMessage)
			currentLog = append(currentLog, normalErrorMessage)
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
		currentLog = append(currentLog, normalErrors...)
	}
	if len(panicErrors) != 0 {
		message := fmt.Sprintf("💀 W %d plikach wystąpiły nietypowe błędy (panic error)", len(panicErrors))
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
			foundCorpusFiles = append(foundCorpusFiles, corpus.FindCorpusFiles(dirOrFile.path)...)
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
							macroFilesTochange := []string{}
							macrosToRename := map[string]string{}
							for i, e := range MacrosToChangeEntries {
								macroFilesTochange = append(macroFilesTochange, e.Text)
								if MacrosToChangeReNamesEntriesBool[i].Checked {
									macrosToRename[MacrosToChangeNamesEntries[i].Text] = MacrosToChangeReNamesEntries[i].Text
								}
							}
							macroNamesOverrides := []*string{}
							for _, e := range MacrosToChangeNamesEntries {
								name := string(e.Text)
								macroNamesOverrides = append(macroNamesOverrides, &name)
							}
							alwaysConvertLocalToGlobal := a.Preferences().Bool("alwaysConvertLocalToGlobal")
							verbose := a.Preferences().Bool("verbose")
							minify := a.Preferences().Bool("minify")
							makroRootPath := a.Preferences().String("makroSearchPath")
							WriteOutput(logData, foundCorpusFiles, outputPath.Text, macroFilesTochange, macroNamesOverrides, macrosToRename, alwaysConvertLocalToGlobal, verbose, minify, &makroRootPath, corpus.MakroCollectionCache.GetMakroMappings())
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
