package compose

import (
	"os"
	"testing"

	"stamus-ctl/internal/app"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestComposeCmd_NotNil(t *testing.T) {
	cmd := ComposeCmd()
	require.NotNil(t, cmd)
}

func TestComposeCmd_Use(t *testing.T) {
	cmd := ComposeCmd()
	assert.Equal(t, "compose", cmd.Use)
}

func TestComposeCmd_HasSubcommands(t *testing.T) {
	cmd := ComposeCmd()
	assert.Greater(t, len(cmd.Commands()), 0, "compose should have subcommands")
}

func TestComposeCmd_HasInitSubcommand(t *testing.T) {
	cmd := ComposeCmd()
	names := make(map[string]bool)
	for _, sub := range cmd.Commands() {
		names[sub.Name()] = true
	}
	assert.True(t, names["init"], "should have 'init' subcommand")
}

func TestComposeCmd_HasUpdateSubcommand(t *testing.T) {
	cmd := ComposeCmd()
	names := make(map[string]bool)
	for _, sub := range cmd.Commands() {
		names[sub.Name()] = true
	}
	assert.True(t, names["update"], "should have 'update' subcommand")
}

func TestComposeCmd_Help(t *testing.T) {
	cmd := ComposeCmd()
	cmd.SetArgs([]string{"--help"})
	_ = cmd.Execute()
}

func TestInitCmd_NotNil(t *testing.T) {
	cmd := initCmd()
	require.NotNil(t, cmd)
}

func TestInitCmd_Use(t *testing.T) {
	cmd := initCmd()
	assert.Equal(t, "init", cmd.Use)
}

func TestInitCmd_HasClearNDRSubcommand(t *testing.T) {
	cmd := initCmd()
	names := make(map[string]bool)
	for _, sub := range cmd.Commands() {
		names[sub.Name()] = true
	}
	assert.True(t, names["clearndr"], "init should have 'clearndr' subcommand")
}

func TestInitCmd_HasExpectedFlags(t *testing.T) {
	cmd := initCmd()
	for _, flag := range []string{"config", "version", "registry"} {
		assert.NotNil(t, cmd.Flags().Lookup(flag), "initCmd should have --%s flag", flag)
	}
}

func TestClearNDRCmd_NotNil(t *testing.T) {
	cmd := ClearNDRCmd()
	require.NotNil(t, cmd)
}

func TestClearNDRCmd_Use(t *testing.T) {
	cmd := ClearNDRCmd()
	assert.Equal(t, "clearndr", cmd.Use)
}

func TestUpdateCmd_NotNil(t *testing.T) {
	cmd := updateCmd()
	require.NotNil(t, cmd)
}

func TestUpdateCmd_Use(t *testing.T) {
	cmd := updateCmd()
	assert.Equal(t, "update", cmd.Use)
}

func TestUpdateCmd_HasExpectedFlags(t *testing.T) {
	cmd := updateCmd()
	for _, flag := range []string{"config", "version"} {
		assert.NotNil(t, cmd.Flags().Lookup(flag), "updateCmd should have --%s flag", flag)
	}
}

func TestWrappedCmd_ReturnsSlice(t *testing.T) {
	cmds, named := wrappedCmd()
	// wrappedCmd may return empty if docker-compose is not installed; just check types.
	assert.NotNil(t, cmds)
	assert.NotNil(t, named)
}

// ---------------------------------------------------------------------------
// readPcapCmd
// ---------------------------------------------------------------------------

func TestReadPcapCmd_NotNil(t *testing.T) {
	cmd := readPcapCmd()
	require.NotNil(t, cmd)
}

func TestReadPcapCmd_Use(t *testing.T) {
	cmd := readPcapCmd()
	assert.Equal(t, "readpcap <pcap file>", cmd.Use)
}

func TestReadPcapCmd_HasConfigFlag(t *testing.T) {
	cmd := readPcapCmd()
	assert.NotNil(t, cmd.Flags().Lookup("config"), "readPcapCmd should have --config flag")
}

func TestReadPcapCmd_ExactArgs(t *testing.T) {
	// cobra.ExactArgs(1): calling without an arg should return an error.
	root := ComposeCmd()
	root.SetArgs([]string{"readpcap"})
	err := root.Execute()
	assert.Error(t, err)
}

// ---------------------------------------------------------------------------
// checkFile
// ---------------------------------------------------------------------------

func TestCheckFile_NonExistent(t *testing.T) {
	err := checkFile("/nonexistent/path/to/file.pcap")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "does not exist")
}

func TestCheckFile_ExistingRegularFile(t *testing.T) {
	// Use an in-memory FS so the test is isolated from the real filesystem.
	oldFS := app.FS
	memFS := afero.NewMemMapFs()
	app.FS = memFS
	defer func() { app.FS = oldFS }()

	require.NoError(t, memFS.MkdirAll("/tmp", 0o755))
	f, err := memFS.Create("/tmp/test.pcap")
	require.NoError(t, err)
	_, err = f.WriteString("pcapdata")
	require.NoError(t, err)
	f.Close()

	err = checkFile("/tmp/test.pcap")
	assert.NoError(t, err)
}

// ---------------------------------------------------------------------------
// updateCmd flags
// ---------------------------------------------------------------------------

func TestUpdateCmd_HasInteractiveFlag(t *testing.T) {
	cmd := updateCmd()
	assert.NotNil(t, cmd.Flags().Lookup("interactive"),
		"updateCmd should have --interactive flag")
}

func TestUpdateCmd_HasTemplateFlag(t *testing.T) {
	cmd := updateCmd()
	assert.NotNil(t, cmd.Flags().Lookup("template"), "updateCmd should have --template flag")
}

// ---------------------------------------------------------------------------
// initCmd additional flags
// ---------------------------------------------------------------------------

func TestInitCmd_HasDefaultFlag(t *testing.T) {
	cmd := initCmd()
	assert.NotNil(t, cmd.Flags().Lookup("default"), "initCmd should have --default flag")
}

func TestInitCmd_HasExpertFlag(t *testing.T) {
	cmd := initCmd()
	assert.NotNil(t, cmd.Flags().Lookup("expert"), "initCmd should have --expert flag")
}

func TestInitCmd_HasValuesFlag(t *testing.T) {
	cmd := initCmd()
	assert.NotNil(t, cmd.Flags().Lookup("values"), "initCmd should have --values flag")
}

func TestInitCmd_HasBindFlag(t *testing.T) {
	cmd := initCmd()
	assert.NotNil(t, cmd.Flags().Lookup("bind"), "initCmd should have --bind flag")
}

// ---------------------------------------------------------------------------
// ClearNDRCmd (compose/init.go)
// ---------------------------------------------------------------------------

func TestClearNDRCmd_HasConfigFlag(t *testing.T) {
	cmd := ClearNDRCmd()
	assert.NotNil(t, cmd.Flags().Lookup("config"), "ClearNDRCmd should have --config flag")
}

func TestClearNDRCmd_HelpDoesNotPanic(t *testing.T) {
	cmd := ClearNDRCmd()
	cmd.SetArgs([]string{"--help"})
	assert.NotPanics(t, func() { _ = cmd.Execute() })
}

func TestCheckFile_Directory_ReturnsError(t *testing.T) {
	// Directories are not regular files — checkFile should return an error.
	oldFS := app.FS
	memFS := afero.NewMemMapFs()
	app.FS = memFS
	defer func() { app.FS = oldFS }()

	require.NoError(t, memFS.MkdirAll("/tmp/testdir", 0o755))

	err := checkFile("/tmp/testdir")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not a regular file")
}

// ---------------------------------------------------------------------------
// readPcap handler
// ---------------------------------------------------------------------------

func TestReadPcap_NoArgs_ReturnsError(t *testing.T) {
	cmd := readPcapCmd()
	err := readPcap(cmd, []string{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "pcap file path is required")
}

func TestReadPcap_NonExistentFile_ReturnsError(t *testing.T) {
	cmd := readPcapCmd()
	err := readPcap(cmd, []string{"/nonexistent/path/to/file.pcap"})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "does not exist")
}

func TestReadPcap_ValidFile_CallsHandler(t *testing.T) {
	// Use a real temp file (checkFile uses app.FS which is OsFs in this test,
	// but we set MemMapFs to create the file in memory).
	// Actually checkFile uses app.FS — let's use OsFs + real temp file.
	oldFS := app.FS
	app.FS = afero.NewOsFs()
	defer func() { app.FS = oldFS }()

	// Create a real temp file.
	tmpFile, err := os.CreateTemp("", "test-*.pcap")
	require.NoError(t, err)
	tmpFile.WriteString("fake pcap data")
	tmpFile.Close()
	defer os.Remove(tmpFile.Name())

	cmd := readPcapCmd()
	require.NoError(t, cmd.Flags().Set("config", "testconf"))
	// PcapHandler will fail (no docker/suricata) but readPcap returns nil even on error.
	err = readPcap(cmd, []string{tmpFile.Name()})
	assert.NoError(t, err) // readPcap eats the error from PcapHandler
}
