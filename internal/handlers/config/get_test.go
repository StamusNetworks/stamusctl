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

// setupGetTestFS sets up a real temporary directory backed by afero.OsFs.
// GetVersion uses os.ReadFile, so we need real disk files.
func setupGetTestFS(t *testing.T) (tmpDir string, cleanup func()) {
	t.Helper()
	tmpDir, err := os.MkdirTemp("", "stamus-get-test-*")
	require.NoError(t, err)

	origFS := app.FS
	origName := app.Name
	origConfigsFolder := app.ConfigsFolder

	app.FS = afero.NewOsFs()
	app.Name = app.CtlName // CLI mode: IsCtl() == true
	app.ConfigsFolder = tmpDir

	cleanup = func() {
		app.FS = origFS
		app.Name = origName
		app.ConfigsFolder = origConfigsFolder
		os.RemoveAll(tmpDir)
	}
	return tmpDir, cleanup
}

// ---- GetVersion ----

func TestGetVersion_FileFound(t *testing.T) {
	tmpDir, cleanup := setupGetTestFS(t)
	defer cleanup()

	// Create a "version" file inside a config directory
	confDir := filepath.Join(tmpDir, "myconf")
	require.NoError(t, os.MkdirAll(confDir, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(confDir, "version"), []byte("1.2.3"), 0o644))

	// In CLI mode, GetVersion uses the path directly as given.
	result := GetVersion(confDir)
	assert.Equal(t, "1.2.3", result)
}

func TestGetVersion_FileNotFound(t *testing.T) {
	tmpDir, cleanup := setupGetTestFS(t)
	defer cleanup()

	nonExistentDir := filepath.Join(tmpDir, "no-such-conf")
	result := GetVersion(nonExistentDir)
	assert.Equal(t, "Could not read the version file", result)
}

func TestGetVersion_EmptyVersionFile(t *testing.T) {
	tmpDir, cleanup := setupGetTestFS(t)
	defer cleanup()

	confDir := filepath.Join(tmpDir, "emptyver")
	require.NoError(t, os.MkdirAll(confDir, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(confDir, "version"), []byte(""), 0o644))

	result := GetVersion(confDir)
	assert.Equal(t, "", result)
}

func TestGetVersion_MultilineVersion(t *testing.T) {
	tmpDir, cleanup := setupGetTestFS(t)
	defer cleanup()

	confDir := filepath.Join(tmpDir, "mlver")
	require.NoError(t, os.MkdirAll(confDir, 0o755))
	content := "3.0.0-beta\nbuilt from main"
	require.NoError(t, os.WriteFile(filepath.Join(confDir, "version"), []byte(content), 0o644))

	result := GetVersion(confDir)
	assert.Equal(t, content, result)
}

// ---- GetConfigsList ----

func TestGetConfigsList_Empty(t *testing.T) {
	_, cleanup := setupGetTestFS(t)
	defer cleanup()

	// The configs folder exists but is empty.
	// stamus.GetConfigsList reads from app.FS (which is OsFs pointing at tmpDir).
	configs, err := GetConfigsList()
	assert.NoError(t, err)
	assert.Empty(t, configs)
}

func TestGetConfigsList_MultipleConfigs(t *testing.T) {
	tmpDir, cleanup := setupGetTestFS(t)
	defer cleanup()

	// Create three config directories
	for _, name := range []string{"alpha", "beta", "gamma"} {
		require.NoError(t, os.MkdirAll(filepath.Join(tmpDir, name), 0o755))
	}
	// Also create a regular file — it should not appear in the list
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "notadir.txt"), []byte("x"), 0o644))

	configs, err := GetConfigsList()
	assert.NoError(t, err)
	assert.Len(t, configs, 3)
	assert.ElementsMatch(t, []string{"alpha", "beta", "gamma"}, configs)
}

func TestGetConfigsList_SingleConfig(t *testing.T) {
	tmpDir, cleanup := setupGetTestFS(t)
	defer cleanup()

	require.NoError(t, os.MkdirAll(filepath.Join(tmpDir, "only"), 0o755))

	configs, err := GetConfigsList()
	assert.NoError(t, err)
	require.Len(t, configs, 1)
	assert.Equal(t, "only", configs[0])
}

func TestGetConfigsList_ConfigsFolderDoesNotExist(t *testing.T) {
	origFS := app.FS
	origName := app.Name
	origConfigsFolder := app.ConfigsFolder
	defer func() {
		app.FS = origFS
		app.Name = origName
		app.ConfigsFolder = origConfigsFolder
	}()

	// Point ConfigsFolder at a path that does not exist.
	// stamus.GetConfigsList will create it automatically (returns empty list).
	app.FS = afero.NewMemMapFs()
	app.Name = app.CtlName
	app.ConfigsFolder = "/nonexistent-stamus-configs-path"

	configs, err := GetConfigsList()
	assert.NoError(t, err)
	assert.Empty(t, configs)
}

// ---- GetGroupedConfig error path ----

func TestGetGroupedConfig_MissingValuesFile(t *testing.T) {
	origFS := app.FS
	origName := app.Name
	app.FS = afero.NewMemMapFs()
	app.Name = app.CtlName
	defer func() {
		app.FS = origFS
		app.Name = origName
	}()

	// Pass a path with no values.yaml → should error
	_, err := GetGroupedConfig("/nonexistent", nil, false)
	assert.Error(t, err)
}

// ---- GetGroupedContent ----

func TestGetGroupedContent_ValidFolder(t *testing.T) {
	tmpDir, cleanup := setupGetTestFS(t)
	defer cleanup()

	// Create files on the real OS (GetGroupedContent uses filepath.Walk on real FS)
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "values.yaml"), []byte("key: val"), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "docker-compose.yaml"), []byte("version: '3'"), 0o644))

	grouped, err := GetGroupedContent(tmpDir, nil)
	assert.NoError(t, err)
	assert.NotNil(t, grouped)
}

func TestGetGroupedContent_NonExistentFolder(t *testing.T) {
	_, cleanup := setupGetTestFS(t)
	defer cleanup()

	_, err := GetGroupedContent("/absolutely-does-not-exist-xyz", nil)
	assert.Error(t, err)
}

func TestGetGroupedContent_WithFilter(t *testing.T) {
	tmpDir, cleanup := setupGetTestFS(t)
	defer cleanup()

	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "values.yaml"), []byte("x: y"), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "other.txt"), []byte("z"), 0o644))

	// Only files containing "values" should remain after filtering
	grouped, err := GetGroupedContent(tmpDir, []string{"values"})
	assert.NoError(t, err)
	assert.NotNil(t, grouped)
}

// ---- GetParamsList error path ----

func TestGetParamsList_MissingValuesFile(t *testing.T) {
	origFS := app.FS
	origName := app.Name
	app.FS = afero.NewMemMapFs()
	app.Name = app.CtlName
	defer func() {
		app.FS = origFS
		app.Name = origName
	}()

	_, err := GetParamsList("/nonexistent")
	assert.Error(t, err)
}

// setupMinimalConfig writes the minimal YAML files needed by LoadConfigFrom into app.FS.
// It returns the path to the config directory (which contains values.yaml).
// configDir must already exist in app.FS.
func setupMinimalConfig(t *testing.T, configDir string) {
	t.Helper()

	// The template config directory — referenced by stamus.config inside values.yaml
	templateDir := configDir + "/template"
	require.NoError(t, app.FS.MkdirAll(templateDir, 0o755))

	// Minimal config.yaml (parameter definitions)
	configYAML := `param1:
  type: string
  usage: A test parameter
  default: hello
`
	require.NoError(t, afero.WriteFile(app.FS, templateDir+"/config.yaml", []byte(configYAML), 0o644))

	// Minimal values.yaml — stamus.config points to template dir, stamus.project is set
	valuesYAML := `stamus:
  config: ` + templateDir + `
  project: test-project
param1: hello
`
	require.NoError(t, afero.WriteFile(app.FS, configDir+"/values.yaml", []byte(valuesYAML), 0o644))
}

func TestGetGroupedConfig_Success(t *testing.T) {
	origFS := app.FS
	origName := app.Name
	app.FS = afero.NewMemMapFs()
	app.Name = app.CtlName
	defer func() {
		app.FS = origFS
		app.Name = origName
	}()

	configDir := "/testconf"
	require.NoError(t, app.FS.MkdirAll(configDir, 0o755))
	setupMinimalConfig(t, configDir)

	grouped, err := GetGroupedConfig(configDir, nil, true)
	assert.NoError(t, err)
	assert.NotNil(t, grouped)
}

func TestGetParamsList_Success(t *testing.T) {
	origFS := app.FS
	origName := app.Name
	app.FS = afero.NewMemMapFs()
	app.Name = app.CtlName
	defer func() {
		app.FS = origFS
		app.Name = origName
	}()

	configDir := "/paramsconf"
	require.NoError(t, app.FS.MkdirAll(configDir, 0o755))
	setupMinimalConfig(t, configDir)

	params, err := GetParamsList(configDir)
	assert.NoError(t, err)
	assert.NotNil(t, params)
}

// ---------------------------------------------------------------------------
// Daemon-mode paths (app.Name != "stamusctl", so IsCtl() returns false).
// These tests cover the `if !app.IsCtl() { conf = app.GetConfigsFolder(conf) }`
// branch in each exported function.
// ---------------------------------------------------------------------------

// TestGetVersion_DaemonMode exercises the daemon-mode path where GetVersion
// prepends the ConfigsFolder to the supplied config name.
func TestGetVersion_DaemonMode(t *testing.T) {
	origName := app.Name
	origConfigsFolder := app.ConfigsFolder
	defer func() {
		app.Name = origName
		app.ConfigsFolder = origConfigsFolder
	}()

	// Create a real temp dir that contains a named config subdirectory with a version file.
	tmpDir, err := os.MkdirTemp("", "stamus-getver-daemon-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	confName := "myconf"
	confDir := filepath.Join(tmpDir, confName)
	require.NoError(t, os.MkdirAll(confDir, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(confDir, "version"), []byte("9.9.9"), 0o644))

	app.Name = "stamusd"             // IsCtl() == false
	app.ConfigsFolder = tmpDir + "/" // GetConfigsFolder joins this with confName

	result := GetVersion(confName)
	assert.Equal(t, "9.9.9", result)
}

// TestGetGroupedConfig_DaemonMode exercises the daemon path where GetGroupedConfig
// prepends ConfigsFolder to the config name.
func TestGetGroupedConfig_DaemonMode(t *testing.T) {
	origFS := app.FS
	origName := app.Name
	origConfigsFolder := app.ConfigsFolder
	defer func() {
		app.FS = origFS
		app.Name = origName
		app.ConfigsFolder = origConfigsFolder
	}()

	app.FS = afero.NewMemMapFs()
	app.Name = "stamusd"

	// ConfigsFolder + confName is where GetGroupedConfig will look.
	app.ConfigsFolder = "/dconfigs/"
	confName := "testconf"
	configDir := app.ConfigsFolder + confName
	require.NoError(t, app.FS.MkdirAll(configDir, 0o755))
	setupMinimalConfig(t, configDir)

	grouped, err := GetGroupedConfig(confName, nil, true)
	assert.NoError(t, err)
	assert.NotNil(t, grouped)
}

// TestGetGroupedContent_DaemonMode exercises the daemon path where GetGroupedContent
// prepends ConfigsFolder to the config name.
func TestGetGroupedContent_DaemonMode(t *testing.T) {
	origName := app.Name
	origConfigsFolder := app.ConfigsFolder
	defer func() {
		app.Name = origName
		app.ConfigsFolder = origConfigsFolder
	}()

	// GetGroupedContent calls utils.ListFilesInFolder which uses the real OS filesystem,
	// so we need actual disk files.
	tmpDir, err := os.MkdirTemp("", "stamus-getcontent-daemon-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	confName := "dcontent"
	confDir := filepath.Join(tmpDir, confName)
	require.NoError(t, os.MkdirAll(confDir, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(confDir, "values.yaml"), []byte("key: val"), 0o644))

	app.Name = "stamusd"
	app.ConfigsFolder = tmpDir + "/"

	grouped, err := GetGroupedContent(confName, nil)
	assert.NoError(t, err)
	assert.NotNil(t, grouped)
}

// TestGetParamsList_DaemonMode exercises the daemon path where GetParamsList
// prepends ConfigsFolder to the config name.
func TestGetParamsList_DaemonMode(t *testing.T) {
	origFS := app.FS
	origName := app.Name
	origConfigsFolder := app.ConfigsFolder
	defer func() {
		app.FS = origFS
		app.Name = origName
		app.ConfigsFolder = origConfigsFolder
	}()

	app.FS = afero.NewMemMapFs()
	app.Name = "stamusd"
	app.ConfigsFolder = "/dparams/"

	confName := "pconf"
	configDir := app.ConfigsFolder + confName
	require.NoError(t, app.FS.MkdirAll(configDir, 0o755))
	setupMinimalConfig(t, configDir)

	params, err := GetParamsList(confName)
	assert.NoError(t, err)
	assert.NotNil(t, params)
}

// ---------------------------------------------------------------------------
// Error path coverage for CreateFile validation failures
// ---------------------------------------------------------------------------

// TestGetGroupedConfig_InvalidPath exercises the models.CreateFile error path (line 37)
// by passing a path with traversal sequences that fail SanitizePath.
func TestGetGroupedConfig_InvalidPath(t *testing.T) {
	origFS := app.FS
	origName := app.Name
	app.FS = afero.NewMemMapFs()
	app.Name = app.CtlName
	defer func() {
		app.FS = origFS
		app.Name = origName
	}()

	// A path with null byte fails SanitizePath and causes CreateFile to error.
	_, err := GetGroupedConfig("/path\x00invalid", nil, false)
	assert.Error(t, err)
}

// TestGetParamsList_InvalidPath exercises the models.CreateFile error path (line 99)
// by passing a path with a null byte that fails SanitizePath validation.
func TestGetParamsList_InvalidPath(t *testing.T) {
	origFS := app.FS
	origName := app.Name
	app.FS = afero.NewMemMapFs()
	app.Name = app.CtlName
	defer func() {
		app.FS = origFS
		app.Name = origName
	}()

	_, err := GetParamsList("/path\x00invalid")
	assert.Error(t, err)
}
