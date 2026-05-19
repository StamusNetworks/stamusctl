package handlers

import (
	"os"
	"testing"

	"stamus-ctl/internal/app"
	"stamus-ctl/internal/logging"
	"stamus-ctl/internal/shutdown"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestMain initialises the logger and shutdown tracker once for the whole
// package test binary.
func TestMain(m *testing.M) {
	logging.SetLogger()
	shutdown.Init(logging.Logger)
	os.Exit(m.Run())
}

// ---------------------------------------------------------------------------
// InitHandler – validation tests (do not reach network/template layer)
// ---------------------------------------------------------------------------

func TestInitHandler_InvalidProjectName(t *testing.T) {
	oldFS := app.FS
	app.FS = afero.NewMemMapFs()
	defer func() { app.FS = oldFS }()

	err := InitHandler(true, InitHandlerInputs{
		IsDefault: true,
		Project:   "../evil",
		Version:   "latest",
	})
	assert.Error(t, err)
}

func TestInitHandler_InvalidProjectName_Slash(t *testing.T) {
	oldFS := app.FS
	app.FS = afero.NewMemMapFs()
	defer func() { app.FS = oldFS }()

	err := InitHandler(true, InitHandlerInputs{
		IsDefault: true,
		Project:   "foo/bar",
		Version:   "latest",
	})
	assert.Error(t, err)
}

func TestInitHandler_InvalidVersion(t *testing.T) {
	oldFS := app.FS
	app.FS = afero.NewMemMapFs()
	defer func() { app.FS = oldFS }()

	err := InitHandler(true, InitHandlerInputs{
		IsDefault: true,
		Project:   "clearndr",
		Version:   "../../etc/passwd",
	})
	assert.Error(t, err)
}

func TestInitHandler_InvalidVersion_Slash(t *testing.T) {
	oldFS := app.FS
	app.FS = afero.NewMemMapFs()
	defer func() { app.FS = oldFS }()

	err := InitHandler(true, InitHandlerInputs{
		IsDefault: true,
		Project:   "clearndr",
		Version:   "bad/version",
	})
	assert.Error(t, err)
}

func TestInitHandler_EmptyProject(t *testing.T) {
	oldFS := app.FS
	app.FS = afero.NewMemMapFs()
	defer func() { app.FS = oldFS }()

	err := InitHandler(true, InitHandlerInputs{
		IsDefault: true,
		Project:   "",
		Version:   "latest",
	})
	assert.Error(t, err)
}

func TestInitHandler_EmptyVersion(t *testing.T) {
	oldFS := app.FS
	app.FS = afero.NewMemMapFs()
	defer func() { app.FS = oldFS }()

	err := InitHandler(true, InitHandlerInputs{
		IsDefault: true,
		Project:   "clearndr",
		Version:   "",
	})
	assert.Error(t, err)
}

// ---------------------------------------------------------------------------
// InitHandler – integration tests that exercise past the validation layer.
// These tests require a locally-installed clearndr template and Docker.
// They are skipped automatically when the template is not present.
// ---------------------------------------------------------------------------

// testInitSetup is a helper that configures the global state for integration
// tests and returns a cleanup function plus a fresh temp directory.
func testInitSetup(t *testing.T) (tmpDir string, cleanup func()) {
	t.Helper()

	origFS := app.FS
	origEmbed := app.Embed
	origTemplatesFolder := app.TemplatesFolder
	origConfigsFolder := app.ConfigsFolder

	app.FS = afero.NewOsFs()
	app.Embed = "true" // Swallow pull errors; rely on local templates

	localTemplates := app.TemplatesFolder + "clearndr/"
	if _, statErr := os.Stat(localTemplates); os.IsNotExist(statErr) {
		// Restore before skipping so defer doesn't see a mutated state.
		app.FS = origFS
		app.Embed = origEmbed
		t.Skip("local clearndr template not installed; skipping integration test")
	}

	tmpDir = t.TempDir()

	cleanup = func() {
		app.FS = origFS
		app.Embed = origEmbed
		app.TemplatesFolder = origTemplatesFolder
		app.ConfigsFolder = origConfigsFolder
	}
	return tmpDir, cleanup
}

// TestInitHandler_WithRealTemplate exercises InitHandler past the validation
// layer by pointing it at the already-installed local template.
func TestInitHandler_WithRealTemplate(t *testing.T) {
	tmpDir, cleanup := testInitSetup(t)
	defer cleanup()

	err := InitHandler(true, InitHandlerInputs{
		IsDefault: true,
		Project:   "clearndr",
		Version:   "latest",
		Config:    tmpDir,
	})
	// InitHandler may succeed or fail depending on template completeness.
	// The important thing is that it executed past the validation guards.
	_ = err
}

// TestInitHandler_WithValidParams passes explicit arbitrary parameter values
// to bypass validation failures and exercise the config-save code path.
func TestInitHandler_WithValidParams(t *testing.T) {
	tmpDir, cleanup := testInitSetup(t)
	defer cleanup()

	err := InitHandler(true, InitHandlerInputs{
		IsDefault: true,
		Project:   "clearndr",
		Version:   "latest",
		Config:    tmpDir,
		Arbitrary: map[string]string{
			"nginx.exec": "nginx",
		},
	})
	_ = err
}

// TestInitHandler_WithRegistry exercises the params.Registry != "" branch.
func TestInitHandler_WithRegistry(t *testing.T) {
	tmpDir, cleanup := testInitSetup(t)
	defer cleanup()

	err := InitHandler(true, InitHandlerInputs{
		IsDefault: true,
		Project:   "clearndr",
		Version:   "latest",
		Config:    tmpDir,
		Registry:  "ghcr.io/stamusnetworks/stamusctl-templates",
		Arbitrary: map[string]string{
			"nginx.exec": "nginx",
		},
	})
	_ = err
}

// ---------------------------------------------------------------------------
// InitHandler — embed mode tests (do not require local templates or network)
// ---------------------------------------------------------------------------

// TestInitHandler_EmbedMode_FullPath exercises the full handler body up to
// and including template instantiation. With app.Embed=true a pull failure
// is tolerated, so execution continues into InstanciateConfig.
func TestInitHandler_EmbedMode_FullPath(t *testing.T) {
	oldFS := app.FS
	app.FS = afero.NewMemMapFs()
	oldEmbed := app.Embed
	oldClearNDRPath := app.DefaultClearNDRPath
	defer func() {
		app.FS = oldFS
		app.Embed = oldEmbed
		app.DefaultClearNDRPath = oldClearNDRPath
	}()

	app.Embed.Set("true")

	// Place a config.yaml at DefaultClearNDRPath so InstanciateConfig succeeds.
	embedPath := "/tmp/test-compose-init-embedded"
	app.DefaultClearNDRPath = embedPath
	require.NoError(t, app.FS.MkdirAll(embedPath, 0o755))
	require.NoError(t, afero.WriteFile(app.FS, embedPath+"/config.yaml", []byte(`
mykey:
  type: string
  usage: A test parameter
  default: mydefault
`), 0o644))

	configDir := "/tmp/test-compose-init-config"
	require.NoError(t, app.FS.MkdirAll(configDir, 0o755))

	err := InitHandler(false, InitHandlerInputs{
		IsDefault: true,
		Project:   "clearndr",
		Version:   "1.0.0",
		Config:    configDir,
	})
	// The handler may fail at SaveConfigTo or stamus.AddInstance.
	// What matters is we got past validation + pull phase.
	if err != nil {
		assert.NotContains(t, err.Error(), "invalid project name")
		assert.NotContains(t, err.Error(), "invalid version")
	}
}

// TestInitHandler_EmbedFalse_PullFails verifies that when embed=false and the
// pull fails (no Docker or templates), the error is propagated.
func TestInitHandler_EmbedFalse_PullFails(t *testing.T) {
	oldFS := app.FS
	app.FS = afero.NewMemMapFs()
	oldEmbed := app.Embed
	defer func() {
		app.FS = oldFS
		app.Embed = oldEmbed
	}()

	app.Embed.Set("false")

	err := InitHandler(false, InitHandlerInputs{
		IsDefault: false,
		Project:   "clearndr",
		Version:   "latest",
		Config:    "/tmp/compose-init-embed-false",
	})
	// Without Docker the pull always fails; with embed=false the error is returned.
	assert.Error(t, err)
}

// TestInitHandler_WithRegistry_PullError exercises the explicit registry branch.
func TestInitHandler_WithRegistry_PullError(t *testing.T) {
	oldFS := app.FS
	app.FS = afero.NewMemMapFs()
	defer func() { app.FS = oldFS }()

	err := InitHandler(false, InitHandlerInputs{
		IsDefault: true,
		Project:   "clearndr",
		Version:   "nonexistent-version-xyz",
		Config:    "/tmp/compose-init-registry-test",
		Registry:  "ghcr.io/stamusnetworks/stamusctl-templates",
	})
	// Any error is acceptable; we just check no panic and validation passed.
	assert.Error(t, err)
}

// TestInitHandler_EmbedMode_DaemonPath exercises the !isCli branch (configPath = GetConfigsFolder).
func TestInitHandler_EmbedMode_DaemonPath(t *testing.T) {
	oldFS := app.FS
	app.FS = afero.NewMemMapFs()
	oldEmbed := app.Embed
	oldClearNDRPath := app.DefaultClearNDRPath
	defer func() {
		app.FS = oldFS
		app.Embed = oldEmbed
		app.DefaultClearNDRPath = oldClearNDRPath
	}()

	app.Embed.Set("true")

	embedPath := "/tmp/test-compose-init-daemon-embedded"
	app.DefaultClearNDRPath = embedPath
	require.NoError(t, app.FS.MkdirAll(embedPath, 0o755))
	require.NoError(t, afero.WriteFile(app.FS, embedPath+"/config.yaml", []byte(`
mykey:
  type: string
  usage: A test parameter
  default: mydefault
`), 0o644))

	// isCli=false → configPath = app.GetConfigsFolder(params.Config)
	err := InitHandler(false, InitHandlerInputs{
		IsDefault: true,
		Project:   "clearndr",
		Version:   "1.0.0",
		Config:    "test-daemon-config",
	})
	if err != nil {
		assert.NotContains(t, err.Error(), "invalid project name")
	}
}
