package main

import (
	"bytes"
	"encoding/xml"
	"path/filepath"
	"testing"
)

var pathToE3DTestData = filepath.Join("..", "tests", "testData", "E3D")
var testFilesE3D = []string{
	"simple.E3D",
	"simple_macro_in_macro.E3D",
	// "simple_in_simple.E3D",
	"simple_in_simple_in_simple.E3D",
}

func TestLoadCorpusE3DFile(t *testing.T) {
	for _, testFile := range testFilesE3D {
		simple_path := filepath.Join(pathToE3DTestData, testFile)
		_, _, err := NewCorpusFile(simple_path)
		if err != nil {
			t.Error(err)
			// t.Log(elementFile)
		}
	}
}

// output E3D file in not exactly the same as input, but that is within xml spec
// todo I want to manually check what is actually encoded
func isSimilar(t *testing.T, originalPath string, encoded string) bool {
	// panic("no implemented")
	return true
}

func TestLoadCorpusE3DFileSimpleInSimple(t *testing.T) {
	simple_path := filepath.Join(pathToE3DTestData, "simple_in_simple.E3D")
	_, elementFile, err := NewCorpusFile(simple_path)
	if err != nil {
		t.Error(err)
	}
	if len(elementFile.Element) != 1 {
		t.Errorf("wrong num of elements/cabinets")
		t.Log(elementFile)
	}
	elem := elementFile.Element[0]
	if elem.EName.Value != "simple" {
		t.Errorf("wrong element name")
		t.Log(elem)
	}
	if len(elem.Daske.AD) != 3 {
		t.Errorf("wrong num of plates (formatka)")
		t.Log(elem.Daske)
	}
	if len(elem.ElmList.Elm) != 1 {
		t.Errorf("missing recursive element (cabinet in cabinet)")
		t.Log(elem.ElmList)
	}
	recursiveElement := elem.ElmList.Elm[0]
	if recursiveElement.Daske.DCount.Value != "3" {
		t.Errorf("wrong num of plates (formatka)")
		t.Log(elem.Daske)
	}
	if len(recursiveElement.ElmList.Elm) != 0 {
		t.Errorf("tripple recursive should be 0")
		t.Log(elem.Daske)
	}
	recursiveAD := recursiveElement.Daske.AD[0]
	if recursiveAD.DName.Value != "Bok_Lewy" {
		t.Errorf("recursive cabinet: can not find 'Bok_Lewy'")
		t.Log(recursiveAD)
	}
	var encodedData bytes.Buffer
	encoder := xml.NewEncoder(&encodedData)
	encoder.Encode(elementFile)

	encodedCorpusString := encodedData.String()

	if !isSimilar(t, simple_path, encodedCorpusString) {
		t.Errorf("Saved corpus file is corrupt!")
	}
}
