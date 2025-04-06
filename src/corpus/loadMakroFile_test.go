package corpus

import (
	"path/filepath"
	"testing"
)

var pathToCMKTestData = filepath.Join("..", "tests", "testData", "CMK")
var testFilesCMK = []string{
	"simple.CMK",
	"loadMakroByFilename.CMK",
	"loadMakroByNamedMakro.CMK",
}

func TestNewMakroFromCMKFileSimple(t *testing.T) {
	makro, err := NewMakroFromCMKFile(nil,
		filepath.Join(pathToCMKTestData, "simple.CMK"),
		nil, nil,
	)
	if err != nil {
		t.Error(err)
	}
	if makro.MakroName != "simple" {
		t.Errorf("makro name is bad: %s", makro.MakroName)
		t.Log(makro.MakroName)
	}
	// the rest is same as TestPartialNewMakroFromCMKFileSimple

	checkMakroSimple(makro, t)
}

func TestPartialNewMakroFromCMKFileSimple(t *testing.T) {
	makro, err := partialNewMakroFromCMKFile("",
		filepath.Join(pathToCMKTestData, "simple.CMK"),
	)
	if err != nil {
		t.Error(err)
	}
	if makro.MakroName != "" {
		t.Errorf("makro name is bad %s", makro.MakroName)
		t.Log(makro.MakroName)
	}
	checkMakroSimple(makro, t)
}

func checkMakroSimple(makro *M1, t *testing.T) {
	if makro.Varijable.DAT != "x=0" {
		t.Errorf("Varijable is bad %s", makro.Varijable.DAT)
		t.Log(makro.Varijable)
	}
	if makro.Joint.DAT != "CONNECT=23,mindistance=-14,//maxdistance=10" {
		t.Errorf("Joint is bad %s", makro.Joint.DAT)
		t.Log(makro.Joint)
	}
	if makro.Formule.DAT != "nr_narzedzia_dno=obj1.param9876NR_NARZEDZIA_DNO" {
		t.Errorf("Formule is bad %s", makro.Formule.DAT)
		t.Log(makro.Formule)
	}
	if len(makro.Pila) != 1 || makro.Pila[0].DAT != "J=1,GB=if(obj1.param9876FREZ_DNO=0;0;1),\"GN=rowek na dno\",GD=wpust_glebokosc_dno,GX=pmaxx,GY=-5,PX=pmaxx,PY=obj2.maxy+5,GS=obj2.autost,PSP=frez_srednica_dno,PO=0,PS=1,PP=1,PA=0,PMU=1,PMT=nr_narzedzia_dno" {
		t.Errorf("Pila is bad %s", makro.Pila)
	}
	if len(makro.Grupa) != 1 || makro.Grupa[0].DAT != "J=1,GB=if(Testczykolekdodatkowy=1;2+dodaj_nawiert-czy_listwa;0),\"GN=kolki wiercone w obiekcie przylegajacym\",GX=obj1.gr/2,GY=0,GS=obj2.autost,GK=0,GP=0,RX1=0,\"RY1=nawiert od krawedzi+Kolekkonfirmat+ KolekMinifix+ KolekVB35+ KolekVB36 + KolekWkret\",RF1=obj1.param500SK,RD1=obj1.param500GPlus,RX2=0,\"RY2=pmaxy-nawiert od krawedzi-Kolekkonfirmat - KolekMinifix - KolekVB35 - KolekVB36 - KolekWkret\",RF2=obj1.param500SK,RD2=obj1.param500GPlus,RX3=0,\"RY3=(pmaxy-pminy)/2-Kolekkonfirmat - KolekMinifix - KolekVB35 - KolekVB36 - KolekWkret\",RF3=obj1.param500SK,RD3=obj1.param500GPlus" {
		t.Errorf("Grupa is bad %s", makro.Grupa)
	}
	if len(makro.Potrosni) != 1 || makro.Potrosni[0].DAT != "\"//=Frezowanie dna antaro\",J=0,RT=0,GB=1,PP1=1,\"PS1=Frezowanie dna antaro\",PK1=1" {
		t.Errorf("Potrosni is bad %s", makro.Potrosni)
	}
	if len(makro.Pocket) != 1 || makro.Pocket[0].DAT != "J=0,GB=if((Hafele_Zawieszki_Wybor=0)and((Hafele_Zawieszki_plecy=0)or(Hafele_Zawieszki_plecy=1));1;0),GN=Scrapi_Lewa,GD=obj1.grubosc+Hafele_extra_zejscie_freza,GX=(11/2)+wpust_boki,GY=obj1.wysokosc-(42/2)-wpust_wieniec,GS=Hafele_strona_HDF,GK=0,GH=42,GW=11,GCR=Hafele_srednica_freza/2,GSD=0,GXY=80,GFE=5,PMT=Hafele_Numer,CUT=if(Hafele_Zawieszki_plecy=0;0;1)" {
		t.Errorf("Pocket is bad %s", makro.Pocket)
	}
	if len(makro.Raster) != 1 || makro.Raster[0].DAT != "J=1,GB=if(parent.parent.obj1.param8010WL=0;0;2),GN=raster1,GD=parent.parent.obj1.param8010GN,GF=parent.parent.obj1.param8010SN,GX=7,GY=7,GS=obj2.autost,GK=0,GP=parent.parent.obj1.param8010TN,GR=38" {
		t.Errorf("Raster is bad %s", makro.Raster)
	}
	if makro.Makro != nil {
		t.Errorf("Sub-makro is bad")
		t.Log(makro.Makro)
	}
}

func TestNewMakroFromCMKFileLoadMakroByNamedMakroBasicNameMapping(t *testing.T) {
	var makroName *string = nil
	var makroFile string = filepath.Join(pathToCMKTestData, "loadMakroByNamedMakro.CMK")
	makro, err := NewMakroFromCMKFile(makroName, makroFile, nil,
		// makro mapping are relative to Makro folder, ex. The root path is `C:\Tri D Corpus\Corpus 6.0\Makro`, and the value is `./Sevroll Zawiasy/Flod_lewy.CMK`
		MakroMappings{"creative_user_wants_to_load_simple": "simple.CMK"},
	)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	if makro.MakroName != "loadMakroByNamedMakro" {
		t.Errorf("makro name is bad: %s", makro.MakroName)
		t.Log(makro.MakroName)
	}
	if len(makro.Makro) != 1 {
		t.Errorf("submakro is bad")
		t.Log(makro.Makro)
		t.FailNow()
	}
	submakro := makro.Makro[0]
	if submakro.EmbeddedMakroName != "creative_user_wants_to_load_simple" {
		t.Errorf("submakro name is bad: %s", submakro.EmbeddedMakroName)
		checkMakroSimple(submakro.MAK, t)
	}
	if submakro.DAT != "J=0,RT=0,NAME=creative_user_wants_to_load_simple,MB=1,MA=1,INDEX=1,LACZ_BLENDA=,przesuniecie_lewej=,przesuniecie_prawej=,PODAJ_GRUBOSC_PLYTY=,STRONA_NAWIERTU_PUSZKA=,ilosc_nawiertow_srodkowych=" {
		t.Errorf("submakro DAT is bad: %s", submakro.DAT)
	}
}
func TestNewMakroFromCMKFileLoadMakroByNameBasicNameMappingFails(t *testing.T) {
	makro, err := NewMakroFromCMKFile(nil,
		filepath.Join(pathToCMKTestData, "loadMakroByNamedMakro.CMK"),
		nil,
		nil,
	)
	if makro != nil {
		t.Error("makro should be nil")
		t.Log(makro)
	}
	targetErr, ok := err.(*CMKUnknownMakroError)
	if !ok || targetErr.Name != "creative_user_wants_to_load_simple" {
		t.Error("error should be type UnknownMakroError")
		t.Log(err)
	}
}

func TestNewMakroFromCMKFileLoadMakroByNameFindInFilesFails(t *testing.T) {
	var makroRootPath = "path_to_macro_root_that_does_not_exist"
	makro, err := NewMakroFromCMKFile(nil,
		filepath.Join(pathToCMKTestData, "loadMakroByFilename.CMK"),
		&makroRootPath,
		nil,
	)
	if makro != nil {
		t.Error("makro should be nil")
		t.Log(makro)
	}
	targetErr, ok := err.(*CMKUnknownMakroError)
	if !ok || targetErr.Name != "simple" {
		t.Error("error should be type UnknownMakroError")
		t.Log(err)
	}
}

func TestNewMakroFromCMKFileLoadMakroByNameFindInFilesOk(t *testing.T) {
	makro, err := NewMakroFromCMKFile(nil,
		filepath.Join(pathToCMKTestData, "loadMakroByFilename.CMK"),
		nil, nil,
	)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	if makro.MakroName != "loadMakroByFilename" {
		t.Errorf("makro name is bad: %s", makro.MakroName)
		t.Log(makro.MakroName)
	}
	if len(makro.Makro) != 1 {
		t.Errorf("submakro is bad")
		t.Log(makro.Makro)
		t.FailNow()
	}
	submakro := makro.Makro[0]
	if submakro.EmbeddedMakroName != "simple" {
		t.Errorf("submakro name is bad: %s", submakro.EmbeddedMakroName)
		checkMakroSimple(submakro.MAK, t)
	}
	if submakro.DAT != "J=0,RT=0,NAME=simple,MB=1,MA=1,INDEX=1,LACZ_BLENDA=,przesuniecie_lewej=,przesuniecie_prawej=,PODAJ_GRUBOSC_PLYTY=,STRONA_NAWIERTU_PUSZKA=,ilosc_nawiertow_srodkowych=" {
		t.Errorf("submakro DAT is bad: %s", submakro.DAT)
	}
}
