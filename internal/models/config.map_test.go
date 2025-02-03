package models

import (
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

type MockFile struct {
	v *viper.Viper
}

func (m *MockFile) GetViper() *viper.Viper {
	return m.v
}

func TestExtracKeys(t *testing.T) {
	tests := []struct {
		name             string
		configData       map[string]interface{}
		expectedIncludes []string
		expectedParams   []string
	}{
		{
			name: "Test with includes and parameters",
			configData: map[string]interface{}{
				"includes":    []string{"file1", "file2"},
				"param1.key1": "value1",
				"param2.key2": "value2",
			},
			expectedIncludes: []string{"file1", "file2"},
			expectedParams:   []string{"param1", "param2"},
		},
		{
			name: "Test with no includes",
			configData: map[string]interface{}{
				"param1.key1": "value1",
				"param2.key2": "value2",
			},
			expectedIncludes: []string{},
			expectedParams:   []string{"param1", "param2"},
		},
		{
			name: "Test with no parameters",
			configData: map[string]interface{}{
				"includes": []string{"file1", "file2"},
			},
			expectedIncludes: []string{"file1", "file2"},
			expectedParams:   []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			file := &File{
				viperInstance: viper.New(),
			}
			for key, value := range tt.configData {
				file.viperInstance.Set(key, value)
			}
			config := NewConfig(file)

			includes, params := config.extracKeys()

			assert.ElementsMatch(t, tt.expectedIncludes, includes)
			assert.ElementsMatch(t, tt.expectedParams, params)
		})
	}
}

func TestExtractValues(t *testing.T) {
	tests := []struct {
		name       string
		configData map[string]interface{}
		expected   map[string]*Variable
	}{
		{
			name: "Test with string, bool, and int values",
			configData: map[string]interface{}{
				"param1": "value1",
				"param2": true,
				"param3": 123,
			},
			expected: map[string]*Variable{
				"param1": {String: stringPtr("value1"), Bool: boolPtr(false), Int: intPtr(0)},
				"param2": {String: stringPtr("true"), Bool: boolPtr(true), Int: intPtr(1)},
				"param3": {String: stringPtr("123"), Bool: boolPtr(true), Int: intPtr(123)},
			},
		},
		{
			name: "Test with only string values",
			configData: map[string]interface{}{
				"param1": "value1",
				"param2": "value2",
			},
			expected: map[string]*Variable{
				"param1": {String: stringPtr("value1"), Bool: boolPtr(false), Int: intPtr(0)},
				"param2": {String: stringPtr("value2"), Bool: boolPtr(false), Int: intPtr(0)},
			},
		},
		{
			name: "Test with only bool values",
			configData: map[string]interface{}{
				"param1": true,
				"param2": false,
			},
			expected: map[string]*Variable{
				"param1": {String: stringPtr("true"), Bool: boolPtr(true), Int: intPtr(1)},
				"param2": {String: stringPtr("false"), Bool: boolPtr(false), Int: intPtr(0)},
			},
		},
		{
			name: "Test with only int values",
			configData: map[string]interface{}{
				"param1": 123,
				"param2": 456,
			},
			expected: map[string]*Variable{
				"param1": {String: stringPtr("123"), Bool: boolPtr(true), Int: intPtr(123)},
				"param2": {String: stringPtr("456"), Bool: boolPtr(true), Int: intPtr(456)},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			file := &File{
				viperInstance: viper.New(),
			}
			for key, value := range tt.configData {
				file.viperInstance.Set(key, value)
			}
			result := NewConfig(file).ExtractValues()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestExtracParamOverview(t *testing.T) {
	tests := []struct {
		name       string
		paramName  string
		configData map[string]interface{}
		expected   Parameter
	}{
		{
			name:      "Test with string parameter",
			paramName: "param1",
			configData: map[string]interface{}{
				"param1.name":      "param1",
				"param1.shorthand": "p1",
				"param1.type":      "string",
				"param1.usage":     "usage1",
				"param1.validate":  "validate1",
				"param1.default":   "default1",
			},
			expected: Parameter{
				Name:         "param1",
				Shorthand:    "p1",
				Type:         "string",
				Usage:        "usage1",
				ValidateFunc: GetValidateFunc("validate1"),
				Default:      CreateVariableString("default1"),
			},
		},
		{
			name:      "Test with bool parameter",
			paramName: "param2",
			configData: map[string]interface{}{
				"param2.name":      "param2",
				"param2.shorthand": "p2",
				"param2.type":      "bool",
				"param2.usage":     "usage2",
				"param2.validate":  "validate2",
				"param2.default":   true,
			},
			expected: Parameter{
				Name:         "param2",
				Shorthand:    "p2",
				Type:         "bool",
				Usage:        "usage2",
				ValidateFunc: GetValidateFunc("validate2"),
				Default:      CreateVariableBool(true),
			},
		},
		{
			name:      "Test with int parameter",
			paramName: "param3",
			configData: map[string]interface{}{
				"param3.name":      "param3",
				"param3.shorthand": "p3",
				"param3.type":      "int",
				"param3.usage":     "usage3",
				"param3.validate":  "validate3",
				"param3.default":   123,
			},
			expected: Parameter{
				Name:         "param3",
				Shorthand:    "p3",
				Type:         "int",
				Usage:        "usage3",
				ValidateFunc: GetValidateFunc("validate3"),
				Default:      CreateVariableInt(123),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			file := &File{
				viperInstance: viper.New(),
			}
			for key, value := range tt.configData {
				file.viperInstance.Set(key, value)
			}
			config := NewConfig(file)

			result := config.extracParamOverview(tt.paramName)

			// assert.Equal(t, tt.expected, result)
			assert.Equal(t, tt.expected.Name, result.Name)
			assert.Equal(t, tt.expected.Shorthand, result.Shorthand)
			assert.Equal(t, tt.expected.Type, result.Type)
			assert.Equal(t, tt.expected.Usage, result.Usage)
			assert.Equal(t, tt.expected.Default, result.Default)
		})
	}
}

func stringPtr(s string) *string {
	return &s
}
