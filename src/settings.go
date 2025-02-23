package main

import (
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

type ConvertLocalVariablesToGlobal string

const (
	Always ConvertLocalVariablesToGlobal = "always"
	OnlyIfValueIsTheSame
	OnlyIfValueIsNumber
	/*
		{
			[VARIJABLE] // old
			grubosc=32

			[VARIJABLE] // new
			_grubosc=18

			[VARIJABLE] // output
			grubosc=32
		}
		note global value is ignored:
		{
			[VARIJABLE] // output
			_grubosc=32 // 32 is ignored
		}
	*/
	KeepLocal
	/*
		{
			[VARIJABLE] // old
			grubosc=if(sz>50;18;32)

			[VARIJABLE] // new
			_grubosc=18

			[VARIJABLE] // output
			grubosc=if(sz>50;18;32)
		}
	*/
	// KeepLocalIfHasCondition
)

func NewCorpusMakroReplacerSettings(a fyne.App) *widget.Card {
	labelSearch := widget.NewLabel("Domyślna ścieżka szukania Makr")
	makroSearchPath := a.Preferences().StringWithFallback("makroSearchPath", `C:\Tri D Corpus\Corpus 5.0\Makro\`)
	makroSearchEntry := widget.NewEntry()
	makroSearchEntry.SetText(makroSearchPath)
	makroSearchEntry.OnChanged = func(inputPath string) {
		a.Preferences().SetString("makroSearchPath", inputPath)
	}
	label := widget.NewLabel("Opcjonalna ścieżka do MakroCollection.Dat. Ten plik dostarcza mapowanie nazwa makra w Corpus <-> ścieżka pliku. Domyślnie nazwa makra to nazwa pliku. ")
	label.Wrapping = fyne.TextWrapBreak
	errLabel := widget.NewLabel("")
	errLabel.Wrapping = fyne.TextWrapBreak
	makroCollectionPath := a.Preferences().StringWithFallback("makroCollectionPath", `C:\Tri D Corpus\Corpus 5.0\Makro\MakroCollection.dat`)
	makroCollectionEntry := widget.NewEntry()
	makroCollectionEntry.SetText(makroCollectionPath)
	makroCollectionEntry.OnChanged = func(inputPath string) {
		collection, err := NewMakroCollection(inputPath)
		MakroCollectionCache = collection
		errLabel.Show()
		if err != nil {
			errLabel.SetText(fmt.Sprintf("error: %s", err))
			errLabel.Importance = widget.DangerImportance
			errLabel.Refresh()
			return
		} else {
			errLabel.SetText(fmt.Sprintf("MakroCollection.Dat: załadowano %d mapowań", len(collection)))
			errLabel.Importance = widget.MediumImportance
			errLabel.Refresh()
			a.Preferences().SetString("makroCollectionPath", inputPath)
		}

	}
	makroCollectionEntry.OnChanged(makroCollectionPath) // run to report any errors
	return widget.NewCard("Ustawienia makr", "", container.NewVBox(labelSearch, makroSearchEntry, label, makroCollectionEntry, errLabel))

}
