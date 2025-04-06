package main

import (
	"cmp"
	"corpus_macro_replacer/corpus"
	"fmt"
	"log"
	"path/filepath"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

type MacroContainer struct {
	widget.BaseWidget
	all           *fyne.Container
	contentHeader *fyne.Container
	contentRead   *fyne.Container
	contentDiff   *fyne.Container
	isOpen        bool
	parentPlate   *PlateContainer
	oldMakro      *corpus.M1
	newMakro      *corpus.M1

	// for reference when updating
	openButton          *widget.Button
	stats               *widget.Label
	loadThisMacroButton *widget.Button
}

func NewMacroContainer(nestLevel int, parentPlate *PlateContainer) *MacroContainer {
	mc := &MacroContainer{
		all:           nil,
		contentHeader: nil,
		contentRead:   nil,
		contentDiff:   nil,
		isOpen:        false,
		parentPlate:   parentPlate,
	}
	var openButton, loadThisMacroButton *widget.Button
	contentDiff := container.NewVBox(
		container.NewVBox(),          // 0 Varijable
		container.NewVBox(),          // 1 joint
		container.NewVBox(),          // 2 formule
		container.NewVBox(),          // 3 pocket
		container.NewVBox(),          // 4 potrosni
		container.NewVBox(),          // 5 grupa
		container.NewVBox(),          // 6 raster
		container.NewVBox(),          // 7 makro
		container.NewWithoutLayout(), // 8 something like "go to the top"
		container.NewWithoutLayout(),
	)
	contentRead := container.NewVBox(
		container.NewVBox(),          // 0 Varijable
		container.NewVBox(),          // 1 joint
		container.NewVBox(),          // 2 formule
		container.NewVBox(),          // 3 pocket
		container.NewVBox(),          // 4 potrosni
		container.NewVBox(),          // 5 grupa
		container.NewVBox(),          // 6 raster
		container.NewVBox(),          // 7 makro
		container.NewWithoutLayout(), // 8 something like "go to the top"
		container.NewWithoutLayout(),
	)

	openButton = widget.NewButtonWithIcon("", theme.VisibilityIcon(), func() {})
	openButton.OnTapped = func() {
		if mc.isOpen {
			openButton.Icon = theme.VisibilityIcon()
			mc.contentRead.Hide()
			mc.contentDiff.Hide()
		} else {
			openButton.Icon = theme.VisibilityOffIcon()
			if mc.newMakro == nil {
				mc.contentRead.Show()
				mc.contentDiff.Hide()
			} else {
				mc.contentRead.Hide()
				mc.contentDiff.Show()
			}
		}
		openButton.Refresh()
		mc.isOpen = !mc.isOpen
	}
	mc.openButton = openButton
	loadThisMacroButton = widget.NewButtonWithIcon("Zamień to makro", theme.NavigateNextIcon(), func() {
		if mc.oldMakro == nil {
			return
		}
		MacrosDefaultPathNormal := fyne.CurrentApp().Preferences().String("makroSearchPath")
		oldMakroNameWithExtension := mc.oldMakro.MakroName + ".CMK"
		macroFileName := cmp.Or(corpus.MakroCollectionCache.GetMacroFileNameByName(mc.oldMakro.MakroName), &oldMakroNameWithExtension)
		macroGuessedPath := filepath.Join(MacrosDefaultPathNormal, *macroFileName)
		addToLoadedFilesAndRefresh(SelectedPath)
		for _, makroTochangeName := range MacrosToChangeNamesEntries {
			if makroTochangeName.Text == mc.oldMakro.MakroName {
				return
			}
		}
		needCreation := true
		for i, makroToChangeEntry := range MacrosToChangeEntries {
			if makroToChangeEntry.Text == "" {
				makroToChangeEntry.SetText(macroGuessedPath)
				MacrosToChangeNamesEntries[i].SetText(mc.oldMakro.MakroName)
				needCreation = false
				break
			}
		}
		if needCreation {
			AddMakroButton.OnTapped()
			MacrosToChangeEntries[len(MacrosToChangeEntries)-1].SetText(macroGuessedPath)
			MacrosToChangeNamesEntries[len(MacrosToChangeNamesEntries)-1].SetText(mc.oldMakro.MakroName)
		}
		contentRead.Refresh()
		refreshCorpusPreviewFunc()
	},
	)
	mc.loadThisMacroButton = loadThisMacroButton
	leftPadding := canvas.NewRectangle(nil)                        // Empty rectangle
	leftPadding.SetMinSize(fyne.NewSize(float32(20*nestLevel), 0)) // 20 units wide
	h := container.NewHBox(leftPadding, openButton)

	mc.contentHeader = container.NewBorder(nil, nil,
		h,
		loadThisMacroButton,
	)
	mc.contentRead = contentRead
	mc.contentDiff = contentDiff
	mc.all = container.NewVBox(mc.contentHeader, mc.contentRead, mc.contentDiff)
	mc.contentRead.Hide()
	mc.contentDiff.Hide()

	mc.ExtendBaseWidget(mc)
	return mc
}

func smartTextComparison(changes []corpus.Change) (string, string) {
	var oldReformatted strings.Builder
	var newReformatted strings.Builder

	// first some stats
	hasSameValues := 0
	for _, change := range changes {
		switch change.Result {
		case corpus.ValueSame:
			hasSameValues++
		}
	}
	if hasSameValues > 0 {
		newReformatted.WriteString(fmt.Sprintf("Schowano %d takie same wartości\n\n", hasSameValues))
		oldReformatted.WriteString("\n")
	}

	// actual diff
	for _, change := range changes {
		switch change.Result {
		case corpus.ValueAdded:
			// oldReformatted.WriteString("\n")
			newReformatted.WriteString(fmt.Sprintf("%s=%s \t// nowa zmienna dodana\n", *change.NewName, change.NewValue))
		case corpus.ValueSame:
			oldReformatted.WriteString(fmt.Sprintf("%s=%s\n", *change.OldName, change.OldValue))
			newReformatted.WriteString("\n")
		default:
			oldReformatted.WriteString(fmt.Sprintf("%s=%s\n", *change.OldName, change.OldValue))
			newReformatted.WriteString(fmt.Sprintf("%s=%s \t// zachowano starą wartość (wartość obecna:%s)\n", *change.NewName, *&change.OldValue, change.NewValue))
		}
	}

	return oldReformatted.String(), newReformatted.String()
}

func NewRichTextFromCMK(prefix string, a *corpus.GenericNodeWithDat) *widget.RichText {
	if a == nil {
		return NewRichTextFromCode(prefix, "\n")
	}
	return NewRichTextFromCode(prefix, strings.Join(corpus.DecodeAllCMKLines(a.DAT), "\n")+"\n")
}

func NewRichTextFromCode(prefix string, code string) *widget.RichText {
	newText := fmt.Sprintf("```\n%s%s```", prefix, code)
	newRichText := widget.NewRichTextFromMarkdown(newText)
	newRichText.Wrapping = fyne.TextWrap(fyne.TextTruncateClip)
	return newRichText
}

func NewHSplitFromCMK(leading fyne.CanvasObject, trailing fyne.CanvasObject) *container.Split {
	split := container.NewHSplit(leading, trailing)
	split.Offset = 0.4
	return split
}

func (mc *MacroContainer) UpdateMacroForDiff(newMakro *corpus.M1, compact bool) {
	mc.newMakro = newMakro
	oldMakro := mc.oldMakro
	if newMakro == nil {
		mc.Update(mc.oldMakro, compact)
		if mc.isOpen {
			mc.contentDiff.Hide()
			mc.contentRead.Show()
		}
		return
	}
	if mc.isOpen {
		mc.contentDiff.Show()
	}
	mc.contentRead.Hide()

	{
		var renamed *string = nil
		for i, changeEntry := range MacrosToChangeNamesEntries {
			if MacrosToChangeReNamesEntriesBool[i].Checked && changeEntry.Text == oldMakro.MakroName {
				renamed = &MacrosToChangeReNamesEntries[i].Text
			}
		}
		var makroNameLabel string
		if renamed != nil && renamed != &oldMakro.MakroName {
			makroNameLabel = cmp.Or(oldMakro.MakroName, "<Brak nazwy>") + " -> " + *renamed + " (porównanie z plikiem .CMK)"
		} else {
			makroNameLabel = cmp.Or(oldMakro.MakroName, "<Brak nazwy>") + " (porównanie z plikiem .CMK)"
		}
		if compact {
			mc.openButton.SetText(mc.parentPlate.openButton.Text + "/" + makroNameLabel)
		} else {
			mc.openButton.SetText(makroNameLabel)
		}
	}
	mc.contentHeader.Refresh()

	{ // VARIJABLE
		// todo only varijable changes are returned

		newMakroCopyUntilIFixTheUpdateMakro := *newMakro
		varijableChanges := corpus.UpdateMakro(mc.oldMakro, &newMakroCopyUntilIFixTheUpdateMakro, nil, false)
		old, new := smartTextComparison(varijableChanges)
		newRichText := NewRichTextFromCode("[VARIJABLE]\n// po zaktualizowaniu z pliku .CMK, ", new)
		oldRichText := NewRichTextFromCode("[VARIJABLE]\n// wczytane z Corpusa\n", old)
		mc.contentDiff.Objects[0] = NewHSplitFromCMK(oldRichText, newRichText)
	}
	if oldMakro.Joint != nil || newMakro.Joint != nil {
		newRichText := NewRichTextFromCMK("[JOINT]\n// sekcja JOINT nigdy nie jest aktualizowana\n", oldMakro.Joint)
		oldRichText := NewRichTextFromCMK("[JOINT]\n\n", oldMakro.Joint)
		mc.contentDiff.Objects[1] = NewHSplitFromCMK(oldRichText, newRichText)
	}

	var showAllMakro *widget.Button

	showAllMakro = widget.NewButton("Pokaż całe makro wczytane z pliku .CMK", func() {
		showAllMakro.Hide()

		mc.contentDiff.Objects[2] = NewRichTextFromCMK("[FORMULE]\n// formule jest wczytywane z pliku .CMK\n", newMakro.Formule)

		pocketVBox := mc.contentDiff.Objects[3].(*fyne.Container)
		pocketVBox.RemoveAll()
		for i, item := range newMakro.Pocket {
			pocketVBox.Add(NewRichTextFromCMK(fmt.Sprintf("[POCKET%d]\n// Wczytane z pliku .CMK\n", i), &item))
		}
		potrosniVBox := mc.contentDiff.Objects[4].(*fyne.Container)
		potrosniVBox.RemoveAll()
		for i, item := range newMakro.Potrosni {
			potrosniVBox.Add(NewRichTextFromCMK(fmt.Sprintf("[POTROSNI%d]\n// Wczytane z pliku .CMK\n", i), &item))
		}
		grupaVBox := mc.contentDiff.Objects[5].(*fyne.Container)
		grupaVBox.RemoveAll()
		for i, item := range newMakro.Grupa {
			grupaVBox.Add(NewRichTextFromCMK(fmt.Sprintf("[GRUPA%d]\n// Wczytane z pliku .CMK\n", i), &item))
		}
		rasterVBox := mc.contentDiff.Objects[6].(*fyne.Container)
		rasterVBox.RemoveAll()
		for i, item := range newMakro.Raster {
			rasterVBox.Add(NewRichTextFromCMK(fmt.Sprintf("[RASTER%d]\n// Wczytane z pliku .CMK\n", i), &item))
		}
		makroVBox := mc.contentDiff.Objects[7].(*fyne.Container)
		makroVBox.RemoveAll()
		for i, item := range newMakro.Makro {
			makroVBox.Add(NewRichTextFromCMK(fmt.Sprintf("[MAKRO%d]\n// Wczytane z pliku .CMK\n", i), &item.GenericNodeWithDat))
		}

		mc.contentDiff.Objects[8] = widget.NewButtonWithIcon("Zwiń", theme.VisibilityOffIcon(), func() {
			showAllMakro.Show()
			mc.contentDiff.Hide()
			mc.openButton.SetIcon(theme.VisibilityIcon())
			mc.contentHeader.Refresh()
			mc.isOpen = false
		})
	})

	mc.contentDiff.Objects[9] = showAllMakro
}

func (mc *MacroContainer) Update(oldMakro *corpus.M1, compact bool) {
	mc.oldMakro = oldMakro
	mc.newMakro = nil
	var newMakro *corpus.M1
	for i, makroToChangeName := range MacrosToChangeNamesEntries {
		if oldMakro.MakroName == makroToChangeName.Text {
			makroRootPath := fyne.CurrentApp().Preferences().String("makroSearchPath")
			makro, err := corpus.NewMakroFromCMKFile(nil, MacrosToChangeEntries[i].Text, &makroRootPath, corpus.MakroCollectionCache.GetMakroMappings())
			if err != nil {
				log.Printf("ERROR: reading makro failed: %s\n", err)
			}
			newMakro = makro
			break
		}
	}
	if newMakro != nil {
		mc.UpdateMacroForDiff(newMakro, compact)
		mc.loadThisMacroButton.Hide()
		return
	} else {
		mc.loadThisMacroButton.Show()
	}

	if oldMakro.MakroName == "" {
		mc.loadThisMacroButton.Disable()
		// todo
		if compact {
			mc.openButton.SetText(mc.parentPlate.openButton.Text + "/<Brak nazwy>")
			mc.contentHeader.Show()
		} else {
			mc.openButton.SetText("<Brak nazwy>")
		}
	} else {
		mc.loadThisMacroButton.Enable()
		if compact {
			mc.openButton.SetText(mc.parentPlate.openButton.Text + "/" + oldMakro.MakroName)
			mc.contentHeader.Show()
		} else {
			mc.openButton.SetText(oldMakro.MakroName)
		}
	}
	mc.contentHeader.Refresh()
	mc.contentRead.Objects[0] = NewRichTextFromCMK("[VARIJABLE]\n", &oldMakro.Varijable)
	mc.contentRead.Objects[1] = NewRichTextFromCMK("[JOINT]\n", oldMakro.Joint)

	var showAllMakro *widget.Button
	showAllMakro = widget.NewButton("Pokaż całe makro", func() {
		showAllMakro.Hide()

		mc.contentRead.Objects[2] = NewRichTextFromCMK("[FORMULE]\n", oldMakro.Formule)

		pocketVBox := mc.contentRead.Objects[3].(*fyne.Container)
		pocketVBox.RemoveAll()
		for i, item := range oldMakro.Pocket {
			pocketVBox.Add(NewRichTextFromCMK(fmt.Sprintf("[POCKET%d]\n", i), &item))
		}
		potrosniVBox := mc.contentRead.Objects[4].(*fyne.Container)
		potrosniVBox.RemoveAll()
		for i, item := range oldMakro.Potrosni {
			potrosniVBox.Add(NewRichTextFromCMK(fmt.Sprintf("[POTROSNI%d]\n", i), &item))
		}
		grupaVBox := mc.contentRead.Objects[5].(*fyne.Container)
		grupaVBox.RemoveAll()
		for i, item := range oldMakro.Grupa {
			grupaVBox.Add(NewRichTextFromCMK(fmt.Sprintf("[GRUPA%d]\n", i), &item))
		}
		rasterVBox := mc.contentRead.Objects[6].(*fyne.Container)
		rasterVBox.RemoveAll()
		for i, item := range oldMakro.Raster {
			rasterVBox.Add(NewRichTextFromCMK(fmt.Sprintf("[RASTER%d]\n", i), &item))
		}
		makroVBox := mc.contentRead.Objects[7].(*fyne.Container)
		makroVBox.RemoveAll()
		for i, item := range oldMakro.Makro {
			makroVBox.Add(NewRichTextFromCMK(fmt.Sprintf("[MAKRO%d]\n", i), &item.GenericNodeWithDat))
		}
		mc.contentRead.Objects[8] = widget.NewButtonWithIcon("Zwiń", theme.VisibilityOffIcon(), func() {
			mc.contentRead.Hide()
			pocketVBox.RemoveAll()
			potrosniVBox.RemoveAll()
			grupaVBox.RemoveAll()
			rasterVBox.RemoveAll()
			makroVBox.RemoveAll()
			mc.openButton.SetIcon(theme.VisibilityIcon())
			mc.contentHeader.Refresh()
			mc.isOpen = false
		})
	})
	mc.contentRead.Objects[9] = showAllMakro
}

func (mc *MacroContainer) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(mc.all)
}
