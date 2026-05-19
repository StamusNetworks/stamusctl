package handlers

import (
	"testing"

	"stamus-ctl/internal/app"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ---------------------------------------------------------------------------
// NixISOHandler tests
// ---------------------------------------------------------------------------

func TestNixISOHandler_ConfigNotFound(t *testing.T) {
	oldFS := app.FS
	app.FS = afero.NewMemMapFs()
	defer func() { app.FS = oldFS }()

	err := NixISOHandler(NixISOHandlerInputs{
		Config: "nonexistent",
		Output: "/tmp/out",
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestNixISOHandler_InvalidOutput_PathTraversal(t *testing.T) {
	oldFS := app.FS
	app.FS = afero.NewMemMapFs()
	defer func() { app.FS = oldFS }()

	// Create a valid config directory so we pass the existence check
	configPath := app.GetConfigsFolder("testcfg")
	require.NoError(t, app.FS.MkdirAll(configPath, 0o755))

	// Non-CTL mode: GetConfigsFolder is used for the configPath lookup
	oldName := app.Name
	app.Name = "stamusd" // not CTL — triggers GetConfigsFolder usage
	defer func() { app.Name = oldName }()

	err := NixISOHandler(NixISOHandlerInputs{
		Config: "testcfg",
		Output: "../../etc/passwd",
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid output path")
}

func TestNixISOHandler_ConfigExistsButBuildFails(t *testing.T) {
	oldFS := app.FS
	app.FS = afero.NewMemMapFs()
	defer func() { app.FS = oldFS }()

	// Use CTL mode (IsCtl() → true) so configPath = params.Config directly.
	oldName := app.Name
	app.Name = app.CtlName
	defer func() { app.Name = oldName }()

	configPath := "/tmp/testconfig"
	require.NoError(t, app.FS.MkdirAll(configPath, 0o755))

	// nix-build does not exist in CI; we expect an error from BuildISO.
	err := NixISOHandler(NixISOHandlerInputs{
		Config: configPath,
		Output: "/tmp/iso-output",
	})
	// Config is found, output path is valid — error from BuildISO (nix-build not found).
	assert.Error(t, err)
}

// ---------------------------------------------------------------------------
// NixInfectHandler tests
// ---------------------------------------------------------------------------

func TestNixInfectHandler_OnNonNixOS(t *testing.T) {
	oldFS := app.FS
	app.FS = afero.NewMemMapFs()
	defer func() { app.FS = oldFS }()

	// /etc/NIXOS does not exist in the in-memory FS → IsNixOS() = false.
	// nix.Infect always returns "not yet implemented".
	err := NixInfectHandler(NixInfectHandlerInputs{Config: "myconfig"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not yet implemented")
}

func TestNixInfectHandler_OnNixOS(t *testing.T) {
	oldFS := app.FS
	app.FS = afero.NewMemMapFs()
	defer func() { app.FS = oldFS }()

	// Place the NixOS marker so IsNixOS() returns true.
	require.NoError(t, app.FS.MkdirAll("/etc", 0o755))
	require.NoError(t, afero.WriteFile(app.FS, "/etc/NIXOS", []byte(""), 0o644))

	// When already on NixOS, NixInfectHandler returns the "already running NixOS" error.
	err := NixInfectHandler(NixInfectHandlerInputs{Config: "myconfig"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "already running NixOS")
}
