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

func (p *Parameters) GetParameters(keys ...string) map[string]*Parameter {
	values := make(map[string]*Parameter)
	for key, param := range *p {
		// if keys are provided, only return values for keys that start with the provided keys
		if len(keys) > 0 {
			for _, k := range keys {
				if strings.HasPrefix(key, k) {
					values[key] = param
				}
			}
		} else {
			values[key] = param
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

// Returns only the customized (non-default) variable values
// A value is considered customized if it differs from its default
func (p *Parameters) GetCustomizedValues(keys ...string) map[string]*Variable {
	values := make(map[string]*Variable)
	for key, param := range *p {
		// Skip if keys filter is provided and key doesn't match
		if len(keys) > 0 {
			matched := false
			for _, k := range keys {
				if strings.HasPrefix(key, k) {
					matched = true
					break
				}
			}
			if !matched {
				continue
			}
		}

		// Only include if the variable differs from the default
		// This indicates the user customized this value
		if !param.Variable.IsNil() && !param.Default.IsNil() {
			// Compare variable against default to detect customization
			varStr := param.Variable.AsString()
			defStr := param.Default.AsString()
			if varStr != defStr {
				values[key] = &param.Variable
			}
		} else if !param.Variable.IsNil() && param.Default.IsNil() {
			// If there's a variable but no default, preserve it
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
