package models

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidatePath(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		expected bool
	}{
		{
			name:     "Valid path with file",
			path:     "./config/file.yaml",
			expected: true,
		},
		{
			name:     "Valid path with nested directories",
			path:     "./some/nested/path/file.txt",
			expected: true,
		},
		{
			name:     "Valid path with underscores",
			path:     "./my_config/my_file.yaml",
			expected: true,
		},
		{
			name:     "Invalid path without leading ./",
			path:     "config/file.yaml",
			expected: false,
		},
		{
			name:     "Invalid path without extension",
			path:     "./config/file",
			expected: false,
		},
		{
			name:     "Invalid path with only directory",
			path:     "./config/",
			expected: false,
		},
		{
			name:     "Invalid absolute path",
			path:     "/absolute/path/file.txt",
			expected: false,
		},
		{
			name:     "Invalid path with spaces",
			path:     "./config/my file.txt",
			expected: false,
		},
		{
			name:     "Empty path",
			path:     "",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ValidatePath(tt.path)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestValidateMemoryUsage(t *testing.T) {
	tests := []struct {
		name     string
		memory   Variable
		expected bool
	}{
		{
			name:     "Any memory value returns true (commented out logic)",
			memory:   CreateVariableString("1024m"),
			expected: true,
		},
		{
			name:     "Empty memory value returns true",
			memory:   CreateVariableString(""),
			expected: true,
		},
		{
			name:     "Nil string returns true",
			memory:   Variable{},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ValidateMemoryUsage(tt.memory)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestValidateRestartMode(t *testing.T) {
	tests := []struct {
		name     string
		restart  Variable
		expected bool
	}{
		{
			name:     "Valid restart mode: no",
			restart:  CreateVariableString("no"),
			expected: true,
		},
		{
			name:     "Valid restart mode: always",
			restart:  CreateVariableString("always"),
			expected: true,
		},
		{
			name:     "Valid restart mode: on-failure",
			restart:  CreateVariableString("on-failure"),
			expected: true,
		},
		{
			name:     "Valid restart mode: unless-stopped",
			restart:  CreateVariableString("unless-stopped"),
			expected: true,
		},
		{
			name:     "Invalid restart mode",
			restart:  CreateVariableString("invalid-mode"),
			expected: false,
		},
		{
			name:     "Empty restart mode",
			restart:  CreateVariableString(""),
			expected: false,
		},
		{
			name:     "Nil string pointer",
			restart:  Variable{},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ValidateRestartMode(tt.restart)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGetValidateFunc(t *testing.T) {
	tests := []struct {
		name         string
		validatorKey string
		testValue    Variable
		expected     bool
	}{
		{
			name:         "Empty validator key always returns true",
			validatorKey: "",
			testValue:    CreateVariableString("anything"),
			expected:     true,
		},
		{
			name:         "Memory validator returns true (current implementation)",
			validatorKey: "memory",
			testValue:    CreateVariableString("1024m"),
			expected:     true,
		},
		{
			name:         "Restart validator with valid value",
			validatorKey: "restart",
			testValue:    CreateVariableString("always"),
			expected:     true,
		},
		{
			name:         "Restart validator with invalid value",
			validatorKey: "restart",
			testValue:    CreateVariableString("invalid"),
			expected:     false,
		},
		{
			name:         "Unknown validator returns false",
			validatorKey: "unknown-validator",
			testValue:    CreateVariableString("value"),
			expected:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			validateFunc := GetValidateFunc(tt.validatorKey)
			assert.NotNil(t, validateFunc)
			result := validateFunc(tt.testValue)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGetValidateFunc_FunctionBehavior(t *testing.T) {
	t.Run("Empty validator always returns true", func(t *testing.T) {
		fn := GetValidateFunc("")
		assert.True(t, fn(CreateVariableString("test")))
		assert.True(t, fn(CreateVariableInt(123)))
		assert.True(t, fn(Variable{}))
	})

	t.Run("Memory validator always returns true", func(t *testing.T) {
		fn := GetValidateFunc("memory")
		assert.True(t, fn(CreateVariableString("1024m")))
		assert.True(t, fn(CreateVariableString("invalid")))
		assert.True(t, fn(Variable{}))
	})

	t.Run("Restart validator validates correctly", func(t *testing.T) {
		fn := GetValidateFunc("restart")
		assert.True(t, fn(CreateVariableString("no")))
		assert.True(t, fn(CreateVariableString("always")))
		assert.True(t, fn(CreateVariableString("on-failure")))
		assert.True(t, fn(CreateVariableString("unless-stopped")))
		assert.False(t, fn(CreateVariableString("invalid")))
		assert.False(t, fn(Variable{}))
	})

	t.Run("Unknown validator always returns false", func(t *testing.T) {
		fn := GetValidateFunc("nonexistent")
		assert.False(t, fn(CreateVariableString("test")))
		assert.False(t, fn(CreateVariableInt(123)))
		assert.False(t, fn(Variable{}))
	})
}
