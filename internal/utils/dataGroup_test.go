package utils

import (
	"testing"

	"stamus-ctl/internal/models"

	"github.com/stretchr/testify/assert"
)

func TestGroupValues(t *testing.T) {
	tests := []struct {
		name     string
		params   *models.Parameters
		args     []string
		expected map[string]interface{}
	}{
		{
			name: "simple flat values",
			params: func() *models.Parameters {
				p := models.Parameters{
					"key1": &models.Parameter{
						Name:     "key1",
						Default:  models.CreateVariableString("value1"),
						Variable: models.CreateVariableString("value1"),
					},
					"key2": &models.Parameter{
						Name:     "key2",
						Default:  models.CreateVariableString("value2"),
						Variable: models.CreateVariableString("value2"),
					},
				}
				return &p
			}(),
			args: []string{"key1", "key2"},
			expected: map[string]interface{}{
				"key1": "value1",
				"key2": "value2",
			},
		},
		{
			name: "nested values with dots",
			params: func() *models.Parameters {
				p := models.Parameters{
					"parent.child": &models.Parameter{
						Name:     "parent.child",
						Default:  models.CreateVariableString("nested_value"),
						Variable: models.CreateVariableString("nested_value"),
					},
				}
				return &p
			}(),
			args: []string{"parent.child"},
			expected: map[string]interface{}{
				"parent": map[string]interface{}{
					"child": "nested_value",
				},
			},
		},
		{
			name: "multiple nested levels",
			params: func() *models.Parameters {
				p := models.Parameters{
					"a.b.c": &models.Parameter{
						Name:     "a.b.c",
						Default:  models.CreateVariableString("deep_value"),
						Variable: models.CreateVariableString("deep_value"),
					},
				}
				return &p
			}(),
			args: []string{"a.b.c"},
			expected: map[string]interface{}{
				"a": map[string]interface{}{
					"b": map[string]interface{}{
						"c": "deep_value",
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GroupValues(tt.params, tt.args)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGroupStuff(t *testing.T) {
	tests := []struct {
		name     string
		stuff    map[string]string
		expected map[string]interface{}
	}{
		{
			name:     "empty map",
			stuff:    map[string]string{},
			expected: map[string]interface{}{},
		},
		{
			name: "flat keys",
			stuff: map[string]string{
				"key1": "value1",
				"key2": "value2",
			},
			expected: map[string]interface{}{
				"key1": "value1",
				"key2": "value2",
			},
		},
		{
			name: "nested with slashes",
			stuff: map[string]string{
				"parent/child": "nested_value",
			},
			expected: map[string]interface{}{
				"parent": map[string]interface{}{
					"child": "nested_value",
				},
			},
		},
		{
			name: "multiple nested levels",
			stuff: map[string]string{
				"a/b/c": "deep_value",
			},
			expected: map[string]interface{}{
				"a": map[string]interface{}{
					"b": map[string]interface{}{
						"c": "deep_value",
					},
				},
			},
		},
		{
			name: "mixed nested and flat",
			stuff: map[string]string{
				"flat":       "flat_value",
				"nest/child": "nested_value",
			},
			expected: map[string]interface{}{
				"flat": "flat_value",
				"nest": map[string]interface{}{
					"child": "nested_value",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GroupStuff(tt.stuff)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestAddToGroup(t *testing.T) {
	tests := []struct {
		name     string
		parts    []string
		value    string
		initial  map[string]interface{}
		expected map[string]interface{}
	}{
		{
			name:    "single part",
			parts:   []string{"key"},
			value:   "value",
			initial: make(map[string]interface{}),
			expected: map[string]interface{}{
				"key": "value",
			},
		},
		{
			name:    "two parts",
			parts:   []string{"parent", "child"},
			value:   "value",
			initial: make(map[string]interface{}),
			expected: map[string]interface{}{
				"parent": map[string]interface{}{
					"child": "value",
				},
			},
		},
		{
			name:    "three parts",
			parts:   []string{"a", "b", "c"},
			value:   "value",
			initial: make(map[string]interface{}),
			expected: map[string]interface{}{
				"a": map[string]interface{}{
					"b": map[string]interface{}{
						"c": "value",
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			addToGroup(tt.parts, tt.value, tt.initial)
			assert.Equal(t, tt.expected, tt.initial)
		})
	}
}
