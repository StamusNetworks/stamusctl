package handlers

import (
	"os"
	"testing"

	"stamus-ctl/internal/app"
	"stamus-ctl/internal/logging"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
)

func TestMain(m *testing.M) {
	logging.SetLogger()
	os.Exit(m.Run())
}

func TestNixInitHandler_InvalidProjectName(t *testing.T) {
	oldFS := app.FS
	app.FS = afero.NewMemMapFs()
	defer func() { app.FS = oldFS }()

	err := NixInitHandler(true, NixInitHandlerInputs{
		IsDefault: true,
		Project:   "../evil",
		Version:   "latest",
	})
	assert.Error(t, err)
}

func TestNixInitHandler_InvalidVersion(t *testing.T) {
	oldFS := app.FS
	app.FS = afero.NewMemMapFs()
	defer func() { app.FS = oldFS }()

	err := NixInitHandler(true, NixInitHandlerInputs{
		IsDefault: true,
		Project:   "clearndr",
		Version:   "../../../etc/passwd",
	})
	assert.Error(t, err)
}

func TestNixInitHandler_EmptyProject(t *testing.T) {
	oldFS := app.FS
	app.FS = afero.NewMemMapFs()
	defer func() { app.FS = oldFS }()

	err := NixInitHandler(true, NixInitHandlerInputs{
		IsDefault: true,
		Project:   "",
		Version:   "latest",
	})
	assert.Error(t, err)
}
