package handlers

import (
	"fmt"
	"path/filepath"
	"testing"

	"stamus-ctl/internal/app"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNixTestHandler_NotNixOS(t *testing.T) {
	oldFS := app.FS
	app.FS = afero.NewMemMapFs()
	defer func() { app.FS = oldFS }()

	err := NixTestHandler(NixTestHandlerInputs{Config: "myconfig"})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not running NixOS")
}

func TestNixTestHandler_ConfigNotFound(t *testing.T) {
	oldFS := app.FS
	app.FS = afero.NewMemMapFs()
	defer func() { app.FS = oldFS }()

	require.NoError(t, app.FS.MkdirAll("/etc", 0o755))
	require.NoError(t, afero.WriteFile(app.FS, "/etc/NIXOS", []byte(""), 0o644))

	oldName := app.Name
	app.Name = app.CtlName
	defer func() { app.Name = oldName }()

	err := NixTestHandler(NixTestHandlerInputs{Config: "/nonexistent"})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestNixTestHandler_NoTestsDir(t *testing.T) {
	oldFS := app.FS
	app.FS = afero.NewMemMapFs()
	defer func() { app.FS = oldFS }()

	require.NoError(t, app.FS.MkdirAll("/etc", 0o755))
	require.NoError(t, afero.WriteFile(app.FS, "/etc/NIXOS", []byte(""), 0o644))

	oldName := app.Name
	app.Name = app.CtlName
	defer func() { app.Name = oldName }()

	configPath := "/tmp/testconfig"
	require.NoError(t, app.FS.MkdirAll(configPath, 0o755))

	err := NixTestHandler(NixTestHandlerInputs{Config: configPath})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no tests directory")
}

func TestNixTestHandler_NoTestFiles(t *testing.T) {
	oldFS := app.FS
	app.FS = afero.NewMemMapFs()
	defer func() { app.FS = oldFS }()

	require.NoError(t, app.FS.MkdirAll("/etc", 0o755))
	require.NoError(t, afero.WriteFile(app.FS, "/etc/NIXOS", []byte(""), 0o644))

	oldName := app.Name
	app.Name = app.CtlName
	defer func() { app.Name = oldName }()

	configPath := "/tmp/testconfig"
	require.NoError(t, app.FS.MkdirAll(filepath.Join(configPath, "tests"), 0o755))

	err := NixTestHandler(NixTestHandlerInputs{Config: configPath})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no test files")
}

func TestNixTestHandler_FilterNoMatch(t *testing.T) {
	oldFS := app.FS
	app.FS = afero.NewMemMapFs()
	defer func() { app.FS = oldFS }()

	require.NoError(t, app.FS.MkdirAll("/etc", 0o755))
	require.NoError(t, afero.WriteFile(app.FS, "/etc/NIXOS", []byte(""), 0o644))

	oldName := app.Name
	app.Name = app.CtlName
	defer func() { app.Name = oldName }()

	configPath := "/tmp/testconfig"
	testsDir := filepath.Join(configPath, "tests")
	require.NoError(t, app.FS.MkdirAll(testsDir, 0o755))
	require.NoError(t, afero.WriteFile(app.FS, filepath.Join(testsDir, "check.sh"), []byte("#!/bin/bash\nexit 0"), 0o755))

	err := NixTestHandler(NixTestHandlerInputs{Config: configPath, Filter: "*.nix"})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no test files matching filter")
}

func TestDiscoverTests_SortsAlphabetically(t *testing.T) {
	oldFS := app.FS
	app.FS = afero.NewMemMapFs()
	defer func() { app.FS = oldFS }()

	testsDir := "/tmp/tests"
	require.NoError(t, app.FS.MkdirAll(testsDir, 0o755))
	for _, name := range []string{"z-test.sh", "a-test.sh", "m-test.nix"} {
		require.NoError(t, afero.WriteFile(app.FS, filepath.Join(testsDir, name), []byte(""), 0o644))
	}

	files, err := discoverTests(testsDir, "")
	require.NoError(t, err)
	assert.Len(t, files, 3)
	assert.Equal(t, filepath.Join(testsDir, "a-test.sh"), files[0])
	assert.Equal(t, filepath.Join(testsDir, "m-test.nix"), files[1])
	assert.Equal(t, filepath.Join(testsDir, "z-test.sh"), files[2])
}

func TestDiscoverTests_IgnoresNonTestFiles(t *testing.T) {
	oldFS := app.FS
	app.FS = afero.NewMemMapFs()
	defer func() { app.FS = oldFS }()

	testsDir := "/tmp/tests"
	require.NoError(t, app.FS.MkdirAll(testsDir, 0o755))
	for _, name := range []string{"test.sh", "test.nix", "readme.md", "data.json"} {
		require.NoError(t, afero.WriteFile(app.FS, filepath.Join(testsDir, name), []byte(""), 0o644))
	}

	files, err := discoverTests(testsDir, "")
	require.NoError(t, err)
	assert.Len(t, files, 2)
}

func TestDiscoverTests_FilterWorks(t *testing.T) {
	oldFS := app.FS
	app.FS = afero.NewMemMapFs()
	defer func() { app.FS = oldFS }()

	testsDir := "/tmp/tests"
	require.NoError(t, app.FS.MkdirAll(testsDir, 0o755))
	for _, name := range []string{"check-syntax.sh", "check-modules.nix", "validate.sh"} {
		require.NoError(t, afero.WriteFile(app.FS, filepath.Join(testsDir, name), []byte(""), 0o644))
	}

	files, err := discoverTests(testsDir, "check-*")
	require.NoError(t, err)
	assert.Len(t, files, 2)
}

func TestDiscoverTests_IgnoresSubdirectories(t *testing.T) {
	oldFS := app.FS
	app.FS = afero.NewMemMapFs()
	defer func() { app.FS = oldFS }()

	testsDir := "/tmp/tests"
	require.NoError(t, app.FS.MkdirAll(filepath.Join(testsDir, "subdir"), 0o755))
	require.NoError(t, afero.WriteFile(app.FS, filepath.Join(testsDir, "test.sh"), []byte(""), 0o644))

	files, err := discoverTests(testsDir, "")
	require.NoError(t, err)
	assert.Len(t, files, 1)
}

func TestPrintSummary_AllPass(t *testing.T) {
	results := []TestResult{
		{Name: "a.sh", Passed: true},
		{Name: "b.nix", Passed: true},
	}
	err := printSummary(results)
	assert.NoError(t, err)
}

func TestPrintSummary_SomeFail(t *testing.T) {
	results := []TestResult{
		{Name: "a.sh", Passed: true},
		{Name: "b.nix", Passed: false, Err: fmt.Errorf("eval error")},
	}
	err := printSummary(results)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "1 test(s) failed")
}
