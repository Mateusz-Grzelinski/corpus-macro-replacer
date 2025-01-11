package main

import (
	"fmt"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

type ElementContainer struct {
	widget.BaseWidget
	contentHeader *fyne.Container
	content       *fyne.Container
	isOpen        bool
	nestLevel     int

	// used for rendering
	all *fyne.Container

	// for reference when updating
	openButton       *widget.Button
	stats            *widget.Label
	platesContainers *fyne.Container
	nestedContent    *fyne.Container
}

func NewElement(element *Element, nestLevel int) *ElementContainer {
	leftPadding := canvas.NewRectangle(nil)                        // Empty rectangle
	leftPadding.SetMinSize(fyne.NewSize(float32(20*nestLevel), 0)) // 20 units wide
	h := container.NewHBox(leftPadding)
	content := container.NewVBox()
	p := container.NewVBox()
	nested := container.NewWithoutLayout()
	ec := &ElementContainer{
		contentHeader:    h,
		content:          content,
		platesContainers: p,
		nestedContent:    nested,
		isOpen:           false,
		nestLevel:        nestLevel,
		stats:            nil,
		all:              container.NewVBox(h, content),
	}
	ec.content.Hide()
	var openButton *widget.Button
	openButton = widget.NewButtonWithIcon(
		"",
		theme.MenuExpandIcon(),
		func() {
			ec.isOpen = !ec.isOpen
			if ec.isOpen {
				openButton.Icon = theme.MenuDropDownIcon()
				ec.content.Show()
			} else {
				openButton.Icon = theme.MenuExpandIcon()
				ec.content.Hide()
			}
			openButton.Refresh()
		})
	var icon *fyne.StaticResource
	switch nestLevel {
	case 0:
		icon = resourceCabinetSvg
	case 1:
		icon = resourceBoxRecurse1Svg
	case 2:
		icon = resourceBoxRecurse2Svg
	case 3:
		icon = resourceBoxRecurse3Svg
	default:
		icon = resourceBox3Svg
	}
	iconWidget := widget.NewIcon(icon)
	h.Add(iconWidget)
	h.Add(openButton)
	ec.openButton = openButton
	h.Add(layout.NewSpacer())
	stats := widget.NewLabel("")
	ec.stats = stats
	h.Add(stats)

	// c.Add(widget.NewLabel("Todo"))
	for adIndex, _ := range element.Daske.AD {
		_c := NewPlate(element, adIndex, nestLevel)
		content.Add(_c)
		p.Add(_c)
	}
	ec.ExtendBaseWidget(ec)

	for _, element := range element.ElmList.Elm {
		_c := NewElement(&element, nestLevel+1)
		content.Add(_c)
		nested.Add(_c)
	}
	content.Add(widget.NewSeparator())
	ec.Update(*element, nestLevel) // todo do spearate to Update and UpdateAll?
	return ec
}

func (ec *ElementContainer) Show() {
	ec.BaseWidget.Show()
	ec.all.Show()
	// ec.all.Refresh()
}
func (ec *ElementContainer) Hide() {
	ec.BaseWidget.Hide()
	ec.all.Hide()
	// ec.all.Refresh()
}
func (ec *ElementContainer) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(ec.all)
}

func (ec *ElementContainer) Update(element Element, nestLevel int) bool {
	totalMacros := ElementTotalNumOfMacros(&element)
	if Settings.hideElementsWithZeroMacros {
		if totalMacros == 0 {
			ec.Hide()
			return true
		}
	} else {
		ec.Show()
	}
	title := ""
	switch nestLevel {
	case 0:
		title = "Szafka"
	default:
		title = "Grupa"
	}
	ec.openButton.SetText(fmt.Sprintf("%s: %s", title, element.EName.Value))

	hiddenPlates := 0
	for adIndex, _ := range element.Daske.AD {
		plateCon := ec.platesContainers.Objects[adIndex].(*PlateContainer)
		if plateCon.Update(&element, adIndex) {
			hiddenPlates++
		}
	}
	hiddenElements := 0
	for elementListIndex, element := range element.ElmList.Elm {
		nestedCon := ec.nestedContent.Objects[elementListIndex].(*ElementContainer)
		if nestedCon.Update(element, nestLevel+1) {
			hiddenElements++
		}
	}

	var subGroupsText string
	if hiddenElements > 0 {
		subGroupsText = fmt.Sprintf("Podgrupy: %d/%d", len(element.ElmList.Elm)-hiddenElements, len(element.ElmList.Elm))
	} else {
		subGroupsText = fmt.Sprintf("Podgrupy: %d", len(element.ElmList.Elm))
	}
	var hiddenPlatesText string
	if hiddenPlates > 0 {
		hiddenPlatesText = fmt.Sprintf("Formatek: %d/%d", len(ec.platesContainers.Objects)-hiddenPlates, len(ec.platesContainers.Objects))
	} else {
		hiddenPlatesText = fmt.Sprintf("Formatek: %d", len(ec.platesContainers.Objects))
	}
	ec.stats.SetText(strings.Join([]string{
		subGroupsText,
		hiddenPlatesText,
		fmt.Sprintf("Makra: %d", totalMacros),
	}, ", "))
	return false
}
