package main

import (
	"corpus_macro_replacer/corpus"
	"encoding/xml"
	"log"
	"regexp"
	"strings"
	"testing"
)

const pathToOriginal = "integrationTest/Kontenerek 9 w 1-original.E3D"

// const pathToDesired = "integrationTest/Kontenerek 9 w 1-desired-blum_standard.E3D"

var variablesToRemoveTest []string = []string{
	"ilosc_zawiasow",
	"producent_zawiasu",
	"typ_prowadnika_rodzaj",
	"typ_zawiasu_s",
	"zmiana_pozycji_dolnego",
	"zmiana_pozycji_gornego",
	"zmiana_pozycji_3",
	"zmiana_pozycji_4",
	"zmiana_pozycji_5",
	"zmiana_pozycji_6",
	"przesuniecie_prowadnik",
	"przesuniecie_zawias",
}

// var variablesToPreserveTest []string = []string{
// 	"ilosc_zawiasow_1",
// 	"typ_zawiasu_1",
// 	"rodzaj_zawiasu_1",
// 	"Przeliczenie_Blum",
// 	"mocowanie_puszki_1",
// 	"zawias_1",
// 	"typ_prowadnika_1",
// 	"mocowanie_prowadnika_1",
// 	"blumotion_1",
// 	"zmiana_pozycji_dolnego",
// 	"zmiana_pozycji_gornego",
// 	"zmiana_pozycji_3",
// 	"zmiana_pozycji_4",
// 	"zmiana_pozycji_5",
// 	"zmiana_pozycji_6",
// 	"przesuniecie_prowadnik",
// 	"przesuniecie_zawias",
// 	"pierwszy_zawiasBlum",
// 	"drugi_zawiasBlum",
// 	"trzeci_zawiasBlum",
// 	"czwarty_zawiasBlum",
// 	"piaty_zawiasBlum",
// 	"szosty_zawiasBlum",
// }

// func checkForWantedVars(t *testing.T, attrs []xml.Attr) {
// patterns_loop:
// 	for _, p := range variablesToPreserveTest {
// 		for _, attr := range attrs {
// 			if attr.Value == p {
// 				continue patterns_loop
// 			}
// 		}
// 		t.Logf("there is no attribute named %s", p)
// 		t.Fail()
// 	}
// }

func checkForUnwantedVars(t *testing.T, attrs []xml.Attr) {
	for _, attr := range attrs {
		for _, p := range variablesToRemoveTest {
			if strings.HasPrefix(attr.Value, p) {
				t.Logf("attribute %s should not have this variable: %s", attr, p)
				t.Fail()
			}
		}
	}
}

func TestRemoveVariableIntegration(t *testing.T) {
	var result *corpus.ElementFile
	{ // initialization
		projectFile, elementFile, err := corpus.NewCorpusFile(pathToOriginal)
		if err != nil {
			t.Logf("missing test data? %s", err)
			t.FailNow()
		}
		if projectFile != nil {
			t.Logf("projectFile should be empty %s", projectFile.FILE.Value)
			t.FailNow()
		} else if elementFile != nil {
			regexes := make([]*regexp.Regexp, len(variablesToRemove))
			for i, pattern := range variablesToRemove {
				r := regexp.MustCompile(pattern)
				regexes[i] = r
				// regexes = append(regexes, r)
			}
			result, err = RemoveVariablesFromFile(elementFile, regexes)
			if err != nil {
				t.Logf("there should be no errors: %s", err)
				t.FailNow()
			}
		} else {
			t.Log("should not happen")
			t.FailNow()
		}
	}

	for _, e := range result.Element {
		if strings.HasPrefix(e.EName.Value, "FR") {
			// checkForWantedVars(t, e.Evar.Attr)
			log.Printf("%s has %d variables", e.EName.Value, len(e.Evar.Attr))
		}
		checkForUnwantedVars(t, e.Evar.Attr)
		for _, sube := range e.ElmList.Elm {
			if strings.HasPrefix(sube.EName.Value, "FR") {
				// checkForWantedVars(t, sube.Evar.Attr)
				log.Printf("%s has %d variables", sube.EName.Value, len(sube.Evar.Attr))
			}
			checkForUnwantedVars(t, sube.Evar.Attr)
		}
	}

	// { // compare with desired (not needed for now)
	// 	projectFile, elementFile, err := corpus.NewCorpusFile(pathToDesired)
	// 	if err != nil {
	// 		t.Logf("missing test data? %s", err)
	// 		t.FailNow()
	// 	}
	// 	if projectFile != nil {
	// 		t.Logf("projectFile should be empty %s", projectFile.FILE.Value)
	// 		t.FailNow()
	// 	} else if elementFile != nil {
	// 	} else {
	// 		t.Log("should not happen")
	// 		t.FailNow()
	// 	}
	// }
}
