package main

import (
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

type MultiStateCheck struct {
	widget.BaseWidget
	States     []string
	Icons      []fyne.Resource // Icons for each state
	CurrentIdx int
	OnChange   func(newState string) // Callback when the state changes
}

func NewMultiStateCheck(states []string, icons []fyne.Resource, initialIdx int, onChange func(newState string)) *MultiStateCheck {
	if len(states) == 0 {
		panic("states cannot be empty")
	}
	if len(icons) != 0 && len(icons) != len(states) {
		panic("icons must be empty or match the number of states")
	}

	w := &MultiStateCheck{
		States:     states,
		Icons:      icons,
		CurrentIdx: initialIdx,
		OnChange:   onChange,
	}
	w.ExtendBaseWidget(w)
	return w
}

// Tapped cycles through the states
func (m *MultiStateCheck) Tapped(_ *fyne.PointEvent) {
	m.CurrentIdx = (m.CurrentIdx + 1) % len(m.States)
	m.Refresh()
	if m.OnChange != nil {
		m.OnChange(m.States[m.CurrentIdx])
	}
}

// CreateRenderer draws the widget with text and optional icon
func (m *MultiStateCheck) CreateRenderer() fyne.WidgetRenderer {
	th := m.Theme()
	stateText := canvas.NewText(m.States[m.CurrentIdx], theme.ForegroundColor())
	stateText.Alignment = fyne.TextAlignCenter

	bg := canvas.NewImageFromResource(th.Icon(theme.IconNameCheckButtonFill))
	var stateIcon *canvas.Image
	if len(m.Icons) > 0 {
		stateIcon = canvas.NewImageFromResource(m.Icons[m.CurrentIdx])
		stateIcon.SetMinSize(fyne.NewSize(24, 24)) // Set a fixed size for the icon
	}

	objects := []fyne.CanvasObject{stateText}
	if stateIcon != nil {
		objects = append(objects, stateIcon)
	}

	return &multiStateCheckRenderer{
		widget:  m,
		label:   stateText,
		icon:    stateIcon,
		bg:      bg,
		objects: objects,
	}
}

type multiStateCheckRenderer struct {
	widget   *MultiStateCheck
	label    *canvas.Text
	bg, icon *canvas.Image
	objects  []fyne.CanvasObject
}

func (r *multiStateCheckRenderer) Layout(size fyne.Size) {
	iconSize := fyne.NewSize(24, 24)
	textSize := r.label.MinSize()

	// Layout with icon and text
	iconPosition := fyne.NewPos((size.Width-iconSize.Width)/2, 0)
	textPosition := fyne.NewPos((size.Width-textSize.Width)/2, iconSize.Height+4)

	r.icon.Resize(iconSize)
	r.icon.Move(iconPosition)
	r.bg.Resize(iconSize)
	r.bg.Move(iconPosition)
	r.label.Resize(textSize)
	r.label.Move(textPosition)
}

func (r *multiStateCheckRenderer) MinSize() fyne.Size {
	textSize := r.label.MinSize()
	// Add space for icon and text
	return fyne.NewSize(
		fyne.Max(24, textSize.Width),
		24+textSize.Height+4,
	)
}

func (r *multiStateCheckRenderer) Refresh() {
	r.label.Text = r.widget.States[r.widget.CurrentIdx]
	r.label.Refresh()
	// r.icon.Resource = r.widget.Icons[r.widget.CurrentIdx]
	r.updateResource()
	r.icon.Refresh()
}

// must be called while holding c.check.propertyLock for reading
func (c *multiStateCheckRenderer) updateResource() {
	th := c.widget.Theme()
	res := theme.NewThemedResource(th.Icon(theme.IconNameCheckButton))
	res.ColorName = theme.ColorNameInputBorder
	bgRes := theme.NewThemedResource(th.Icon(theme.IconNameCheckButtonFill))
	bgRes.ColorName = theme.ColorNameInputBackground

	switch c.widget.States[c.widget.CurrentIdx] {
	case "ok":
		res = theme.NewThemedResource(th.Icon(theme.IconNameCheckButtonChecked))
		res.ColorName = theme.ColorNamePrimary
		bgRes.ColorName = theme.ColorNameBackground
	case "off":
		res = theme.NewThemedResource(theme.CheckButtonIcon())
		res.ColorName = theme.ColorNameDisabled
		bgRes.ColorName = theme.ColorNameBackground
	default:
		res = theme.NewThemedResource(theme.FolderIcon())
		res.ColorName = theme.ColorNamePrimary
		bgRes.ColorName = theme.ColorNameBackground
	}
	c.icon.Resource = res
	c.icon.Refresh()
	c.bg.Resource = bgRes
	c.bg.Refresh()
}

func (r *multiStateCheckRenderer) Destroy() {}

func (r *multiStateCheckRenderer) Objects() []fyne.CanvasObject {
	return r.objects
}

func main2() {
	a := app.NewWithID("demo.selectableFileTree")
	w := a.NewWindow("Multi-State Check with Icons")

	states := []string{"off", "ok", "Paused"}
	// bg := canvas.NewImageFromResource(theme.Icon(theme.IconNameCheckButtonFill))
	of := theme.NewThemedResource(theme.Icon(theme.IconNameCheckButton))
	// of.ColorName = theme.ColorNameInputBorder
	on := theme.NewThemedResource(theme.Icon(theme.IconNameCheckButtonChecked))
	// on.ColorName = theme.ColorNameInputBackground
	icons := []fyne.Resource{
		// theme.NewThemedResource(theme.Icon(theme.IconNameCheckButtonChecked)),
		// theme.NewThemedResource(theme.Icon(theme.IconNameCheckButton)),
		on, of,
		theme.MediaPauseIcon(), // Example icon for "Paused"
	}

	multiCheck := NewMultiStateCheck(states, icons, 0, func(newState string) {
		fmt.Printf("State changed to: " + newState)
	})

	w.SetContent(container.NewVBox(
		multiCheck,
		widget.NewLabel("Click on the widget to cycle through states"),
	))

	w.Resize(fyne.NewSize(300, 200))
	w.ShowAndRun()
}
