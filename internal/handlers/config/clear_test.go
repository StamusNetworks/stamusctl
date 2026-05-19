package config

import (
	"os"
	"path/filepath"
	"testing"

	"stamus-ctl/internal/app"

	"github.com/spf13/afero"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// setupClearTestFS creates a real temporary directory for deleteFolder tests.
// deleteFolder uses os.RemoveAll, which operates on the real filesystem.
func setupClearTestFS(t *testing.T) (tmpDir string, cleanup func()) {
	t.Helper()
	tmpDir, err := os.MkdirTemp("", "stamus-clear-test-*")
	require.NoError(t, err)

	origConfigsFolder := app.ConfigsFolder
	origName := app.Name
	// Use ctl mode so deleteFolder uses app.ConfigsFolder directly
	app.Name = app.CtlName
	app.ConfigsFolder = tmpDir

	cleanup = func() {
		app.ConfigsFolder = origConfigsFolder
		app.Name = origName
		os.RemoveAll(tmpDir)
	}
	return tmpDir, cleanup
}

func TestDeleteFolder_ValidPath(t *testing.T) {
	tmpDir, cleanup := setupClearTestFS(t)
	defer cleanup()

	// Create a sub-directory inside the configs folder
	target := filepath.Join(tmpDir, "myconfig")
	require.NoError(t, os.MkdirAll(target, 0o755))
	// Write a file inside it
	require.NoError(t, os.WriteFile(filepath.Join(target, "values.yaml"), []byte("key: val"), 0o644))

	err := deleteFolder(target)
	assert.NoError(t, err)

	// Directory should be gone
	_, statErr := os.Stat(target)
	assert.True(t, os.IsNotExist(statErr))
}

func TestDeleteFolder_ProtectedPaths(t *testing.T) {
	tmpDir, cleanup := setupClearTestFS(t)
	defer cleanup()

	// These paths must never be deleted; SanitizePath should reject them because
	// they escape the base configsFolder or are the protected-path check in deleteFolder.
	tests := []struct {
		name string
		path string
	}{
		{"root", "/"},
		{"etc", "/etc"},
		{"var", "/var"},
		{"home", "/home"},
		{"usr", "/usr"},
		{"traversal above configs", filepath.Join(tmpDir, "..", "evil")},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := deleteFolder(tt.path)
			assert.Error(t, err, "expected error for protected path %q", tt.path)
		})
	}
}

func TestDeleteFolder_EmptyPathRejected(t *testing.T) {
	_, cleanup := setupClearTestFS(t)
	defer cleanup()

	// SanitizePath returns an error for empty path
	err := deleteFolder("")
	assert.Error(t, err)
}

func TestDeleteFolder_PathTraversal(t *testing.T) {
	tmpDir, cleanup := setupClearTestFS(t)
	defer cleanup()

	// A traversal attempt: configs/../../../tmp
	traversal := filepath.Join(tmpDir, "..", "..", "..", "tmp")
	err := deleteFolder(traversal)
	assert.Error(t, err, "expected error for path traversal attempt")
}

func TestDeleteFolder_NonExistentPathWithinConfigs(t *testing.T) {
	tmpDir, cleanup := setupClearTestFS(t)
	defer cleanup()

	// A non-existent but otherwise valid path inside configsFolder.
	// os.RemoveAll is a no-op for non-existent paths, so no error is expected.
	target := filepath.Join(tmpDir, "does-not-exist")
	err := deleteFolder(target)
	assert.NoError(t, err)
}

func TestDeleteFolder_NestedSubdirectories(t *testing.T) {
	tmpDir, cleanup := setupClearTestFS(t)
	defer cleanup()

	// Create deeply nested structure
	target := filepath.Join(tmpDir, "conf", "sub", "deep")
	require.NoError(t, os.MkdirAll(target, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(target, "file.txt"), []byte("data"), 0o644))

	// Delete the top-level config dir (conf) inside tmpDir
	topLevel := filepath.Join(tmpDir, "conf")
	err := deleteFolder(topLevel)
	assert.NoError(t, err)

	_, statErr := os.Stat(topLevel)
	assert.True(t, os.IsNotExist(statErr))
}

// TestDeleteFolder_DaemonMode tests deleteFolder when app is in daemon mode
// (app.Name != "stamusctl"). Daemon mode validates against app.ConfigsFolder.
func TestDeleteFolder_DaemonMode(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "stamus-daemon-clear-test-*")
	require.NoError(t, err)

	origName := app.Name
	origConfigsFolder := app.ConfigsFolder
	// Switch to daemon mode
	app.Name = "stamusd"
	app.ConfigsFolder = tmpDir
	defer func() {
		app.Name = origName
		app.ConfigsFolder = origConfigsFolder
		os.RemoveAll(tmpDir)
	}()

	// Create valid target within configs folder
	target := filepath.Join(tmpDir, "myconf")
	require.NoError(t, os.MkdirAll(target, 0o755))

	err = deleteFolder(target)
	assert.NoError(t, err)
	_, statErr := os.Stat(target)
	assert.True(t, os.IsNotExist(statErr))
}

// TestDeleteFolder_DaemonMode_EmptyConfigsFolder tests that daemon mode with empty
// ConfigsFolder returns an error before even trying to sanitize.
func TestDeleteFolder_DaemonMode_EmptyConfigsFolder(t *testing.T) {
	origName := app.Name
	origConfigsFolder := app.ConfigsFolder
	app.Name = "stamusd"
	app.ConfigsFolder = "" // empty — should trigger the guard
	defer func() {
		app.Name = origName
		app.ConfigsFolder = origConfigsFolder
	}()

	err := deleteFolder("/some/path")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "configs folder not configured")
}

// TestClear_TestMode exercises the Clear function using app.Mode = "test" so that
// wrapper.HandleDown uses the mocker (no real docker needed).
// We use a real tmpdir because deleteFolder uses os.RemoveAll.
func TestClear_TestMode(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "stamus-clear-func-test-*")
	require.NoError(t, err)

	origName := app.Name
	origMode := app.Mode
	origConfigsFolder := app.ConfigsFolder
	origFS := app.FS
	defer func() {
		app.Name = origName
		app.Mode = origMode
		app.ConfigsFolder = origConfigsFolder
		app.FS = origFS
		os.RemoveAll(tmpDir)
	}()

	// Use CLI mode so paths are used directly
	app.Name = app.CtlName
	app.Mode = "test"
	app.ConfigsFolder = tmpDir
	// Use real OS FS so deleteFolder (os.RemoveAll) sees the directory
	app.FS = afero.NewOsFs()

	// Create the config directory with a docker-compose.yaml (required by mocker.Down)
	confDir := filepath.Join(tmpDir, "myconf")
	require.NoError(t, os.MkdirAll(confDir, 0o755))
	dcContent := `services:
  web:
    image: nginx
`
	require.NoError(t, os.WriteFile(filepath.Join(confDir, "docker-compose.yaml"), []byte(dcContent), 0o644))

	err = Clear(confDir)
	// Clear will try backup (expected to fail gracefully), then HandleDown (test mode),
	// deleteFolder, and RemoveInstance. The overall result may or may not error
	// depending on stamus config state — we just ensure no panic and that the
	// directory is removed.
	_ = err

	// Verify the config directory was deleted
	_, statErr := os.Stat(confDir)
	assert.True(t, os.IsNotExist(statErr), "config directory should be deleted after Clear")
}

// TestClear_NoDockerCompose verifies Clear completes even without a docker-compose.yaml.
// The test-mode mocker.Down ignores missing compose files and succeeds, so Clear
// proceeds to deleteFolder and removes the config directory.
func TestClear_NoDockerCompose(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "stamus-clear-nodc-test-*")
	require.NoError(t, err)

	origName := app.Name
	origMode := app.Mode
	origConfigsFolder := app.ConfigsFolder
	origFS := app.FS
	defer func() {
		app.Name = origName
		app.Mode = origMode
		app.ConfigsFolder = origConfigsFolder
		app.FS = origFS
		os.RemoveAll(tmpDir)
	}()

	app.Name = app.CtlName
	app.Mode = "test"
	app.ConfigsFolder = tmpDir
	app.FS = afero.NewOsFs()

	// Config dir exists but has no docker-compose.yaml
	// mocker.Down silently ignores the missing file
	confDir := filepath.Join(tmpDir, "nodc")
	require.NoError(t, os.MkdirAll(confDir, 0o755))

	err = Clear(confDir)
	_ = err // success or stamus RemoveInstance error are both acceptable

	// Either way, the config directory itself should be gone
	_, statErr := os.Stat(confDir)
	assert.True(t, os.IsNotExist(statErr), "config directory should be removed by Clear")
}

// ---------------------------------------------------------------------------
// deleteFolder — protected path check when ConfigsFolder is empty (CLI mode).
// With baseDir="", SanitizePath("/", "") returns "/" without error, and the
// protected-path guard in deleteFolder catches it.
// ---------------------------------------------------------------------------

func TestDeleteFolder_ProtectedPath_EmptyConfigsFolder(t *testing.T) {
	origName := app.Name
	origConfigsFolder := app.ConfigsFolder
	app.Name = app.CtlName // CLI mode
	app.ConfigsFolder = "" // empty → SanitizePath uses "" as baseDir
	defer func() {
		app.Name = origName
		app.ConfigsFolder = origConfigsFolder
	}()

	// "/" is a protected directory; with empty ConfigsFolder it reaches the guard.
	err := deleteFolder("/")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "refusing to delete protected directory")
}

// TestClear_DaemonMode exercises the daemon-mode path where Clear prepends
// ConfigsFolder to the supplied config name. Uses test mode so HandleDown
// calls the mocker instead of real docker-compose.
func TestClear_DaemonMode(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "stamus-clear-daemon-*")
	require.NoError(t, err)

	origName := app.Name
	origMode := app.Mode
	origConfigsFolder := app.ConfigsFolder
	origFS := app.FS
	defer func() {
		app.Name = origName
		app.Mode = origMode
		app.ConfigsFolder = origConfigsFolder
		app.FS = origFS
		os.RemoveAll(tmpDir)
	}()

	// Daemon mode: Clear will prepend ConfigsFolder to the config name.
	app.Name = "stamusd"
	app.Mode = "test"
	app.ConfigsFolder = tmpDir + "/"
	app.FS = afero.NewOsFs()

	confName := "daemon-conf"
	confDir := filepath.Join(tmpDir, confName)
	require.NoError(t, os.MkdirAll(confDir, 0o755))
	dcContent := "services:\n  web:\n    image: nginx\n"
	require.NoError(t, os.WriteFile(filepath.Join(confDir, "docker-compose.yaml"), []byte(dcContent), 0o644))

	// Clear with just the name (not the full path) so daemon path is exercised.
	err = Clear(confName)
	// The config directory should be gone regardless of error.
	_ = err
	_, statErr := os.Stat(confDir)
	assert.True(t, os.IsNotExist(statErr), "config directory should be deleted after Clear in daemon mode")
}
