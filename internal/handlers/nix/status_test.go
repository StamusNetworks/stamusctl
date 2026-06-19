package handlers

import (
	"testing"

	"stamus-ctl/internal/app"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNixStatusHandler_NotNixOS_ConfigNotFound(t *testing.T) {
	oldFS := app.FS
	app.FS = afero.NewMemMapFs()
	defer func() { app.FS = oldFS }()

	oldName := app.Name
	app.Name = app.CtlName
	defer func() { app.Name = oldName }()

	// Status should not error even when nothing is found
	err := NixStatusHandler(NixStatusHandlerInputs{Config: "/nonexistent"})
	assert.NoError(t, err)
}

func TestNixStatusHandler_ConfigExists(t *testing.T) {
	oldFS := app.FS
	app.FS = afero.NewMemMapFs()
	defer func() { app.FS = oldFS }()

	oldName := app.Name
	app.Name = app.CtlName
	defer func() { app.Name = oldName }()

	configPath := "/tmp/test-config"
	require.NoError(t, app.FS.MkdirAll(configPath, 0o755))
	require.NoError(t, afero.WriteFile(app.FS, configPath+"/version", []byte("1.0.0\n"), 0o644))

	err := NixStatusHandler(NixStatusHandlerInputs{Config: configPath})
	assert.NoError(t, err)
}

func TestNixStatusHandler_ConfigExistsNoVersion(t *testing.T) {
	oldFS := app.FS
	app.FS = afero.NewMemMapFs()
	defer func() { app.FS = oldFS }()

	oldName := app.Name
	app.Name = app.CtlName
	defer func() { app.Name = oldName }()

	// Config dir exists but no version file — should still succeed
	configPath := "/tmp/test-config-noversion"
	require.NoError(t, app.FS.MkdirAll(configPath, 0o755))

	err := NixStatusHandler(NixStatusHandlerInputs{Config: configPath})
	assert.NoError(t, err)
}

func TestNixStatusHandler_ConfigExistsEmptyVersion(t *testing.T) {
	oldFS := app.FS
	app.FS = afero.NewMemMapFs()
	defer func() { app.FS = oldFS }()

	oldName := app.Name
	app.Name = app.CtlName
	defer func() { app.Name = oldName }()

	configPath := "/tmp/test-config-emptyver"
	require.NoError(t, app.FS.MkdirAll(configPath, 0o755))
	// Empty version file — should not print version line
	require.NoError(t, afero.WriteFile(app.FS, configPath+"/version", []byte("\n"), 0o644))

	err := NixStatusHandler(NixStatusHandlerInputs{Config: configPath})
	assert.NoError(t, err)
}

func TestNixStatusHandler_DaemonMode_ConfigNotFound(t *testing.T) {
	oldFS := app.FS
	app.FS = afero.NewMemMapFs()
	defer func() { app.FS = oldFS }()

	oldName := app.Name
	app.Name = "stamusd"
	defer func() { app.Name = oldName }()

	// Status should succeed even if config not found in daemon mode
	err := NixStatusHandler(NixStatusHandlerInputs{Config: "nonexistent-daemon-conf"})
	assert.NoError(t, err)
}

func TestNixStatusHandler_NixOS_ConfigNotFound(t *testing.T) {
	oldFS := app.FS
	app.FS = afero.NewMemMapFs()
	defer func() { app.FS = oldFS }()

	// Simulate NixOS
	require.NoError(t, app.FS.MkdirAll("/etc", 0o755))
	require.NoError(t, afero.WriteFile(app.FS, "/etc/NIXOS", []byte(""), 0o644))

	oldName := app.Name
	app.Name = app.CtlName
	defer func() { app.Name = oldName }()

	// On NixOS with no config — still should not error.
	// ListGenerations will fail (no nixos-rebuild in CI) but that's a warning, not an error.
	err := NixStatusHandler(NixStatusHandlerInputs{Config: "/nonexistent"})
	assert.NoError(t, err)
}

func TestNixStatusHandler_NixOS_ConfigExists(t *testing.T) {
	oldFS := app.FS
	app.FS = afero.NewMemMapFs()
	defer func() { app.FS = oldFS }()

	require.NoError(t, app.FS.MkdirAll("/etc", 0o755))
	require.NoError(t, afero.WriteFile(app.FS, "/etc/NIXOS", []byte(""), 0o644))

	oldName := app.Name
	app.Name = app.CtlName
	defer func() { app.Name = oldName }()

	configPath := "/tmp/nixos-status-test"
	require.NoError(t, app.FS.MkdirAll(configPath, 0o755))
	require.NoError(t, afero.WriteFile(app.FS, configPath+"/version", []byte("2.0.0\n"), 0o644))

	// Should not error. ListGenerations will warn (no nixos-rebuild) but not fail.
	err := NixStatusHandler(NixStatusHandlerInputs{Config: configPath})
	assert.NoError(t, err)
}
