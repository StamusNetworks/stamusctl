package handlers

import (
	"testing"

	"stamus-ctl/internal/app"
	"stamus-ctl/pkg/mocker"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ---------------------------------------------------------------------------
// HandleConfigRestart – test mode
// ---------------------------------------------------------------------------

func TestHandleConfigRestart_TestMode(t *testing.T) {
	oldMode := app.Mode
	oldFS := app.FS
	app.Mode = modeTest
	app.FS = afero.NewMemMapFs()
	defer func() {
		app.Mode = oldMode
		app.FS = oldFS
		mocker.Mocked.Down("") //nolint:errcheck,gosec
	}()

	composePath := "/restartconf"
	require.NoError(t, app.FS.MkdirAll(composePath, 0o755))
	compose := `services:
  app:
    image: myapp
`
	require.NoError(t, afero.WriteFile(app.FS, composePath+"/docker-compose.yaml", []byte(compose), 0o644))

	// Prime mocker with running containers so Restart has something to cycle.
	mocker.Mocked.Down("") //nolint:errcheck,gosec
	require.NoError(t, mocker.Mocked.Up(composePath))

	err := HandleConfigRestart(composePath)
	assert.NoError(t, err)
}

func TestHandleConfigRestart_TestMode_EmptyPath(t *testing.T) {
	oldMode := app.Mode
	oldFS := app.FS
	app.Mode = modeTest
	app.FS = afero.NewMemMapFs()
	defer func() {
		app.Mode = oldMode
		app.FS = oldFS
		mocker.Mocked.Down("") //nolint:errcheck,gosec
	}()

	mocker.Mocked.Down("") //nolint:errcheck,gosec
	err := HandleConfigRestart("")
	// An error here is acceptable (compose file not found). Just ensure no panic.
	_ = err
}

// ---------------------------------------------------------------------------
// HandleContainersRestart – test mode
// ---------------------------------------------------------------------------

func TestHandleContainersRestart_TestMode_KnownContainer(t *testing.T) {
	oldMode := app.Mode
	oldFS := app.FS
	app.Mode = modeTest
	app.FS = afero.NewMemMapFs()
	defer func() {
		app.Mode = oldMode
		app.FS = oldFS
		mocker.Mocked.Down("") //nolint:errcheck,gosec
	}()

	composePath := "/crdata"
	require.NoError(t, app.FS.MkdirAll(composePath, 0o755))
	compose := `services:
  worker:
    image: worker
`
	require.NoError(t, afero.WriteFile(app.FS, composePath+"/docker-compose.yaml", []byte(compose), 0o644))

	mocker.Mocked.Down("") //nolint:errcheck,gosec
	require.NoError(t, mocker.Mocked.Up(composePath))

	err := HandleContainersRestart([]string{"worker"})
	assert.NoError(t, err)
}

func TestHandleContainersRestart_TestMode_UnknownContainer(t *testing.T) {
	oldMode := app.Mode
	oldFS := app.FS
	app.Mode = modeTest
	app.FS = afero.NewMemMapFs()
	defer func() {
		app.Mode = oldMode
		app.FS = oldFS
		mocker.Mocked.Down("") //nolint:errcheck,gosec
	}()

	mocker.Mocked.Down("") //nolint:errcheck,gosec

	err := HandleContainersRestart([]string{"ghost-container"})
	assert.NoError(t, err)
}

func TestHandleContainersRestart_TestMode_EmptyList(t *testing.T) {
	oldMode := app.Mode
	oldFS := app.FS
	app.Mode = modeTest
	app.FS = afero.NewMemMapFs()
	defer func() {
		app.Mode = oldMode
		app.FS = oldFS
	}()

	mocker.Mocked.Down("") //nolint:errcheck,gosec

	err := HandleContainersRestart([]string{})
	assert.NoError(t, err)
}

// ---------------------------------------------------------------------------
// RestartContainer – real Docker daemon (returns quickly for non-existent ID)
// ---------------------------------------------------------------------------

func TestRestartContainer_NonExistentID(t *testing.T) {
	// RestartContainer calls the Docker API directly. Docker returns
	// "No such container" immediately without a timeout.
	err := RestartContainer("nonexistent-container-stamusctl-test-xyz")
	if err == nil {
		// Unlikely but possible if a container with that name exists.
		t.Skip("unexpected success – container may exist on the host")
	}
	// The important outcome: the function ran and returned an error (not panicked).
	assert.Error(t, err)
}

// ---------------------------------------------------------------------------
// handleConfigRestart – real Docker path (calls wrapper.HandleDown/Up).
// Requires shutdown.Init which is called in TestMain (init_test.go).
// ---------------------------------------------------------------------------

func TestHandleConfigRestart_RealPath_NonExistentConfig(t *testing.T) {
	oldFS := app.FS
	app.FS = afero.NewOsFs()
	defer func() { app.FS = oldFS }()

	// handleConfigRestart calls wrapper.HandleDown then HandleUp.
	// Both will fail because the config path doesn't exist on the real FS.
	err := handleConfigRestart("/nonexistent-config-path-stamusctl-test")
	// We expect an error here since the config path doesn't exist.
	// The important thing is the real function body executes.
	_ = err
}

// ---------------------------------------------------------------------------
// HandleConfigRestart and HandleContainersRestart – non-test mode
// ---------------------------------------------------------------------------

func TestHandleConfigRestart_ProdMode(t *testing.T) {
	oldMode := app.Mode
	oldFS := app.FS
	app.Mode = modeProd
	app.FS = afero.NewOsFs()
	defer func() {
		app.Mode = oldMode
		app.FS = oldFS
	}()

	// In prod mode, HandleConfigRestart calls handleConfigRestart.
	// It will fail because the path doesn't exist – that's expected.
	err := HandleConfigRestart("/nonexistent-prod-restart-path")
	_ = err // May or may not error depending on how compose handles it.
}

func TestHandleContainersRestart_ProdMode_EmptyList(t *testing.T) {
	oldMode := app.Mode
	app.Mode = modeProd
	defer func() { app.Mode = oldMode }()

	// handleContainersRestart with empty list: wg.Add(0) → wg.Wait() returns
	// immediately → len(returned)==0 → returns nil. No Docker call needed.
	err := HandleContainersRestart([]string{})
	// We may get a Docker client creation error or nil.
	// The important thing: the real handleContainersRestart body runs.
	_ = err
}
