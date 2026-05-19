package completion

import (
	"bytes"
	"io"
	"os"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// captureStdout redirects os.Stdout during fn and returns the captured output.
func captureStdout(fn func()) string {
	r, w, err := os.Pipe()
	if err != nil {
		panic(err)
	}
	old := os.Stdout
	os.Stdout = w

	fn()

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	io.Copy(&buf, r)
	return buf.String()
}

// ---------------------------------------------------------------------------
// CompletionCmd structure
// ---------------------------------------------------------------------------

func TestCompletionCmd_NotNil(t *testing.T) {
	cmd := CompletionCmd()
	require.NotNil(t, cmd)
}

func TestCompletionCmd_Use(t *testing.T) {
	cmd := CompletionCmd()
	assert.Equal(t, "completion [bash|zsh|fish|powershell]", cmd.Use)
}

func TestCompletionCmd_ShortDescription(t *testing.T) {
	cmd := CompletionCmd()
	assert.NotEmpty(t, cmd.Short)
}

func TestCompletionCmd_LongDescription(t *testing.T) {
	cmd := CompletionCmd()
	assert.NotEmpty(t, cmd.Long)
}

func TestCompletionCmd_DisableFlagsInUseLine(t *testing.T) {
	cmd := CompletionCmd()
	assert.True(t, cmd.DisableFlagsInUseLine)
}

func TestCompletionCmd_ValidArgs(t *testing.T) {
	cmd := CompletionCmd()
	validArgs := cmd.ValidArgs
	assert.Contains(t, validArgs, "bash")
	assert.Contains(t, validArgs, "zsh")
	assert.Contains(t, validArgs, "fish")
	assert.Contains(t, validArgs, "powershell")
}

func TestCompletionCmd_HelpDoesNotPanic(t *testing.T) {
	root := &cobra.Command{Use: "stamusctl"}
	root.AddCommand(CompletionCmd())
	root.SetArgs([]string{"completion", "--help"})
	assert.NotPanics(t, func() { _ = root.Execute() })
}

// makeRoot creates a root command with the completion subcommand attached
// so that cmd.Root() is set and GenXxxCompletion works correctly.
func makeRoot() *cobra.Command {
	root := &cobra.Command{Use: "stamusctl"}
	root.AddCommand(CompletionCmd())
	return root
}

// ---------------------------------------------------------------------------
// Shell completion generation — the generators write to os.Stdout.
// ---------------------------------------------------------------------------

func TestCompletionCmd_Bash(t *testing.T) {
	root := makeRoot()
	root.SetArgs([]string{"completion", "bash"})
	var err error
	out := captureStdout(func() {
		err = root.Execute()
	})
	require.NoError(t, err)
	assert.NotEmpty(t, out)
}

func TestCompletionCmd_Zsh(t *testing.T) {
	root := makeRoot()
	root.SetArgs([]string{"completion", "zsh"})
	var err error
	out := captureStdout(func() {
		err = root.Execute()
	})
	require.NoError(t, err)
	assert.NotEmpty(t, out)
}

func TestCompletionCmd_Fish(t *testing.T) {
	root := makeRoot()
	root.SetArgs([]string{"completion", "fish"})
	var err error
	out := captureStdout(func() {
		err = root.Execute()
	})
	require.NoError(t, err)
	assert.NotEmpty(t, out)
}

func TestCompletionCmd_PowerShell(t *testing.T) {
	root := makeRoot()
	root.SetArgs([]string{"completion", "powershell"})
	var err error
	out := captureStdout(func() {
		err = root.Execute()
	})
	require.NoError(t, err)
	assert.NotEmpty(t, out)
}

func TestCompletionCmd_InvalidShell(t *testing.T) {
	root := makeRoot()
	root.SetArgs([]string{"completion", "invalid-shell"})
	err := root.Execute()
	assert.Error(t, err, "invalid shell argument should return an error")
}

func TestCompletionCmd_NoArgs(t *testing.T) {
	root := makeRoot()
	root.SetArgs([]string{"completion"})
	err := root.Execute()
	assert.Error(t, err, "no argument should return an error")
}
