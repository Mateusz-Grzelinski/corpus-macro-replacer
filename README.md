# Batch update makro for CORPUS

This program performs batch update of macro for given models, but respects already existing `[VARIJABLE]` and `[JOINT]` sections. 
This functionality is the same as double tics in macro edit window, but this one actually works.

Input:
- makro to update, for example: `Nawierty_uniwersalne_28mm.CMK`
- input dir: `.\elmsav\00_STOLARZ_BAZA_2022` (recursive)
- output dir: `.\elmsav\00_STOLARZ_BAZA_2022-changed`

Output:
- updates all `.E3D` files that references `Nawierty_uniwersalne_28mm` and saves it in output dir

# Usage

Run from command line.

```
This program is used to update makro in Copus (.E3D) files. 
It is alternative to doule ticks in macro editor that actually works: 
- it does not edit [JOINT] section
- does a smart merge on [VARIJABLE] section, see README

Usage of C:\Users\grzel\go-projects\demo\src\__debug_bin3043078883.exe:
  -alwaysConvertLocalToGlobal
    	default: false. Global variable start with "_" prefix - it takes value from "evar". 
    	Default logic allows adding "_" prefix to variables that consists only from integers (no if statements, no +-* operations). It prevents from erasing your custom logic.
  -force
    	default: false. Specify to override file specified in -output
  -input string
    	required. File or dir, must exist. If dir then changes macro recursively for all .E3D files.
  -makro string
    	required. Path to macro that should be replaced. Usually one of files in "C:\Tri D Corpus\Corpus 5.0\Makro"
  -minify
    	default: false. Reduce file size by deleting spaces, (~7% size reduction)
  -output string
    	required. File or dir, does not need to exist. 
    	If input is dir then output must be dir, but will be created if does not exist. Directory structure of input is mirrored.
    	If input is file the output can be file (must end with .E3D) or directory.
  -v	print version
```

Example output (multiple makros are read because the `custom.CMK` includes them in `[MAKRO1]` section):

```bash
❯ .\Corpus_Macro_Replacer.exe --force --input "C:\Tri D Corpus\Corpus 5.0\elmsav\_modifications\simple_original_custom_with_submacro.E3D" --output "C:\Tri D Corpus\Corpus 5.0\elmsav\_modifications\simple_original_custom_with_submacro_output.E3D" --makro "C:\Tri D Corpus\Corpus 5.0\Makro\custom.CMK"
2025/01/03 18:18:19 Reading makro: 'C:\Tri D Corpus\Corpus 5.0\Makro\custom.CMK'
2025/01/03 18:18:19 Reading makro: 'C:\Tri D Corpus\Corpus 5.0\Makro\Blenda.CMK'
2025/01/03 18:18:19 Reading makro: 'C:\Tri D Corpus\Corpus 5.0\Makro\Blenda_dodatkowa.CMK'
2025/01/03 18:18:19 Reading Corpus file: 'C:\Tri D Corpus\Corpus 5.0\elmsav\_modifications\simple_original_custom_with_submacro.E3D'
2025/01/03 18:18:19 Done writing file  : 'C:\Tri D Corpus\Corpus 5.0\elmsav\_modifications\simple_original_custom_with_submacro_output.E3D'
2025/01/03 18:18:19   Updated 1 macros, 0 skipped
```

# Install

Download from releases page https://github.com/Mateusz-Grzelinski/corpus-macro-replacer/releases

## Features

Update old makro in smart way:

- do not touch old JOINT section
- merge old and new VARIJABLE section:
-- there is new variable (append new var to the end)
-- reorder variable names the same as new makro
-- discard variables that are no longer present in new makro
-- discard comments from old makro
- discard other old sections (groupa, potrosni, makro, pila)
- names are not case sensitive: prefer names from new macro
- convert local to global variables, watch for corner cases:
-- todo

# Corner cases

This program edits Corpus files with external tool. ALWAYS MAKE A BACKUP!

General points:

- .CMK files are usually encoded with `Windows 1250` 
- `.E3D` files might be utf-8 with BOM (unclear)
- file name must be the same as makro name (settings from `MakroCollection.dat` are ignored)
- makro file extension must be `.CMD` (must be capitalized)

Converting local variables to global (evar) variables might be not what you want:

```
[VARIJABLE] // old
grubosc=18

[VARIJABLE] // new
_grubosc=18

[VARIJABLE] // output
_grubosc=18 // ok, now values is taken from global setting (18 is ignored)
```

```
[VARIJABLE] // old
grubosc=32 // modified

[VARIJABLE] // new
_grubosc=18

[VARIJABLE] // output 
grubosc=evar.grubosc+12 // ok, now values is taken from global setting but local modification is preserved
```

```
[VARIJABLE] // old
grubosc=obj1.gr

[VARIJABLE] // new
_grubosc=18 // ugh, you just lost old value. the actual value is taken from evar.grubosc

[VARIJABLE] // output
grubosc=evar.grubosc+obj1.gr // you can preserve it like this
```



# Testing

Tested manually using Corpus 5.0