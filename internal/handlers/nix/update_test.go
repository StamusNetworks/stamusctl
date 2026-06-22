package handlers

import (
	"os"
	"os/exec"
	"testing"

	"stamus-ctl/internal/app"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestNixUpdateHandler_PreRunOutputSurvives proves that the YAML emitted by a
// template's sbin/pre-run script actually influences the updated config. The
// pre-run script migrates the existing values; the handler must reload those
// migrated values before the smart merge, otherwise SaveConfigTo overwrites the
// pre-run output with values loaded *before* the script ran and the pre-run
// feature is silently dead.
func TestNixUpdateHandler_PreRunOutputSurvives(t *testing.T) {
	oldName := app.Name
	oldFS := app.FS
	oldEmbed := app.Embed
	oldTemplates := app.TemplatesFolder
	oldCmd := execCommandFunc
	defer func() {
		app.Name = oldName
		app.FS = oldFS
		app.Embed = oldEmbed
		app.TemplatesFolder = oldTemplates
		execCommandFunc = oldCmd
	}()

	app.Name = app.CtlName
	app.FS = afero.NewOsFs()
	app.Embed = "true" // tolerate the (failing) template pull
	tmpDir := t.TempDir()
	app.TemplatesFolder = tmpDir + "/"

	// Template with one string parameter (default "hello"). Kept OUTSIDE the
	// pulled-template location (tmpDir/clearndr) so the (failing) pull cannot
	// clobber it with real template files.
	templateDir := tmpDir + "/mytemplate"
	require.NoError(t, os.MkdirAll(templateDir, 0o755))
	configYAML := "param1:\n  type: string\n  usage: A test parameter\n  default: hello\n"
	require.NoError(t, os.WriteFile(templateDir+"/config.yaml", []byte(configYAML), 0o644))

	// Existing config: param1 customized to "original". An unreachable registry
	// forces the pull to fail fast and deterministically (tolerated by embed).
	const registry = "127.0.0.1:1/none"
	configPath := tmpDir + "/conf"
	require.NoError(t, os.MkdirAll(configPath, 0o755))
	existingValues := "param1: original\nstamus:\n  project: clearndr\n  registry: " + registry + "\n  config: " + templateDir + "\n"
	require.NoError(t, os.WriteFile(configPath+"/values.yaml", []byte(existingValues), 0o644))

	// Pre-run script: present on disk so runScript proceeds to execCommandFunc.
	prerunDir := tmpDir + "/clearndr/sbin"
	require.NoError(t, os.MkdirAll(prerunDir, 0o755))
	require.NoError(t, os.WriteFile(prerunDir+"/pre-run", []byte("#!/bin/sh\n"), 0o755))

	// The pre-run script "migrates" param1 to "migrated". Mock its execution so
	// it deterministically emits the migrated values.yaml on stdout.
	migratedValues := "param1: migrated\nstamus:\n  project: clearndr\n  registry: " + registry + "\n  config: " + templateDir + "\n"
	execCommandFunc = func(name string, arg ...string) *exec.Cmd {
		return exec.Command("printf", "%s", migratedValues)
	}

	err := NixUpdateHandler(NixUpdateHandlerInputs{
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

// TestNixUpdateHandler_NoPreRunPreservesExistingValues guards the pre-run reload
// fix: when no pre-run script exists, the handler must NOT truncate values.yaml
// and reload an empty config (which would wipe the user's customizations). The
// in-memory existing config must drive the smart merge.
func TestNixUpdateHandler_NoPreRunPreservesExistingValues(t *testing.T) {
	oldName := app.Name
	oldFS := app.FS
	oldEmbed := app.Embed
	oldTemplates := app.TemplatesFolder
	defer func() {
		app.Name = oldName
		app.FS = oldFS
		app.Embed = oldEmbed
		app.TemplatesFolder = oldTemplates
	}()

	app.Name = app.CtlName
	app.FS = afero.NewOsFs()
	app.Embed = "true"
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

	err := NixUpdateHandler(NixUpdateHandlerInputs{
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

func TestNixUpdateHandler_InvalidVersion(t *testing.T) {
	oldFS := app.FS
	app.FS = afero.NewMemMapFs()
	defer func() { app.FS = oldFS }()

	oldName := app.Name
	app.Name = app.CtlName
	defer func() { app.Name = oldName }()

	err := NixUpdateHandler(NixUpdateHandlerInputs{
		Config:  "/tmp/test",
		Version: "../../../etc/passwd",
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid version")
}

func TestNixUpdateHandler_ConfigMissing(t *testing.T) {
	oldFS := app.FS
	app.FS = afero.NewMemMapFs()
	defer func() { app.FS = oldFS }()

	oldName := app.Name
	app.Name = app.CtlName
	defer func() { app.Name = oldName }()

	err := NixUpdateHandler(NixUpdateHandlerInputs{
		Config:  "/nonexistent",
		Version: "latest",
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cannot read config file")
}

func TestNixUpdateHandler_EmptyVersion(t *testing.T) {
	oldFS := app.FS
	app.FS = afero.NewMemMapFs()
	defer func() { app.FS = oldFS }()

	oldName := app.Name
	app.Name = app.CtlName
	defer func() { app.Name = oldName }()

	err := NixUpdateHandler(NixUpdateHandlerInputs{
		Config:  "/tmp/test",
		Version: "",
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid version")
}

func TestNixUpdateHandler_InvalidProjectName(t *testing.T) {
	oldName := app.Name
	app.Name = app.CtlName
	defer func() { app.Name = oldName }()

	// Viper reads from the real filesystem, so use a real temp dir
	configPath := t.TempDir()
	valuesContent := "stamus:\n  project: ../../../etc/passwd\n"
	require.NoError(t, os.WriteFile(configPath+"/values.yaml", []byte(valuesContent), 0o644))

	err := NixUpdateHandler(NixUpdateHandlerInputs{
		Config:  configPath,
		Version: "latest",
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid project name")
}

func TestNixUpdateHandler_DaemonMode_ConfigMissing(t *testing.T) {
	oldFS := app.FS
	app.FS = afero.NewMemMapFs()
	defer func() { app.FS = oldFS }()

	oldName := app.Name
	app.Name = "stamusd"
	defer func() { app.Name = oldName }()

	err := NixUpdateHandler(NixUpdateHandlerInputs{
		Config:  "nonexistent-daemon-conf",
		Version: "latest",
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cannot read config file")
}

func TestNixUpdateHandler_ValidConfig_PullFails(t *testing.T) {
	oldName := app.Name
	app.Name = app.CtlName
	defer func() { app.Name = oldName }()

	// Viper reads from real filesystem, so use a real temp dir
	configPath := t.TempDir()
	valuesContent := "stamus:\n  project: clearndr\n  registry: \"\"\n"
	require.NoError(t, os.WriteFile(configPath+"/values.yaml", []byte(valuesContent), 0o644))

	// Embed mode is off (default), so pull failure is fatal.
	// The handler will fail trying to load config from values.yaml (no params section)
	// or fail at the pull step — either way it errors.
	err := NixUpdateHandler(NixUpdateHandlerInputs{
		Config:  configPath,
		Version: "latest",
	})
	assert.Error(t, err)
}
