package main

import (
	"fmt"
	"strconv"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

type PlateContainer struct {
	widget.BaseWidget
	contentHeader *fyne.Container
	content       *fyne.Container
	isOpen        bool

	// used for rendering
	all *fyne.Container

	// for reference when updating
	openButton       *widget.Button
	stats            *widget.Label
	macrosContainers *fyne.Container // MacroContainer
}

// formatka/Daske
func NewPlate(element *Element, adIndex int, nestLevel int) *PlateContainer {
	leftPadding := canvas.NewRectangle(nil)                        // Empty rectangle
	leftPadding.SetMinSize(fyne.NewSize(float32(20*nestLevel), 0)) // 20 units wide
	h := container.NewHBox(leftPadding)
	c := container.NewVBox()
	macrosContainer := container.NewWithoutLayout()
	c.Hide()

	pc := &PlateContainer{
		contentHeader:    h,
		content:          c,
		isOpen:           false,
		macrosContainers: macrosContainer,
		all:              container.NewVBox(h, c),
	}
	daske := element.Daske
	ad := daske.AD[adIndex]
	var openButton *widget.Button
	openButton = widget.NewButtonWithIcon(
		fmt.Sprintf("Formatka: %s", ad.DName.Value),
		theme.MenuExpandIcon(),
		func() {
			pc.isOpen = !pc.isOpen
			if pc.isOpen {
				openButton.Icon = theme.MenuDropDownIcon()
				pc.content.Show()
			} else {
				openButton.Icon = theme.MenuExpandIcon()
				pc.content.Hide()
			}
			openButton.Refresh()
		})

	iconWidget := widget.NewIcon(resourceBox3Svg)
	h.Add(iconWidget)
	h.Add(openButton)
	pc.openButton = openButton
	h.Add(layout.NewSpacer())

	howMenyMacros := 0
	for _, spoj := range element.Elinks.Spoj {
		_adIndex, _ := strconv.Atoi(spoj.O1.Value)
		if _adIndex != adIndex {
			continue
		}
		howMenyMacros++
		_con := NewMacroContainer(nestLevel, pc)
		c.Add(_con)
		macrosContainer.Add(_con)
		_con.Update(&spoj.Makro1) // todo update here?
	}
	stats := widget.NewLabel(fmt.Sprintf("Makra: %d", howMenyMacros))
	h.Add(stats)
	pc.stats = stats
	pc.ExtendBaseWidget(pc)
	return pc
}

func (ec *PlateContainer) Show() {
	ec.BaseWidget.Show()
	ec.all.Show()
	// ec.all.Refresh()
}
func (ec *PlateContainer) Hide() {
	ec.BaseWidget.Hide()
	ec.all.Hide()
	// ec.all.Refresh()
}
func (pc *PlateContainer) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(pc.all)
}

func DaskeTotalNumOfMacros(element *Element, adIndex int) int {
	howManyMacros := 0
	for _, spoj := range element.Elinks.Spoj {
		_adIndex, _ := strconv.Atoi(spoj.O1.Value)
		if _adIndex != adIndex {
			continue
		}
		howManyMacros++
	}
	return howManyMacros
}

func (pc *PlateContainer) Update(element *Element, adIndex int) bool {
	if Settings.hideElementsWithZeroMacros {
		if DaskeTotalNumOfMacros(element, adIndex) == 0 {
			pc.Hide()
			return true
		}
	} else {
		pc.Show()
	}
	daske := element.Daske
	ad := daske.AD[adIndex]
	if Settings.compact {
		pc.contentHeader.Hide()
		pc.isOpen = true
		pc.content.Show()
		pc.openButton.SetText(ad.DName.Value)
	} else {
		pc.openButton.SetText(fmt.Sprintf("Formatka: %s", ad.DName.Value))
		pc.isOpen = false
		pc.contentHeader.Show()
		// pc.content.Hide()
	}

	howManyMacros := 0
	for _, spoj := range element.Elinks.Spoj {
		_adIndex, _ := strconv.Atoi(spoj.O1.Value)
		if _adIndex != adIndex {
			continue
		}
		macroCon := pc.macrosContainers.Objects[howManyMacros].(*MacroContainer)
		macroCon.Update(&spoj.Makro1)
		howManyMacros++
	}
	pc.stats.SetText(fmt.Sprintf("Makra: %d", howManyMacros))
	return false
}
