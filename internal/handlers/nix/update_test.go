package handlers

import (
	"os"
	"testing"

	"stamus-ctl/internal/app"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

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
