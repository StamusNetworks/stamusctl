package config

import (
	"bytes"
	"strings"
	"testing"

	"stamus-ctl/internal/app"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
)

// minimalConfigYAML is a self-contained config.yaml with two typed parameters.
const minimalConfigYAML = `param1:
  usage: first param usage
  type: string
  default: hello
param2:
  usage: second param usage
  type: int
  default: 42
`

func TestKeysHandler(t *testing.T) {
	// Swap filesystem to an in-memory one.
	oldFS := app.FS
	app.FS = afero.NewMemMapFs()
	defer func() { app.FS = oldFS }()

	// Redirect output to a buffer.
	buf := &bytes.Buffer{}
	oldWriter := outputWriter
	outputWriter = buf
	defer func() { outputWriter = oldWriter }()

	// Write the config file into the in-memory FS.
	configDir := "/testtemplate"
	err := app.FS.MkdirAll(configDir, 0o755)
	assert.NoError(t, err)
	err = afero.WriteFile(app.FS, configDir+"/config.yaml", []byte(minimalConfigYAML), 0o644)
	assert.NoError(t, err)

	err = KeysHandler(configDir, false)
	assert.NoError(t, err)

	output := buf.String()
	// The table must list both parameter names, their defaults and usage strings.
	assert.True(t, strings.Contains(output, "param1"), "expected param1 in output, got: %s", output)
	assert.True(t, strings.Contains(output, "hello"), "expected default 'hello' in output, got: %s", output)
	assert.True(t, strings.Contains(output, "first param usage"),
		"expected usage in output, got: %s", output)
	assert.True(t, strings.Contains(output, "param2"), "expected param2 in output, got: %s", output)
	assert.True(t, strings.Contains(output, "42"), "expected default '42' in output, got: %s", output)
	assert.True(t, strings.Contains(output, "second param usage"),
		"expected usage in output, got: %s", output)
}

func TestKeysHandler_Markdown(t *testing.T) {
	oldFS := app.FS
	app.FS = afero.NewMemMapFs()
	defer func() { app.FS = oldFS }()

	buf := &bytes.Buffer{}
	oldWriter := outputWriter
	outputWriter = buf
	defer func() { outputWriter = oldWriter }()

	configDir := "/mdtemplate"
	err := app.FS.MkdirAll(configDir, 0o755)
	assert.NoError(t, err)
	err = afero.WriteFile(app.FS, configDir+"/config.yaml", []byte(minimalConfigYAML), 0o644)
	assert.NoError(t, err)

	err = KeysHandler(configDir, true)
	assert.NoError(t, err)

	output := buf.String()
	// Markdown tables use pipe characters.
	assert.True(t, strings.Contains(output, "|"), "expected markdown pipe characters in output, got: %s", output)
	assert.True(t, strings.Contains(output, "param1"), "expected param1 in markdown output, got: %s", output)
}

func TestKeysHandler_MissingConfig(t *testing.T) {
	oldFS := app.FS
	app.FS = afero.NewMemMapFs()
	defer func() { app.FS = oldFS }()

	// Provide a directory that contains no config.yaml.
	err := KeysHandler("/nonexistent", false)
	assert.Error(t, err)
}
