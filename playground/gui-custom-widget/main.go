package main

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

// MacroContainer is a custom widget with an isOpen state
type MacroContainer struct {
	widget.BaseWidget
	header  *fyne.Container
	content *fyne.Container
	isOpen  bool
}

// NewMacroContainer creates a new instance of MacroContainer
func NewMacroContainer(objects ...fyne.CanvasObject) *MacroContainer {
	mc := &MacroContainer{
		header:  container.NewVBox(objects...),
		content: container.NewVBox(objects...),
		isOpen:  false,
	}
	mc.ExtendBaseWidget(mc) // Initialize BaseWidget
	return mc
}

// ToggleOpen toggles the open/close state
func (mc *MacroContainer) ToggleOpen() {
	mc.isOpen = !mc.isOpen
	label2 := widget.NewLabel("Label 2")
	mc.content.Add(label2)
	if mc.isOpen {
		mc.Show()
	} else {
		// for _, o := range mc.content.Objects {
		// 	o.Hide()
		// }
		// mc.content.Hide()
		mc.Hide()
	}
	mc.Refresh()
}

// SetOpen explicitly sets the open/close state
func (mc *MacroContainer) SetOpen(open bool) {
	mc.isOpen = open
	if open {
		mc.content.Show()
	} else {
		mc.content.Hide()
	}
	mc.Refresh()
}

// AddObject adds a new object to the MacroContainer
func (mc *MacroContainer) AddObject(obj fyne.CanvasObject) {
	mc.content.Add(obj)
	mc.Refresh()
}

// RemoveObject removes an object from the MacroContainer
func (mc *MacroContainer) RemoveObject(obj fyne.CanvasObject) {
	mc.content.Remove(obj)
	mc.Refresh()
}

// CreateRenderer implements the fyne.WidgetRenderer interface
func (mc *MacroContainer) CreateRenderer() fyne.WidgetRenderer {
	return &macroContainerRenderer{
		macroContainer: mc,
		objects:        mc.content.Objects,
	}
}

// macroContainerRenderer is the custom renderer for MacroContainer
type macroContainerRenderer struct {
	macroContainer *MacroContainer
	objects        []fyne.CanvasObject
}

// Layout arranges the child objects
func (r *macroContainerRenderer) Layout(size fyne.Size) {
	r.macroContainer.content.Resize(size)
}

// MinSize calculates the minimum size of the widget
func (r *macroContainerRenderer) MinSize() fyne.Size {
	return r.macroContainer.content.MinSize()
}

// Refresh redraws the widget
func (r *macroContainerRenderer) Refresh() {
	for _, obj := range r.objects {
		obj.Refresh()
	}
}

// Objects returns the child objects for rendering
func (r *macroContainerRenderer) Objects() []fyne.CanvasObject {
	return r.macroContainer.content.Objects
}

// Destroy is called when the widget is destroyed
func (r *macroContainerRenderer) Destroy() {}

func main() {
	// Example usage of MacroContainer
	myApp := app.New()
	myWindow := myApp.NewWindow("MacroContainer Widget Example")

	label1 := widget.NewLabel("Label 1")
	label2 := widget.NewLabel("Label 2")
	label2.Hide()
	button := widget.NewButton("Toggle Open/Close", nil)

	var mainContainer *fyne.Container
	macroContainer := NewMacroContainer(label1, label2)
	button.OnTapped = func() {
		macroContainer.ToggleOpen()
		mainContainer.Refresh()
		mainContainer.Show()
	}

	mainContainer = container.NewVBox(button, macroContainer)

	myWindow.SetContent(mainContainer)
	myWindow.Resize(fyne.NewSize(400, 300))
	myWindow.ShowAndRun()
}

/*
package main

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

type MacroContainer struct {
	*fyne.Container
	isOpen bool
}

// NewMacroContainer creates a new MacroContainer
func NewMacroContainer(objects ...fyne.CanvasObject) *MacroContainer {
	return &MacroContainer{
		Container: container.NewVBox(objects...), // Use VBox as the base layout
		isOpen:    false,
	}
}

// AddObject adds a new object to the container
// func (mc *MacroContainer) AddObject(obj fyne.CanvasObject) {
// 	mc.Add(obj)
// }

// // RemoveObject removes an object from the container
// func (mc *MacroContainer) RemoveObject(obj fyne.CanvasObject) {
// 	mc.Remove(obj)
// }

func main() {
	// Example usage of MacroContainer
	myApp := app.New()
	myWindow := myApp.NewWindow("MacroContainer Example")

	label1 := widget.NewLabel("Label 1")
	label2 := widget.NewLabel("Label 2")
	button := widget.NewButton("Toggle Open/Close", nil)

	macroContainer := NewMacroContainer(label1, label2)
	button.OnTapped = func() {
		macroContainer.Add(widget.NewLabel("Label 3"))
		macroContainer.Refresh()
	}

	mainContainer := container.NewBorder(button, nil, nil, nil, macroContainer)

	myWindow.SetContent(mainContainer)
	myWindow.Resize(fyne.NewSize(400, 300))
	myWindow.ShowAndRun()
}
*/
