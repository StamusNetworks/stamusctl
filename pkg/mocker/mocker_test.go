package mocker

import (
	"testing"

	"stamus-ctl/internal/app"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const testComposeYAML = `services:
  web:
    image: nginx
  db:
    image: postgres
`

// setupMockerFS sets up an in-memory FS with a docker-compose.yaml at the given path.
func setupMockerFS(t *testing.T, composePath, composeContent string) (cleanup func()) {
	t.Helper()
	origFS := app.FS
	memFS := afero.NewMemMapFs()
	app.FS = memFS

	require.NoError(t, memFS.MkdirAll(composePath, 0o755))
	require.NoError(t, afero.WriteFile(memFS, composePath+"/docker-compose.yaml", []byte(composeContent), 0o644))

	return func() { app.FS = origFS }
}

// ---------------------------------------------------------------------------
// createMocked
// ---------------------------------------------------------------------------

func TestCreateMocked_ReturnsEmptyMap(t *testing.T) {
	m := createMocked()
	assert.Empty(t, m)
}

// ---------------------------------------------------------------------------
// Up / Down
// ---------------------------------------------------------------------------

func TestUp_CreatesContainers(t *testing.T) {
	cleanup := setupMockerFS(t, "/test/conf", testComposeYAML)
	defer cleanup()

	m := createMocked()
	err := m.Up("/test/conf")
	assert.NoError(t, err)
	assert.Len(t, m, 2)
}

func TestUp_MissingComposeFile_ReturnsError(t *testing.T) {
	origFS := app.FS
	app.FS = afero.NewMemMapFs()
	defer func() { app.FS = origFS }()

	m := createMocked()
	err := m.Up("/no/such/path")
	assert.Error(t, err)
}

func TestDown_ClearsContainers(t *testing.T) {
	cleanup := setupMockerFS(t, "/test/down", testComposeYAML)
	defer cleanup()

	m := createMocked()
	_ = m.Up("/test/down")
	assert.NotEmpty(t, m)

	err := m.Down("/test/down")
	assert.NoError(t, err)
	assert.Empty(t, m)
}

func TestDown_MissingComposeFile_StillClearsMap(t *testing.T) {
	origFS := app.FS
	app.FS = afero.NewMemMapFs()
	defer func() { app.FS = origFS }()

	m := createMocked()
	// Down with missing file still resets the map
	err := m.Down("/no/such")
	assert.NoError(t, err)
	assert.Empty(t, m)
}

// ---------------------------------------------------------------------------
// Restart
// ---------------------------------------------------------------------------

func TestRestart_RecreatesContainers(t *testing.T) {
	cleanup := setupMockerFS(t, "/test/restart", testComposeYAML)
	defer cleanup()

	m := createMocked()
	err := m.Restart("/test/restart")
	assert.NoError(t, err)
	assert.Len(t, m, 2)
}

// ---------------------------------------------------------------------------
// RestartContainers
// ---------------------------------------------------------------------------

func TestRestartContainers_ExistingContainer(t *testing.T) {
	cleanup := setupMockerFS(t, "/test/restartc", testComposeYAML)
	defer cleanup()

	m := createMocked()
	_ = m.Up("/test/restartc")

	// Get one container name
	var name string
	for _, c := range m {
		name = c.Names[0][1:] // strip leading "/"
		break
	}

	err := m.RestartContainers([]string{name})
	assert.NoError(t, err)
}

func TestRestartContainers_NonExistentContainer(t *testing.T) {
	m := createMocked()

	// Restarting a container not in the map is a no-op
	err := m.RestartContainers([]string{"nonexistent"})
	assert.NoError(t, err)
}

// ---------------------------------------------------------------------------
// Ps
// ---------------------------------------------------------------------------

func TestPs_Empty(t *testing.T) {
	m := createMocked()
	containers, err := m.Ps()
	assert.NoError(t, err)
	assert.Empty(t, containers)
}

func TestPs_WithContainers(t *testing.T) {
	cleanup := setupMockerFS(t, "/test/ps", testComposeYAML)
	defer cleanup()

	m := createMocked()
	_ = m.Up("/test/ps")

	containers, err := m.Ps()
	assert.NoError(t, err)
	assert.Len(t, containers, 2)
}

// ---------------------------------------------------------------------------
// Logs
// ---------------------------------------------------------------------------

func TestLogs_Empty(t *testing.T) {
	m := createMocked()
	logs, err := m.Logs()
	assert.NoError(t, err)
	assert.Empty(t, logs.Containers)
}

func TestLogs_WithContainers(t *testing.T) {
	cleanup := setupMockerFS(t, "/test/logs", testComposeYAML)
	defer cleanup()

	m := createMocked()
	_ = m.Up("/test/logs")

	logs, err := m.Logs()
	assert.NoError(t, err)
	assert.Len(t, logs.Containers, 2)
	for _, cl := range logs.Containers {
		assert.NotEmpty(t, cl.Logs)
	}
}

// ---------------------------------------------------------------------------
// randomContainerId
// ---------------------------------------------------------------------------

func TestRandomContainerId_IsUnique(t *testing.T) {
	id1 := randomContainerId()
	id2 := randomContainerId()
	assert.NotEqual(t, id1, id2)
	assert.Len(t, id1, 64) // Docker IDs are 64 hex chars
}

// ---------------------------------------------------------------------------
// createContainer
// ---------------------------------------------------------------------------

func TestCreateContainer_HasExpectedFields(t *testing.T) {
	c := createContainer("myservice")
	require.Len(t, c.Names, 1)
	assert.Equal(t, "/myservice", c.Names[0])
	assert.Equal(t, "myservice", c.Image)
	assert.Equal(t, "running", c.State)
	assert.NotEmpty(t, c.ID)
	assert.Len(t, c.ID, 64)
}

// ---------------------------------------------------------------------------
// getServices
// ---------------------------------------------------------------------------

func TestGetServices_ParsesComposeFile(t *testing.T) {
	cleanup := setupMockerFS(t, "/test/svc", testComposeYAML)
	defer cleanup()

	services, err := getServices("/test/svc")
	assert.NoError(t, err)
	assert.Len(t, services, 2)
	assert.Contains(t, services, "web")
	assert.Contains(t, services, "db")
}

func TestGetServices_MissingFile_ReturnsError(t *testing.T) {
	origFS := app.FS
	app.FS = afero.NewMemMapFs()
	defer func() { app.FS = origFS }()

	_, err := getServices("/no/such/path")
	assert.Error(t, err)
}
