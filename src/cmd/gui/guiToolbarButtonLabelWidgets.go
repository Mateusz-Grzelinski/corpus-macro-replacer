package main

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

type toolbarButton struct {
	*widget.Button
}

func NewToolbarButtonWithIcon(label string, icon fyne.Resource, buttonFunc func()) widget.ToolbarItem {
	b := widget.NewButtonWithIcon(label, icon, buttonFunc)
	return &toolbarButton{b}
}

func (t *toolbarButton) ToolbarObject() fyne.CanvasObject {
	return t.Button
}

type toolbarLabel struct {
	*widget.Label
}

func NewToolbarLabel(label string) *toolbarLabel {
	l := widget.NewLabelWithStyle(label, fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	l.Truncation = fyne.TextTruncateClip
	return &toolbarLabel{l}
}

func (t *toolbarLabel) ToolbarObject() fyne.CanvasObject {
	return container.NewHBox(
		widget.NewIcon(theme.FileIcon()),
		t.Label,
	)
	// return t.Label
}

// ThreeStateCheckbox represents a 3-state checkbox
type ThreeStateCheckbox struct {
	widget.BaseWidget
	state    int             // 0: Unchecked, 1: Checked, 2: Indeterminate
	label    string          // Text for the checkbox
	callback func(state int) // Callback when state changes
}

// NewThreeStateCheckbox creates a new 3-state checkbox
func NewThreeStateCheckbox(label string) *ThreeStateCheckbox {
	cb := &ThreeStateCheckbox{
		label: label,
		state: 0, // Initial state is unchecked
		callback: func(state int) {
			println(label+" state changed to:", state)
		},
	}
	cb.ExtendBaseWidget(cb)
	return cb
}

// CreateRenderer creates the renderer for the 3-state checkbox
func (cb *ThreeStateCheckbox) CreateRenderer() fyne.WidgetRenderer {
	label := widget.NewLabel(cb.label)
	button := widget.NewButton("Unchecked", func() {
		// Cycle through states: 0 -> 1 -> 2 -> 0
		cb.state = (cb.state + 1) % 3
		// cb.updateButton(button) // todo
		if cb.callback != nil {
			cb.callback(cb.state)
		}
	})
	cb.updateButton(button)

	container := container.NewVBox(label, button)
	return widget.NewSimpleRenderer(container)
}

// updateButton updates the button text based on the current state
func (cb *ThreeStateCheckbox) updateButton(button *widget.Button) {
	switch cb.state {
	case 0:
		button.SetText("Unchecked")
	case 1:
		button.SetText("Checked")
	case 2:
		button.SetText("Indeterminate")
	}
}
