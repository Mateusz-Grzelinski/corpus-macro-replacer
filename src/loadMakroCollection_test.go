package main

import (
	"image/color"
	"path/filepath"
	"testing"
)

var pathToTestMakroCollection = filepath.Join("..", "tests", "makroCollection")
var testFilesMakroCollection = []string{
	"MakroCollectionMinimal.dat",
	"MakroCollection2Items.dat",
	"MakroCollection2ItemsWithSameName.dat",
	"MakroCollectionCorpus5CompleteExampleStolarz.dat",
}

func TestLoadMakroCollectionMinimal(t *testing.T) {
	path := filepath.Join(pathToTestMakroCollection, testFilesMakroCollection[0])
	out, err := NewMakroCollection(path)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	if len(out) != 1 {
		t.Log(out)
		t.Error("wrong len")
		t.FailNow()
	}
	item := out[0]
	if item.Name != "custom" {
		t.Errorf("wrong name: %s", item.Name)
	}
	if item.Category != "Kategoria" {
		t.Errorf("wrong category: %s", item.Category)
	}
	if item.FileName != "custom.CMK" {
		t.Errorf("wrong file name: %s", item.FileName)
	}
	colorFG := color.NRGBA{R: 0x01, G: 0x01, B: 0x01, A: 0x0}
	if item.TextColorFG != colorFG {
		t.Log(item.TextColorFG)
		t.Errorf("wrong FG color: %s", item.TextColorFG)
	}
	colorBG := color.NRGBA{R: 0xF2, G: 0xF2, B: 0xF2, A: 0x0}
	if item.TextColorBG != colorBG {
		t.Log(item.TextColorBG)
		t.Errorf("wrong BG color: %s", item.TextColorBG)
	}
}
func TestLoadMakroCollection2Items(t *testing.T) {
	path := filepath.Join(pathToTestMakroCollection, testFilesMakroCollection[1])
	out, err := NewMakroCollection(path)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	if len(out) != 2 {
		t.Log(out)
		t.Errorf("wrong len: %d", len(out))
		t.FailNow()
	}
	item := out[0]
	if item.Name != "Blenda" {
		t.Errorf("wrong name: %s", item.Name)
	}
	if item.Category != "BlendaKat" {
		t.Errorf("wrong category: %s", item.Category)
	}
	if item.FileName != "Blenda.CMK" {
		t.Errorf("wrong file name: %s", item.FileName)
	}

	item = out[1]
	if item.Name != "custom" {
		t.Errorf("wrong name: %s", item.Name)
	}
	if item.Category != "Kategoria" {
		t.Errorf("wrong category: %s", item.Category)
	}
	if item.FileName != "custom.CMK" {
		t.Errorf("wrong file name: %s", item.FileName)
	}
}
func TestLoadMakroCollection2ItemsSameNameMissingFileds(t *testing.T) {
	path := filepath.Join(pathToTestMakroCollection, testFilesMakroCollection[2])
	out, err := NewMakroCollection(path)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	if len(out) != 2 {
		t.Log(out)
		t.Error("wrong len")
		t.FailNow()
	}
	item := out[0]
	if item.Name != "1" {
		t.Errorf("wrong name: %s", item.Name)
	}
	if item.Category != "" {
		t.Errorf("wrong category: %s", item.Category)
	}
	if item.FileName != "" {
		t.Errorf("wrong file name: %s", item.FileName)
	}

	item = out[1]
	if item.Name != "1" {
		t.Errorf("wrong name: %s", item.Name)
	}
	if item.Category != "" {
		t.Errorf("wrong category: %s", item.Category)
	}
	if item.FileName != "" {
		t.Errorf("wrong file name: %s", item.FileName)
	}
}
func TestLoadMakroCollectionCompleteExample(t *testing.T) {
	path := filepath.Join(pathToTestMakroCollection, testFilesMakroCollection[3])
	out, err := NewMakroCollection(path)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	if len(out) != 63 {
		t.Log(out)
		t.Error("wrong len")
		t.FailNow()
	}
	item := out[27]
	if item.Name != "Blum Zawiasy Równoległe" {
		t.Errorf("wrong name: %s", item.Name)
	}
	if item.Category != "DRZWI" {
		t.Errorf("wrong category: %s", item.Category)
	}
	if item.FileName != "Blum Zawiasy\\Równoległy.CMK" {
		t.Errorf("wrong file name: %s", item.FileName)
	}
	// Blum Zawiasy Równoległe cat DRZWI
}
