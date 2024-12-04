package models

import (
	"fmt"
	"strings"
)

// Return list of config files to include and list of parameters for current config
func (f *Config) extracKeys() ([]string, []string) {
	// Extract parameters list
	parametersMap := map[string]bool{}
	for _, key := range f.viperInstance.AllKeys() {
		// Extract the parameter name
		parameterAsArray := strings.Split(key, ".")
		parameter := strings.Join(parameterAsArray[:len(parameterAsArray)-1], ".")
		if len(parameter) != 0 {
			parametersMap[parameter] = true
		}
	}
	// Convert map to list
	parametersList := []string{}
	for key := range parametersMap {
		parametersList = append(parametersList, key)
	}
	return f.viperInstance.GetStringSlice("includes"), parametersList
}

// Extract values from the config file
func (f *Config) ExtractValues() map[string]*Variable {
	// Extract parameters list
	parametersList := f.viperInstance.AllKeys()
	// Extract values
	var paramMap = make(map[string]*Variable)
	for _, parameter := range parametersList {
		str := f.viperInstance.GetString(parameter)
		boolean := f.viperInstance.GetBool(parameter)
		integer := f.viperInstance.GetInt(parameter)
		paramMap[parameter] = &Variable{
			String: &str,
			Bool:   &boolean,
			Int:    &integer,
		}
	}
	return paramMap
}

func (f *Config) extracParamOverview(paramName string) Parameter {
	// Extract parameter
	param := Parameter{
		Name:         f.getStringParamValue(paramName, "name"),
		Shorthand:    f.getStringParamValue(paramName, "shorthand"),
		Type:         f.getStringParamValue(paramName, "type"),
		Usage:        f.getStringParamValue(paramName, "usage"),
		ValidateFunc: GetValidateFunc(f.getStringParamValue(paramName, "validate")),
	}
	// Extract default
	switch param.Type {
	case "string":
		param.Default = CreateVariableString(f.getStringParamValue(paramName, "default"))
	case "bool", "optional":
		param.Default = CreateVariableBool(f.getBoolParamValue(paramName, "default"))
	case "int":
		param.Default = CreateVariableInt(f.getIntParamValue(paramName, "default"))
	}
	return param
}

func (f *Config) getStringParamValue(name string, param string) string {
	return f.viperInstance.GetString(fmt.Sprintf("%s.%s", name, param))
}
func (f *Config) getBoolParamValue(name string, param string) bool {
	return f.viperInstance.GetBool(fmt.Sprintf("%s.%s", name, param))
}
func (f *Config) getIntParamValue(name string, param string) int {
	return f.viperInstance.GetInt(fmt.Sprintf("%s.%s", name, param))
}

func (f *Config) GetParams() *Parameters {
	return f.parameters
}

func (f *Config) GetArbitrary() *Arbitrary {
	return f.arbitrary
}

func (f *Config) MergeArbitrary(arbitrary map[string]any) {
	for key, value := range arbitrary {
		f.arbitrary.Set(key, value)
	}
}

func (f *Config) SetProject(project string) {
	f.project = project
}
