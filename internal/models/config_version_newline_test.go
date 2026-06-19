package models

import (
	"testing"

	"stamus-ctl/internal/app"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestSaveParamsTo_LatestVersionTrailingNewline reproduces a bug where a
// version file with a trailing newline (common for files produced by echo or
// CI tooling) embeds the \n into the stamus.config path stored in
// values.yaml.  Subsequent LoadConfigFrom then passes that path through
// SanitizePath which rejects control characters, causing:
//
//	Error: invalid file path: path contains forbidden control characters
func TestSaveParamsTo_LatestVersionTrailingNewline(t *testing.T) {
	// Create a config file under a "latest" directory so saveParamsTo
	// enters the version-rewriting branch.
	file, err := CreateFile("/tpl/latest", "config.yaml")
	require.NoError(t, err)

	// Write the version file WITH a trailing newline — this is the norm
	// for files written by shell commands (echo "1.0.0" > version).
	err = afero.WriteFile(app.FS, "/tpl/latest/version", []byte("1.0.0\n"), 0o644)
	require.NoError(t, err)

	config := &Config{
		file:      file,
		project:   "test_project",
		arbitrary: &Arbitrary{},
		parameters: &Parameters{
			"param1": &Parameter{
				Variable: CreateVariableString("value1"),
				Type:     "string",
			},
		},
	}

	destFile, err := CreateFile("/dest_newline", "config.yaml")
	require.NoError(t, err)

	err = config.saveParamsTo(destFile)
	require.NoError(t, err)

	// Re-read the written config to inspect what got persisted.
	viperInstance, err := destFile.InstanciateViper()
	require.NoError(t, err)

	stamusConfig := viperInstance.GetString("stamus.config")

	// The stored path must not contain any newline characters.
	assert.NotContains(t, stamusConfig, "\n",
		"stamus.config must not contain a newline — got %q", stamusConfig)
	// Positive check: the path should resolve to the trimmed version.
	assert.Equal(t, "/tpl/1.0.0", stamusConfig)
}

// TestGetStamusFile_TrailingNewlineInStoredPath reproduces the situation of a
// user whose values.yaml was written by an OLDER binary (before the
// trailing-newline write fix): the stored stamus.config path already contains
// an embedded "\n". GetStamusFile must trim it so the existing, on-disk config
// remains loadable instead of failing with:
//
//	invalid file path: path contains forbidden control characters
func TestGetStamusFile_TrailingNewlineInStoredPath(t *testing.T) {
	stamusConf := CreateVariableString("/tpl/1.0.0\n")
	values := map[string]*Variable{
		"stamus.config": &stamusConf,
	}

	file, err := GetStamusFile(values)
	require.NoError(t, err, "GetStamusFile must tolerate a trailing newline in the stored path")
	require.NotNil(t, file)
	assert.Equal(t, "/tpl/1.0.0", file.Path,
		"the newline must be trimmed from the resolved path")
}

// TestLoadConfigFrom_VersionNewlineRoundTrip is the end-to-end reproduction
// of the CI failure: compose init writes a values.yaml whose stamus.config
// contains a trailing newline, then config get keys calls LoadConfigFrom
// which fails at GetStamusFile → CreateFile → SanitizePath.
func TestLoadConfigFrom_VersionNewlineRoundTrip(t *testing.T) {
	// --- Set up template config.yaml (the "origin" config) ---
	templateDir := "/roundtrip/tpl/1.0.0"
	templateConfigContent := `
param1:
    usage: "A test param"
    type: "string"
    default: "hello"
`
	err := app.FS.MkdirAll(templateDir, 0o755)
	require.NoError(t, err)
	err = afero.WriteFile(app.FS, templateDir+"/config.yaml", []byte(templateConfigContent), 0o644)
	require.NoError(t, err)

	// --- Set up the "latest" template dir with a newline-terminated version file ---
	latestDir := "/roundtrip/tpl/latest"
	err = app.FS.MkdirAll(latestDir, 0o755)
	require.NoError(t, err)
	err = afero.WriteFile(app.FS, latestDir+"/config.yaml", []byte(templateConfigContent), 0o644)
	require.NoError(t, err)
	err = afero.WriteFile(app.FS, latestDir+"/version", []byte("1.0.0\n"), 0o644)
	require.NoError(t, err)

	// --- Build a Config pointing at "latest" and save it ---
	srcFile, err := CreateFile(latestDir, "config.yaml")
	require.NoError(t, err)

	config := &Config{
		file:      srcFile,
		project:   "roundtrip",
		seed:      "testseed1234",
		arbitrary: &Arbitrary{},
		parameters: &Parameters{
			"param1": &Parameter{
				Variable: CreateVariableString("world"),
				Type:     "string",
			},
		},
	}

	destDir := "/roundtrip/config"
	err = app.FS.MkdirAll(destDir, 0o755)
	require.NoError(t, err)
	destFile, err := CreateFile(destDir, "values.yaml")
	require.NoError(t, err)

	err = config.saveParamsTo(destFile)
	require.NoError(t, err)

	// --- Now attempt to reload the saved config (the failing operation) ---
	reloadFile, err := CreateFile(destDir, "values.yaml")
	require.NoError(t, err)

	loaded, err := LoadConfigFrom(reloadFile, true)
	assert.NoError(t, err, "LoadConfigFrom must not fail due to newline in stamus.config path")
	if loaded != nil {
		assert.Equal(t, "roundtrip", loaded.project)
	}
}
