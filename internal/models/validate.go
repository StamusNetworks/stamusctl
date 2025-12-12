package models

import (
	"regexp"
	"slices"

	"stamus-ctl/internal/logging"
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
	// Exists
	if restart.String == nil {
		return false
	}
	possibleValues := []string{"no", "always", "on-failure", "unless-stopped"}
	// Valid value
	return slices.Contains(possibleValues, *restart.String)
}

func GetValidateFunc(name string) func(Variable) bool {
	switch name {
	case "":
		// Empty string means no validation specified - allow all
		return func(Variable) bool {
			return true
		}
	case "memory":
		return ValidateMemoryUsage
	case "restart":
		return ValidateRestartMode
	default:
		// Unknown validator type - log warning and reject
		logging.Sugar.Warnf("Unknown validator type '%s' - validation will fail. Valid validators: memory, restart", name)
		return func(v Variable) bool {
			// For safety, unknown validators should fail validation
			// This prevents typos in validator names from silently accepting invalid input
			logging.Sugar.Warnf("Validation failed due to unknown validator type: %s", name)
			return false
		}
	}
}
