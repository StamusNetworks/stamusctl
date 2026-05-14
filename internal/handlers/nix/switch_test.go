package handlers

import (
	"testing"

	"stamus-ctl/internal/app"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
)

func TestNixSwitchHandler_ConfigNotFound(t *testing.T) {
	oldFS := app.FS
	app.FS = afero.NewMemMapFs()
	defer func() { app.FS = oldFS }()

	err := NixSwitchHandler(NixSwitchHandlerInputs{Config: "nonexistent"})
	assert.Error(t, err)
}

func TestNixSwitchHandler_NotNixOS(t *testing.T) {
	oldFS := app.FS
	app.FS = afero.NewMemMapFs()
	defer func() { app.FS = oldFS }()

	configPath := app.GetConfigsFolder("test")
	err := app.FS.MkdirAll(configPath, 0o755)
	assert.NoError(t, err)

	err = NixSwitchHandler(NixSwitchHandlerInputs{Config: configPath})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not running NixOS")
}
