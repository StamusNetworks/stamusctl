package backup

// handlers_test.go — covers createHandler, listHandler, deleteHandler,
// restoreHandler with real temp dirs.
//
// IMPORTANT: deleteForce and restoreForce are package-level vars bound
// to cobra flags. Calling deleteCmd()/restoreCmd() resets them to the
// default (false). So we must set deleteForce/restoreForce AFTER
// constructing the command.

import (
	"os"
	"testing"

	"stamus-ctl/internal/app"
	handlers "stamus-ctl/internal/handlers/backup"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// setupBackupHandlerEnv sets up a real temp dir for both configs and backups.
func setupBackupHandlerEnv(t *testing.T) (cleanup func()) {
	t.Helper()
	tmpDir, err := os.MkdirTemp("", "ctl-backup-handler-*")
	require.NoError(t, err)

	oldConfigFolder := app.ConfigFolder
	oldConfigsFolder := app.ConfigsFolder
	oldFS := app.FS

	app.ConfigFolder = tmpDir + "/"
	app.ConfigsFolder = tmpDir + "/configs/"
	app.FS = afero.NewOsFs()
	require.NoError(t, os.MkdirAll(app.ConfigsFolder, 0o755))

	return func() {
		app.ConfigFolder = oldConfigFolder
		app.ConfigsFolder = oldConfigsFolder
		app.FS = oldFS
		os.RemoveAll(tmpDir)
	}
}

// createNamedConfig creates a config dir by name under app.ConfigsFolder.
func createNamedConfig(t *testing.T, name string) {
	t.Helper()
	confDir := app.GetConfigsFolder(name)
	require.NoError(t, os.MkdirAll(confDir, 0o755))
	require.NoError(t, os.WriteFile(confDir+"/test.txt", []byte("data"), 0o644))
}

// ---------------------------------------------------------------------------
// createHandler — success path
// ---------------------------------------------------------------------------

func TestCreateHandler_Success(t *testing.T) {
	cleanup := setupBackupHandlerEnv(t)
	defer cleanup()

	createNamedConfig(t, "testconf")

	cmd := createCmd()
	require.NoError(t, cmd.Flags().Set("config", "testconf"))

	err := createHandler()
	assert.NoError(t, err)
}

func TestCreateHandler_Error(t *testing.T) {
	cleanup := setupBackupHandlerEnv(t)
	defer cleanup()

	cmd := createCmd()
	require.NoError(t, cmd.Flags().Set("config", "nonexistentconf"))

	err := createHandler()
	assert.Error(t, err)
}

// ---------------------------------------------------------------------------
// listHandler
// ---------------------------------------------------------------------------

func TestListHandler_WithBackups(t *testing.T) {
	cleanup := setupBackupHandlerEnv(t)
	defer cleanup()

	createNamedConfig(t, "myconf")

	_, err := handlers.HandleCreate(handlers.CreateHandlerInputs{Config: "myconf"})
	require.NoError(t, err)

	cmd := listCmd()
	require.NoError(t, cmd.Flags().Set("config", "myconf"))

	err = listHandler()
	assert.NoError(t, err)
}

func TestListHandler_EmptyConfig(t *testing.T) {
	cleanup := setupBackupHandlerEnv(t)
	defer cleanup()

	createNamedConfig(t, "emptyconf")

	cmd := listCmd()
	require.NoError(t, cmd.Flags().Set("config", "emptyconf"))

	err := listHandler()
	assert.NoError(t, err)
}

// ---------------------------------------------------------------------------
// deleteHandler — force mode
// IMPORTANT: Set deleteForce AFTER constructing deleteCmd(), because
// cmd.Flags().BoolVarP(&deleteForce, ..., false, ...) resets it to false.
// ---------------------------------------------------------------------------

func TestDeleteHandler_Force_Success(t *testing.T) {
	cleanup := setupBackupHandlerEnv(t)
	defer cleanup()

	createNamedConfig(t, "delconf")

	_, err := handlers.HandleCreate(handlers.CreateHandlerInputs{Config: "delconf"})
	require.NoError(t, err)

	backupList, err := handlers.HandleList(handlers.ListHandlerInputs{Config: "delconf"})
	require.NoError(t, err)
	require.Len(t, backupList, 1)

	// Construct cmd first (resets deleteForce), then set force mode
	cmd := deleteCmd()
	require.NoError(t, cmd.Flags().Set("config", "delconf"))
	deleteForce = true
	defer func() { deleteForce = false }()

	err = deleteHandler(backupList[0].Timestamp)
	assert.NoError(t, err)
}

// ---------------------------------------------------------------------------
// restoreHandler — force mode
// ---------------------------------------------------------------------------

func TestRestoreHandler_Force_Success(t *testing.T) {
	cleanup := setupBackupHandlerEnv(t)
	defer cleanup()

	createNamedConfig(t, "restoreconf")

	_, err := handlers.HandleCreate(handlers.CreateHandlerInputs{Config: "restoreconf"})
	require.NoError(t, err)

	confDir := app.GetConfigsFolder("restoreconf")
	require.NoError(t, os.WriteFile(confDir+"/test.txt", []byte("modified"), 0o644))

	backupList, err := handlers.HandleList(handlers.ListHandlerInputs{Config: "restoreconf"})
	require.NoError(t, err)
	require.Len(t, backupList, 1)

	// Construct cmd first (resets restoreForce), then set force mode
	cmd := restoreCmd()
	require.NoError(t, cmd.Flags().Set("config", "restoreconf"))
	restoreForce = true
	defer func() { restoreForce = false }()

	err = restoreHandler(backupList[0].Timestamp)
	assert.NoError(t, err)
}

// ---------------------------------------------------------------------------
// deleteHandler / restoreHandler — interactive (non-force) paths via stdin pipe
// ---------------------------------------------------------------------------

// withStdinPipe replaces os.Stdin with a pipe containing the given input,
// calls f, then restores os.Stdin. The write side of the pipe is closed
// before calling f so bufio.ReadString returns EOF (which the handler treats
// as an error) — OR the provided content is written and then closed.
func withStdinContent(t *testing.T, content string, f func()) {
	t.Helper()
	r, w, err := os.Pipe()
	require.NoError(t, err)

	origStdin := os.Stdin
	os.Stdin = r
	defer func() {
		os.Stdin = origStdin
		r.Close()
	}()

	// Write content then close write end so the reader gets EOF.
	_, err = w.WriteString(content)
	require.NoError(t, err)
	w.Close()

	f()
}

func TestDeleteHandler_NoForce_NoResponse(t *testing.T) {
	cleanup := setupBackupHandlerEnv(t)
	defer cleanup()

	createNamedConfig(t, "delconf2")

	_, err := handlers.HandleCreate(handlers.CreateHandlerInputs{Config: "delconf2"})
	require.NoError(t, err)

	backupList, err := handlers.HandleList(handlers.ListHandlerInputs{Config: "delconf2"})
	require.NoError(t, err)
	require.Len(t, backupList, 1)

	cmd := deleteCmd()
	require.NoError(t, cmd.Flags().Set("config", "delconf2"))
	// deleteForce is already false from cmd construction.

	var handlerErr error
	withStdinContent(t, "no\n", func() {
		handlerErr = deleteHandler(backupList[0].Timestamp)
	})
	// "no" response → "Delete cancelled" → nil error
	assert.NoError(t, handlerErr)
}

func TestDeleteHandler_NoForce_YesResponse(t *testing.T) {
	cleanup := setupBackupHandlerEnv(t)
	defer cleanup()

	createNamedConfig(t, "delconf3")

	_, err := handlers.HandleCreate(handlers.CreateHandlerInputs{Config: "delconf3"})
	require.NoError(t, err)

	backupList, err := handlers.HandleList(handlers.ListHandlerInputs{Config: "delconf3"})
	require.NoError(t, err)
	require.Len(t, backupList, 1)

	cmd := deleteCmd()
	require.NoError(t, cmd.Flags().Set("config", "delconf3"))

	var handlerErr error
	withStdinContent(t, "yes\n", func() {
		handlerErr = deleteHandler(backupList[0].Timestamp)
	})
	// "yes" response → calls HandleDelete → success
	assert.NoError(t, handlerErr)
}

func TestRestoreHandler_NoForce_NoResponse(t *testing.T) {
	cleanup := setupBackupHandlerEnv(t)
	defer cleanup()

	createNamedConfig(t, "restoreconf2")

	_, err := handlers.HandleCreate(handlers.CreateHandlerInputs{Config: "restoreconf2"})
	require.NoError(t, err)

	backupList, err := handlers.HandleList(handlers.ListHandlerInputs{Config: "restoreconf2"})
	require.NoError(t, err)
	require.Len(t, backupList, 1)

	cmd := restoreCmd()
	require.NoError(t, cmd.Flags().Set("config", "restoreconf2"))
	// restoreForce is already false from cmd construction.

	var handlerErr error
	withStdinContent(t, "no\n", func() {
		handlerErr = restoreHandler(backupList[0].Timestamp)
	})
	// "no" response → "Restore cancelled" → nil error
	assert.NoError(t, handlerErr)
}

func TestRestoreHandler_NoForce_YesResponse(t *testing.T) {
	cleanup := setupBackupHandlerEnv(t)
	defer cleanup()

	createNamedConfig(t, "restoreconf3")

	_, err := handlers.HandleCreate(handlers.CreateHandlerInputs{Config: "restoreconf3"})
	require.NoError(t, err)

	confDir := app.GetConfigsFolder("restoreconf3")
	require.NoError(t, os.WriteFile(confDir+"/test.txt", []byte("modified"), 0o644))

	backupList, err := handlers.HandleList(handlers.ListHandlerInputs{Config: "restoreconf3"})
	require.NoError(t, err)
	require.Len(t, backupList, 1)

	cmd := restoreCmd()
	require.NoError(t, cmd.Flags().Set("config", "restoreconf3"))

	var handlerErr error
	withStdinContent(t, "yes\n", func() {
		handlerErr = restoreHandler(backupList[0].Timestamp)
	})
	// "yes" → calls HandleRestore → success
	assert.NoError(t, handlerErr)
}
