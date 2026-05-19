package config

import (
	"testing"

	"stamus-ctl/internal/app"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ---------------------------------------------------------------------------
// TemplateCmd
// ---------------------------------------------------------------------------

func TestTemplateCmd_NotNil(t *testing.T) {
	cmd := TemplateCmd()
	require.NotNil(t, cmd)
}

func TestTemplateCmd_Use(t *testing.T) {
	cmd := TemplateCmd()
	assert.Equal(t, "template", cmd.Use)
}

func TestTemplateCmd_ShortDescription(t *testing.T) {
	cmd := TemplateCmd()
	assert.NotEmpty(t, cmd.Short)
}

func TestTemplateCmd_HasSubcommands(t *testing.T) {
	cmd := TemplateCmd()
	assert.Greater(t, len(cmd.Commands()), 0, "template should have subcommands")
}

func TestTemplateCmd_HasKeysSubcommand(t *testing.T) {
	cmd := TemplateCmd()
	names := make(map[string]bool)
	for _, sub := range cmd.Commands() {
		names[sub.Name()] = true
	}
	assert.True(t, names["keys"], "template should have 'keys' subcommand")
}

func TestTemplateCmd_HelpDoesNotPanic(t *testing.T) {
	cmd := TemplateCmd()
	cmd.SetArgs([]string{"--help"})
	assert.NotPanics(t, func() { _ = cmd.Execute() })
}

// ---------------------------------------------------------------------------
// keysCmd
// ---------------------------------------------------------------------------

func TestKeysCmd_NotNil(t *testing.T) {
	cmd := keysCmd()
	require.NotNil(t, cmd)
}

func TestKeysCmd_Use(t *testing.T) {
	cmd := keysCmd()
	assert.Equal(t, "keys", cmd.Use)
}

func TestKeysCmd_ShortDescription(t *testing.T) {
	cmd := keysCmd()
	assert.NotEmpty(t, cmd.Short)
}

func TestKeysCmd_HasTemplateFlag(t *testing.T) {
	cmd := keysCmd()
	assert.NotNil(t, cmd.Flags().Lookup("template"), "keysCmd should have --template flag")
}

func TestKeysCmd_HasMarkdownFlag(t *testing.T) {
	cmd := keysCmd()
	assert.NotNil(t, cmd.Flags().Lookup("markdown"), "keysCmd should have --markdown flag")
}

// ---------------------------------------------------------------------------
// keysHandler — requires a valid template folder; without one it returns an
// error ("template folder is required"), which is the expected path to cover.
// ---------------------------------------------------------------------------

func TestKeysHandler_NoTemplateFlag_ReturnsError(t *testing.T) {
	err := keysHandler()
	assert.Error(t, err, "keysHandler with no template flag should return an error")
	assert.Contains(t, err.Error(), "template folder is required")
}

// ---------------------------------------------------------------------------
// keysHandler — with a valid template folder.
// Uses afero.MemMapFs (which is what app.FS uses under tests) to write files.
// ---------------------------------------------------------------------------

const configYAMLFixture = `param1:
  type: string
  usage: A parameter
  default: hello
`

func TestKeysHandler_WithValidTemplate(t *testing.T) {
	origFS := app.FS
	memFS := afero.NewMemMapFs()
	app.FS = memFS
	defer func() { app.FS = origFS }()

	const templatePath = "/test-template-dir"
	require.NoError(t, memFS.MkdirAll(templatePath, 0o755))
	require.NoError(t, afero.WriteFile(memFS, templatePath+"/config.yaml", []byte(configYAMLFixture), 0o644))

	cmd := keysCmd()
	require.NoError(t, cmd.Flags().Set("template", templatePath))
	require.NoError(t, cmd.Flags().Set("markdown", "false"))

	err := keysHandler()
	assert.NoError(t, err)
}

func TestKeysHandler_WithValidTemplate_Markdown(t *testing.T) {
	origFS := app.FS
	memFS := afero.NewMemMapFs()
	app.FS = memFS
	defer func() { app.FS = origFS }()

	const templatePath = "/test-template-md-dir"
	require.NoError(t, memFS.MkdirAll(templatePath, 0o755))
	require.NoError(t, afero.WriteFile(memFS, templatePath+"/config.yaml", []byte(configYAMLFixture), 0o644))

	cmd := keysCmd()
	require.NoError(t, cmd.Flags().Set("template", templatePath))
	require.NoError(t, cmd.Flags().Set("markdown", "true"))

	err := keysHandler()
	assert.NoError(t, err)
}

func TestKeysHandler_InvalidTemplate_ReturnsError(t *testing.T) {
	origFS := app.FS
	app.FS = afero.NewMemMapFs()
	defer func() { app.FS = origFS }()

	cmd := keysCmd()
	require.NoError(t, cmd.Flags().Set("template", "/nonexistent-template-dir"))
	require.NoError(t, cmd.Flags().Set("markdown", "false"))

	err := keysHandler()
	assert.Error(t, err)
}
