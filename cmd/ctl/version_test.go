package ctl

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"

	"stamus-ctl/internal/app"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// captureOutput captures anything written to os.Stdout during fn.
func captureOutput(fn func()) string {
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

func TestVersionCmd_NotNil(t *testing.T) {
	cmd := versionCmd()
	require.NotNil(t, cmd)
}

func TestVersionCmd_Use(t *testing.T) {
	cmd := versionCmd()
	assert.Equal(t, "version", cmd.Use)
}

func TestVersionCmd_ShortDescription(t *testing.T) {
	cmd := versionCmd()
	assert.NotEmpty(t, cmd.Short)
}

func TestPrintVersion_DoesNotPanic(t *testing.T) {
	assert.NotPanics(t, func() {
		captureOutput(printVersion)
	})
}

func TestPrintVersion_ContainsExpectedKeys(t *testing.T) {
	out := captureOutput(printVersion)
	assert.True(t, strings.Contains(out, "version:"), "output should contain 'version:'")
	assert.True(t, strings.Contains(out, "arch:"), "output should contain 'arch:'")
	assert.True(t, strings.Contains(out, "commit:"), "output should contain 'commit:'")
}

func TestPrintVersion_ReflectsAppVersion(t *testing.T) {
	old := app.Version
	app.Version = "sentinel-test-version"
	defer func() { app.Version = old }()

	out := captureOutput(printVersion)
	assert.Contains(t, out, "sentinel-test-version")
}

func TestVersionCmd_RunDoesNotPanic(t *testing.T) {
	cmd := versionCmd()
	assert.NotPanics(t, func() {
		captureOutput(func() {
			_ = cmd.RunE // nil — version uses Run, not RunE
			cmd.Run(cmd, nil)
		})
	})
}
