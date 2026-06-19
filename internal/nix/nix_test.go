package nix

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	// Internal
	"stamus-ctl/internal/app"

	// External
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

// ---------------------------------------------------------------------------
// DiffClosures
// ---------------------------------------------------------------------------

func TestDiffClosures_Success(t *testing.T) {
	old := execCommand
	execCommand = func(_ string, _ ...string) *exec.Cmd {
		return exec.Command("true")
	}
	defer func() { execCommand = old }()

	err := DiffClosures("/run/current-system", "./result")
	assert.NoError(t, err)
}

func TestDiffClosures_Failure(t *testing.T) {
	old := execCommand
	execCommand = func(_ string, _ ...string) *exec.Cmd {
		return exec.Command("false")
	}
	defer func() { execCommand = old }()

	err := DiffClosures("/run/current-system", "./result")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "nix store diff-closures failed")
}

func TestDiffClosures_PassesCorrectArgs(t *testing.T) {
	old := execCommand
	var capturedName string
	var capturedArgs []string
	execCommand = func(name string, args ...string) *exec.Cmd {
		capturedName = name
		capturedArgs = args
		return exec.Command("true")
	}
	defer func() { execCommand = old }()

	err := DiffClosures("/run/current-system", "/nix/store/abc-nixos")
	assert.NoError(t, err)
	assert.Equal(t, "nix", capturedName)
	assert.Equal(t, []string{"store", "diff-closures", "/run/current-system", "/nix/store/abc-nixos"}, capturedArgs)
}

// ---------------------------------------------------------------------------
// ListGenerations
// ---------------------------------------------------------------------------

func TestListGenerations_Success(t *testing.T) {
	old := execCommand
	execCommand = func(_ string, _ ...string) *exec.Cmd {
		return exec.Command("echo", "1 current")
	}
	defer func() { execCommand = old }()

	out, err := ListGenerations()
	assert.NoError(t, err)
	assert.Contains(t, out, "1 current")
}

func TestListGenerations_Failure(t *testing.T) {
	old := execCommand
	execCommand = func(_ string, _ ...string) *exec.Cmd {
		return exec.Command("false")
	}
	defer func() { execCommand = old }()

	_, err := ListGenerations()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "nixos-rebuild list-generations failed")
}

func TestListGenerations_CapturesMultilineOutput(t *testing.T) {
	old := execCommand
	execCommand = func(_ string, _ ...string) *exec.Cmd {
		return exec.Command("printf", "  1   2025-01-01 (current)\n  2   2025-01-02\n")
	}
	defer func() { execCommand = old }()

	out, err := ListGenerations()
	assert.NoError(t, err)
	assert.Contains(t, out, "2025-01-01")
	assert.Contains(t, out, "2025-01-02")
}

func TestListGenerations_EmptyOutput(t *testing.T) {
	old := execCommand
	execCommand = func(_ string, _ ...string) *exec.Cmd {
		return exec.Command("true") // outputs nothing
	}
	defer func() { execCommand = old }()

	out, err := ListGenerations()
	assert.NoError(t, err)
	assert.Empty(t, out)
}

// ---------------------------------------------------------------------------
// FindAndRunVM
// ---------------------------------------------------------------------------

func TestFindAndRunVM_NoScript(t *testing.T) {
	dir := t.TempDir()
	err := FindAndRunVM(dir)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no run-*-vm script found")
}

func TestFindAndRunVM_NoBinDir(t *testing.T) {
	dir := t.TempDir()
	// bin/ doesn't exist — glob returns no matches
	err := FindAndRunVM(dir)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no run-*-vm script found")
}

func TestFindAndRunVM_FindsScript(t *testing.T) {
	dir := t.TempDir()
	binDir := filepath.Join(dir, "bin")
	require.NoError(t, os.MkdirAll(binDir, 0o755))

	// Create a fake run-nixos-vm script
	scriptPath := filepath.Join(binDir, "run-nixos-vm")
	require.NoError(t, os.WriteFile(scriptPath, []byte("#!/bin/sh\nexit 0\n"), 0o755))

	old := execCommand
	var capturedScript string
	execCommand = func(name string, _ ...string) *exec.Cmd {
		capturedScript = name
		return exec.Command("true")
	}
	defer func() { execCommand = old }()

	err := FindAndRunVM(dir)
	assert.NoError(t, err)
	assert.Equal(t, scriptPath, capturedScript)
}

func TestFindAndRunVM_ScriptFails(t *testing.T) {
	dir := t.TempDir()
	binDir := filepath.Join(dir, "bin")
	require.NoError(t, os.MkdirAll(binDir, 0o755))

	scriptPath := filepath.Join(binDir, "run-test-vm")
	require.NoError(t, os.WriteFile(scriptPath, []byte("#!/bin/sh\nexit 1\n"), 0o755))

	old := execCommand
	execCommand = func(_ string, _ ...string) *exec.Cmd {
		return exec.Command("false")
	}
	defer func() { execCommand = old }()

	err := FindAndRunVM(dir)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "VM script")
}

func TestFindAndRunVM_PicksFirstMatch(t *testing.T) {
	dir := t.TempDir()
	binDir := filepath.Join(dir, "bin")
	require.NoError(t, os.MkdirAll(binDir, 0o755))

	// Create two matching scripts — should pick the first alphabetically
	for _, name := range []string{"run-alpha-vm", "run-beta-vm"} {
		require.NoError(t, os.WriteFile(filepath.Join(binDir, name), []byte("#!/bin/sh\n"), 0o755))
	}

	old := execCommand
	var capturedScript string
	execCommand = func(name string, _ ...string) *exec.Cmd {
		capturedScript = name
		return exec.Command("true")
	}
	defer func() { execCommand = old }()

	err := FindAndRunVM(dir)
	assert.NoError(t, err)
	assert.Equal(t, filepath.Join(binDir, "run-alpha-vm"), capturedScript)
}

func TestInfect(t *testing.T) {
	err := Infect("")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not yet implemented")
}
