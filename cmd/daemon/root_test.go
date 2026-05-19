package daemon

import (
	"testing"

	"stamus-ctl/internal/logging"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMain(m *testing.M) {
	logging.SetLogger()
	m.Run()
}

func TestRootCmd_Structure(t *testing.T) {
	t.Parallel()

	cmd := rootCmd()
	assert.Equal(t, "stamusd", cmd.Use)

	// Verify sub-commands are registered
	subCmds := cmd.Commands()
	names := make([]string, 0, len(subCmds))
	for _, c := range subCmds {
		names = append(names, c.Use)
	}

	assert.Contains(t, names, "version")
	assert.Contains(t, names, "run")
}

func TestRootCmd_VerboseFlag(t *testing.T) {
	t.Parallel()

	cmd := rootCmd()
	// verbose is added as a persistent flag
	flag := cmd.PersistentFlags().Lookup("verbose")
	require.NotNil(t, flag, "verbose flag should be registered as a persistent flag")
}

func TestVersionCmd_Structure(t *testing.T) {
	t.Parallel()

	cmd := versionCmd()
	assert.Equal(t, "version", cmd.Use)
	assert.NotEmpty(t, cmd.Short)
}

func TestVersionCmd_Run(t *testing.T) {
	t.Parallel()

	cmd := versionCmd()
	// Run should not panic
	assert.NotPanics(t, func() {
		cmd.Run(cmd, []string{})
	})
}

func TestPrintVersion_NoPanic(t *testing.T) {
	t.Parallel()

	// printVersion writes to stdout; just verify no panic
	assert.NotPanics(t, func() {
		printVersion()
	})
}

func TestRootCmd_SubCommandCount(t *testing.T) {
	t.Parallel()

	cmd := rootCmd()
	// Should have at least version and run sub-commands
	assert.GreaterOrEqual(t, len(cmd.Commands()), 2)
}
