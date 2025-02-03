package models

import (
	"sort"
	"strings"
)

func (p *Parameters) GetValues(keys ...string) map[string]string {
	values := make(map[string]string)
	for key, param := range *p {
		// if keys are provided, only return values for keys that start with the provided keys
		if len(keys) > 0 {
			for _, k := range keys {
				if strings.HasPrefix(key, k) {
					values[key] = param.Variable.AsString()
				}
			}
		} else {
			values[key] = param.Variable.AsString()
		}
	}
	return values
}

func (p *Parameters) GetVariablesValues(keys ...string) map[string]*Variable {
	values := make(map[string]*Variable)
	for key, param := range *p {
		// if keys are provided, only return values for keys that start with the provided keys
		if len(keys) > 0 {
			for _, k := range keys {
				if strings.HasPrefix(key, k) {
					values[key] = &param.Variable
				}
			}
		} else {
			values[key] = &param.Variable
		}
	}
	return values
}

// Returns an ordered slices of the parameters keys
func (p *Parameters) GetOrdered() []string {
	keys := make([]string, 0, len(*p))
	for key := range *p {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}
