package main

import (
	"corpus_macro_replacer/corpus"
	"path/filepath"
	"regexp"
	"testing"
)

var pathToE3DTestData = filepath.Join("..", "..", "..", "tests", "testData", "E3D-version-16")

func TestRemoveVariable(t *testing.T) {
	projectFile, elementFile, err := corpus.NewCorpusFile(filepath.Join(pathToE3DTestData, "nested_variables.E3D"))
	if err != nil {
		t.Logf("missing test data? %s", err)
		t.FailNow()
	}
	if projectFile != nil {
		t.Logf("projectFile should be empty %s", projectFile.FILE.Value)
		t.FailNow()
	} else if elementFile != nil {
		r := regexp.MustCompile(".*")
		regexes := []*regexp.Regexp{r}
		result, err := RemoveVariablesFromFile(elementFile, regexes)
		if err != nil {
			t.Logf("there should be no errors: %s", err)
			t.FailNow()
		}
		for _, e := range result.Element {
			if len(e.Evar.Attr) != 0 {
				t.Logf("%s: those attributes should be emtpy: %s", e.EName.Value, e.Evar)
				t.Fail()
			}
			for _, sube := range e.ElmList.Elm {
				if len(sube.Evar.Attr) != 0 {
					t.Logf("%s->%s: those attributes should be emtpy: %s", e.EName.Value, sube.EName.Value, sube.Evar)
					t.Fail()
				}
			}
		}
	} else {
		t.Log("should not happen")
		t.FailNow()
	}
}
