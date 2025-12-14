package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestExtractArgs(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		expected map[string]string
	}{
		{
			name:     "empty args",
			args:     []string{},
			expected: map[string]string{},
		},
		{
			name: "single valid arg",
			args: []string{"key=value"},
			expected: map[string]string{
				"key": "value",
			},
		},
		{
			name: "multiple valid args",
			args: []string{"key1=value1", "key2=value2", "key3=value3"},
			expected: map[string]string{
				"key1": "value1",
				"key2": "value2",
				"key3": "value3",
			},
		},
		{
			name: "args with spaces in values",
			args: []string{"key=value with spaces"},
			expected: map[string]string{
				"key": "value with spaces",
			},
		},
		{
			name:     "invalid arg without equals",
			args:     []string{"invalid"},
			expected: map[string]string{},
		},
		{
			name:     "invalid arg with multiple equals",
			args:     []string{"key=value=extra"},
			expected: map[string]string{},
		},
		{
			name: "mixed valid and invalid args",
			args: []string{"valid=value", "invalid", "another=test"},
			expected: map[string]string{
				"valid":   "value",
				"another": "test",
			},
		},
		{
			name: "empty key or value",
			args: []string{"key="},
			expected: map[string]string{
				"key": "",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ExtractArgs(tt.args)
			assert.Equal(t, tt.expected, result)
		})
	}
}
