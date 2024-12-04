package utils

import (
	"stamus-ctl/internal/models"
	"strings"
)

// Utility function to group values from the config to nested maps
func GroupValues(config *models.Config, args []string) map[string]interface{} {
	values := config.GetParams().GetValues(args...)
	groupedValues := make(map[string]interface{})
	for _, param := range config.GetParams().GetOrdered() {
		if value, ok := values[param]; ok {
			parts := strings.Split(param, ".")
			addToGroup(parts, value, groupedValues)
		}
	}
	return groupedValues
}

func GroupStuff(stuff map[string]string) map[string]interface{} {
	groupedStuff := make(map[string]interface{})
	for key, value := range stuff {
		parts := strings.Split(key, "/")
		addToGroup(parts, value, groupedStuff)
	}
	return groupedStuff
}

func addToGroup(parts []string, value string, group map[string]interface{}) {
	if len(parts) == 1 {
		group[parts[0]] = value
	} else {
		if _, ok := group[parts[0]]; !ok {
			group[parts[0]] = make(map[string]interface{})
		}
		addToGroup(parts[1:], value, group[parts[0]].(map[string]interface{}))
	}
}
