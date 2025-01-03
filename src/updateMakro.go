package main

import (
	"log"
	"strings"
)

func CMKFindOldName(oldVariablesNames []string, name string) (string, bool) {
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

/*
	Update old makro in smart way. Modifies newMacro in place

- do not touch old JOINT section
- do not touch old VARIJABLE, unless
-- there is new variable
-- maybe in future suport deleting unused variable
- discard other old sections (groupa, potrosni, makro, pila)
- todo handle case insensitive and global names: _VAR==VAR==var==vAr
*/
func UpdateMakro(oldMacro *M1, newMacro *M1) {
	// load everything related to old
	oldVariablesKeys := []string{} // not needed for now
	lastName := ""
	oldValues := map[string]string{}
	for _, line := range strings.Split(oldMacro.Varijable.DAT, CMKLineSeparator) {
		lineTrimmed := decodeCMKLine(line)
		isComment := strings.HasPrefix(lineTrimmed, "//")
		if isComment {
			// discard old comments
			continue
		} else if strings.Contains(lineTrimmed, "=") {
			nameValue := strings.SplitN(lineTrimmed, "=", 2)
			lastName = nameValue[0]
			oldVariablesKeys = append(oldVariablesKeys, lastName)
			oldValues[lastName] = nameValue[1]
		} else {
			log.Printf("DEBUG: unknown line when updating macro: '%s'", line)
		}
	}

	// load everything related to new
	newVariablesKeys := []string{}
	newValues := map[string]string{}
	newVariablesComments := map[string][]string{}
	lastName = ""
	for i, line := range strings.Split(newMacro.Varijable.DAT, CMKLineSeparator) {
		lineTrimmed := decodeCMKLine(line)
		isComment := strings.HasPrefix(lineTrimmed, "//")
		if isComment {
			newVariablesComments[lastName] = append(newVariablesComments[lastName], lineTrimmed)
		} else if strings.Contains(lineTrimmed, "=") {
			nameValue := strings.SplitN(lineTrimmed, "=", 2)
			lastName = nameValue[0]
			newVariablesKeys = append(newVariablesKeys, nameValue[0])
			newValues[lastName] = nameValue[1]
		} else {
			log.Printf("DEBUG: unknown line: %s in index %d, %s", line, i, newMacro.Varijable.DAT)
		}
	}

	// combine old and new in "smart way"
	var outputVarijable strings.Builder
	// write initial comment
	for _, line := range newVariablesComments[M1InitialMacroMarker] {
		outputVarijable.WriteString(encodeCMKLine(line))
	}
	delete(newVariablesComments, M1InitialMacroMarker) // not really necessary
	for _, newName := range newVariablesKeys {
		oldName, _ := CMKFindOldName(oldVariablesKeys, newName)
		// setting: convert local values to global: always; if value stays the same; convert to evar expression; Keep as is
		// todo if oldValue is not integer, do not allow to convert it to global value
		// one=4
		// one=evar.one+20
		// _one=4//4 is ignored
		oldValue, ok := oldValues[oldName]
		if ok {
			outputVarijable.WriteString(encodeCMKLine(newName + "=" + oldValue))
		} else {
			outputVarijable.WriteString(encodeCMKLine(newName + "=" + newValues[newName]))
		}
		delete(newValues, newName)
		for _, comment := range newVariablesComments[newName] {
			outputVarijable.WriteString(encodeCMKLine(comment))
		}
		delete(newVariablesComments, newName)
	}

	// append missing new variables
	for name, newValue := range newValues {
		outputVarijable.WriteString(encodeCMKLine(name + "=" + newValue))
		for _, comment := range newVariablesComments[name] {
			outputVarijable.WriteString(encodeCMKLine(comment))
		}
	}

	// old variables that no longer exist are discarded

	newMacro.Varijable.DAT, _ = strings.CutSuffix(outputVarijable.String(), CMKLineSeparator)
	newMacro.Joint = oldMacro.Joint
}
