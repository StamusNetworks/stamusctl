package handlers

import (
	"testing"

	"stamus-ctl/internal/app"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNixBuildVMHandler_NotNixOS(t *testing.T) {
	oldFS := app.FS
	app.FS = afero.NewMemMapFs()
	defer func() { app.FS = oldFS }()

	oldName := app.Name
	app.Name = app.CtlName
	defer func() { app.Name = oldName }()

	err := NixBuildVMHandler(NixBuildVMHandlerInputs{Config: "/tmp/test"})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not running NixOS")
}

func TestNixBuildVMHandler_ConfigNotFound(t *testing.T) {
	oldFS := app.FS
	app.FS = afero.NewMemMapFs()
	defer func() { app.FS = oldFS }()

	oldName := app.Name
	app.Name = app.CtlName
	defer func() { app.Name = oldName }()

	require.NoError(t, app.FS.MkdirAll("/etc", 0o755))
	require.NoError(t, afero.WriteFile(app.FS, "/etc/NIXOS", []byte(""), 0o644))

	err := NixBuildVMHandler(NixBuildVMHandlerInputs{Config: "/nonexistent"})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestNixBuildVMHandler_NixOS_ConfigExists_BuildFails(t *testing.T) {
	oldFS := app.FS
	app.FS = afero.NewMemMapFs()
	defer func() { app.FS = oldFS }()

	require.NoError(t, app.FS.MkdirAll("/etc", 0o755))
	require.NoError(t, afero.WriteFile(app.FS, "/etc/NIXOS", []byte(""), 0o644))

	configPath := "/etc/nixos-buildvm-test"
	require.NoError(t, app.FS.MkdirAll(configPath, 0o755))

	oldName := app.Name
	app.Name = app.CtlName
	defer func() { app.Name = oldName }()

	// nixos-rebuild is not in PATH → NixosRebuild("build-vm") fails
	err := NixBuildVMHandler(NixBuildVMHandlerInputs{Config: configPath})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "nixos-rebuild")
}

func TestNixBuildVMHandler_RunFlagFalse_SkipsVM(t *testing.T) {
	oldFS := app.FS
	app.FS = afero.NewMemMapFs()
	defer func() { app.FS = oldFS }()

	require.NoError(t, app.FS.MkdirAll("/etc", 0o755))
	require.NoError(t, afero.WriteFile(app.FS, "/etc/NIXOS", []byte(""), 0o644))

	configPath := "/etc/nixos-buildvm-norun"
	require.NoError(t, app.FS.MkdirAll(configPath, 0o755))

	oldName := app.Name
	app.Name = app.CtlName
	defer func() { app.Name = oldName }()

	// With Run=false, even if the build were to succeed,
	// FindAndRunVM should not be called. Since nixos-rebuild
	// isn't in PATH, it fails at the build step.
	err := NixBuildVMHandler(NixBuildVMHandlerInputs{
		Config: configPath,
		Run:    false,
	})
	require.Error(t, err)
	// Error is from build-vm, not from FindAndRunVM
	assert.Contains(t, err.Error(), "nixos-rebuild")
}

func TestNixBuildVMHandler_DaemonMode_ConfigNotFound(t *testing.T) {
	oldFS := app.FS
	app.FS = afero.NewMemMapFs()
	defer func() { app.FS = oldFS }()

	require.NoError(t, app.FS.MkdirAll("/etc", 0o755))
	require.NoError(t, afero.WriteFile(app.FS, "/etc/NIXOS", []byte(""), 0o644))

	oldName := app.Name
	app.Name = "stamusd"
	defer func() { app.Name = oldName }()

	err := NixBuildVMHandler(NixBuildVMHandlerInputs{Config: "nonexistent-daemon-conf"})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}
