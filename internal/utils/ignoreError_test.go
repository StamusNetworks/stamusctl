package utils

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIgnoreError(t *testing.T) {
	tests := []struct {
		name     string
		value    interface{}
		err      error
		expected interface{}
	}{
		{
			name:     "no error with string",
			value:    "test-value",
			err:      nil,
			expected: "test-value",
		},
		{
			name:     "with error with string",
			value:    "test-value",
			err:      errors.New("some error"),
			expected: "test-value",
		},
		{
			name:     "no error with int",
			value:    42,
			err:      nil,
			expected: 42,
		},
		{
			name:     "with error with int",
			value:    42,
			err:      errors.New("some error"),
			expected: 42,
		},
		{
			name:     "no error with bool",
			value:    true,
			err:      nil,
			expected: true,
		},
		{
			name:     "with error with bool",
			value:    false,
			err:      errors.New("some error"),
			expected: false,
		},
		{
			name:     "no error with nil",
			value:    nil,
			err:      nil,
			expected: nil,
		},
		{
			name:     "with error with nil",
			value:    nil,
			err:      errors.New("some error"),
			expected: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IgnoreError(tt.value, tt.err)
			assert.Equal(t, tt.expected, result)
		})
	}
}
