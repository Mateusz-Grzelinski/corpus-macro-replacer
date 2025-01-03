PL/ENG text below

# Batch update makro for CORPUS

This program performs batch update of macro for given models, but respects already existing `[VARIJABLE]` and `[JOINT]` sections. 
This functionality is the same as double tics in macro edit window, but this one actually works.

Input:
- makro to update, for example: `Nawierty_uniwersalne_28mm.CMK`
- input dir: `.\elmsav\00_STOLARZ_BAZA_2022` (recursive)
- output dir: `.\elmsav\00_STOLARZ_BAZA_2022-changed`

Output:
- updates all `.E3D` files that references `Nawierty_uniwersalne_28mm` and saves it in output dir

## Features

Update old makro in smart way:
- do not touch old JOINT section
- do not touch old VARIJABLE, unless:
-- there is new variable (append new var to the end)
-- reorder variable names the same as new makro
-- use comments from updated makro
-- not supported: deleting unused variable
- discard other old sections (groupa, potrosni, makro, pila)
- todo handle case insensitive and global names: _VAR==VAR==var==vAr
- replace all other sections

# Corner cases

This program edits Corpus files with external tool. ALWAYS MAKE A BACKUP!

- watch out: 
-- .CMK files are usually encoded with `Windows 1250` 
-- `.E3D` files might be utf-8 with BOM (unclear)
- file name must be the same as makro name (settings from `MakroCollection.dat` are ignored)
- makro file name extension must be `.CMD` (must be capitalized)

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