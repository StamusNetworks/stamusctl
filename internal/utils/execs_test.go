package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetExecVersion(t *testing.T) {
	tests := []struct {
		name        string
		executable  string
		flags       []string
		expectError bool
	}{
		{
			name:        "non-existing executable",
			executable:  "non-existing-command-12345",
			flags:       []string{},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			version, err := GetExecVersion(tt.executable, tt.flags...)
			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, version)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, version)
			}
		})
	}
}
