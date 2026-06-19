package handlers

import (
	"testing"

	"stamus-ctl/internal/app"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNixDiffHandler_NotNixOS(t *testing.T) {
	oldFS := app.FS
	app.FS = afero.NewMemMapFs()
	defer func() { app.FS = oldFS }()

	oldName := app.Name
	app.Name = app.CtlName
	defer func() { app.Name = oldName }()

	err := NixDiffHandler(NixDiffHandlerInputs{Config: "/tmp/test"})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not running NixOS")
}

func TestNixDiffHandler_ConfigNotFound(t *testing.T) {
	oldFS := app.FS
	app.FS = afero.NewMemMapFs()
	defer func() { app.FS = oldFS }()

	oldName := app.Name
	app.Name = app.CtlName
	defer func() { app.Name = oldName }()

	require.NoError(t, app.FS.MkdirAll("/etc", 0o755))
	require.NoError(t, afero.WriteFile(app.FS, "/etc/NIXOS", []byte(""), 0o644))

	err := NixDiffHandler(NixDiffHandlerInputs{Config: "/nonexistent"})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestNixDiffHandler_NixOS_ConfigExists_BuildFails(t *testing.T) {
	oldFS := app.FS
	app.FS = afero.NewMemMapFs()
	defer func() { app.FS = oldFS }()

	require.NoError(t, app.FS.MkdirAll("/etc", 0o755))
	require.NoError(t, afero.WriteFile(app.FS, "/etc/NIXOS", []byte(""), 0o644))

	configPath := "/etc/nixos-diff-test"
	require.NoError(t, app.FS.MkdirAll(configPath, 0o755))

	oldName := app.Name
	app.Name = app.CtlName
	defer func() { app.Name = oldName }()

	// nixos-rebuild is not in PATH → NixosRebuild("build") fails
	err := NixDiffHandler(NixDiffHandlerInputs{Config: configPath})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "nixos-rebuild")
}

func TestNixDiffHandler_DaemonMode_ConfigNotFound(t *testing.T) {
	oldFS := app.FS
	app.FS = afero.NewMemMapFs()
	defer func() { app.FS = oldFS }()

	require.NoError(t, app.FS.MkdirAll("/etc", 0o755))
	require.NoError(t, afero.WriteFile(app.FS, "/etc/NIXOS", []byte(""), 0o644))

	oldName := app.Name
	app.Name = "stamusd" // not IsCtl → uses GetConfigsFolder
	defer func() { app.Name = oldName }()

	err := NixDiffHandler(NixDiffHandlerInputs{Config: "nonexistent-daemon-conf"})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}
