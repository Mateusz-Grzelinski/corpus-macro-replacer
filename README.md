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

- watch out: 
-- .CMK files are usually encoded with `Windows 1250` 
-- `.E3D` files might be utf-8 with BOM (unclear)
- file name must be the same as makro name (settings from `MakroCollection.dat` are ignored)
- makro file name extension must be `.CMD` (must be capitalized)

# Testing

Tested manually using Corpus 5.0