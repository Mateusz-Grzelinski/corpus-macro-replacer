package main

import (
	"strconv"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

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
	elemCon := efc.root.Objects[elementIndex].(*fyne.Container)
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
