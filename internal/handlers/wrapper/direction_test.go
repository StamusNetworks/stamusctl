package wrapper

import (
	"testing"

	"stamus-ctl/internal/app"
	"stamus-ctl/pkg/mocker"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
)

const minimalCompose = `services:
  web:
    image: nginx
`

// resetMocker drains the global Mocked map by calling Down with a path that
// contains no compose file; Down always resets the map regardless of whether
// getServices succeeds.
func resetMocker() {
	_ = mocker.Mocked.Down("/nonexistent-reset-path")
}

func TestHandleUp_TestMode(t *testing.T) {
	oldFS := app.FS
	app.FS = afero.NewMemMapFs()
	oldMode := app.Mode
	app.Mode = app.ModeStruct("test")
	oldName := app.Name
	app.Name = app.CtlName
	defer func() {
		app.FS = oldFS
		app.Mode = oldMode
		app.Name = oldName
		resetMocker()
	}()

	configPath := "/testconfig"
	err := afero.WriteFile(app.FS, configPath+"/docker-compose.yaml", []byte(minimalCompose), 0o644)
	assert.NoError(t, err)

	err = HandleUp(configPath, true)
	assert.NoError(t, err)

	containers, err := mocker.Mocked.Ps()
	assert.NoError(t, err)
	assert.NotEmpty(t, containers)
}

func TestHandleDown_TestMode(t *testing.T) {
	oldFS := app.FS
	app.FS = afero.NewMemMapFs()
	oldMode := app.Mode
	app.Mode = app.ModeStruct("test")
	oldName := app.Name
	app.Name = app.CtlName
	defer func() {
		app.FS = oldFS
		app.Mode = oldMode
		app.Name = oldName
		resetMocker()
	}()

	configPath := "/testconfig"
	err := afero.WriteFile(app.FS, configPath+"/docker-compose.yaml", []byte(minimalCompose), 0o644)
	assert.NoError(t, err)

	// Bring everything up first.
	err = HandleUp(configPath, true)
	assert.NoError(t, err)

	containers, err := mocker.Mocked.Ps()
	assert.NoError(t, err)
	assert.NotEmpty(t, containers, "expected containers after Up")

	// Now bring them down.
	err = HandleDown(configPath, false, false)
	assert.NoError(t, err)

	containers, err = mocker.Mocked.Ps()
	assert.NoError(t, err)
	assert.Empty(t, containers, "expected no containers after Down")
}

func TestHandleUp_PathResolution(t *testing.T) {
	oldFS := app.FS
	app.FS = afero.NewMemMapFs()
	oldMode := app.Mode
	app.Mode = app.ModeStruct("test")
	oldName := app.Name
	// Daemon mode: HandleUp must prepend ConfigsFolder to the supplied name.
	app.Name = "stamusd"
	defer func() {
		app.FS = oldFS
		app.Mode = oldMode
		app.Name = oldName
		resetMocker()
	}()

	name := "mydeployment"
	resolvedPath := app.GetConfigsFolder(name)
	err := afero.WriteFile(app.FS, resolvedPath+"/docker-compose.yaml", []byte(minimalCompose), 0o644)
	assert.NoError(t, err)

	err = HandleUp(name, true)
	assert.NoError(t, err)

	containers, err := mocker.Mocked.Ps()
	assert.NoError(t, err)
	assert.NotEmpty(t, containers)
}

func TestHandleDown_PathResolution_DaemonMode(t *testing.T) {
	oldFS := app.FS
	app.FS = afero.NewMemMapFs()
	oldMode := app.Mode
	app.Mode = app.ModeStruct("test")
	oldName := app.Name
	// Daemon mode: HandleDown must prepend ConfigsFolder to the supplied name.
	app.Name = "stamusd"
	defer func() {
		app.FS = oldFS
		app.Mode = oldMode
		app.Name = oldName
		resetMocker()
	}()

	name := "daemondeployment"
	resolvedPath := app.GetConfigsFolder(name)
	err := afero.WriteFile(app.FS, resolvedPath+"/docker-compose.yaml", []byte(minimalCompose), 0o644)
	assert.NoError(t, err)

	// Bring up first (daemon mode path resolution).
	err = HandleUp(name, true)
	assert.NoError(t, err)

	// Now bring down (exercises the !IsCtl() branch in HandleDown).
	err = HandleDown(name, false, false)
	assert.NoError(t, err)

	containers, err := mocker.Mocked.Ps()
	assert.NoError(t, err)
	assert.Empty(t, containers, "expected no containers after Down in daemon mode")
}
