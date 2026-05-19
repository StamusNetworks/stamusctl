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

func TestGetExecVersion_Git(t *testing.T) {
	// git version outputs "git version X.Y.Z" — the last word is semver-parseable.
	version, err := GetExecVersion("git")
	if err != nil {
		// git may not be installed in all CI environments; skip gracefully.
		t.Skipf("skipping: git not available: %v", err)
	}
	assert.NotNil(t, version)
	assert.Positive(t, version.Major()+version.Minor())
}

// TestGetExecVersion_InvalidSemverOutput exercises the semver parse-error path (line 37).
// "true" accepts any arguments and exits 0 with no output. GetExecVersion prepends
// "version" to flags, so the invocation is: true version → exits 0, stdout = "".
// strings.Split("", " ") = [""], last element after Trim is "", which is not a valid
// semver string, so semver.NewVersion returns an error, covering line 37-39.
func TestGetExecVersion_InvalidSemverOutput(t *testing.T) {
	version, err := GetExecVersion("true")
	assert.Error(t, err)
	assert.Nil(t, version)
	assert.Contains(t, err.Error(), "cannot parse true version")
}
