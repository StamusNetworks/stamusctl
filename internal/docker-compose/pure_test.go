package compose

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// ---------------------------------------------------------------------------
// shouldCreateBackup
// ---------------------------------------------------------------------------

func TestShouldCreateBackup(t *testing.T) {
	tests := []struct {
		name        string
		cmdName     string
		volumeValue string
		want        bool
	}{
		{
			name:        "down with volumes true creates backup",
			cmdName:     "down",
			volumeValue: "true",
			want:        true,
		},
		{
			name:        "down with volumes false does not create backup",
			cmdName:     "down",
			volumeValue: "false",
			want:        false,
		},
		{
			name:        "down with empty volumes does not create backup",
			cmdName:     "down",
			volumeValue: "",
			want:        false,
		},
		{
			name:        "up with volumes true does not create backup",
			cmdName:     "up",
			volumeValue: "true",
			want:        false,
		},
		{
			name:        "restart with volumes true does not create backup",
			cmdName:     "restart",
			volumeValue: "true",
			want:        false,
		},
		{
			name:        "pull with volumes true does not create backup",
			cmdName:     "pull",
			volumeValue: "true",
			want:        false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := shouldCreateBackup(tc.cmdName, tc.volumeValue)
			assert.Equal(t, tc.want, got)
		})
	}
}

// ---------------------------------------------------------------------------
// ComposeFlags.Contains
// ---------------------------------------------------------------------------

func TestComposeFlags_Contains(t *testing.T) {
	tests := []struct {
		cmd  string
		want bool
	}{
		{"up", true},
		{"down", true},
		{"restart", true},
		{"exec", true},
		{"ps", true},
		{"logs", true},
		{"pull", true},
		{"images", true},
		{"nonexistent", false},
		{"", false},
		{"UP", false}, // case-sensitive
	}

	for _, tc := range tests {
		t.Run(tc.cmd, func(t *testing.T) {
			got := ComposeFlags.Contains(tc.cmd)
			assert.Equal(t, tc.want, got)
		})
	}
}

// ---------------------------------------------------------------------------
// isValideRestartModeDocker
// ---------------------------------------------------------------------------

func TestIsValideRestartModeDocker(t *testing.T) {
	tests := []struct {
		mode string
		want bool
	}{
		{"no", true},
		{"always", true},
		{"on-failure", true},
		{"unless-stopped", true},
		{"invalid", false},
		{"", false},
		{"Never", false},
		{"ALWAYS", false},
	}

	for _, tc := range tests {
		t.Run(tc.mode, func(t *testing.T) {
			got := isValideRestartModeDocker(tc.mode)
			assert.Equal(t, tc.want, got)
		})
	}
}

// ---------------------------------------------------------------------------
// isValideRamSizeDocker
//
// The source blacklist is: "", "0", "0m", "0g\t" (note the trailing tab!),
// "0k", "0t", "0p".  The string "0g" (without a tab) is NOT in the blacklist
// and therefore returns true (considered a valid RAM size).
// ---------------------------------------------------------------------------

func TestIsValideRamSizeDocker(t *testing.T) {
	tests := []struct {
		ram  string
		want bool
	}{
		{"256m", true},
		{"1g", true},
		{"512k", true},
		{"2t", true},
		{"100p", true},
		// "0g" (no trailing tab) is NOT blacklisted → valid
		{"0g", true},
		// the exact blacklisted string (with trailing tab) is invalid
		{"0g\t", false},
		// explicit blacklist entries
		{"", false},
		{"0", false},
		{"0m", false},
		{"0k", false},
		{"0t", false},
		{"0p", false},
	}

	for _, tc := range tests {
		name := tc.ram
		if name == "" {
			name = "<empty>"
		}
		t.Run(name, func(t *testing.T) {
			got := isValideRamSizeDocker(tc.ram)
			assert.Equal(t, tc.want, got)
		})
	}
}
