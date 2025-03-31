package main

import (
	"bytes"
	"compress/zlib"
	"encoding/base64"
	"encoding/xml"
	"fmt"
	"io"
	"log"
	"strings"
)

type GenericNode struct {
	XMLName  xml.Name
	Attr     []xml.Attr    `xml:",any,attr"`
	Content  []GenericNode `xml:",any"`
	Chardata string        `xml:",chardata"`
	// Comment  xml.Comment   `xml:",comment"`
}

// E3D file
type ElementFile struct {
	XMLName xml.Name `xml:"ELEMENTFILE"`
	GenericNode
	FILE    xml.Attr  `xml:"FILE,attr"`
	VER     xml.Attr  `xml:"VER,attr"`
	Element []Element `xml:"ELEMENT"`
}

// does nothing special, only fill in data for children
func (ef *ElementFile) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	type Alias ElementFile
	temp := &struct{ *Alias }{Alias: (*Alias)(ef)}

	if err := d.DecodeElement(&temp, &start); err != nil {
		return err
	}
	ef = (*ElementFile)(temp.Alias)
	for i := range ef.Element {
		ef.Element[i].Elinks.EncodeVersion = ef.VER.Value
	}
	return nil
}

// S3D file
type ProjectFile struct {
	XMLName xml.Name `xml:"PROJECTFILE"`
	ElementFile
}

// single cabinet
type Element struct {
	GenericNode
	EName   xml.Attr `xml:"ENAME,attr"`
	Daske   Daske    `xml:"DASKE"`
	ElmList ElmList  `xml:"ELMLIST"`
	Elinks  Elinks   `xml:"ELINKS"`
}

// list of nested elements
type ElmList struct {
	GenericNode
	ECount xml.Attr  `xml:"ECOUNT,attr"`
	Elm    []Element `xml:"ELM"`
}
type Daske struct {
	GenericNode
	DCount xml.Attr `xml:"DCOUNT,attr"`
	AD     []AD     `xml:"AD"`
}

// plank
type AD struct {
	GenericNode
	Potrosni GenericNode `xml:"POTROSNI"`
	// curves
	Krivulje GenericNode `xml:"KRIVULJE"`
	DName    xml.Attr    `xml:"DNAME,attr"`
}

type Elinks struct {
	EncodeVersion string `xml:"-"`
	GenericNode
	COUNT xml.Attr `xml:"COUNT,attr"`
	// version 16
	Spoj []Spoj `xml:"SPOJ"`
	// version 17
	MakLink []MakLink `xml:"MAKLINK"`
}

// encode
func (elink *Elinks) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
	// Start the main element
	var elinkCopy Elinks = *elink

	// 16 is our default version
	if elinkCopy.EncodeVersion == "16" || elinkCopy.EncodeVersion == "" {
		elinkCopy.MakLink = []MakLink{}
		if err := e.EncodeElement(elinkCopy, start); err != nil {
			return err
		}
	} else if elinkCopy.EncodeVersion == "17" {
		elinkCopy.Spoj = []Spoj{}
		if err := e.EncodeElement(elinkCopy, start); err != nil {
			return err
		}
	} else {
		return fmt.Errorf("saving to unsupported version: %s", elinkCopy.EncodeVersion)
	}

	// Close the CustomElement tag
	// return e.EncodeToken(start.End())
	return nil
}

// version 17
type MakLink struct {
	GenericNode
	/* index of: element.Daske.AD[O1] */
	OB1 xml.Attr `xml:"OB1,attr"`
	OB2 xml.Attr `xml:"OB2,attr"`
	CSP xml.Attr `xml:"CSP,attr"`
	/* unknown */
	SP  xml.Attr `xml:"SP,attr"`
	MM1 MM1      `xml:"MM1"`
	/* unknown, M2 seems to be unused */
	MM2 GenericNode `xml:"MM2"`
}

func NewMakLink(spoj *Spoj) (*MakLink, error) {
	makLink := MakLink{}
	makLink.GenericNode = spoj.GenericNode
	makLink.OB1.Value = spoj.O1.Value
	makLink.OB2.Value = spoj.O2.Value
	makLink.CSP.Value = spoj.SP.Value
	makNew, err := NewMM1(&spoj.Makro1)
	if err != nil {
		return nil, err
	}
	makLink.MM1 = *makNew
	// todo MM2 not supported...
	makLink.MM2 = spoj.Makro2
	return &makLink, nil
}

// version 16
type Spoj struct {
	GenericNode
	/* index of: element.Daske.AD[O1] */
	O1 xml.Attr `xml:"O1,attr"`
	/* unknown */
	O2 xml.Attr `xml:"O2,attr"`
	/* unknown */
	SP     xml.Attr `xml:"SP,attr"`
	Makro1 M1       `xml:"M1"`
	/* unknown, M2 seems to be unused */
	Makro2 GenericNode `xml:"M2"`
}

// used for version conversion: 17 -> 16
// version 16 scheme is treated as source of truth
func NewSpoj(makLink *MakLink) (*Spoj, error) {
	spoj := Spoj{}
	spoj.GenericNode = makLink.GenericNode
	spoj.O1.Value = makLink.OB1.Value
	spoj.O2.Value = makLink.OB2.Value
	spoj.SP.Value = makLink.CSP.Value
	makNew, err := NewM1(&makLink.MM1)
	if err != nil {
		return nil, err
	}
	spoj.Makro1 = *makNew
	// todo M2 not supported...
	spoj.Makro2 = GenericNode{}
	return &spoj, nil
}

type GenericNodeWithDat struct { // Varijable
	GenericNode
	DAT string `xml:"DAT,attr"`
}

// version 17
type GenericNodeWithC6Dat struct { // Varijable
	GenericNodeWithDat
	// explicitely ignore DAT, they usually do not occur together
	DAT   string `xml:"DAT,attr,omitempty"`
	C6DAT string `xml:"C6DAT,attr"`
}

func (gn *GenericNodeWithC6Dat) DecodeC6Dat() (string, error) {
	decoded, err := base64.StdEncoding.DecodeString(gn.C6DAT)
	if err != nil {
		return "", err
	}

	// Decompress using zlib
	reader, err := zlib.NewReader(bytes.NewReader(decoded))
	if err != nil {
		return "", err
	}
	defer reader.Close()

	// Read decompressed data
	var output bytes.Buffer
	_, err = io.Copy(&output, reader)
	if err != nil {
		return "", err
	}

	return output.String(), nil
}

func EncodeC6Dat(input string) (*string, error) {
	// Compress using zlib
	var compressed bytes.Buffer
	writer := zlib.NewWriter(&compressed)
	_, err := writer.Write([]byte(input))
	if err != nil {
		return nil, err
	}
	writer.Close()

	// Encode to base64
	C6DAT := base64.StdEncoding.EncodeToString(compressed.Bytes())
	return &C6DAT, nil
}

// version 16
type M1EmbeddedMakro struct {
	GenericNodeWithDat
	EmbeddedMakroName string `xml:"-"`
	MAK               *M1    `xml:"MAK,omitempty"`
}

// version 17
type MM1EmbeddedMakro struct {
	GenericNodeWithDat        // lol, still uses DAT, not C6DAT
	EmbeddedMakroName  string `xml:"-"`
	MAK                *MM1   `xml:"MAK,omitempty"`
}

// version 17 -> 16
func NewM1EmbeddedMakro(mm1EmbeddedMakro *MM1EmbeddedMakro) (*M1EmbeddedMakro, error) {
	m1 := M1EmbeddedMakro{}
	m1.Attr = mm1EmbeddedMakro.Attr
	m1.DAT = mm1EmbeddedMakro.DAT
	m1.Content = mm1EmbeddedMakro.Content
	m1.Chardata = mm1EmbeddedMakro.Chardata
	m1.EmbeddedMakroName = mm1EmbeddedMakro.EmbeddedMakroName
	if mm1EmbeddedMakro.MAK != nil {
		mak, err := NewM1(mm1EmbeddedMakro.MAK)
		if err != nil {
			return nil, err
		}
		m1.MAK = mak
	}
	return &m1, nil
}

// version 16 -> 17
func NewMM1EmbeddedMakro(m1EmbeddedMakro *M1EmbeddedMakro) (*MM1EmbeddedMakro, error) {
	mm1 := MM1EmbeddedMakro{}
	mm1.Attr = m1EmbeddedMakro.Attr
	mm1.DAT = m1EmbeddedMakro.DAT
	mm1.Content = m1EmbeddedMakro.Content
	mm1.Chardata = m1EmbeddedMakro.Chardata
	mm1.EmbeddedMakroName = m1EmbeddedMakro.EmbeddedMakroName
	if m1EmbeddedMakro.MAK != nil {
		mak, err := NewMM1(m1EmbeddedMakro.MAK)
		if err != nil {
			return nil, err
		}
		mm1.MAK = mak
	}
	return &mm1, nil
}

// extracts name of submacro that is used in Corpus call in [MAKRO] section
func (em *M1EmbeddedMakro) CalledWith() string {
	for _, line := range decodeAllCMKLines(em.DAT) {
		nameAndValue := strings.SplitN(line, "=", 2)
		if strings.ToLower(nameAndValue[0]) == "name" {
			if len(nameAndValue) != 2 {
				log.Printf("Error, I do not know how to read name from EmbeddedMakro. Is there '=' in file name? %s", line)
				return ""
			}
			return nameAndValue[1]
		}
	}
	return ""
}

/*
Represents makro as defined in Corpus 5.0 (reverse engineered).
Help is available only in Corpus makro editor.

Note: names come from Croatian language

Example value:

	{
		<M1 MN="">
			<MSVA DAT="WPUST_GLEBOKOSC=13,NUMER_NARZEDZIA=155"></MSVA>
			<MSFO DAT="pila_grubosc=4"></MSFO>
			<MSPI DAT="J=1,GB=if(typ_plecow=3;1;0),&#34;GN=frezowanie pila&#34;,GD=wpust_glebokosc,..."></MSPI>
			<MSJO DAT="CONNECT=2345,mindistance=-16,maxdistance=10"></MSJO>
		</M1>
	}

version 16
*/
type M1 struct {
	GenericNode
	/* MN is not obligatory. Empty names means that makro is not save in any file. */
	MakroName string               `xml:"MN,attr"`
	Varijable GenericNodeWithDat   `xml:"MSVA"`
	Formule   *GenericNodeWithDat  `xml:"MSFO,omitempty"`
	Pila      []GenericNodeWithDat `xml:"MSPI,omitempty"`
	Joint     *GenericNodeWithDat  `xml:"MSJO,omitempty"`
	Grupa     []GenericNodeWithDat `xml:"MSGR,omitempty"`
	Potrosni  []GenericNodeWithDat `xml:"MSPO,omitempty"`
	Pocket    []GenericNodeWithDat `xml:"MSPOCK,omitempty"`
	Raster    []GenericNodeWithDat `xml:"MSRA,omitempty"`
	Makro     []M1EmbeddedMakro    `xml:"MSMA,omitempty"`
}

// version 17
type MM1 struct {
	GenericNode
	/* MN is not obligatory. Empty names means that makro is not save in any file. */
	MakroName string                 `xml:"MN,attr"`
	Varijable GenericNodeWithC6Dat   `xml:"MSVA"`
	Formule   *GenericNodeWithC6Dat  `xml:"MSFO,omitempty"`
	Pila      []GenericNodeWithC6Dat `xml:"MSPI,omitempty"`
	Joint     *GenericNodeWithC6Dat  `xml:"MSJO,omitempty"`
	Grupa     []GenericNodeWithC6Dat `xml:"MSGR,omitempty"`
	Potrosni  []GenericNodeWithC6Dat `xml:"MSPO,omitempty"`
	Pocket    []GenericNodeWithC6Dat `xml:"MSPOCK,omitempty"`
	Raster    []GenericNodeWithC6Dat `xml:"MSRA,omitempty"`
	Makro     []MM1EmbeddedMakro     `xml:"MSMA,omitempty"`
}

// decoding version 17 -> 16
func NewM1(mm1 *MM1) (*M1, error) {
	m1 := M1{
		GenericNode: mm1.GenericNode,
		MakroName:   mm1.MakroName,
	}
	{
		decoded, err := mm1.Varijable.DecodeC6Dat()
		if err != nil {
			return nil, err
		}
		m1.Varijable = GenericNodeWithDat{GenericNode: mm1.Varijable.GenericNode, DAT: decoded}
	}

	if mm1.Formule != nil {
		decoded, err := mm1.Formule.DecodeC6Dat()
		if err != nil {
			return nil, err
		}
		m1.Formule = &GenericNodeWithDat{GenericNode: mm1.Formule.GenericNode, DAT: decoded}
	}

	m1.Pila = make([]GenericNodeWithDat, len(mm1.Pila))
	for i, pilaEncoded := range mm1.Pila {
		decoded, err := pilaEncoded.DecodeC6Dat()
		if err != nil {
			return nil, err
		}
		m1.Pila[i].DAT = decoded
	}

	if mm1.Joint != nil {
		decoded, err := mm1.Joint.DecodeC6Dat()
		if err != nil {
			return nil, err
		}
		m1.Joint = &GenericNodeWithDat{GenericNode: mm1.Joint.GenericNode, DAT: decoded}
	}

	m1.Grupa = make([]GenericNodeWithDat, len(mm1.Grupa))
	for i, encodedItem := range mm1.Grupa {
		decoded, err := encodedItem.DecodeC6Dat()
		if err != nil {
			return nil, err
		}
		m1.Grupa[i].GenericNode = encodedItem.GenericNode
		m1.Grupa[i].DAT = decoded
	}

	m1.Potrosni = make([]GenericNodeWithDat, len(mm1.Potrosni))
	for i, encodedItem := range mm1.Potrosni {
		decoded, err := encodedItem.DecodeC6Dat()
		if err != nil {
			return nil, err
		}
		m1.Potrosni[i].GenericNode = encodedItem.GenericNode
		m1.Potrosni[i].DAT = decoded
	}

	m1.Pocket = make([]GenericNodeWithDat, len(mm1.Pocket))
	for i, encodedItem := range mm1.Pocket {
		decoded, err := encodedItem.DecodeC6Dat()
		if err != nil {
			return nil, err
		}
		m1.Pocket[i].GenericNode = encodedItem.GenericNode
		m1.Pocket[i].DAT = decoded
	}

	m1.Raster = make([]GenericNodeWithDat, len(mm1.Raster))
	for i, encodedItem := range mm1.Raster {
		decoded, err := encodedItem.DecodeC6Dat()
		if err != nil {
			return nil, err
		}
		m1.Raster[i].GenericNode = encodedItem.GenericNode
		m1.Raster[i].DAT = decoded
	}

	m1.Makro = make([]M1EmbeddedMakro, len(mm1.Makro))
	for i, mm1EmbeddedMakro := range mm1.Makro {
		// decoded, err := encodedItem.DecodeC6Dat()
		embedded, err := NewM1EmbeddedMakro(&mm1EmbeddedMakro)
		if err != nil {
			return nil, err
		}
		m1.Makro[i] = *embedded
	}

	return &m1, nil
}

// decoding version 16 -> 17
func NewMM1(m1 *M1) (*MM1, error) {
	mm1 := MM1{
		GenericNode: m1.GenericNode,
		MakroName:   m1.MakroName,
	}
	{
		encoded, err := EncodeC6Dat(m1.Varijable.DAT)
		if err != nil {
			return nil, err
		}
		mm1.Varijable = GenericNodeWithC6Dat{GenericNodeWithDat: m1.Varijable, C6DAT: *encoded}
		mm1.Varijable.DAT = ""
	}

	if m1.Formule != nil {
		encoded, err := EncodeC6Dat(m1.Formule.DAT)
		if err != nil {
			return nil, err
		}
		mm1.Formule = &GenericNodeWithC6Dat{GenericNodeWithDat: *m1.Formule, C6DAT: *encoded}
	}

	mm1.Pila = make([]GenericNodeWithC6Dat, len(m1.Pila))
	for i, pilaEncoded := range m1.Pila {
		encoded, err := EncodeC6Dat(pilaEncoded.DAT)
		if err != nil {
			return nil, err
		}
		mm1.Pila[i].DAT = *encoded
	}

	if m1.Joint != nil {
		encoded, err := EncodeC6Dat(m1.Joint.DAT)
		if err != nil {
			return nil, err
		}
		mm1.Joint = &GenericNodeWithC6Dat{GenericNodeWithDat: *m1.Joint, C6DAT: *encoded}
	}

	mm1.Grupa = make([]GenericNodeWithC6Dat, len(m1.Grupa))
	for i, encodedItem := range m1.Grupa {
		encoded, err := EncodeC6Dat(encodedItem.DAT)
		if err != nil {
			return nil, err
		}
		mm1.Grupa[i].GenericNode = encodedItem.GenericNode
		mm1.Grupa[i].DAT = *encoded
	}

	mm1.Potrosni = make([]GenericNodeWithC6Dat, len(m1.Potrosni))
	for i, encodedItem := range m1.Potrosni {
		encoded, err := EncodeC6Dat(encodedItem.DAT)
		if err != nil {
			return nil, err
		}
		mm1.Potrosni[i].GenericNode = encodedItem.GenericNode
		mm1.Potrosni[i].DAT = *encoded
	}

	mm1.Pocket = make([]GenericNodeWithC6Dat, len(m1.Pocket))
	for i, encodedItem := range m1.Pocket {
		encoded, err := EncodeC6Dat(encodedItem.DAT)
		if err != nil {
			return nil, err
		}
		mm1.Pocket[i].GenericNode = encodedItem.GenericNode
		mm1.Pocket[i].DAT = *encoded
	}

	mm1.Raster = make([]GenericNodeWithC6Dat, len(m1.Raster))
	for i, encodedItem := range m1.Raster {
		encoded, err := EncodeC6Dat(encodedItem.DAT)
		if err != nil {
			return nil, err
		}
		mm1.Raster[i].GenericNode = encodedItem.GenericNode
		mm1.Raster[i].DAT = *encoded
	}

	mm1.Makro = make([]MM1EmbeddedMakro, len(m1.Makro))
	for i, m1EmbeddedMakro := range m1.Makro {
		embedded, err := NewMM1EmbeddedMakro(&m1EmbeddedMakro)
		if err != nil {
			return nil, err
		}
		mm1.Makro[i] = *embedded
	}

	return &mm1, nil
}

// save makro as CMK file
func (m *M1) Save(w io.Writer) error {
	if _, err := w.Write([]byte("[VARIJABLE]\n")); err != nil {
		return err
	}
	{
		text := strings.Join(decodeAllCMKLines(m.Varijable.DAT), "\n")
		if _, err := w.Write([]byte(text)); err != nil {
			return err
		}
	}
	if m.Joint != nil {
		if _, err := w.Write([]byte("\n\n[JOINT]\n")); err != nil {
			return err
		}
		text := strings.Join(decodeAllCMKLines(m.Joint.DAT), "\n")
		if _, err := w.Write([]byte(text)); err != nil {
			return err
		}
	}
	if m.Formule != nil {
		if _, err := w.Write([]byte("\n\n[FORMULE]\n")); err != nil {
			return err
		}
		text := strings.Join(decodeAllCMKLines(m.Formule.DAT), "\n")
		if _, err := w.Write([]byte(text)); err != nil {
			return err
		}
	}

	for i, item := range m.Pocket {
		section := fmt.Sprintf("\n\n[POCKET%d]\n", i)
		if _, err := w.Write([]byte(section)); err != nil {
			return err
		}
		text := strings.Join(decodeAllCMKLines(item.DAT), "\n")
		if _, err := w.Write([]byte(text)); err != nil {
			return err
		}
	}
	for i, item := range m.Potrosni {
		section := fmt.Sprintf("\n\n[POTROSNI%d]\n", i)
		if _, err := w.Write([]byte(section)); err != nil {
			return err
		}
		text := strings.Join(decodeAllCMKLines(item.DAT), "\n")
		if _, err := w.Write([]byte(text)); err != nil {
			return err
		}
	}
	for i, item := range m.Grupa {
		section := fmt.Sprintf("\n\n[GRUPA%d]\n", i)
		if _, err := w.Write([]byte(section)); err != nil {
			return err
		}
		text := strings.Join(decodeAllCMKLines(item.DAT), "\n")
		if _, err := w.Write([]byte(text)); err != nil {
			return err
		}
	}
	for i, item := range m.Raster {
		section := fmt.Sprintf("\n\n[RASTER%d]\n", i)
		if _, err := w.Write([]byte(section)); err != nil {
			return err
		}
		text := strings.Join(decodeAllCMKLines(item.DAT), "\n")
		if _, err := w.Write([]byte(text)); err != nil {
			return err
		}
	}
	for i, item := range m.Makro {
		section := fmt.Sprintf("\n\n[MAKRO%d]\n", i)
		if _, err := w.Write([]byte(section)); err != nil {
			return err
		}
		text := strings.Join(decodeAllCMKLines(item.DAT), "\n")
		if _, err := w.Write([]byte(text)); err != nil {
			return err
		}
	}

	return nil
}

func (ef *ElementFile) VisitElementsAndSubelements(f func(*Element)) {
	for _, e := range ef.Element {
		e.VisitElementsAndSubelements(f)
	}
}

func (e *Element) VisitElementsAndSubelements(f func(*Element)) {
	f(e)
	for _, sube := range e.ElmList.Elm {
		sube.VisitElementsAndSubelements(f)
	}
}

// visit self and submakros
func (m *M1) VisitSubmakros(f func(parent *M1, embededParent *M1EmbeddedMakro, child *M1EmbeddedMakro)) {
	f(m, nil, nil) // entrypoiny
	m.partialVisitSubmakros(nil, f)
}

func (m *M1) partialVisitSubmakros(embededParent *M1EmbeddedMakro, f func(parent *M1, embededParent *M1EmbeddedMakro, child *M1EmbeddedMakro)) {
	for _, submacro := range m.Makro {
		f(m, embededParent, &submacro)
		submacro.MAK.partialVisitSubmakros(&submacro, f)
	}
}

// // visit self and submakros
// func (em *EmbeddedMakro) VisitSubmakros(f func(parent *M1, child *EmbeddedMakro)) {
// 	em.MAK.VisitSubmakros(f)
// }
