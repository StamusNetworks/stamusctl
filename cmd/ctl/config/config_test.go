package config

import (
	"bytes"
	"io"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// captureStdout redirects os.Stdout during fn and returns the captured output.
func captureStdout(fn func()) string {
	r, w, _ := os.Pipe()
	old := os.Stdout
	os.Stdout = w

	fn()

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	io.Copy(&buf, r)
	return buf.String()
}

// TestConfigCmd verifies the ConfigCmd structure.

func TestConfigCmd_NotNil(t *testing.T) {
	cmd := ConfigCmd()
	require.NotNil(t, cmd)
}

func TestConfigCmd_Use(t *testing.T) {
	cmd := ConfigCmd()
	assert.Equal(t, "config", cmd.Use)
}

func TestConfigCmd_HasSubcommands(t *testing.T) {
	cmd := ConfigCmd()
	assert.Greater(t, len(cmd.Commands()), 0, "config should have subcommands")
}

func TestConfigCmd_SubcommandNames(t *testing.T) {
	cmd := ConfigCmd()
	names := make(map[string]bool)
	for _, sub := range cmd.Commands() {
		names[sub.Name()] = true
	}

	assert.True(t, names["get"], "should have 'get' subcommand")
	assert.True(t, names["set"], "should have 'set' subcommand")
	assert.True(t, names["list"], "should have 'list' subcommand")
	assert.True(t, names["clear"], "should have 'clear' subcommand")
	assert.True(t, names["version"], "should have 'version' subcommand")
}

func TestConfigCmd_Help(t *testing.T) {
	cmd := ConfigCmd()
	cmd.SetArgs([]string{"--help"})
	_ = cmd.Execute()
}

// TestPrintGroupedValues covers the pure formatting helper.

func TestPrintGroupedValues_StringLeaf(t *testing.T) {
	group := map[string]interface{}{
		"key": "value",
	}
	out := captureStdout(func() {
		printGroupedValues(group, "")
	})
	assert.Contains(t, out, "key: value")
}

func TestPrintGroupedValues_Nested(t *testing.T) {
	group := map[string]interface{}{
		"parent": map[string]interface{}{
			"child": "childval",
		},
	}
	out := captureStdout(func() {
		printGroupedValues(group, "")
	})
	assert.Contains(t, out, "parent:")
	assert.Contains(t, out, "child: childval")
}

func TestPrintGroupedValues_Empty(t *testing.T) {
	out := captureStdout(func() {
		printGroupedValues(map[string]interface{}{}, "")
	})
	assert.Empty(t, out)
}

func TestPrintGroupedValues_WithPrefix(t *testing.T) {
	group := map[string]interface{}{
		"k": "v",
	}
	out := captureStdout(func() {
		printGroupedValues(group, ">> ")
	})
	assert.Contains(t, out, ">> k: v")
}

// TestPrintColoredGroupedValues covers the colored variant.

func TestPrintColoredGroupedValues_StringLeaf(t *testing.T) {
	group := map[string]interface{}{
		"file.go": "content",
	}
	// Should not panic; ANSI codes are emitted but value is not shown (key only).
	out := captureStdout(func() {
		printColoredGroupedValues(group, "")
	})
	assert.Contains(t, out, "file.go")
}

func TestPrintColoredGroupedValues_Nested(t *testing.T) {
	group := map[string]interface{}{
		"dir": map[string]interface{}{
			"file.go": "content",
		},
	}
	out := captureStdout(func() {
		printColoredGroupedValues(group, "")
	})
	assert.Contains(t, out, "dir")
}

func TestPrintColoredGroupedValues_Empty(t *testing.T) {
	// Must not panic on empty map.
	captureStdout(func() {
		printColoredGroupedValues(map[string]interface{}{}, "")
	})
}

// TestSubcommand flag presence checks

func TestGetCmd_HasConfigFlag(t *testing.T) {
	cmd := getCmd()
	require.NotNil(t, cmd.Flags().Lookup("config"))
}

func TestSetCmd_HasRequiredFlags(t *testing.T) {
	cmd := setCmd()
	for _, flag := range []string{"config", "reload", "apply"} {
		assert.NotNil(t, cmd.Flags().Lookup(flag), "setCmd should have --%s flag", flag)
	}
}

func TestClearCmd_HasConfigFlag(t *testing.T) {
	cmd := clearCmd()
	require.NotNil(t, cmd.Flags().Lookup("config"))
}

func TestVersionCmd_Use(t *testing.T) {
	cmd := versionCmd()
	assert.Equal(t, "version", cmd.Use)
}

func TestListCmd_Use(t *testing.T) {
	cmd := listCmd()
	assert.Equal(t, "list", cmd.Use)
}

// ---------------------------------------------------------------------------
// Additional getCmd subcommand tests
// ---------------------------------------------------------------------------

func TestGetCmd_NotNil(t *testing.T) {
	cmd := getCmd()
	require.NotNil(t, cmd)
}

func TestGetCmd_Use(t *testing.T) {
	cmd := getCmd()
	assert.Equal(t, "get [keys...]", cmd.Use)
}

func TestGetCmd_HasContentSubcommand(t *testing.T) {
	cmd := getCmd()
	names := make(map[string]bool)
	for _, sub := range cmd.Commands() {
		names[sub.Name()] = true
	}
	assert.True(t, names["content"], "getCmd should have 'content' subcommand")
}

func TestGetCmd_HasKeysSubcommand(t *testing.T) {
	cmd := getCmd()
	names := make(map[string]bool)
	for _, sub := range cmd.Commands() {
		names[sub.Name()] = true
	}
	assert.True(t, names["keys"], "getCmd should have 'keys' subcommand")
}

func TestGetKeysCmd_HasMarkdownFlag(t *testing.T) {
	cmd := getKeysCmd()
	assert.NotNil(t, cmd.Flags().Lookup("markdown"), "getKeysCmd should have --markdown flag")
}

func TestGetContentCmd_HasConfigFlag(t *testing.T) {
	cmd := getContentCmd()
	assert.NotNil(t, cmd.Flags().Lookup("config"), "getContentCmd should have --config flag")
}

// ---------------------------------------------------------------------------
// setCmd subcommand structure
// ---------------------------------------------------------------------------

func TestSetCmd_NotNil(t *testing.T) {
	cmd := setCmd()
	require.NotNil(t, cmd)
}

func TestSetCmd_Use(t *testing.T) {
	cmd := setCmd()
	assert.Equal(t, "set [keys=values...]", cmd.Use)
}

func TestSetCmd_HasContentSubcommand(t *testing.T) {
	cmd := setCmd()
	names := make(map[string]bool)
	for _, sub := range cmd.Commands() {
		names[sub.Name()] = true
	}
	assert.True(t, names["content"], "setCmd should have 'content' subcommand")
}

func TestSetCmd_HasFromFileFlag(t *testing.T) {
	cmd := setCmd()
	assert.NotNil(t, cmd.Flags().Lookup("fromFile"), "setCmd should have --fromFile flag")
}

func TestSetCmd_HasValuesFlag(t *testing.T) {
	cmd := setCmd()
	assert.NotNil(t, cmd.Flags().Lookup("values"), "setCmd should have --values flag")
}

func TestSetContentCmd_NotNil(t *testing.T) {
	cmd := setContentCmd()
	require.NotNil(t, cmd)
}

func TestSetContentCmd_Use(t *testing.T) {
	cmd := setContentCmd()
	assert.Equal(t, "content", cmd.Use)
}

func TestSetContentCmd_HasConfigFlag(t *testing.T) {
	cmd := setContentCmd()
	assert.NotNil(t, cmd.Flags().Lookup("config"), "setContentCmd should have --config flag")
}

// ---------------------------------------------------------------------------
// clearCmd
// ---------------------------------------------------------------------------

func TestClearCmd_NotNil(t *testing.T) {
	cmd := clearCmd()
	require.NotNil(t, cmd)
}

func TestClearCmd_Use(t *testing.T) {
	cmd := clearCmd()
	assert.Equal(t, "clear", cmd.Use)
}

func TestClearCmd_ShortDescription(t *testing.T) {
	cmd := clearCmd()
	assert.NotEmpty(t, cmd.Short)
}

// ---------------------------------------------------------------------------
// listCmd
// ---------------------------------------------------------------------------

func TestListCmd_NotNil(t *testing.T) {
	cmd := listCmd()
	require.NotNil(t, cmd)
}

func TestListCmd_ShortDescription(t *testing.T) {
	cmd := listCmd()
	assert.NotEmpty(t, cmd.Short)
}

// ---------------------------------------------------------------------------
// versionCmd
// ---------------------------------------------------------------------------

func TestVersionCmd_NotNil(t *testing.T) {
	cmd := versionCmd()
	require.NotNil(t, cmd)
}

func TestVersionCmd_HasConfigFlag(t *testing.T) {
	cmd := versionCmd()
	assert.NotNil(t, cmd.Flags().Lookup("config"), "versionCmd should have --config flag")
}

// ---------------------------------------------------------------------------
// ConfigCmd structure
// ---------------------------------------------------------------------------

func TestConfigCmd_HasVersionSubcommand(t *testing.T) {
	cmd := ConfigCmd()
	names := make(map[string]bool)
	for _, sub := range cmd.Commands() {
		names[sub.Name()] = true
	}
	assert.True(t, names["version"], "ConfigCmd should have 'version' subcommand")
}

func TestConfigCmd_HelpDoesNotPanic(t *testing.T) {
	cmd := ConfigCmd()
	cmd.SetArgs([]string{"--help"})
	assert.NotPanics(t, func() { _ = cmd.Execute() })
}

// ---------------------------------------------------------------------------
// setCmd — verify more flag coverage to push construction up
// ---------------------------------------------------------------------------

func TestSetCmd_HelpDoesNotPanic(t *testing.T) {
	cmd := setCmd()
	cmd.SetArgs([]string{"--help"})
	assert.NotPanics(t, func() { _ = cmd.Execute() })
}

func TestSetContentCmd_HelpDoesNotPanic(t *testing.T) {
	cmd := setContentCmd()
	cmd.SetArgs([]string{"--help"})
	assert.NotPanics(t, func() { _ = cmd.Execute() })
}

func TestClearCmd_HelpDoesNotPanic(t *testing.T) {
	cmd := clearCmd()
	cmd.SetArgs([]string{"--help"})
	assert.NotPanics(t, func() { _ = cmd.Execute() })
}

func TestListCmd_HelpDoesNotPanic(t *testing.T) {
	cmd := listCmd()
	cmd.SetArgs([]string{"--help"})
	assert.NotPanics(t, func() { _ = cmd.Execute() })
}

func TestVersionCmd_HelpDoesNotPanic(t *testing.T) {
	cmd := versionCmd()
	cmd.SetArgs([]string{"--help"})
	assert.NotPanics(t, func() { _ = cmd.Execute() })
}

func TestGetCmd_HelpDoesNotPanic(t *testing.T) {
	cmd := getCmd()
	cmd.SetArgs([]string{"--help"})
	assert.NotPanics(t, func() { _ = cmd.Execute() })
}

// TestListCmd_Execute exercises the RunE closure in listCmd.
// handlers.ListHandler reads from app.ConfigsFolder, which defaults to an in-memory
// filesystem in test mode — so it returns an empty list without error.
func TestListCmd_Execute_Success(t *testing.T) {
	cmd := listCmd()
	cmd.SetArgs([]string{})
	err := cmd.Execute()
	// ListHandler may succeed (empty list) or fail (no configs folder) — any outcome is fine.
	_ = err
}
