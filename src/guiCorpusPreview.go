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

var macroIcon *canvas.Image = canvas.NewImageFromResource(resourceMacroSvg)

type MacroCacheKey struct {
	Path         string
	ElementIndex int
	ADIndex      int
	MacroIndex   int
}

type MacroContainer struct {
	widget.BaseWidget
	all           *fyne.Container
	contentHeader *fyne.Container
	contentRead   *fyne.Container
	contentDiff   *fyne.Container
	isOpen        bool
	oldMakro      *M1
	newMakro      *M1
}

// NewMacroContainer creates a new instance of MacroContainer
func NewMacroContainer() *MacroContainer {
	mc := &MacroContainer{
		all:           nil,
		contentHeader: nil,
		contentRead:   nil,
		contentDiff:   nil,
		isOpen:        false,
	}
	var previewButton, loadThisMacroButton *widget.Button
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

	previewButton = widget.NewButtonWithIcon("", theme.VisibilityIcon(), func() {})
	previewButton.OnTapped = func() {
		if mc.isOpen {
			previewButton.Icon = theme.VisibilityIcon()
			mc.contentRead.Hide()
			mc.contentDiff.Hide()
		} else {
			previewButton.Icon = theme.VisibilityOffIcon()
			if mc.newMakro == nil {
				mc.contentRead.Show()
				mc.contentDiff.Hide()
			} else {
				mc.contentRead.Hide()
				mc.contentDiff.Show()
			}
		}
		previewButton.Refresh()
		mc.isOpen = !mc.isOpen
	}
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
	mc.contentHeader = container.NewBorder(nil, nil,
		previewButton,
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

	hasSameValues := 0
	for _, change := range changes {
		switch change.result {
		case ValueAdded:
			oldReformatted.WriteString("\n")
			newReformatted.WriteString(fmt.Sprintf("%s=%s \t// nowa zmienna dodana\n", *change.newName, change.newValue))
		case ValueSame:
			hasSameValues++
			oldReformatted.WriteString(fmt.Sprintf("%s=%s\n", *change.oldName, change.oldValue))
			newReformatted.WriteString("\n")
		default:
			oldReformatted.WriteString(fmt.Sprintf("%s=%s\n", *change.oldName, change.oldValue))
			newReformatted.WriteString(fmt.Sprintf("%s=%s \t// zachowano starą wartość (wartość obecna:%s)\n", *change.newName, *&change.oldValue, change.newValue))
		}
	}
	if hasSameValues > 0 {
		newReformatted.WriteString(fmt.Sprintf("Schowano %d takie same wartości\n", hasSameValues))
	}

	return oldReformatted.String(), newReformatted.String()
}

func (mc *MacroContainer) SetMacroForDiff(newMakro *M1) {
	con := mc
	con.newMakro = newMakro
	oldMakro := con.oldMakro
	if newMakro == nil {
		mc.SetMacro(mc.oldMakro)
		if mc.isOpen {
			con.contentDiff.Hide()
			con.contentRead.Show()
		}
		return
	}
	if mc.isOpen {
		con.contentDiff.Show()
	}
	con.contentRead.Hide()

	con.contentHeader.Objects[0].(*widget.Button).SetText(newMakro.MakroName + " (porównanie z plikiem .CMK)")
	con.contentHeader.Refresh()

	// todo only varijable changes are returned
	{
		varijableChanges := UpdateMakro(con.oldMakro, con.newMakro, false)
		old, new := smartTextComparison(varijableChanges)
		newText := "`[VARIJABLE] // po zaktualizowaniu z pliku .CMK`\n\n```\n" + new + "\n```"
		newVarijableRichText := widget.NewRichTextFromMarkdown(newText)
		newVarijableRichText.Wrapping = fyne.TextWrap(fyne.TextTruncateClip)
		oldText := "`[VARIJABLE] // wczytane z corpusa`\n\n```\n" + old + "\n```"
		oldVarijableRichText := widget.NewRichTextFromMarkdown(oldText)
		oldVarijableRichText.Wrapping = fyne.TextWrap(fyne.TextTruncateClip)
		split := container.NewHSplit(oldVarijableRichText, newVarijableRichText)
		split.Offset = 0.4
		con.contentDiff.Objects[0] = split
	}
	{
		new, old := "", ""
		if newMakro.Joint == nil {
			new = strings.Join(decodeAllCMKLines(cmp.Or(newMakro.Joint.DAT, "")), "\n")
		}
		if oldMakro.Joint == nil {
			old = strings.Join(decodeAllCMKLines(cmp.Or(oldMakro.Joint.DAT, "")), "\n")
		}
		if new != "" || old != "" {
			newText := "`[JOINT]` // po zaktualizowaniu z pliku .CMK\n\n```\n" + new + "\n```"
			newVarijableRichText := widget.NewRichTextFromMarkdown(newText)
			newVarijableRichText.Wrapping = fyne.TextWrap(fyne.TextTruncateClip)
			oldText := "`[JOINT] // wczytane z corpusa`\n\n```\n" + old + "\n```"
			oldVarijableRichText := widget.NewRichTextFromMarkdown(oldText)
			oldVarijableRichText.Wrapping = fyne.TextWrap(fyne.TextTruncateClip)
			split := container.NewHSplit(oldVarijableRichText, newVarijableRichText)
			split.Offset = 0.4
			con.contentDiff.Objects[1] = split
		}
	}
	var showAllMakro *widget.Button
	showAllMakro = widget.NewButton("Pokaż całe makro", func() {
		showAllMakro.Hide()
		if oldMakro.Formule != nil {
			textOld := "`[FORMULE] // formule jest zawsze wczytywane z pliku .CMK`\n\n```\n" + strings.Join(decodeAllCMKLines(cmp.Or(newMakro.Formule.DAT, "")), "\n") + "\n```"
			multilineOld := widget.NewRichTextFromMarkdown(textOld)
			multilineOld.Wrapping = fyne.TextWrap(fyne.TextTruncateClip)
			con.contentRead.Objects[2] = multilineOld
		}
		con.contentRead.Objects[8] = widget.NewButtonWithIcon("Zwiń", theme.VisibilityOffIcon(), func() {
			con.contentDiff.Hide()
			con.contentHeader.Objects[0].(*widget.Button).SetIcon(theme.VisibilityIcon())
			con.contentHeader.Refresh()
			mc.isOpen = false
		})
	})
	con.contentRead.Objects[9] = showAllMakro
}

func (mc *MacroContainer) SetMacro(oldMakro *M1) {
	con := mc
	con.oldMakro = oldMakro
	con.newMakro = nil
	con.contentHeader.Objects[0].(*widget.Button).SetText(oldMakro.MakroName)
	con.contentHeader.Refresh()
	{
		textOld := "`[VARIJABLE]`\n\n```\n" + strings.Join(decodeAllCMKLines(oldMakro.Varijable.DAT), "\n") + "\n```"
		multilineOld := widget.NewRichTextFromMarkdown(textOld)
		multilineOld.Wrapping = fyne.TextWrap(fyne.TextTruncateClip)
		con.contentRead.Objects[0] = multilineOld
	}
	if oldMakro.Joint != nil {
		textOld := "`[JOINT]`\n\n```\n" + strings.Join(decodeAllCMKLines(cmp.Or(oldMakro.Joint.DAT, "")), "\n") + "\n```"
		multilineOld := widget.NewRichTextFromMarkdown(textOld)
		multilineOld.Wrapping = fyne.TextWrap(fyne.TextTruncateClip)
		con.contentRead.Objects[1] = multilineOld
	}
	var showAllMakro *widget.Button
	showAllMakro = widget.NewButton("Pokaż całe makro", func() {
		showAllMakro.Hide()
		if oldMakro.Formule != nil {
			text := "`[FORMULE]`\n\n```\n" + strings.Join(decodeAllCMKLines(cmp.Or(oldMakro.Formule.DAT, "")), "\n") + "\n```"
			multiline := widget.NewRichTextFromMarkdown(text)
			multiline.Wrapping = fyne.TextWrap(fyne.TextTruncateClip)
			con.contentRead.Objects[2] = multiline
		}
		pocketVBox := con.contentRead.Objects[3].(*fyne.Container)
		pocketVBox.RemoveAll()
		for i, item := range oldMakro.Pocket {
			text := "`[POCKET" + strconv.Itoa(i) + "]`\n\n```\n" + strings.Join(decodeAllCMKLines(cmp.Or(item.DAT, "")), "\n") + "\n```"
			multiline := widget.NewRichTextFromMarkdown(text)
			multiline.Wrapping = fyne.TextWrap(fyne.TextTruncateClip)
			pocketVBox.Add(multiline)
		}
		potrosniVBox := con.contentRead.Objects[4].(*fyne.Container)
		potrosniVBox.RemoveAll()
		for i, item := range oldMakro.Potrosni {
			text := "`[POTROSNI" + strconv.Itoa(i) + "]`\n\n```\n" + strings.Join(decodeAllCMKLines(cmp.Or(item.DAT, "")), "\n") + "\n```"
			multiline := widget.NewRichTextFromMarkdown(text)
			multiline.Wrapping = fyne.TextWrap(fyne.TextTruncateClip)
			pocketVBox.Add(multiline)
		}
		grupaVBox := con.contentRead.Objects[5].(*fyne.Container)
		grupaVBox.RemoveAll()
		for i, item := range oldMakro.Grupa {
			text := "`[GRUPA" + strconv.Itoa(i) + "]`\n\n```\n" + strings.Join(decodeAllCMKLines(cmp.Or(item.DAT, "")), "\n") + "\n```"
			multiline := widget.NewRichTextFromMarkdown(text)
			multiline.Wrapping = fyne.TextWrap(fyne.TextTruncateClip)
			grupaVBox.Add(multiline)
		}
		rasterVBox := con.contentRead.Objects[6].(*fyne.Container)
		rasterVBox.RemoveAll()
		for i, item := range oldMakro.Raster {
			text := "`[RASTER" + strconv.Itoa(i) + "]`\n\n```\n" + strings.Join(decodeAllCMKLines(cmp.Or(item.DAT, "")), "\n") + "\n```"
			multiline := widget.NewRichTextFromMarkdown(text)
			multiline.Wrapping = fyne.TextWrap(fyne.TextTruncateClip)
			rasterVBox.Add(multiline)
		}
		makroVBox := con.contentRead.Objects[7].(*fyne.Container)
		makroVBox.RemoveAll()
		for i, item := range oldMakro.Makro {
			textOld := "`[MAKRO" + strconv.Itoa(i) + "]`\n\n```\n" + strings.Join(decodeAllCMKLines(cmp.Or(item.DAT, "")), "\n") + "\n```"
			multiline := widget.NewRichTextFromMarkdown(textOld)
			multiline.Wrapping = fyne.TextWrap(fyne.TextTruncateClip)
			makroVBox.Add(multiline)
		}
		con.contentRead.Objects[8] = widget.NewButtonWithIcon("Zwiń", theme.VisibilityOffIcon(), func() {
			con.contentRead.Hide()
			pocketVBox.RemoveAll()
			potrosniVBox.RemoveAll()
			grupaVBox.RemoveAll()
			rasterVBox.RemoveAll()
			makroVBox.RemoveAll()
			con.contentHeader.Objects[0].(*widget.Button).SetIcon(theme.VisibilityIcon())
			con.contentHeader.Refresh()
			mc.isOpen = false
		})
	})
	con.contentRead.Objects[9] = showAllMakro
}

func (mc *MacroContainer) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(mc.all)
}

func (cm *MacroContainer) Refresh() {
	con := cm
	oldMakro := cm.oldMakro
	con.SetMacro(oldMakro)

	// todo very slow
	var newMakro *M1
	for _, makroTochangeName := range MacrosToChangeEntries {
		newMakroName := getMacroName(makroTochangeName.Text)
		if oldMakro.MakroName == newMakroName {
			makro, err := LoadMakroFromCMKFile(makroTochangeName.Text)
			newMakro = makro
			if err != nil {
				log.Println(err)
			}
			break
		}
	}
	con.SetMacroForDiff(newMakro)
	con.contentHeader.Refresh()
	con.contentRead.Refresh()
	con.contentDiff.Refresh()
}

// formatka/Daske
func NewPlate(element *Element, elementIndex int, adIndex int) *fyne.Container {
	var con *fyne.Container = container.NewVBox()
	accordionContent := container.NewVBox()

	for _, spoj := range element.Elinks.Spoj {
		_adIndex, _ := strconv.Atoi(spoj.O1.Value)
		if _adIndex != adIndex {
			continue
		}
		_con := NewMacroContainer()
		_con.SetMacro(&spoj.Makro1)
		accordionContent.Add(_con)
	}
	item1 := widget.NewAccordionItem("", accordionContent)
	con.Add(widget.NewAccordion(item1))
	return con
}

func (efc *ElementFileContainer) SetPlate(element *Element, elementIndex int, adIndex int) {
	offset := 1
	elemCon := efc.elements.Objects[elementIndex].(*fyne.Container)
	plateCon := elemCon.Objects[adIndex+offset].(*fyne.Container)
	accordionItem := plateCon.Objects[0].(*widget.Accordion).Items[0]

	daskeName := element.Daske.AD[adIndex].DName.Value
	numMakros := 0
	for _, spoj := range element.Elinks.Spoj {
		_adIndex, _ := strconv.Atoi(spoj.O1.Value)
		if _adIndex == adIndex {
			numMakros++
		}
	}
	accordionItem.Title = "▧ Formatka: '" + daskeName + "' (makra: " + strconv.Itoa(numMakros) + ")"

	for _, spoj := range element.Elinks.Spoj {
		_adIndex, _ := strconv.Atoi(spoj.O1.Value)
		if _adIndex != adIndex {
			continue
		}
		accordionDetail := accordionItem.Detail.(*fyne.Container)
		for _, macroCon := range accordionDetail.Objects {
			macroCon.Refresh()
		}
	}
}

var cabinetIcon = canvas.NewImageFromResource(resourceCabinetSvg)

// todo make refresh instead add
func NewElement(element *Element, elementIndex int) *fyne.Container {
	var con *fyne.Container = container.NewVBox()
	cabinetIcon.SetMinSize(fyne.NewSquareSize(25))
	cabinetTitle := widget.NewRichTextFromMarkdown("")
	con.Add(
		container.NewHBox(
			cabinetIcon,
			cabinetTitle,
		),
	)
	for adIndex, _ := range element.Daske.AD {
		_c := NewPlate(element, elementIndex, adIndex)
		con.Add(_c)
	}
	// if len(element.Daske.AD) == 0 {
	// 	_c := NewPlate(element, elementIndex, 0)
	// 	con.Add(_c)
	// }
	con.Add(widget.NewSeparator())
	return con
}

func (efc *ElementFileContainer) SetElement(element *Element, elementIndex int) {
	elemCon := efc.elements.Objects[elementIndex].(*fyne.Container)
	// cabinetTitle := elemCon.Objects[1].(*widget.RichText)
	hbox := elemCon.Objects[0].(*fyne.Container)
	hbox.Objects[1] = widget.NewRichTextFromMarkdown(
		fmt.Sprintf(
			"## Szafka: %s (Formatek: %s, Makr: %s)", element.EName.Value, element.Daske.DCount.Value, element.Elinks.COUNT.Value,
		),
	)
	for adIndex, _ := range element.Daske.AD {
		efc.SetPlate(element, elementIndex, adIndex)
	}
}

type ElementFileContainer struct {
	widget.BaseWidget
	// 1 element = 1 container
	elements    *fyne.Container
	elementFile *ElementFile
}

func NewElementFileContainer(ef *ElementFile, objects ...fyne.CanvasObject) *ElementFileContainer {
	c := container.NewVBox(objects...)
	mc := &ElementFileContainer{elements: c, elementFile: ef}

	for elementIndex, element := range ef.Element {
		c := NewElement(&element, elementIndex)
		mc.elements.Add(c)
		mc.SetElement(&element, elementIndex)
		c.Refresh() // todo refresh here?
	}

	mc.ExtendBaseWidget(mc)
	return mc
}

func (efc *ElementFileContainer) Refresh() {
	// todo if number of elements, planks or macros changes, Refresh will not handle that!
	for elementIndex, element := range efc.elementFile.Element {
		efc.SetElement(&element, elementIndex)
	}
	efc.elements.Refresh()
}

func (mc *ElementFileContainer) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(mc.elements)
}

// func RefreshElementFile(con *fyne.Container) {
// 	for elementIndex, element := range loaddedFileForPreview.Element {
// 		NewElement(con, &element, elementIndex)
// 	}
// }
