package main

import (
	"log"
	"regexp"
	"strings"
)

func CMKFindName(oldVariablesNames []string, name string) (string, bool) {
	cleanupName, _ := strings.CutPrefix(name, "_")
	cleanupName = strings.ToLower(cleanupName)
	for index, possibleMatch := range oldVariablesNames {
		cleanupPossibleMatch, _ := strings.CutPrefix(possibleMatch, "_")
		cleanupPossibleMatch = strings.ToLower(cleanupPossibleMatch)
		if cleanupName == cleanupPossibleMatch {
			return oldVariablesNames[index], true
		}
	}
	return name, false
}

func loadValuesFromSection(DAT string) ([]string, map[string]string, map[string][]string) {
	variablesKeys := []string{} // not needed for now
	variablesComments := map[string][]string{}
	lastName := ""
	values := map[string]string{}
	for _, line := range strings.Split(DAT, CMKLineSeparator) {
		lineTrimmed := decodeCMKLine(line)
		isComment := strings.HasPrefix(lineTrimmed, "//")
		if isComment {
			variablesComments[lastName] = append(variablesComments[lastName], lineTrimmed)
		} else if strings.Contains(lineTrimmed, "=") {
			nameValue := strings.SplitN(lineTrimmed, "=", 2)
			lastName = nameValue[0]
			variablesKeys = append(variablesKeys, lastName)
			values[lastName] = nameValue[1]
		} else {
			log.Printf("DEBUG: unknown line when updating macro: '%s'", line)
		}
	}
	return variablesKeys, values, variablesComments
}

var OnlyDigitsRegex = regexp.MustCompile(`\d*`)

type UpdateResult int

const (
	ValueDeleted UpdateResult = iota
	ValueAdded
	ValueSame
	ValueChanged
	ValueChangedConvertedToGlobal
	ValueChangedRemainedToLocal
)

type Change struct {
	oldName  *string
	newName  *string
	oldValue string
	newValue string
	result   UpdateResult
}

/*
	Update old makro in smart way. Modifies newMacro in place

- do not touch old JOINT section
- do not touch old VARIJABLE, unless
-- there is new variable
-- maybe in future suport deleting unused variable
- discard other old sections (groupa, potrosni, makro, pila)
- todo handle case insensitive and global names: _VAR==VAR==var==vAr
*/
func UpdateMakro(oldMacro *M1, newMacro *M1, alwaysConvertLocalToGlobal bool) []Change {
	oldVariablesKeys, oldValues, _ := loadValuesFromSection(oldMacro.Varijable.DAT)
	newVariablesKeys, newValues, newVariablesComments := loadValuesFromSection(newMacro.Varijable.DAT)

	// combine old and new in "smart way"
	var outputVarijable strings.Builder
	// write initial comment
	for _, line := range newVariablesComments[M1InitialMacroMarker] {
		outputVarijable.WriteString(encodeCMKLine(line))
	}
	delete(newVariablesComments, M1InitialMacroMarker) // not really necessary

	updateResultVarijable := []Change{}
	// old=1 // deleted
	// old=2 // new value, converted to global
	log.Println("Updating [VARIJABLE]")
	for _, newName := range newVariablesKeys {
		oldName, _ := CMKFindName(oldVariablesKeys, newName)
		newValue, _ := newValues[newName]
		// setting: convert local values to global: always; if value stays the same; convert to evar expression; Keep as is
		// todo if oldValue is not integer, do not allow to convert it to global value
		// one=4
		// one=evar.one+20
		// _one=4//4 is ignored
		oldValue, oldValueExists := oldValues[oldName]
		if oldValueExists {
			name := newName
			convertedToGlobal := alwaysConvertLocalToGlobal
			if !strings.HasPrefix(oldName, `_`) && strings.HasPrefix(newName, `_`) {
				if !alwaysConvertLocalToGlobal {
					if !OnlyDigitsRegex.Match([]byte(oldValue)) {
						// old expression does not contain only digits, it is not allowed to be made global
						name, _ = strings.CutPrefix(newName, `_`)
						convertedToGlobal = true
					}
				}
				if convertedToGlobal {
					updateResultVarijable = append(updateResultVarijable, Change{&oldName, &newName, oldValue, newValue, ValueChangedConvertedToGlobal})
					log.Printf("  Copied old value: '%s=%s' (was local, now is global)\n", name, oldValue)
				} else {
					updateResultVarijable = append(updateResultVarijable, Change{&oldName, &newName, oldValue, newValue, ValueChangedRemainedToLocal})
					log.Printf("  Copied old value: '%s=%s' (variable remained local, new global name was: %s)\n", name, oldValue, newName)
				}
			} else if newValues[newName] != oldValue {
				updateResultVarijable = append(updateResultVarijable, Change{&oldName, &newName, oldValue, newValue, ValueChanged})
				log.Printf("  Copied old value: '%s=%s'\n", name, oldValue)
			} else {
				updateResultVarijable = append(updateResultVarijable, Change{&oldName, &newName, oldValue, newValue, ValueSame})
			}
			outputVarijable.WriteString(encodeCMKLine(name + "=" + oldValue))
		} else {
			log.Printf("  Added value: '%s=%s'\n", newName, newValue)
			updateResultVarijable = append(updateResultVarijable, Change{&oldName, &newName, oldValue, newValue, ValueAdded})
			outputVarijable.WriteString(encodeCMKLine(newName + "=" + newValue))
		}
		for _, comment := range newVariablesComments[newName] {
			outputVarijable.WriteString(encodeCMKLine(comment))
		}
	}

	// append missing new variables
	// for name, newValue := range newValues {
	// 	outputVarijable.WriteString(encodeCMKLine(name + "=" + newValue))
	// 	for _, comment := range newVariablesComments[name] {
	// 		outputVarijable.WriteString(encodeCMKLine(comment))
	// 	}
	// }

	// old variables that no longer exist are discarded
	for _, oldName := range oldVariablesKeys {
		newName, _ := CMKFindName(newVariablesKeys, oldName)
		newValue, _ := newValues[newName]
		oldValue, newValueExists := newValues[newName]
		if !newValueExists {
			log.Printf("  Deleted value: '%s=%s (was )'\n", oldName, oldValue)
			updateResultVarijable = append(updateResultVarijable, Change{&oldName, &newName, oldValue, newValue, ValueDeleted})
		}
	}

	newMacro.Varijable.DAT, _ = strings.CutSuffix(outputVarijable.String(), CMKLineSeparator)
	newMacro.Joint = oldMacro.Joint
	return updateResultVarijable
}
