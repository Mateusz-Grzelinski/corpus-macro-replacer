package main

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

type ElementFileContainer struct {
	widget.BaseWidget
	content *fyne.Container
	// actual corpus file, maybe not needed?
	elementFile *ElementFile
	// filters
	// hideElementsWithZeroMacros bool

	/*
		{
			szafka1
			szafka2
				podGrupa2
			szafka3
				podGrupa3
					podPodGrupa3
		}
	*/

	// only reference to containers
	elementContainer []*ElementContainer
}

func NewElementFileContainer(ef *ElementFile, compact bool, hideElementsWithZeroMacros bool) *ElementFileContainer {
	c := container.NewVBox()
	efc := &ElementFileContainer{
		content:          c,
		elementFile:      ef,
		elementContainer: []*ElementContainer{},
	}

	for _, element := range ef.Element {
		c := NewElement(&element, 0, compact, hideElementsWithZeroMacros)
		efc.content.Add(c)
		efc.elementContainer = append(efc.elementContainer, c)
	}

	efc.ExtendBaseWidget(efc)
	return efc
}

// todo is it needed?
// func (efc *ElementFileContainer) Refresh() {
// 	// todo if number of elements, planks or macros changes, Refresh will not handle that!
// 	for elementIndex, element := range efc.elementFile.Element {
// 		efc.Update(&element, elementIndex)
// 	}
// 	efc.root.Refresh()
// }

func (mc *ElementFileContainer) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(mc.content)
}
func ElementTotalNumOfMacros(element *Element) int {
	return _elementTotalNumOfMacros(element, 0)
}
func _elementTotalNumOfMacros(element *Element, accumulate int) int {
	numOfMacros := len(element.Elinks.Spoj) + accumulate
	for _, elem := range element.ElmList.Elm {
		numOfMacros = _elementTotalNumOfMacros(&elem, numOfMacros)
	}
	return numOfMacros
}

func (efc *ElementFileContainer) Update(elementFile *ElementFile, compact bool, hideElementsWithZeroMacros bool) {
	efc.elementFile = elementFile
	for elementIndex, elementCon := range efc.elementContainer {
		element := elementFile.Element[elementIndex]
		elementCon.Update(element, 0, compact, hideElementsWithZeroMacros)
	}
	// efc.Refresh()
}
