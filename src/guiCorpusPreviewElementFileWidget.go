package main

import (
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

func (efc *ElementFileContainer) SetElement(element *Element, elementIndex int) {
	elemCon := efc.root.Objects[elementIndex].(*fyne.Container)
	// cabinetTitle := elemCon.Objects[1].(*widget.RichText)
	hbox := elemCon.Objects[0].(*fyne.Container)
	hbox.Objects[1] = widget.NewRichTextFromMarkdown(
		fmt.Sprintf(
			"## Szafka: %s (Formatek: %s, Makr: %s)\n\nPodgrup: %d", element.EName.Value, element.Daske.DCount.Value, element.Elinks.COUNT.Value, len(element.ElmList.Elm),
		),
	)
	for adIndex, _ := range element.Daske.AD {
		efc.SetPlate(element, elementIndex, adIndex)
	}
}

type ElementFileContainerKey struct {
	elementIndex int  // cabinet
	elmListIndex int  // recursion level
	adIndex      *int // plate. In nil: get container for plate
	macroIndex   *int // recursion level
}

type ElementFileContainer struct {
	widget.BaseWidget
	// 1 element = 1 container
	root        *fyne.Container
	elementFile *ElementFile
	// containerCabinets
	// containerTree map[ElementFileContainerKey]*fyne.Container
}

func NewElementFileContainer(ef *ElementFile, objects ...fyne.CanvasObject) *ElementFileContainer {
	c := container.NewVBox(objects...)
	mc := &ElementFileContainer{root: c, elementFile: ef}

	for elementIndex, element := range ef.Element {
		c := NewElement(&element, elementIndex)
		mc.root.Add(c)
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
	efc.root.Refresh()
}

func (mc *ElementFileContainer) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(mc.root)
}

// func RefreshElementFile(con *fyne.Container) {
// 	for elementIndex, element := range loaddedFileForPreview.Element {
// 		NewElement(con, &element, elementIndex)
// 	}
// }
