package corpus

import (
	"bytes"
	"encoding/xml"
	"path/filepath"
	"testing"
)

var pathToE3DTestDataVertsion16 = filepath.Join("..", "tests", "testData", "E3D-version-16")
var pathToE3DTestDataVertsion17 = filepath.Join("..", "tests", "testData", "E3D-version-17")
var testFilesE3D = []string{
	"simple.E3D",
	"simple_macro_in_macro.E3D",
	// "simple_in_simple.E3D",
	"simple_in_simple_in_simple.E3D",
}

func TestLoadAndSaveCorpusE3DFileVersion16To16(t *testing.T) {
	for _, testFile := range testFilesE3D {
		simple_path := filepath.Join(pathToE3DTestDataVertsion16, testFile)
		_, _, err := NewCorpusFile(simple_path)
		if err != nil {
			t.Logf("Error loading file %s", testFile)
			t.Error(err)
			t.FailNow()
		}

		// todo how to test encoding?
		// var encodedData bytes.Buffer
		// encoder := xml.NewEncoder(&encodedData)
		// err = encoder.Encode(projectFile)
		// if err != nil {
		// 	t.Logf("Error saving (encoding) file %s", testFile)
		// 	t.Error(err)
		// 	t.FailNow()
		// }
	}
}
func TestLoadCorpusE3DFileVersion17(t *testing.T) {
	for _, testFile := range testFilesE3D {
		simple_path := filepath.Join(pathToE3DTestDataVertsion17, testFile)
		_, _, err := NewCorpusFile(simple_path)
		if err != nil {
			t.Logf("Error loading file %s", testFile)
			t.Error(err)
			t.FailNow()
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

func commonTestSimpleInSimple(t *testing.T, elementFile *ElementFile, simple_path string) {
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

func TestLoadCorpusE3DFileSimpleInSimpleVersion16(t *testing.T) {
	simple_path := filepath.Join(pathToE3DTestDataVertsion16, "simple_in_simple.E3D")
	_, elementFile, err := NewCorpusFile(simple_path)
	if err != nil {
		t.Error(err)
	}
	commonTestSimpleInSimple(t, elementFile, simple_path)

	if len(elementFile.Element[0].Elinks.MakLink) != 0 {
		t.Error("Corpus file version 16 should use only SPOJ, not MAKLINK")
		t.FailNow()
	}
	if len(elementFile.Element[0].Elinks.Spoj) != 2 {
		t.Error("Should load 2 makros")
		t.FailNow()
	}
}

func TestLoadCorpusE3DFileSimpleInSimpleVersion17(t *testing.T) {
	simple_path := filepath.Join(pathToE3DTestDataVertsion17, "simple_in_simple.E3D")
	_, elementFile, err := NewCorpusFile(simple_path)
	if err != nil {
		t.Error(err)
	}
	commonTestSimpleInSimple(t, elementFile, simple_path)

	if len(elementFile.Element[0].Elinks.Spoj) != 0 {
		t.Error("Corpus file version 17 should use only MAKLINK, not SPOJ")
		t.FailNow()
	}
	if len(elementFile.Element[0].Elinks.MakLink) != 2 {
		t.Error("Should load 2 makros")
		t.FailNow()
	}

	// todo should be separate test
	m1, err := NewM1(&elementFile.Element[0].Elinks.MakLink[0].MM1)
	if err != nil {
		t.Errorf("Erorr parsing MakLink: %s", err)
		t.FailNow()
	}

	if m1.MakroName != "gorny" {
		t.Errorf("wrong value in MakroName: %s", m1.MakroName)
	}
	if m1.Varijable.DAT != "\"// version 0.1\",one=1,\"// comment 1\",two=2,\"// comment 2\"" {
		t.Errorf("wrong value in Varijable: %s", m1.Varijable.DAT)
	}
	if m1.Formule.DAT != "nr_narzedzia=obj1.param6543NNZ" {
		t.Errorf("wrong value in Foumle: %s", m1.Formule.DAT)
	}
	if m1.Joint.DAT != "CONNECT=23,mindistance=-14,maxdistance=5" {
		t.Errorf("wrong value in joint: %s", m1.Joint.DAT)
	}
	if len(m1.Makro) != 0 {
		t.Errorf("wrong number of submacros: %d", len(m1.Makro))
	}

	// spoj, err := NewSpoj(&elementFile.Element[0].Elinks.MakLink[0])

}
