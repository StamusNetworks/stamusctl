package utils

import (
	"os"
	"path/filepath"
	"testing"

	"stamus-ctl/internal/app"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ---------------------------------------------------------------------------
// GetComposeFilePath — all four filename variants + fallback
// ---------------------------------------------------------------------------

func TestGetComposeFilePath_DockerComposeYAML(t *testing.T) {
	oldFS := app.FS
	app.FS = afero.NewMemMapFs()
	defer func() { app.FS = oldFS }()

	dir := "/GetComposeFilePath_yaml"
	require.NoError(t, app.FS.MkdirAll(dir, 0o750))
	require.NoError(t, afero.WriteFile(app.FS, filepath.Join(dir, "docker-compose.yaml"), []byte(""), 0o600))

	result := GetComposeFilePath(dir)
	assert.Equal(t, filepath.Join(dir, "docker-compose.yaml"), result)
}

func TestGetComposeFilePath_DockerComposeYML(t *testing.T) {
	oldFS := app.FS
	app.FS = afero.NewMemMapFs()
	defer func() { app.FS = oldFS }()

	dir := "/GetComposeFilePath_yml"
	require.NoError(t, app.FS.MkdirAll(dir, 0o750))
	require.NoError(t, afero.WriteFile(app.FS, filepath.Join(dir, "docker-compose.yml"), []byte(""), 0o600))

	result := GetComposeFilePath(dir)
	assert.Equal(t, filepath.Join(dir, "docker-compose.yml"), result)
}

func TestGetComposeFilePath_ComposeYAML(t *testing.T) {
	oldFS := app.FS
	app.FS = afero.NewMemMapFs()
	defer func() { app.FS = oldFS }()

	dir := "/GetComposeFilePath_compose_yaml"
	require.NoError(t, app.FS.MkdirAll(dir, 0o750))
	require.NoError(t, afero.WriteFile(app.FS, filepath.Join(dir, "compose.yaml"), []byte(""), 0o600))

	result := GetComposeFilePath(dir)
	assert.Equal(t, filepath.Join(dir, "compose.yaml"), result)
}

func TestGetComposeFilePath_ComposeYML(t *testing.T) {
	oldFS := app.FS
	app.FS = afero.NewMemMapFs()
	defer func() { app.FS = oldFS }()

	dir := "/GetComposeFilePath_compose_yml"
	require.NoError(t, app.FS.MkdirAll(dir, 0o750))
	require.NoError(t, afero.WriteFile(app.FS, filepath.Join(dir, "compose.yml"), []byte(""), 0o600))

	result := GetComposeFilePath(dir)
	assert.Equal(t, filepath.Join(dir, "compose.yml"), result)
}

func TestGetComposeFilePath_Fallback(t *testing.T) {
	oldFS := app.FS
	app.FS = afero.NewMemMapFs()
	defer func() { app.FS = oldFS }()

	// No compose file exists → should return the default fallback.
	dir := "/GetComposeFilePath_fallback"
	require.NoError(t, app.FS.MkdirAll(dir, 0o750))

	result := GetComposeFilePath(dir)
	assert.Equal(t, filepath.Join(dir, "docker-compose.yaml"), result)
}

// ---------------------------------------------------------------------------
// ListFilesInFolder — happy path using the real OS filesystem (temp dir)
// ---------------------------------------------------------------------------

func TestListFilesInFolder_HappyPath(t *testing.T) {
	// ListFilesInFolder uses filepath.Walk (real OS), so we need a real temp dir.
	dir := t.TempDir()

	// Create a couple of files using restricted permissions.
	require.NoError(t, os.WriteFile(filepath.Join(dir, "a.txt"), []byte("a"), 0o600))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "b.yaml"), []byte("b"), 0o600))

	// Create a sub-directory with one file.
	subDir := filepath.Join(dir, "sub")
	require.NoError(t, os.MkdirAll(subDir, 0o750))
	require.NoError(t, os.WriteFile(filepath.Join(subDir, "c.json"), []byte("c"), 0o600))

	result, err := ListFilesInFolder(dir)
	require.NoError(t, err)

	// Three files total (directories are excluded).
	assert.Len(t, result, 3)
	assert.Equal(t, "a.txt", result["a.txt"])
	assert.Equal(t, "b.yaml", result["b.yaml"])
	assert.Equal(t, "c.json", result[filepath.Join("sub", "c.json")])
}

func TestListFilesInFolder_EmptyDir(t *testing.T) {
	dir := t.TempDir()

	result, err := ListFilesInFolder(dir)
	require.NoError(t, err)
	assert.Empty(t, result)
}
