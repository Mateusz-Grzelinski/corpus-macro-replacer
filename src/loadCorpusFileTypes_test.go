package main

import "testing"

func TestGenericNodeWithC6DatDecodeAndEncode(t *testing.T) {
	gn := GenericNodeWithC6Dat{C6DAT: "eJxz9vfzc3UOsTUy1snNzEvJLC5JzEtOtdU1NNHJTayA800BB+cNZg=="}
	decoded, err := gn.DecodeC6Dat()

	if err != nil {
		t.Errorf("Decoding failed: %s", err)
		t.FailNow()
	}

	if decoded != "CONNECT=23,mindistance=-14,maxdistance=5" {
		t.Errorf("Decoding failed, got: %s", decoded)
		t.FailNow()
	}

	encoded, err := EncodeC6Dat(decoded)
	if err != nil {
		t.Errorf("Encoding failed, got: %s", decoded)
		t.FailNow()
	}
	if *encoded != gn.C6DAT {
		t.Errorf("Encoding produced different output fomr original: \n%s - expected, got: \n%s", gn.C6DAT, decoded)
		t.FailNow()
	}
}
