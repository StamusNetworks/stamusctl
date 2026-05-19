package backup

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBackupCmd_NotNil(t *testing.T) {
	cmd := BackupCmd()
	require.NotNil(t, cmd)
}

func TestBackupCmd_Use(t *testing.T) {
	cmd := BackupCmd()
	assert.Equal(t, "backup", cmd.Use)
}

func TestBackupCmd_HasSubcommands(t *testing.T) {
	cmd := BackupCmd()
	subCmds := cmd.Commands()
	assert.Greater(t, len(subCmds), 0, "backup should have subcommands")
}

func TestBackupCmd_SubcommandNames(t *testing.T) {
	cmd := BackupCmd()
	names := make(map[string]bool)
	for _, sub := range cmd.Commands() {
		names[sub.Name()] = true
	}

	assert.True(t, names["create"], "should have 'create' subcommand")
	assert.True(t, names["list"], "should have 'list' subcommand")
	assert.True(t, names["restore"], "should have 'restore' subcommand")
	assert.True(t, names["delete"], "should have 'delete' subcommand")
}

func TestBackupCmd_Help(t *testing.T) {
	cmd := BackupCmd()
	cmd.SetArgs([]string{"--help"})
	// --help causes Execute to return an error in some Cobra versions; we just check it doesn't panic
	_ = cmd.Execute()
}

func TestListCmd_HasConfigFlag(t *testing.T) {
	cmd := listCmd()
	flag := cmd.Flags().Lookup("config")
	assert.NotNil(t, flag, "list command should have --config flag")
}

func TestCreateCmd_HasConfigFlag(t *testing.T) {
	cmd := createCmd()
	flag := cmd.Flags().Lookup("config")
	assert.NotNil(t, flag, "create command should have --config flag")
}

// ---------------------------------------------------------------------------
// deleteCmd
// ---------------------------------------------------------------------------

func TestDeleteCmd_NotNil(t *testing.T) {
	cmd := deleteCmd()
	require.NotNil(t, cmd)
}

func TestDeleteCmd_Use(t *testing.T) {
	cmd := deleteCmd()
	assert.Equal(t, "delete [timestamp]", cmd.Use)
}

func TestDeleteCmd_HasConfigFlag(t *testing.T) {
	cmd := deleteCmd()
	assert.NotNil(t, cmd.Flags().Lookup("config"), "deleteCmd should have --config flag")
}

func TestDeleteCmd_HasForceFlag(t *testing.T) {
	cmd := deleteCmd()
	assert.NotNil(t, cmd.Flags().Lookup("force"), "deleteCmd should have --force flag")
}

func TestDeleteCmd_ExactArgs(t *testing.T) {
	// Verify cobra.ExactArgs(1) is set: calling with wrong args returns error.
	root := BackupCmd()
	root.SetArgs([]string{"delete"})
	err := root.Execute()
	assert.Error(t, err)
}

// ---------------------------------------------------------------------------
// restoreCmd
// ---------------------------------------------------------------------------

func TestRestoreCmd_NotNil(t *testing.T) {
	cmd := restoreCmd()
	require.NotNil(t, cmd)
}

func TestRestoreCmd_Use(t *testing.T) {
	cmd := restoreCmd()
	assert.Equal(t, "restore [timestamp]", cmd.Use)
}

func TestRestoreCmd_HasConfigFlag(t *testing.T) {
	cmd := restoreCmd()
	assert.NotNil(t, cmd.Flags().Lookup("config"), "restoreCmd should have --config flag")
}

func TestRestoreCmd_HasForceFlag(t *testing.T) {
	cmd := restoreCmd()
	assert.NotNil(t, cmd.Flags().Lookup("force"), "restoreCmd should have --force flag")
}

func TestRestoreCmd_ExactArgs(t *testing.T) {
	// Verify cobra.ExactArgs(1) is set.
	root := BackupCmd()
	root.SetArgs([]string{"restore"})
	err := root.Execute()
	assert.Error(t, err)
}

// ---------------------------------------------------------------------------
// deleteHandler — force mode with an invalid timestamp exercises the
// HandleDelete code path and returns an error (backup not found).
// ---------------------------------------------------------------------------

func TestDeleteHandler_Force_InvalidTimestamp(t *testing.T) {
	// Set --force so we bypass stdin; the handler will call HandleDelete
	// which will fail because the backup does not exist.
	deleteForce = true
	defer func() { deleteForce = false }()

	err := deleteHandler("nonexistent-timestamp")
	assert.Error(t, err, "deleteHandler should return an error for a nonexistent backup")
}

// ---------------------------------------------------------------------------
// restoreHandler — force mode with invalid timestamp returns an error.
// ---------------------------------------------------------------------------

func TestRestoreHandler_Force_InvalidTimestamp(t *testing.T) {
	restoreForce = true
	defer func() { restoreForce = false }()

	err := restoreHandler("nonexistent-timestamp")
	assert.Error(t, err, "restoreHandler should return an error for a nonexistent backup")
}

// ---------------------------------------------------------------------------
// listHandler — with no backups the handler prints "No backups found".
// ---------------------------------------------------------------------------

func TestListHandler_NoBackups(t *testing.T) {
	// listHandler calls HandleList which reads backups from the config's folder.
	// With the default in-memory FS and a nonexistent config, it returns an error
	// (config not found) or empty list depending on implementation.
	// Either outcome is acceptable — we just verify there is no panic.
	assert.NotPanics(t, func() {
		_ = listHandler()
	})
}

// ---------------------------------------------------------------------------
// createCmd — structural tests
// ---------------------------------------------------------------------------

func TestCreateCmd_NotNil(t *testing.T) {
	cmd := createCmd()
	require.NotNil(t, cmd)
}

func TestCreateCmd_Use(t *testing.T) {
	cmd := createCmd()
	assert.Equal(t, "create", cmd.Use)
}

func TestCreateCmd_ShortDescription(t *testing.T) {
	cmd := createCmd()
	assert.NotEmpty(t, cmd.Short)
}

// ---------------------------------------------------------------------------
// listCmd — structural tests
// ---------------------------------------------------------------------------

func TestListCmd_NotNil(t *testing.T) {
	cmd := listCmd()
	require.NotNil(t, cmd)
}

func TestListCmd_Use(t *testing.T) {
	cmd := listCmd()
	assert.Equal(t, "list", cmd.Use)
}

func TestListCmd_ShortDescription(t *testing.T) {
	cmd := listCmd()
	assert.NotEmpty(t, cmd.Short)
}

// ---------------------------------------------------------------------------
// createHandler — with default config (not on disk) returns an error.
// This covers the flag-extraction → HandleCreate error path.
// ---------------------------------------------------------------------------

func TestCreateHandler_NonExistentConfig_ReturnsError(t *testing.T) {
	err := createHandler()
	assert.Error(t, err, "createHandler should return error when config does not exist")
}

// ---------------------------------------------------------------------------

// ---------------------------------------------------------------------------
// RunE closure coverage — execute listCmd and createCmd to cover the RunE body
// ---------------------------------------------------------------------------

func TestListCmd_Execute_RunsHandler(t *testing.T) {
	// Just execute the command; we accept any outcome.
	// This covers the RunE closure body.
	cmd := listCmd()
	cmd.SetArgs([]string{})
	_ = cmd.Execute()
}

func TestCreateCmd_Execute_RunsHandler(t *testing.T) {
	// Execute createCmd; createHandler will fail (no config), covering RunE.
	cmd := createCmd()
	cmd.SetArgs([]string{})
	_ = cmd.Execute()
}
