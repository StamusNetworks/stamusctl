package nix

import (
	"os/exec"
	"testing"

	// Internal
	"stamus-ctl/internal/app"

	// External
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
)

func TestIsNixOS_True(t *testing.T) {
	oldFS := app.FS
	app.FS = afero.NewMemMapFs()
	defer func() { app.FS = oldFS }()

	_, err := app.FS.Create(nixosMarker)
	assert.NoError(t, err)

	assert.True(t, IsNixOS())
}

func TestIsNixOS_False(t *testing.T) {
	oldFS := app.FS
	app.FS = afero.NewMemMapFs()
	defer func() { app.FS = oldFS }()

	assert.False(t, IsNixOS())
}

func TestNixosRebuild_InvalidAction(t *testing.T) {
	cases := []struct {
		name   string
		action string
	}{
		{"empty string", ""},
		{"unknown word", "reboot"},
		{"mixed case", "Switch"},
		{"partial match", "swit"},
		{"extra space", " switch"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := NixosRebuild("/some/path", tc.action)
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "invalid nixos-rebuild action")
		})
	}
}

func TestNixosRebuild_ValidActions(t *testing.T) {
	validActions := []string{"switch", "boot", "test", "build", "build-vm"}

	for _, action := range validActions {
		action := action
		t.Run(action, func(t *testing.T) {
			old := execCommand
			execCommand = func(name string, args ...string) *exec.Cmd {
				return exec.Command("true")
			}
			defer func() { execCommand = old }()

			err := NixosRebuild("/some/config", action)
			assert.NoError(t, err)
		})
	}
}

func TestNixosRebuild_CommandFailure(t *testing.T) {
	old := execCommand
	execCommand = func(name string, args ...string) *exec.Cmd {
		return exec.Command("false")
	}
	defer func() { execCommand = old }()

	err := NixosRebuild("/some/config", "switch")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "nixos-rebuild switch failed")
}

func TestBuildISO_Success(t *testing.T) {
	old := execCommand
	execCommand = func(name string, args ...string) *exec.Cmd {
		return exec.Command("true")
	}
	defer func() { execCommand = old }()

	err := BuildISO("/some/config", "/some/output")
	assert.NoError(t, err)
}

func TestBuildISO_Failure(t *testing.T) {
	old := execCommand
	execCommand = func(name string, args ...string) *exec.Cmd {
		return exec.Command("false")
	}
	defer func() { execCommand = old }()

	err := BuildISO("/some/config", "/some/output")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "nix-build iso failed")
}

func TestRunShellTest_Success(t *testing.T) {
	old := execCommand
	execCommand = func(name string, args ...string) *exec.Cmd {
		return exec.Command("true")
	}
	defer func() { execCommand = old }()

	err := RunShellTest("/some/test.sh", "/some/config")
	assert.NoError(t, err)
}

func TestRunShellTest_Failure(t *testing.T) {
	old := execCommand
	execCommand = func(name string, args ...string) *exec.Cmd {
		return exec.Command("false")
	}
	defer func() { execCommand = old }()

	err := RunShellTest("/some/test.sh", "/some/config")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "shell test")
}

func TestRunNixTest_Success(t *testing.T) {
	old := execCommand
	execCommand = func(name string, args ...string) *exec.Cmd {
		return exec.Command("true")
	}
	defer func() { execCommand = old }()

	err := RunNixTest("/some/test.nix", "/some/config")
	assert.NoError(t, err)
}

func TestRunNixTest_Failure(t *testing.T) {
	old := execCommand
	execCommand = func(name string, args ...string) *exec.Cmd {
		return exec.Command("false")
	}
	defer func() { execCommand = old }()

	err := RunNixTest("/some/test.nix", "/some/config")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "nix test")
}

func TestInfect(t *testing.T) {
	err := Infect("")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not yet implemented")
}
