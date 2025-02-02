# Batch update makro for CORPUS

This program performs batch update of macro for given models, but respects already existing `[VARIJABLE]` and `[JOINT]` sections. 
This functionality is the same as double tics in macro edit window, but this one actually works.

Input:
- makro to update, for example: `Nawierty_uniwersalne_28mm.CMK`
- input dir: `.\elmsav\00_STOLARZ_BAZA_2022` (recursive)
- output dir: `.\elmsav\00_STOLARZ_BAZA_2022-changed`

Output:
- updates all `.E3D` files that references `Nawierty_uniwersalne_28mm` and saves it in output dir

There are still bugs in code, expect some crashes.

# Graphical interface

This program is available with UI or via command line. Just run with no arguments to start UI.

Terminal output of GUI is usefull for debugging.

1. Make sure that your Makro search path and MakroCollection.dat is set to your corpus version
2. Open folder that contains Corpus files S3D and E3D are tested
3. Click folder of file and add files to selected (down right menu)
4. Run the program (top right button)

![image](https://github.com/user-attachments/assets/c6382ecf-fa37-4229-acdd-0c62d041304e)

# Command line usage

Run from command line.

Example output (multiple makros are read because the `custom.CMK` includes them in `[MAKRO1]` section):

```powershell
❯ .\Corpus_Macro_Replacer.exe --force --input "C:\Tri D Corpus\Corpus 5.0\elmsav\_modifications\simple_original_custom_with_submacro.E3D" --output "C:\Tri D 2025/01/04 20:50:29 Reading makro: 'C:\Tri D Corpus\Corpus 5.0\Makro\custom.CMK'
2025/01/04 20:50:29 Reading makro: 'C:\Tri D Corpus\Corpus 5.0\Makro\Blenda.CMK'
2025/01/04 20:50:29 Reading makro: 'C:\Tri D Corpus\Corpus 5.0\Makro\Blenda_dodatkowa.CMK'
2025/01/04 20:50:29 Reading Corpus file: 'C:\Users\grzel\go-projects\demo\playground\hello_xml\lewy_gorny.E3D.xml'
2025/01/04 20:50:29   Cabinet 'simple_original_custom'
2025/01/04 20:50:29     Updated 0 macros, 1 skipped in plate 'Wieniec_Gorny'
2025/01/04 20:50:29     Updated 0 macros, 1 skipped in plate 'Bok_Lewy'
2025/01/04 20:50:29   Summary: updated 0 macros, 2 skipped
2025/01/04 20:50:29 Done writing file  : 'C:\Tri D Corpus\Corpus 5.0\elmsav\_modifications\lewy_gorny.E3D.xml'
```

See all options:

```powershell
❯ .\Corpus_Macro_Replacer.exe --help
This program is used to update makro in Copus (.E3D) files. 
It is alternative to doule ticks in macro editor that actually works: 
- it does not edit [JOINT] section
- does a smart merge on [VARIJABLE] section, see README: https://github.com/Mateusz-Grzelinski/corpus-macro-replacer

Usage of Corpus_Macro_Replacer.exe -input <PATH> -output <PATH> -makro <PATH>:
  -alwaysConvertLocalToGlobal
    	default: false. Global variable start with "_" prefix - it takes value from "evar". 
    	Default logic allows adding "_" prefix to variables that consists only from integers (no if statements, no +-* operations). It prevents from erasing your custom logic.
  -force
    	default: false. Specify to override file specified in -output
  -input string
    	required. File or dir, must exist. If dir then changes macro recursively for all .E3D files.
  -makro value
    	required. Path to macro that should be replaced. Can be specified multiple times. Usually one of files in "C:\Tri D Corpus\Corpus 5.0\Makro"
  -minify
    	default: false. Reduce file size by deleting spaces, (~7% size reduction)
  -output string
    	required. File or dir, does not need to exist. 
    	If input is dir then output must be dir, but will be created if does not exist. Directory structure of input is mirrored.
    	If input is file the output can be file (must end with .E3D) or directory.
  -v	print version
```

To save output to file (for later inspection) use this syntax:

```powershell
.\Corpus_Macro_Replacer.exe --<options ...> | tee "log.txt"
```

# Install

Download from releases page https://github.com/Mateusz-Grzelinski/corpus-macro-replacer/releases

## Features

Update specified makro in smart way:

- use old JOINT section
- merge old and new VARIJABLE section:

	- keep old values
	- if there is new variable append it to the end of VARIJABLE
	- discard variables that are no longer present in new makro
	- reorder variable names the same as new makro
	- discard old comments
	- update name if it changes (variable names are not case sensitive)
	- handle the case when `_` is appended to variable name, see option `-alwaysConvertLocalToGlobal`

- discard other old sections (formule, grupa, potrosni, makro, pila), load sections from new version of file
- handle correctly nested macros
- update one file or all files in directory
- does not override files unless `-force` is specified

# Corner cases

This program edits Corpus files with external tool. ALWAYS MAKE A BACKUP!

General points:

- .CMK files are usually encoded with `Windows 1250` 
- `.E3D` files might be utf-8 with BOM (unclear)
- file name must be the same as makro name (settings from `MakroCollection.dat` are ignored)
- makro file extension must be `.CMD` (must be capitalized)

## Converting variable names from local to gloval

Adding `_` to variable name cases it to become global (global variablers are also accessed via `evar`).

Converting local variables to global (evar) variables might be not what you want - it will discard you local changes, see example below example 1:

```
[VARIJABLE] // old
grubosc=18

[VARIJABLE] // new
_grubosc=18

[VARIJABLE] // output
_grubosc=18 // ok, now values is taken from global setting (18 is ignored)
```

Example 2:

```
[VARIJABLE] // old
grubosc=32

[VARIJABLE] // new
_grubosc=18 // note value 18 is ignored, the value is taken from evar.grubosc

[VARIJABLE] // output: default behavior
grubosc=32

[VARIJABLE] // output when using -alwaysConvertLocalToGlobal
_grubosc=32 // old value is preserved but actual value is taken from evar.grubosc

[VARIJABLE] // not impleneted: smart resolution
grubosc=evar.grubosc+12 // ok, now values is taken from global setting but local modification is preserved
```

Example 3:

```
[VARIJABLE] // old
grubosc=obj1.gr

[VARIJABLE] // new
_grubosc=18

[VARIJABLE] // output: default behavior
grubosc=obj1.gr

[VARIJABLE] // output when using -alwaysConvertLocalToGlobal
_grubosc=obj1.gr
```

# Testing

Tested manually using Corpus 5.0.

# License 

Not set yet. If you make modifications you must share them with author.

# Communication

Use github issues: https://github.com/Mateusz-Grzelinski/corpus-macro-replacer/issues

# Roadmap

- [ ] handle corpus files (currently only E3D)
- [ ] handle (recursive) group of object!!!
- [ ] make cabinet (szafka) collapsible
- [ ] allow to manually edit makro
	- make it so that this edit can be applied to all files
- [ ] read makro names from .dat files
- [ ] save log to file
- [ ] add settings like: minify, default makro path
- [ ] add light/dark theme
- [ ] read multiple corpus files at the time
- [ ] indicate errors with color
- [ ] add various filtering options to UI
- [ ] better comparison view for makro
