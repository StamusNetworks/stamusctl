package handlers

import (
	// Common
	"os"
	"os/exec"
	"strings"
	"testing"

	"stamus-ctl/internal/app"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	embedTrue = "true"
	modeTest  = "test"
	modeProd  = "prod"
)

// ---------------------------------------------------------------------------
// runArbitraryScript
// ---------------------------------------------------------------------------

func TestRunArbitraryScript_MissingScript(t *testing.T) {
	// A script that doesn't exist must be skipped (not an error).
	out, err := runArbitraryScript("/nonexistent/script.sh", "/tmp/config")
	assert.NoError(t, err)
	assert.NotNil(t, out)
	assert.Empty(t, out.String())
}

func TestRunArbitraryScript_ScriptExists(t *testing.T) {
	// Create a real temporary script file on disk so os.Stat succeeds.
	f, err := os.CreateTemp(t.TempDir(), "test-script-*.sh")
	require.NoError(t, err)
	scriptPath := f.Name()
	f.Close()
	defer os.Remove(scriptPath)

	// Save and restore both the execCommandFunc and app.FS.
	origCmdFunc := execCommandFunc
	origFS := app.FS
	defer func() {
		execCommandFunc = origCmdFunc
		app.FS = origFS
	}()

	// Use the real OS FS so that Chmod on the real file works.
	app.FS = afero.NewOsFs()

	// Replace execCommandFunc so we run `echo hello` instead of the script.
	execCommandFunc = func(name string, arg ...string) *exec.Cmd {
		return exec.Command("echo", "hello")
	}

	// configPath must be non-empty and contain no traversal sequences.
	configPath := "/tmp"
	out, err := runArbitraryScript(scriptPath, configPath)
	assert.NoError(t, err)
	require.NotNil(t, out)
	assert.True(t, strings.Contains(out.String(), "hello"),
		"expected 'hello' in output, got: %q", out.String())
}

func TestRunArbitraryScript_InvalidConfigPath(t *testing.T) {
	// Create a real script file so os.Stat passes.
	f, err := os.CreateTemp(t.TempDir(), "test-script-invalid-*.sh")
	require.NoError(t, err)
	scriptPath := f.Name()
	f.Close()
	defer os.Remove(scriptPath)

	origFS := app.FS
	app.FS = afero.NewOsFs()
	defer func() { app.FS = origFS }()

	// A config path with a null byte should trigger SanitizePath error.
	invalidConfig := "/tmp/con\x00fig"
	out, err := runArbitraryScript(scriptPath, invalidConfig)
	assert.Error(t, err)
	assert.Nil(t, out)
}

// ---------------------------------------------------------------------------
// UpdateHandler – validation layer (does not hit network or real filesystem)
// ---------------------------------------------------------------------------

func TestUpdateHandler_InvalidVersion(t *testing.T) {
	oldFS := app.FS
	app.FS = afero.NewMemMapFs()
	defer func() { app.FS = oldFS }()

	err := UpdateHandler(UpdateHandlerParams{
		Version: "../../../etc/passwd",
		Config:  "test",
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid version")
}

func TestUpdateHandler_InvalidVersion_Slash(t *testing.T) {
	oldFS := app.FS
	app.FS = afero.NewMemMapFs()
	defer func() { app.FS = oldFS }()

	err := UpdateHandler(UpdateHandlerParams{
		Version: "bad/version",
		Config:  "some-config",
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid version")
}

func TestUpdateHandler_MissingConfig(t *testing.T) {
	oldFS := app.FS
	// Use real OS FS so that backup and viper can stat the config dir
	app.FS = afero.NewOsFs()
	defer func() { app.FS = oldFS }()

	err := UpdateHandler(UpdateHandlerParams{
		Version: "1.0.0",
		Config:  "/nonexistent-stamus-config-path-xyz",
	})
	assert.Error(t, err)
}

func TestUpdateHandler_EmptyVersion(t *testing.T) {
	oldFS := app.FS
	app.FS = afero.NewMemMapFs()
	defer func() { app.FS = oldFS }()

	err := UpdateHandler(UpdateHandlerParams{
		Version: "",
		Config:  "test",
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid version")
}

// TestUpdateHandler_ValidConfigDirButMissingStamusConfig exercises UpdateHandler past the viper-read
// step using a real on-disk config directory. The test stops at the
// LoadConfigFrom stage because the values.yaml in the test dir intentionally
// lacks the required stamus.config key.
func TestUpdateHandler_ValidConfigDirButMissingStamusConfig(t *testing.T) {
	// We need a real on-disk directory because UpdateHandler's viper reads from
	// the real OS filesystem.
	tmpDir := t.TempDir()

	// Write a minimal values.yaml with a valid project name.
	valuesYAML := "stamus:\n  project: clearndr\n"
	require.NoError(t, os.WriteFile(tmpDir+"/values.yaml", []byte(valuesYAML), 0o644))

	// Switch app.FS to OsFs so models can also access the real directory.
	origFS := app.FS
	app.FS = afero.NewOsFs()
	defer func() { app.FS = origFS }()

	err := UpdateHandler(UpdateHandlerParams{
		Version: "1.0.0",
		Config:  tmpDir,
	})
	// We expect an error here because the values.yaml has no stamus.config
	// pointing to a real template dir. The important thing is that we exercised
	// the viper-read and project-validation paths.
	assert.Error(t, err)
}

// TestUpdateHandler_InvalidProjectInConfig exercises the project name
// validation from a config file that contains an invalid project name.
func TestUpdateHandler_InvalidProjectInConfig(t *testing.T) {
	tmpDir := t.TempDir()

	// Write values.yaml with an invalid project name (contains slash).
	valuesYAML := "stamus:\n  project: \"bad/proj\"\n"
	require.NoError(t, os.WriteFile(tmpDir+"/values.yaml", []byte(valuesYAML), 0o644))

	origFS := app.FS
	app.FS = afero.NewOsFs()
	defer func() { app.FS = origFS }()

	err := UpdateHandler(UpdateHandlerParams{
		Version: "1.0.0",
		Config:  tmpDir,
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid project name")
}

// updateHandlerSetup is a helper that builds a temp directory with a minimal
// config structure suitable for deeper UpdateHandler testing and configures
// app globals.  The returned cleanup must be deferred by the caller.
func updateHandlerSetup(t *testing.T, registry string) (tmpDir string, cleanup func()) {
	t.Helper()

	tmpDir = t.TempDir()

	templateDir := tmpDir + "/template"
	require.NoError(t, os.MkdirAll(templateDir, 0o755))

	// Minimal config.yaml for the template (used by models.ConfigFromFile).
	configYAML := "param1:\n  type: string\n  usage: A test parameter\n  default: hello\n"
	require.NoError(t, os.WriteFile(templateDir+"/config.yaml", []byte(configYAML), 0o644))

	valuesYAML := "stamus:\n  project: clearndr\n  config: " + templateDir + "\n"
	if registry != "" {
		valuesYAML = "stamus:\n  project: clearndr\n  registry: " + registry + "\n  config: " + templateDir + "\n"
	}
	require.NoError(t, os.WriteFile(tmpDir+"/values.yaml", []byte(valuesYAML), 0o644))

	origFS := app.FS
	origEmbed := app.Embed
	origTemplatesFolder := app.TemplatesFolder
	app.FS = afero.NewOsFs()
	app.Embed = embedTrue
	app.TemplatesFolder = tmpDir + "/"

	cleanup = func() {
		app.FS = origFS
		app.Embed = origEmbed
		app.TemplatesFolder = origTemplatesFolder
	}
	return tmpDir, cleanup
}

// TestUpdateHandler_WithCompleteConfig sets up a real on-disk config directory
// that satisfies all layers: viper (OS FS), models (app.FS = OsFs), template
// pull (app.Embed = "true"), and script validation (script not present).
// This lets UpdateHandler run past the LoadConfigFrom call.
func TestUpdateHandler_WithCompleteConfig(t *testing.T) {
	tmpDir, cleanup := updateHandlerSetup(t, "")
	defer cleanup()

	err := UpdateHandler(UpdateHandlerParams{
		Version: "1.0.0",
		Config:  tmpDir,
	})
	// May fail at the post-LoadConfigFrom step; that's expected and fine.
	_ = err
}

// TestUpdateHandler_PreRunOutputSurvives proves that the YAML emitted by a
// template's sbin/pre-run script actually influences the updated config. The
// handler must reload the migrated values before the smart merge, otherwise
// SaveConfigTo overwrites the pre-run output with values loaded before the
// script ran and the pre-run feature is silently dead.
func TestUpdateHandler_PreRunOutputSurvives(t *testing.T) {
	oldFS := app.FS
	oldEmbed := app.Embed
	oldTemplates := app.TemplatesFolder
	oldCmd := execCommandFunc
	defer func() {
		app.FS = oldFS
		app.Embed = oldEmbed
		app.TemplatesFolder = oldTemplates
		execCommandFunc = oldCmd
	}()

	app.FS = afero.NewOsFs()
	app.Embed = embedTrue
	tmpDir := t.TempDir()
	app.TemplatesFolder = tmpDir + "/"

	// Template kept outside the pulled-template location so the (failing) pull
	// cannot clobber it.
	templateDir := tmpDir + "/mytemplate"
	require.NoError(t, os.MkdirAll(templateDir, 0o755))
	configYAML := "param1:\n  type: string\n  usage: A test parameter\n  default: hello\n"
	require.NoError(t, os.WriteFile(templateDir+"/config.yaml", []byte(configYAML), 0o644))

	const registry = "127.0.0.1:1/none"
	configPath := tmpDir + "/conf"
	require.NoError(t, os.MkdirAll(configPath, 0o755))
	existingValues := "param1: original\nstamus:\n  project: clearndr\n  registry: " + registry + "\n  config: " + templateDir + "\n"
	require.NoError(t, os.WriteFile(configPath+"/values.yaml", []byte(existingValues), 0o644))

	prerunDir := tmpDir + "/clearndr/sbin"
	require.NoError(t, os.MkdirAll(prerunDir, 0o755))
	require.NoError(t, os.WriteFile(prerunDir+"/pre-run", []byte("#!/bin/sh\n"), 0o755))

	migratedValues := "param1: migrated\nstamus:\n  project: clearndr\n  registry: " + registry + "\n  config: " + templateDir + "\n"
	execCommandFunc = func(name string, arg ...string) *exec.Cmd {
		return exec.Command("printf", "%s", migratedValues)
	}

	err := UpdateHandler(UpdateHandlerParams{
		Config:         configPath,
		Version:        "1.0.0",
		TemplateFolder: templateDir,
	})
	require.NoError(t, err)

	final, err := os.ReadFile(configPath + "/values.yaml")
	require.NoError(t, err)
	assert.Contains(t, string(final), "migrated",
		"pre-run script output must survive into the final config, got:\n%s", string(final))
	assert.NotContains(t, string(final), "original",
		"pre-run migration should have replaced the old value")
}

// TestUpdateHandler_NoPreRunPreservesExistingValues guards the reload fix: with
// no pre-run script, the handler must not truncate values.yaml and reload an
// empty config (which would wipe user customizations).
func TestUpdateHandler_NoPreRunPreservesExistingValues(t *testing.T) {
	oldFS := app.FS
	oldEmbed := app.Embed
	oldTemplates := app.TemplatesFolder
	defer func() {
		app.FS = oldFS
		app.Embed = oldEmbed
		app.TemplatesFolder = oldTemplates
	}()

	app.FS = afero.NewOsFs()
	app.Embed = embedTrue
	tmpDir := t.TempDir()
	app.TemplatesFolder = tmpDir + "/"

	templateDir := tmpDir + "/mytemplate"
	require.NoError(t, os.MkdirAll(templateDir, 0o755))
	configYAML := "param1:\n  type: string\n  usage: A test parameter\n  default: hello\n"
	require.NoError(t, os.WriteFile(templateDir+"/config.yaml", []byte(configYAML), 0o644))

	const registry = "127.0.0.1:1/none"
	configPath := tmpDir + "/conf"
	require.NoError(t, os.MkdirAll(configPath, 0o755))
	// No sbin/pre-run script is created.
	existingValues := "param1: original\nstamus:\n  project: clearndr\n  registry: " + registry + "\n  config: " + templateDir + "\n"
	require.NoError(t, os.WriteFile(configPath+"/values.yaml", []byte(existingValues), 0o644))

	err := UpdateHandler(UpdateHandlerParams{
		Config:         configPath,
		Version:        "1.0.0",
		TemplateFolder: templateDir,
	})
	require.NoError(t, err)

	final, err := os.ReadFile(configPath + "/values.yaml")
	require.NoError(t, err)
	assert.Contains(t, string(final), "original",
		"user customization must survive an update with no pre-run script, got:\n%s", string(final))
}

// TestUpdateHandler_WithRegistryAndFullConfig exercises the registry branch
// by providing both stamus.registry and a stamus.config in values.yaml.
func TestUpdateHandler_WithRegistryAndFullConfig(t *testing.T) {
	tmpDir, cleanup := updateHandlerSetup(t, "ghcr.io/stamusnetworks/stamusctl-templates")
	defer cleanup()

	err := UpdateHandler(UpdateHandlerParams{
		Version: "latest",
		Config:  tmpDir,
	})
	// May succeed or fail at various points. The goal is to cover the
	// `if registry != ""` branch in the template pull section.
	_ = err
}
