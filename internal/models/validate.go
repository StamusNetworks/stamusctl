package models

import (
	"regexp"
	"slices"
)

// Check if the path is valid
func ValidatePath(path string) bool {
	re := regexp.MustCompile(`^\./[a-zA-Z0-9_/]+\.[a-zA-Z0-9_/]+$`)
	return re.MatchString(path)
}

func ValidateMemoryUsage(memory Variable) bool {
	// //Exists
	// if memory.String == nil {
	// 	return false
	// }
	// logging.Sugar.Info()
	// possibleUnits := []string{"k", "m", "g", "t", "p"}
	// //Extract
	// memoryUnit := (*memory.String)[len(*memory.String)-1:]
	// memoryValue := (*memory.String)[:len(*memory.String)-1]
	// // Valid unit
	// if !slices.Contains(possibleUnits, memoryUnit) {
	// 	return false
	// }
	// // Valid value
	// if _, err := strconv.Atoi(memoryValue); err != nil {
	// 	return false
	// }
	return true
}

func ValidateRestartMode(restart Variable) bool {
	//Exists
	if restart.String == nil {
		return false
	}
	possibleValues := []string{"no", "always", "on-failure", "unless-stopped"}
	// Valid value
	return slices.Contains(possibleValues, *restart.String)
}

func GetValidateFunc(name string) func(Variable) bool {
	switch name {
	case "memory":
		return ValidateMemoryUsage
	case "restart":
		return ValidateRestartMode
	default:
		return func(Variable) bool {
			return true
		}
	}
}
