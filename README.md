PL/ENG text below

# Batch update makro for CORPUS

This program performs batch update of macro for given models, but respects already existing `[VARIJABLE]` and `[JOINT]` sections. 
Input:
- makro to update, for example: `Nawierty_uniwersalne_28mm.CMK`
- input dir
- output dir

Output:
- updates all `.E3D` files that references `Nawierty_uniwersalne_28mm.CMK` and saves it in output dir

## Features: 

Update old makro in smart way:
- do not touch old JOINT section
- do not touch old VARIJABLE, unless
-- there is new variable
-- maybe in future suport deleting unused variable
- discard other old sections (groupa, potrosni, makro, pila)
- todo handle case insensitive and global names: _VAR==VAR==var==vAr


# Corner cases

- file name must be the same as makro name (settings from `MakroCollection.dat` are ignored)
- makro file name extension must be `.CMD` (must be capitalized)