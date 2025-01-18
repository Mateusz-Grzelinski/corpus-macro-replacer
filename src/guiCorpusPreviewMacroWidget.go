package main

import (
	"cmp"
	"fmt"
	"log"
	"path/filepath"
	"strconv"
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
	oldMakro      *M1
	newMakro      *M1

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
		macroGuessedPath := filepath.Join(MacrosDefaultPathNormal, mc.oldMakro.MakroName) + ".CMK"
		addToLoadedFilesAndRefresh(SelectedPath)
		for _, makroTochangeName := range MacrosToChangeEntries {
			if makroTochangeName.Text == macroGuessedPath {
				return
			}
		}
		needCreation := true
		for _, makroTochangeName := range MacrosToChangeEntries {
			if makroTochangeName.Text == "" {
				makroTochangeName.SetText(macroGuessedPath)
				needCreation = false
				break
			}
		}
		if needCreation {
			addMakroButton.OnTapped()
		}
		MacrosToChangeEntries[len(MacrosToChangeEntries)-1].SetText(macroGuessedPath)
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

func smartTextComparison(changes []Change) (string, string) {
	var oldReformatted strings.Builder
	var newReformatted strings.Builder

	// first some stats
	hasSameValues := 0
	for _, change := range changes {
		switch change.result {
		case ValueSame:
			hasSameValues++
		}
	}
	if hasSameValues > 0 {
		newReformatted.WriteString(fmt.Sprintf("Schowano %d takie same wartości\n\n", hasSameValues))
		oldReformatted.WriteString("\n")
	}

	// actual diff
	for _, change := range changes {
		switch change.result {
		case ValueAdded:
			// oldReformatted.WriteString("\n")
			newReformatted.WriteString(fmt.Sprintf("%s=%s \t// nowa zmienna dodana\n", *change.newName, change.newValue))
		case ValueSame:
			oldReformatted.WriteString(fmt.Sprintf("%s=%s\n", *change.oldName, change.oldValue))
			newReformatted.WriteString("\n")
		default:
			oldReformatted.WriteString(fmt.Sprintf("%s=%s\n", *change.oldName, change.oldValue))
			newReformatted.WriteString(fmt.Sprintf("%s=%s \t// zachowano starą wartość (wartość obecna:%s)\n", *change.newName, *&change.oldValue, change.newValue))
		}
	}

	return oldReformatted.String(), newReformatted.String()
}

func (mc *MacroContainer) UpdateMacroForDiff(newMakro *M1, compact bool) {
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

	mc.openButton.SetText(cmp.Or(newMakro.MakroName, "<Brak nazwy>") + " (porównanie z plikiem .CMK)")
	mc.contentHeader.Refresh()

	// todo only varijable changes are returned
	{
		varijableChanges := UpdateMakro(mc.oldMakro, mc.newMakro, false)
		old, new := smartTextComparison(varijableChanges)
		newText := "```\n[VARIJABLE]\n// po zaktualizowaniu z pliku .CMK, " + new + "```"
		newVarijableRichText := widget.NewRichTextFromMarkdown(newText)
		newVarijableRichText.Wrapping = fyne.TextWrap(fyne.TextTruncateClip)
		oldText := "```\n[VARIJABLE]\n// wczytane z corpusa\n" + old + "```"
		oldVarijableRichText := widget.NewRichTextFromMarkdown(oldText)
		oldVarijableRichText.Wrapping = fyne.TextWrap(fyne.TextTruncateClip)
		split := container.NewHSplit(oldVarijableRichText, newVarijableRichText)
		split.Offset = 0.4
		mc.contentDiff.Objects[0] = split
	}
	{
		new, old := "", ""
		if newMakro.Joint != nil {
			new = strings.Join(decodeAllCMKLines(newMakro.Joint.DAT), "\n")
		}
		if oldMakro.Joint != nil {
			old = strings.Join(decodeAllCMKLines(oldMakro.Joint.DAT), "\n")
		}
		if new != "" || old != "" {
			newText := "```\n[JOINT]\n// po zaktualizowaniu z pliku .CMK\n" + new + "\n```"
			newVarijableRichText := widget.NewRichTextFromMarkdown(newText)
			newVarijableRichText.Wrapping = fyne.TextWrap(fyne.TextTruncateClip)
			oldText := "```\n[JOINT]\n// wczytane z corpusa\n" + old + "\n```"
			oldVarijableRichText := widget.NewRichTextFromMarkdown(oldText)
			oldVarijableRichText.Wrapping = fyne.TextWrap(fyne.TextTruncateClip)
			split := container.NewHSplit(oldVarijableRichText, newVarijableRichText)
			split.Offset = 0.4
			mc.contentDiff.Objects[1] = split
		}
	}
	var showAllMakro *widget.Button
	showAllMakro = widget.NewButton("Pokaż całe makro", func() {
		showAllMakro.Hide()
		if oldMakro.Formule != nil {
			textOld := "```[FORMULE]\n// formule jest zawsze wczytywane z pliku .CMK\n" + strings.Join(decodeAllCMKLines(cmp.Or(newMakro.Formule.DAT, "")), "\n") + "\n```"
			multilineOld := widget.NewRichTextFromMarkdown(textOld)
			multilineOld.Wrapping = fyne.TextWrap(fyne.TextTruncateClip)
			mc.contentRead.Objects[2] = multilineOld
		}
		mc.contentRead.Objects[8] = widget.NewButtonWithIcon("Zwiń", theme.VisibilityOffIcon(), func() {
			mc.contentDiff.Hide()
			mc.contentHeader.Objects[0].(*widget.Button).SetIcon(theme.VisibilityIcon())
			mc.contentHeader.Refresh()
			mc.isOpen = false
		})
	})
	mc.contentRead.Objects[9] = showAllMakro
}

func (mc *MacroContainer) Update(oldMakro *M1, compact bool) {
	mc.oldMakro = oldMakro
	mc.newMakro = nil
	var newMakro *M1
	for _, makroToChangeName := range MacrosToChangeEntries {
		newMakroName := getMacroName(makroToChangeName.Text)
		if oldMakro.MakroName == newMakroName {
			makro, err := LoadMakroFromCMKFile(makroToChangeName.Text)
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
	{
		textOld := "`[VARIJABLE]`\n\n```\n" + strings.Join(decodeAllCMKLines(oldMakro.Varijable.DAT), "\n") + "\n```"
		multilineOld := widget.NewRichTextFromMarkdown(textOld)
		multilineOld.Wrapping = fyne.TextWrap(fyne.TextTruncateClip)
		mc.contentRead.Objects[0] = multilineOld
	}
	if oldMakro.Joint != nil {
		textOld := "`[JOINT]`\n\n```\n" + strings.Join(decodeAllCMKLines(cmp.Or(oldMakro.Joint.DAT, "")), "\n") + "\n```"
		multilineOld := widget.NewRichTextFromMarkdown(textOld)
		multilineOld.Wrapping = fyne.TextWrap(fyne.TextTruncateClip)
		mc.contentRead.Objects[1] = multilineOld
	}
	var showAllMakro *widget.Button
	showAllMakro = widget.NewButton("Pokaż całe makro", func() {
		showAllMakro.Hide()
		if oldMakro.Formule != nil {
			text := "`[FORMULE]`\n\n```\n" + strings.Join(decodeAllCMKLines(cmp.Or(oldMakro.Formule.DAT, "")), "\n") + "\n```"
			multiline := widget.NewRichTextFromMarkdown(text)
			multiline.Wrapping = fyne.TextWrap(fyne.TextTruncateClip)
			mc.contentRead.Objects[2] = multiline
		}
		pocketVBox := mc.contentRead.Objects[3].(*fyne.Container)
		pocketVBox.RemoveAll()
		for i, item := range oldMakro.Pocket {
			text := "`[POCKET" + strconv.Itoa(i) + "]`\n\n```\n" + strings.Join(decodeAllCMKLines(cmp.Or(item.DAT, "")), "\n") + "\n```"
			multiline := widget.NewRichTextFromMarkdown(text)
			multiline.Wrapping = fyne.TextWrap(fyne.TextTruncateClip)
			pocketVBox.Add(multiline)
		}
		potrosniVBox := mc.contentRead.Objects[4].(*fyne.Container)
		potrosniVBox.RemoveAll()
		for i, item := range oldMakro.Potrosni {
			text := "`[POTROSNI" + strconv.Itoa(i) + "]`\n\n```\n" + strings.Join(decodeAllCMKLines(cmp.Or(item.DAT, "")), "\n") + "\n```"
			multiline := widget.NewRichTextFromMarkdown(text)
			multiline.Wrapping = fyne.TextWrap(fyne.TextTruncateClip)
			pocketVBox.Add(multiline)
		}
		grupaVBox := mc.contentRead.Objects[5].(*fyne.Container)
		grupaVBox.RemoveAll()
		for i, item := range oldMakro.Grupa {
			text := "`[GRUPA" + strconv.Itoa(i) + "]`\n\n```\n" + strings.Join(decodeAllCMKLines(cmp.Or(item.DAT, "")), "\n") + "\n```"
			multiline := widget.NewRichTextFromMarkdown(text)
			multiline.Wrapping = fyne.TextWrap(fyne.TextTruncateClip)
			grupaVBox.Add(multiline)
		}
		rasterVBox := mc.contentRead.Objects[6].(*fyne.Container)
		rasterVBox.RemoveAll()
		for i, item := range oldMakro.Raster {
			text := "`[RASTER" + strconv.Itoa(i) + "]`\n\n```\n" + strings.Join(decodeAllCMKLines(cmp.Or(item.DAT, "")), "\n") + "\n```"
			multiline := widget.NewRichTextFromMarkdown(text)
			multiline.Wrapping = fyne.TextWrap(fyne.TextTruncateClip)
			rasterVBox.Add(multiline)
		}
		makroVBox := mc.contentRead.Objects[7].(*fyne.Container)
		makroVBox.RemoveAll()
		for i, item := range oldMakro.Makro {
			textOld := "`[MAKRO" + strconv.Itoa(i) + "]`\n\n```\n" + strings.Join(decodeAllCMKLines(cmp.Or(item.DAT, "")), "\n") + "\n```"
			multiline := widget.NewRichTextFromMarkdown(textOld)
			multiline.Wrapping = fyne.TextWrap(fyne.TextTruncateClip)
			makroVBox.Add(multiline)
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

// func (cm *MacroContainer) Refresh() {
// 	con := cm
// 	oldMakro := cm.oldMakro
// 	con.Update(oldMakro)

// 	// todo very slow
// 	var newMakro *M1
// 	for _, makroTochangeName := range MacrosToChangeEntries {
// 		newMakroName := getMacroName(makroTochangeName.Text)
// 		if oldMakro.MakroName == newMakroName {
// 			makro, err := LoadMakroFromCMKFile(makroTochangeName.Text)
// 			newMakro = makro
// 			if err != nil {
// 				log.Println(err)
// 			}
// 			break
// 		}
// 	}
// 	con.UpdateMacroForDiff(newMakro)
// 	con.contentHeader.Refresh()
// 	con.contentRead.Refresh()
// 	con.contentDiff.Refresh()
// }
