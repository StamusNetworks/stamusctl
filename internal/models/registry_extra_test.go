package models

import (
	"path/filepath"
	"testing"

	"stamus-ctl/internal/app"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ---------------------------------------------------------------------------
// ValidateRegistry
// ---------------------------------------------------------------------------

func TestValidateRegistry_Empty(t *testing.T) {
	r := &RegistryInfo{}
	err := r.ValidateRegistry()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "missing registry")
}

func TestValidateRegistry_NonEmpty(t *testing.T) {
	r := &RegistryInfo{Registry: "ghcr.io"}
	assert.NoError(t, r.ValidateRegistry())
}

// ---------------------------------------------------------------------------
// ValidateAllRegistry
// ---------------------------------------------------------------------------

func TestValidateAllRegistry_MissingRegistry(t *testing.T) {
	r := &RegistryInfo{Username: "user", Password: "pass"}
	err := r.ValidateAllRegistry()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "missing registry")
}

func TestValidateAllRegistry_MissingUsername(t *testing.T) {
	r := &RegistryInfo{Registry: "ghcr.io", Password: "pass"}
	err := r.ValidateAllRegistry()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "missing username")
}

func TestValidateAllRegistry_MissingPassword(t *testing.T) {
	r := &RegistryInfo{Registry: "ghcr.io", Username: "user"}
	err := r.ValidateAllRegistry()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "missing password")
}

func TestValidateAllRegistry_Valid(t *testing.T) {
	r := &RegistryInfo{Registry: "ghcr.io", Username: "user", Password: "pass"}
	assert.NoError(t, r.ValidateAllRegistry())
}

// ---------------------------------------------------------------------------
// getRegistryCredentials
// ---------------------------------------------------------------------------

// TestGetRegistryCredentials_ConfigMissing tests the no-config-file path,
// which creates an empty config.json and returns empty credentials.
func TestGetRegistryCredentials_ConfigMissing(t *testing.T) {
	oldFS := app.FS
	app.FS = afero.NewMemMapFs()
	defer func() { app.FS = oldFS }()

	// No config.json exists — function should create it and return empty creds.
	info, err := getRegistryCredentials("ghcr.io")
	require.NoError(t, err)
	assert.Equal(t, "ghcr.io", info.Registry)
	assert.Empty(t, info.Username)
	assert.Empty(t, info.Password)

	// The config file should now exist.
	exists, _ := afero.Exists(app.FS, filepath.Join(app.ConfigFolder, "config.json"))
	assert.True(t, exists)
}

// TestGetRegistryCredentials_ConfigExists_NoMatch tests reading a real
// config file that has registries, none of which match the requested host.
func TestGetRegistryCredentials_ConfigExists_NoMatch(t *testing.T) {
	oldFS := app.FS
	app.FS = afero.NewMemMapFs()
	defer func() { app.FS = oldFS }()

	configDir := app.ConfigFolder
	require.NoError(t, app.FS.MkdirAll(configDir, 0o755))
	configJSON := `{"registries":{"other.io":{"someuser":"somepass"}}}`
	require.NoError(t, afero.WriteFile(app.FS,
		filepath.Join(configDir, "config.json"), []byte(configJSON), 0o644))

	info, err := getRegistryCredentials("ghcr.io")
	require.NoError(t, err)
	assert.Empty(t, info.Username) // anonymous fallback
}

// TestGetRegistryCredentials_ConfigExists_ExactMatch verifies that matching
// credentials are returned.
func TestGetRegistryCredentials_ConfigExists_ExactMatch(t *testing.T) {
	oldFS := app.FS
	app.FS = afero.NewMemMapFs()
	defer func() { app.FS = oldFS }()

	configDir := app.ConfigFolder
	require.NoError(t, app.FS.MkdirAll(configDir, 0o755))
	configJSON := `{"registries":{"ghcr.io":{"myuser":"mysecret"}}}`
	require.NoError(t, afero.WriteFile(app.FS,
		filepath.Join(configDir, "config.json"), []byte(configJSON), 0o644))

	info, err := getRegistryCredentials("ghcr.io")
	require.NoError(t, err)
	assert.Equal(t, "myuser", info.Username)
	assert.Equal(t, "mysecret", info.Password)
}

// TestGetRegistryCredentials_InvalidJSON tests that a malformed config returns an error.
func TestGetRegistryCredentials_InvalidJSON(t *testing.T) {
	oldFS := app.FS
	app.FS = afero.NewMemMapFs()
	defer func() { app.FS = oldFS }()

	configDir := app.ConfigFolder
	require.NoError(t, app.FS.MkdirAll(configDir, 0o755))
	require.NoError(t, afero.WriteFile(app.FS,
		filepath.Join(configDir, "config.json"), []byte("not json"), 0o644))

	_, err := getRegistryCredentials("ghcr.io")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to parse config")
}

// TestGetRegistryCredentials_RegistryWithPort tests that a registry key with port
// still resolves via the stripped hostname.
func TestGetRegistryCredentials_RegistryWithPort(t *testing.T) {
	oldFS := app.FS
	app.FS = afero.NewMemMapFs()
	defer func() { app.FS = oldFS }()

	configDir := app.ConfigFolder
	require.NoError(t, app.FS.MkdirAll(configDir, 0o755))
	// Credentials stored under "localhost" (no port).
	configJSON := `{"registries":{"localhost":{"admin":"secret"}}}`
	require.NoError(t, afero.WriteFile(app.FS,
		filepath.Join(configDir, "config.json"), []byte(configJSON), 0o644))

	// Request credentials for "localhost:5000" — key without port should match.
	info, err := getRegistryCredentials("localhost:5000")
	require.NoError(t, err)
	assert.Equal(t, "admin", info.Username)
}
