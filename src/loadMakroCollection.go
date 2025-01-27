package main

import (
	"bufio"
	"errors"
	"fmt"
	"image/color"
	"io"
	"log"
	"os"
	"path/filepath"
	"unicode/utf8"
)

type MakroCollection []MakroCollectionItem

func (mc *MakroCollection) GetMacroNameByFileName(path string) *string {
	// if MakroCollectionCache == nil {
	// 	return nil
	// }
	for _, mcc := range MakroCollectionCache {
		absCorpusPath, err1 := filepath.Abs(string(mcc.FileName))
		absMakroPath, err2 := filepath.Abs(path)
		if err1 != nil && err2 != nil && absCorpusPath == absMakroPath {
			tmp := string(mcc.Name)
			return &tmp
		}
	}
	return nil
}
func (mc *MakroCollection) GetMacroFileNameByName(name string) *string {
	for _, mcc := range *mc {
		if mcc.Name == name {
			tmp := string(mcc.FileName)
			return &tmp
		}
	}
	return nil
}
func (mc *MakroCollection) GetMacroMappings() MakroMappings {
	out := MakroMappings{}
	for _, mcc := range *mc {
		out[mcc.Name] = mcc.FileName
	}
	return out
}

type MakroCollectionItem struct {
	FileName    string
	Name        string
	Category    string
	TextColorFG color.Color
	TextColorBG color.Color
}

const (
	KWMakroCollection      = "mkc"
	KWItems                = "items"
	KWMakroCollectionItem  = "mki"
	KWMakroName            = "cap"
	KWMakroCategory        = "cat"
	KWMakroFileName        = "fn"
	KWMakroBackgroundColor = "bc"
	KWMakroForegroundColor = "fc"
	KWUnknownUQ            = "UQ"
)

const (
	HexDoNothing = 0x00 // Workaround, if the parsing was 100 correct this would not be needed
	HexPadding   = 0x01 // not sure if this is correct
	// marks unknown seqence of bytes with ascii "UQ"
	HexUnknownUQ            = 0x02
	HexUnknownUQLen         = 2
	HexUnknownMaybeIndex    = 0x03
	HexUnknownMaybeIndexLen = 2
	HexString               = 0x06 // dynamic len
	HexSection              = 0x07 // dynamic len
	HexPadding4Byte         = 0x12
	HexPadding4ByteLen      = 4
	HexStringUtf            = 0x14
)

func LoadMakroCollection(path string) (MakroCollection, error) {
	var result = MakroCollection{}
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}

	br := bufio.NewReader(f)
	var lastItem *MakroCollectionItem = nil
	for {
		b, err := br.ReadByte()
		if err != nil && !errors.Is(err, io.EOF) {
			return nil, err
		}
		switch b {
		case HexUnknownMaybeIndex:
			if ReadSkip(br, HexUnknownMaybeIndexLen) != nil {
				return nil, err
			}
		case HexSection:
			sec, err := ReadLenAndSection(br)
			if err != nil {
				return nil, err
			}
			// fmt.Print("Section: ")
			// fmt.Println(*sec)
			switch *sec {
			case KWMakroCollectionItem:
				lastItem = &MakroCollectionItem{}
				result = append(result, *lastItem)
			}
		case HexString:
			keyWord, err := ReadLenAndString(br)
			if err != nil {
				return nil, err
			}
			// fmt.Print("Key: ")
			// fmt.Println(*keyWord)
			lastValidKW := *keyWord
			switch lastValidKW {
			case KWMakroName:
				value, err := ReadKWAndLenAndString(br)
				if err != nil {
					return nil, err
				}
				// fmt.Printf("Value: %s\n", *value)
				lastItem.Name = *value
			case KWMakroCategory:
				value, err := ReadKWAndLenAndString(br)
				if err != nil {
					return nil, err
				}
				lastItem.Category = *value
			case KWMakroFileName:
				value, err := ReadKWAndLenAndString(br)
				if err != nil {
					return nil, err
				}
				lastItem.FileName = *value
			case KWMakroForegroundColor:
				color, err := ReadLenAndColor(br)
				if err != nil {
					return nil, err
				}
				lastItem.TextColorFG = color
			case KWMakroBackgroundColor:
				color, err := ReadLenAndColor(br)
				if err != nil {
					return nil, err
				}
				lastItem.TextColorBG = color
			}
		case HexDoNothing:
		}
		if err != nil {
			// end of file
			break
		}
	}
	return result, nil
}

func ReadSkip(r io.Reader, n uint8) error {
	_, err := io.ReadAll(io.LimitReader(r, int64(n)))
	return err
}

// read exactly n bytes
func ReadKWAndLenAndString(r *bufio.Reader) (*string, error) {
	b, err := r.ReadByte()
	if err != nil {
		return nil, err
	}
	if b == HexString {
		return ReadLenAndString(r)
	} else if b == HexPadding4Byte {
		err := ReadSkip(r, HexPadding4ByteLen)
		if err == nil {
			temp := ""
			return &temp, nil
		} else {
			return nil, err
		}
	} else if b == HexStringUtf {
		return ReadLenAndUFT8String(r)
	} else {
		return nil, fmt.Errorf("byte does not indicate string: %d", b)
	}
}

func ReadLenAndUFT8String(r *bufio.Reader) (*string, error) {
	b, err := r.ReadByte()
	if err != nil {
		return nil, err
	}
	n := int(uint8(b))

	b, err = r.ReadByte()
	if err != nil {
		return nil, err
	}
	if b != 0 {
		log.Printf("warning when reading utf8 string: not zero")
	}
	b, err = r.ReadByte()
	if err != nil {
		return nil, err
	}
	if b != 0 {
		log.Printf("warning when reading utf8 string: not zero")
	}
	b, err = r.ReadByte()
	if err != nil {
		return nil, err
	}
	if b != 0 {
		log.Printf("warning when reading utf8 string: not zero")
	}
	var runes []rune = []rune{}
	for len(runes) < n-2 {
		// Read a single byte
		b, err := r.ReadByte()
		if err != nil {
			return nil, fmt.Errorf("error reading file: %w", err)
		}

		// Peek more bytes to determine the complete rune
		bytes := []byte{b}
		for !utf8.FullRune(bytes) {
			nextByte, err := r.ReadByte()
			if err != nil {
				return nil, fmt.Errorf("error reading file: %w", err)
			}
			bytes = append(bytes, nextByte)
		}
		r, size := utf8.DecodeRune(bytes)
		if r == utf8.RuneError && size == 1 {
			return nil, fmt.Errorf("invalid UTF-8 sequence encountered")
		}
		runes = append(runes, r)
	}
	tmp := string(runes)
	return &tmp, nil
}

func ReadLenAndString(r *bufio.Reader) (*string, error) {
	b, err := r.ReadByte()
	if err != nil {
		return nil, err
	}
	n := uint8(b)
	bytes, err := io.ReadAll(io.LimitReader(r, int64(n)))
	if err != nil {
		return nil, err
	}
	out := string(bytes)
	return &out, nil
}

// read untill you find 0x00 but take only n bytes and discard the rest
func ReadLenAndSection(r *bufio.Reader) (*string, error) {
	bn, err := r.ReadByte()
	if err != nil {
		return nil, err
	}
	n := uint8(bn)
	bytes, err := r.ReadBytes(0)
	if err != nil {
		return nil, err
	}
	// ignore bytes bigger than n (asssume it is padding)
	out := string(bytes[:n])
	return &out, nil
}

func ReadLenAndColor(r *bufio.Reader) (color.Color, error) {
	colorLength, err := r.ReadByte()
	if err != nil {
		return nil, err
	}
	n := uint8(colorLength) // will be always 4: R,G,B,A
	if n == 4 {
		R, err := r.ReadByte()
		if err != nil {
			return nil, err
		}
		G, err := r.ReadByte()
		if err != nil {
			return nil, err
		}
		B, err := r.ReadByte()
		if err != nil {
			return nil, err
		}
		A, err := r.ReadByte()
		if err != nil {
			return nil, err
		}
		out := color.NRGBA{R: R, G: G, B: B, A: A}
		return out, nil
	} else if n == 3 {
		// todo might be also HSV?
		R, err := r.ReadByte()
		if err != nil {
			return nil, err
		}
		G, err := r.ReadByte()
		if err != nil {
			return nil, err
		}
		B, err := r.ReadByte()
		if err != nil {
			return nil, err
		}
		out := color.NRGBA{R: R, G: G, B: B, A: 0}
		return out, nil
	} else if n == 2 {
		grayscale, err := r.ReadByte()
		if err != nil {
			return nil, err
		}
		A, err := r.ReadByte()
		if err != nil {
			return nil, err
		}
		out := color.NRGBA{R: grayscale, G: grayscale, B: grayscale, A: A}
		return out, nil
	} else {
		return nil, fmt.Errorf("unusual color length: %d", n)
	}
}
