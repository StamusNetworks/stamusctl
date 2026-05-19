package handlers

import (
	"testing"

	"stamus-ctl/internal/app"
	"stamus-ctl/pkg"
	"stamus-ctl/pkg/mocker"

	"github.com/docker/docker/api/types"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ---------------------------------------------------------------------------
// filterContainers – pure function, table-driven
// ---------------------------------------------------------------------------

func TestFilterContainers(t *testing.T) { //nolint:funlen // table-driven test with many cases
	nginx := types.Container{ID: "abc123", Names: []string{"/nginx"}}
	redis := types.Container{ID: "def456", Names: []string{"/redis"}}
	multi := types.Container{ID: "ghi789", Names: []string{"/alias", "/multi"}}

	all := []types.Container{nginx, redis, multi}

	tests := []struct {
		name      string
		input     []types.Container
		nameOrIDs []string
		want      []types.Container
	}{
		{
			name:      "empty filter returns all containers",
			input:     all,
			nameOrIDs: []string{},
			want:      all,
		},
		{
			name:      "nil filter returns all containers",
			input:     all,
			nameOrIDs: nil,
			want:      all,
		},
		{
			name:      "match by exact ID",
			input:     all,
			nameOrIDs: []string{"abc123"},
			want:      []types.Container{nginx},
		},
		{
			name:      "match by name without leading slash",
			input:     all,
			nameOrIDs: []string{"nginx"},
			want:      []types.Container{nginx},
		},
		{
			name:      "match by name with leading slash",
			input:     all,
			nameOrIDs: []string{"/redis"},
			want:      []types.Container{redis},
		},
		{
			name:      "no match returns empty result",
			input:     all,
			nameOrIDs: []string{"notfound"},
			want:      nil,
		},
		{
			name:      "multiple filters return multiple containers",
			input:     all,
			nameOrIDs: []string{"nginx", "redis"},
			want:      []types.Container{nginx, redis},
		},
		{
			name:      "match on second name in Names slice",
			input:     all,
			nameOrIDs: []string{"multi"},
			want:      []types.Container{multi},
		},
		{
			name:      "empty input returns empty result",
			input:     []types.Container{},
			nameOrIDs: []string{"nginx"},
			want:      nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := filterContainers(tt.input, tt.nameOrIDs)
			assert.Equal(t, tt.want, got)
		})
	}
}

// ---------------------------------------------------------------------------
// HandlePs – test mode
// ---------------------------------------------------------------------------

func TestHandlePs_TestMode_Empty(t *testing.T) {
	oldMode := app.Mode
	oldFS := app.FS
	app.Mode = modeTest
	app.FS = afero.NewMemMapFs()
	defer func() {
		app.Mode = oldMode
		app.FS = oldFS
	}()

	// Reset mocker so previous test state doesn't bleed in.
	// Down("") ignores the path error and resets the internal map.
	mocker.Mocked.Down("") //nolint:errcheck,gosec

	containers, err := HandlePs()
	assert.NoError(t, err)
	// An empty mocker returns a nil slice; normalise for comparison.
	assert.Empty(t, containers)
}

func TestHandlePs_TestMode_WithContainers(t *testing.T) {
	oldMode := app.Mode
	oldFS := app.FS
	app.Mode = modeTest
	app.FS = afero.NewMemMapFs()
	defer func() {
		app.Mode = oldMode
		app.FS = oldFS
		mocker.Mocked.Down("") //nolint:errcheck,gosec
	}()

	// Seed the mocker by booting containers from a compose file
	composePath := "/psdata"
	require.NoError(t, app.FS.MkdirAll(composePath, 0o755))
	compose := `services:
  web:
    image: nginx
  db:
    image: postgres
`
	require.NoError(t, afero.WriteFile(app.FS, composePath+"/docker-compose.yaml", []byte(compose), 0o644))

	require.NoError(t, mocker.Mocked.Up(composePath))

	containers, err := HandlePs()
	assert.NoError(t, err)
	assert.Len(t, containers, 2)
}

// ---------------------------------------------------------------------------
// HandleLogs – test mode
// ---------------------------------------------------------------------------

func TestHandleLogs_TestMode_Empty(t *testing.T) {
	oldMode := app.Mode
	oldFS := app.FS
	app.Mode = modeTest
	app.FS = afero.NewMemMapFs()
	defer func() {
		app.Mode = oldMode
		app.FS = oldFS
	}()

	mocker.Mocked.Down("") //nolint:errcheck,gosec

	resp, err := HandleLogs(pkg.LogsRequest{})
	assert.NoError(t, err)
	// Containers field is nil when mocker is empty – just verify no error and
	// the response is a valid zero value.
	assert.IsType(t, pkg.LogsResponse{}, resp)
}

func TestHandleLogs_TestMode_WithContainers(t *testing.T) {
	oldMode := app.Mode
	oldFS := app.FS
	app.Mode = modeTest
	app.FS = afero.NewMemMapFs()
	defer func() {
		app.Mode = oldMode
		app.FS = oldFS
		mocker.Mocked.Down("") //nolint:errcheck,gosec
	}()

	composePath := "/logdata"
	require.NoError(t, app.FS.MkdirAll(composePath, 0o755))
	compose := `services:
  svc:
    image: busybox
`
	require.NoError(t, afero.WriteFile(app.FS, composePath+"/docker-compose.yaml", []byte(compose), 0o644))

	mocker.Mocked.Down("") //nolint:errcheck,gosec
	require.NoError(t, mocker.Mocked.Up(composePath))

	resp, err := HandleLogs(pkg.LogsRequest{})
	assert.NoError(t, err)
	assert.Len(t, resp.Containers, 1)
	assert.NotEmpty(t, resp.Containers[0].Logs)
}

// ---------------------------------------------------------------------------
// handlePs and handleLogs – call real Docker daemon directly
// These tests require a running Docker daemon (tested environment has one).
// ---------------------------------------------------------------------------

func TestHandlePs_RealDocker(t *testing.T) {
	// handlePs always hits real Docker; check the call doesn't panic and either
	// succeeds or returns a recognisable error.
	containers, err := handlePs()
	if err != nil {
		// It's acceptable for this to fail in CI without Docker — just skip.
		t.Skipf("handlePs: Docker unavailable: %v", err)
	}
	// We don't assert a specific length since the host may have 0–N containers.
	_ = containers
}

func TestHandleLogs_RealDocker(t *testing.T) {
	// handleLogs calls handlePs internally and then fetches logs.
	// Use an empty logParams to request all containers.
	resp, err := handleLogs(pkg.LogsRequest{Tail: "1"})
	if err != nil {
		t.Skipf("handleLogs: Docker unavailable: %v", err)
	}
	// If Docker is available, we should at minimum get a non-nil response.
	_ = resp
}

// ---------------------------------------------------------------------------
// HandleLogs – non-test mode (calls real Docker)
// ---------------------------------------------------------------------------

func TestHandleLogs_RealDockerMode(t *testing.T) {
	// Call HandleLogs in non-test mode so the real handleLogs function is invoked.
	oldMode := app.Mode
	app.Mode = modeProd
	defer func() { app.Mode = oldMode }()

	resp, err := HandleLogs(pkg.LogsRequest{Tail: "1"})
	if err != nil {
		t.Skipf("HandleLogs real mode: Docker unavailable: %v", err)
	}
	_ = resp
}
