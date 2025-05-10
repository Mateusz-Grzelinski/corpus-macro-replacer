package main

import (
	"cmp"
	"corpus_macro_replacer/corpus"
	"fmt"
	"os"
	"path/filepath"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
)

func NewDialogMakroRecovery(a fyne.App, w fyne.Window, failedMakroName string) *dialog.CustomDialog {
	makroSearchPath := a.Preferences().String("makroSearchPath")
	missingMakroPath := filepath.Join(makroSearchPath, failedMakroName+".CMK")
	card1Result := widget.NewLabel("")
	card1Result.Wrapping = fyne.TextWrapBreak
	card1Result.Importance = widget.DangerImportance
	card1 := widget.NewCard("1. Szukam brakującego pliku w Makrach", "Szukam w "+makroSearchPath, container.NewVBox(card1Result))
	card2Result := widget.NewLabel("")
	card2Result.Wrapping = fyne.TextWrapBreak
	card2Result.Importance = widget.DangerImportance
	card2 := widget.NewCard("2. Odtwarzam brakujące makro z pliku Corpusa", "Szukam w "+cmp.Or(SelectedPath, ""), container.NewVBox(card2Result))
	card2.Hide()
	cardFail := widget.NewCard("Nie udało się naprawić", "", container.NewVBox(
		widget.NewRichTextFromMarkdown(`
- zaznacz inny plik Corpusa uruchom AutoFix jeszcze raz
- autor mógł zmienić nazwę pliku CMK (duże i małe litery mają znaczenie)
- autor mógł zmienić nazwę makra w pliku MakroCollection.dat (do edycji z poziomu corpusa: )
- plik może już nie istnieć
- edycja nazwy makra nie jest wspierana`),
	))
	cardFail.Hide()
	cardOk := widget.NewCard("Udało się naprawić", "", container.NewVBox(
		widget.NewRichTextFromMarkdown("Makro może mieć niepożadaną (starą) zawartość, polecam je przejrzeć."),
		widget.NewRichTextFromMarkdown("Teraz makro będzie zaczytane ponownie, ale może wyskoczyć kolejny błąd. Klikaj AutoFix do skutku"),
	))
	cardOk.Hide()
	filterDialog := dialog.NewCustom("Szukam brakującego makra", "Zamknij", container.NewVBox(card1, card2, cardFail, cardOk), w)
	continueRecovery := true
	if continueRecovery {
		failedFilename := filepath.Base(failedMakroName)
		foundFile, _ := corpus.FindFile(makroSearchPath, failedFilename+".CMK")
		if foundFile != "" {
			err := corpus.CopyFile(foundFile, missingMakroPath)
			if err != nil {
				card1Result.SetText(fmt.Sprintf("Znaleziono: \"%s\" i ale nie udało się skopiować do \"%s\": %s", foundFile, missingMakroPath, err))
			} else {
				card1Result.SetText(fmt.Sprintf("Znaleziono: \"%s\" i skopiowano do \"%s\"", foundFile, missingMakroPath))
				card1Result.Importance = widget.HighImportance
			}
			continueRecovery = false
		} else {
			card1Result.SetText(fmt.Sprintf("Nie znaleziono: \"%s\" w \"%s\"", failedMakroName+".CMK", makroSearchPath))
		}
	}

	if continueRecovery {
		var elementFile *corpus.ElementFile
		if loadedE3DFileForPreview != nil {
			elementFile = loadedE3DFileForPreview
		} else if loadedS3DFileForPreview != nil {
			elementFile = &loadedS3DFileForPreview.ElementFile
		}
		if elementFile != nil {
			// todo there is no way to break early the visit walk
			elementFile.VisitElementsAndSubelements(func(e *corpus.Element) {
				for _, s := range e.Elinks.Spoj {
					s.Makro1.VisitSubmakros(func(parent *corpus.M1, embededParent *corpus.M1EmbeddedMakro, child *corpus.M1EmbeddedMakro) {
						if !continueRecovery {
							return // lame version of break early
						}
						makroName := ""
						if child != nil {
							makroName = child.CalledWith()
						} else {
							makroName = parent.MakroName
						}
						if makroName != failedMakroName {
							return
						}
						f, err := os.Create(missingMakroPath)
						if err != nil {
							card2Result.SetText(fmt.Sprintf("otwarty plik Corpusa ma makro \"%s\", ale nie można go zapisać: %s", failedMakroName, err))
						} else {
							_, err1 := f.Write([]byte(fmt.Sprintf("// CorpusMakroReplacer: odzyskano z %s\n", SelectedPath)))
							err := s.Makro1.Save(f)
							if err1 != nil && err != nil {
								card2Result.SetText(fmt.Sprintf("odzyskano Makro \"%s\" z Corpusa, ale wystąpił błąd przy zapisywaniu: %s", failedMakroName, err))
							} else {
								card2Result.SetText(fmt.Sprintf("odzyskano Makro \"%s\" z Corpusa i zapisano: %s. Zawartość pliku może być niekatualna.", failedMakroName, missingMakroPath))
								card2Result.Importance = widget.HighImportance
								continueRecovery = false
							}
						}
					})
				}
			})
			if continueRecovery {
				card2Result.SetText(fmt.Sprintf("nie znaleziono Makra \"%s\" w otwartym pliku Corpusa", failedMakroName))
			}
		} else {
			card2Result.SetText(fmt.Sprintf("otwórz plik corpusa aby poszukać w nim Makra: \"%s\"", failedMakroName))
		}
		card2.Show()
	}
	if continueRecovery {
		cardFail.Show()
	} else {
		// makroErrorLabel.Hide()
		// autoFixError.Hide()
		cardOk.Show()
	}
	filterDialog.Resize(DialogSizeDefault)
	return filterDialog
}
