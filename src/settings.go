package main

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

type ProgramSettings struct {
	minify                     bool
	alwaysConvertLocalToGlobal bool
	verbose                    bool
	hideElementsWithZeroMacros bool
	compact                    bool
	// convertLocalVariablesToGlobal ConvertLocalVariablesToGlobal
}
