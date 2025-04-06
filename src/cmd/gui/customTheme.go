package main

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
)

// CustomTheme implements the fyne.Theme interface
type CustomTheme struct {
	darkIcons map[fyne.ThemeIconName]fyne.Resource
	baseTheme fyne.Theme
}

func NewCustomTheme(base fyne.Theme) *CustomTheme {
	customTheme := CustomTheme{
		darkIcons: make(map[fyne.ThemeIconName]fyne.Resource),
		baseTheme: base,
	}

	customTheme.SetIcon("box3", resourceBox3Svg)
	customTheme.SetIcon("boxRecurse1", resourceBoxRecurse1Svg)
	customTheme.SetIcon("boxRecurse2", resourceBoxRecurse2Svg)
	customTheme.SetIcon("boxRecurse3", resourceBoxRecurse3Svg)
	customTheme.SetIcon("cabinet", resourceCabinetSvg)
	customTheme.SetIcon("filter", resourceFilterSvg)
	customTheme.SetIcon("gear", resourceGearSvg)
	customTheme.SetIcon("checkMinux", resourceCheckMinusSvg)

	return &customTheme
}

func (c *CustomTheme) SetIcon(name fyne.ThemeIconName, darkIcon fyne.Resource) {
	c.darkIcons[name] = darkIcon
}

func (c *CustomTheme) Icon(name fyne.ThemeIconName) fyne.Resource {
	if icon, exists := c.darkIcons[name]; exists {
		return theme.NewThemedResource(icon)
	}
	return c.baseTheme.Icon(name)
}

// Delegate other theme methods to the base theme
func (c *CustomTheme) Font(style fyne.TextStyle) fyne.Resource {
	return c.baseTheme.Font(style)
}

func (c *CustomTheme) Color(name fyne.ThemeColorName, variant fyne.ThemeVariant) color.Color {
	return c.baseTheme.Color(name, variant)
}

func (c *CustomTheme) Size(name fyne.ThemeSizeName) float32 {
	return c.baseTheme.Size(name)
}
