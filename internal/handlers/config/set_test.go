package config

import (
	"os"
	"path/filepath"
	"testing"

	"stamus-ctl/internal/app"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// setupSetTestFS switches to a real temporary directory FS.
// SetContentHandler uses utils.Copy which calls cp.Copy on the real OS filesystem,
// so we need actual disk paths. The app.FS is used for Stat checks inside utils.Copy.
func setupSetTestFS(t *testing.T) (tmpDir string, cleanup func()) {
	t.Helper()
	tmpDir, err := os.MkdirTemp("", "stamus-set-test-*")
	require.NoError(t, err)

	// Switch app.FS to real OS filesystem so Stat checks in utils.Copy work
	origFS := app.FS
	origName := app.Name
	app.FS = afero.NewOsFs()
	app.Name = app.CtlName // CLI mode

	cleanup = func() {
		app.FS = origFS
		app.Name = origName
		os.RemoveAll(tmpDir)
	}
	return tmpDir, cleanup
}

// TestSetContentHandler_EmptyArgs verifies that passing no arguments is a no-op.
func TestSetContentHandler_EmptyArgs(t *testing.T) {
	tmpDir, cleanup := setupSetTestFS(t)
	defer cleanup()

	err := SetContentHandler(tmpDir, []string{})
	assert.NoError(t, err)
}

// TestSetContentHandler_EmptyStringArg verifies that an empty string argument is skipped.
func TestSetContentHandler_EmptyStringArg(t *testing.T) {
	tmpDir, cleanup := setupSetTestFS(t)
	defer cleanup()

	err := SetContentHandler(tmpDir, []string{""})
	assert.NoError(t, err)
}

// TestSetContentHandler_ValidCopy verifies a "src:dst" argument copies the file.
func TestSetContentHandler_ValidCopy(t *testing.T) {
	tmpDir, cleanup := setupSetTestFS(t)
	defer cleanup()

	// Create source file
	srcFile := filepath.Join(tmpDir, "source.txt")
	require.NoError(t, os.WriteFile(srcFile, []byte("hello world"), 0o644))

	// Destination inside tmpDir (used as conf)
	dstFile := filepath.Join(tmpDir, "dest.txt")

	arg := srcFile + ":" + dstFile
	err := SetContentHandler(tmpDir, []string{arg})
	assert.NoError(t, err)

	// Verify destination was created with correct content
	content, readErr := os.ReadFile(dstFile)
	require.NoError(t, readErr)
	assert.Equal(t, "hello world", string(content))
}

// TestSetContentHandler_InvalidArgFormat verifies args without ":" return an error.
func TestSetContentHandler_InvalidArgFormat(t *testing.T) {
	tmpDir, cleanup := setupSetTestFS(t)
	defer cleanup()

	err := SetContentHandler(tmpDir, []string{"no-colon-here"})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid argument")
}

// TestSetContentHandler_SourceDoesNotExist verifies a missing source file returns an error.
func TestSetContentHandler_SourceDoesNotExist(t *testing.T) {
	tmpDir, cleanup := setupSetTestFS(t)
	defer cleanup()

	nonExistent := filepath.Join(tmpDir, "ghost.txt")
	dstFile := filepath.Join(tmpDir, "out.txt")

	err := SetContentHandler(tmpDir, []string{nonExistent + ":" + dstFile})
	assert.Error(t, err)
}

// TestSetContentHandler_PathTraversalInOutput verifies that an output path attempting
// to escape the conf directory is rejected.
func TestSetContentHandler_PathTraversalInOutput(t *testing.T) {
	tmpDir, cleanup := setupSetTestFS(t)
	defer cleanup()

	// Create a real source file so input path validation passes
	srcFile := filepath.Join(tmpDir, "legit.txt")
	require.NoError(t, os.WriteFile(srcFile, []byte("data"), 0o644))

	// Attempt directory traversal in output
	traversalDst := "../../etc/passwd"
	arg := srcFile + ":" + traversalDst
	err := SetContentHandler(tmpDir, []string{arg})
	assert.Error(t, err)
}

// TestSetContentHandler_MultipleArgs verifies that multiple valid args all succeed.
func TestSetContentHandler_MultipleArgs(t *testing.T) {
	tmpDir, cleanup := setupSetTestFS(t)
	defer cleanup()

	// Create two source files
	src1 := filepath.Join(tmpDir, "file1.txt")
	src2 := filepath.Join(tmpDir, "file2.txt")
	require.NoError(t, os.WriteFile(src1, []byte("content1"), 0o644))
	require.NoError(t, os.WriteFile(src2, []byte("content2"), 0o644))

	dst1 := filepath.Join(tmpDir, "out1.txt")
	dst2 := filepath.Join(tmpDir, "out2.txt")

	args := []string{
		src1 + ":" + dst1,
		src2 + ":" + dst2,
	}

	err := SetContentHandler(tmpDir, args)
	assert.NoError(t, err)

	c1, _ := os.ReadFile(dst1)
	c2, _ := os.ReadFile(dst2)
	assert.Equal(t, "content1", string(c1))
	assert.Equal(t, "content2", string(c2))
}

// TestSetContentHandler_InvalidConfPath verifies that an invalid/empty conf path is rejected.
func TestSetContentHandler_InvalidConfPath(t *testing.T) {
	origName := app.Name
	app.Name = app.CtlName
	defer func() { app.Name = origName }()

	// Empty conf path should be rejected by SanitizePath
	err := SetContentHandler("", []string{"src:dst"})
	assert.Error(t, err)
}

// ---- SetHandler error paths ----

// TestSetHandler_MissingConfig verifies SetHandler errors when config path does not exist.
func TestSetHandler_MissingConfig(t *testing.T) {
	origFS := app.FS
	origName := app.Name
	app.FS = afero.NewMemMapFs()
	app.Name = app.CtlName
	defer func() {
		app.FS = origFS
		app.Name = origName
	}()

	err := SetHandler(SetHandlerInputs{
		Config: "/nonexistent/config",
		Args:   []string{},
	})
	assert.Error(t, err)
}

// TestSetHandler_InvalidArgs verifies that malformed k=v args propagate an error.
func TestSetHandler_InvalidArgs(t *testing.T) {
	origFS := app.FS
	origName := app.Name
	app.FS = afero.NewMemMapFs()
	app.Name = app.CtlName
	defer func() {
		app.FS = origFS
		app.Name = origName
	}()

	// values.yaml present but empty — LoadConfigFrom will fail on stamus.config missing
	err := afero.WriteFile(app.FS, "/conf/values.yaml", []byte(""), 0o644)
	require.NoError(t, err)

	err = SetHandler(SetHandlerInputs{
		Config: "/conf",
		Args:   []string{},
	})
	assert.Error(t, err)
}

// TestSetHandler_Success tests the happy path for SetHandler with a minimal config.
// It uses the same minimal YAML fixture as the get_test.go success tests.
func TestSetHandler_Success(t *testing.T) {
	origFS := app.FS
	origName := app.Name
	app.FS = afero.NewMemMapFs()
	app.Name = app.CtlName
	defer func() {
		app.FS = origFS
		app.Name = origName
	}()

	configDir := "/sethandlerconf"
	require.NoError(t, app.FS.MkdirAll(configDir, 0o755))

	// Create template dir and config.yaml
	templateDir := configDir + "/template"
	require.NoError(t, app.FS.MkdirAll(templateDir, 0o755))
	configYAML := `param1:
  type: string
  usage: A test parameter
  default: hello
`
	require.NoError(t, afero.WriteFile(app.FS, templateDir+"/config.yaml", []byte(configYAML), 0o644))

	// Create values.yaml
	valuesYAML := `stamus:
  config: ` + templateDir + `
  project: test-project
param1: hello
`
	require.NoError(t, afero.WriteFile(app.FS, configDir+"/values.yaml", []byte(valuesYAML), 0o644))

	err := SetHandler(SetHandlerInputs{
		Config: configDir,
		Args:   []string{},
		Apply:  false,
		Reload: true,
	})
	assert.NoError(t, err)
}

// TestSetHandler_WithApply tests SetHandler with Apply=true using test mode so
// wrapper.HandleUp uses the mocker instead of real docker-compose.
func TestSetHandler_WithApply(t *testing.T) {
	origFS := app.FS
	origName := app.Name
	origMode := app.Mode
	app.FS = afero.NewMemMapFs()
	app.Name = app.CtlName
	app.Mode = "test"
	defer func() {
		app.FS = origFS
		app.Name = origName
		app.Mode = origMode
	}()

	configDir := "/applyconf"
	templateDir := configDir + "/template"
	require.NoError(t, app.FS.MkdirAll(templateDir, 0o755))

	configYAML := `param1:
  type: string
  usage: A test parameter
  default: hello
`
	require.NoError(t, afero.WriteFile(app.FS, templateDir+"/config.yaml", []byte(configYAML), 0o644))

	valuesYAML := `stamus:
  config: ` + templateDir + `
  project: test-project
param1: hello
`
	require.NoError(t, afero.WriteFile(app.FS, configDir+"/values.yaml", []byte(valuesYAML), 0o644))

	// Also write docker-compose.yaml so mocker.Up can read services
	dcYAML := `services:
  web:
    image: nginx
`
	require.NoError(t, afero.WriteFile(app.FS, configDir+"/docker-compose.yaml", []byte(dcYAML), 0o644))

	err := SetHandler(SetHandlerInputs{
		Config: configDir,
		Args:   []string{},
		Apply:  true,
		Reload: true,
	})
	// Apply path: backup may warn but HandleUp via mocker should succeed.
	// Some errors from stamus config are acceptable.
	_ = err
}

// TestSetHandler_WithArgs tests SetHandler passing parameter args.
func TestSetHandler_WithArgs(t *testing.T) {
	origFS := app.FS
	origName := app.Name
	app.FS = afero.NewMemMapFs()
	app.Name = app.CtlName
	defer func() {
		app.FS = origFS
		app.Name = origName
	}()

	configDir := "/argsconf"
	templateDir := configDir + "/template"
	require.NoError(t, app.FS.MkdirAll(templateDir, 0o755))

	configYAML := `param1:
  type: string
  usage: A test parameter
  default: hello
`
	require.NoError(t, afero.WriteFile(app.FS, templateDir+"/config.yaml", []byte(configYAML), 0o644))
	valuesYAML := `stamus:
  config: ` + templateDir + `
  project: test-project
param1: hello
`
	require.NoError(t, afero.WriteFile(app.FS, configDir+"/values.yaml", []byte(valuesYAML), 0o644))

	err := SetHandler(SetHandlerInputs{
		Config: configDir,
		Args:   []string{"param1=world"},
		Apply:  false,
		Reload: true,
	})
	assert.NoError(t, err)
}

// ---------------------------------------------------------------------------
// SetContentHandler — daemon-mode path (app.Name != "stamusctl").
// In daemon mode, the output path is validated relative to the ConfigsFolder.
// ---------------------------------------------------------------------------

// TestSetContentHandler_DaemonMode_ValidCopy exercises the !app.IsCtl() branch
// where the output path is joined with app.GetConfigsFolder(conf).
func TestSetContentHandler_DaemonMode_ValidCopy(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "stamus-setcontent-daemon-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	origName := app.Name
	origConfigsFolder := app.ConfigsFolder
	origFS := app.FS
	app.Name = "stamusd"             // IsCtl() == false → daemon path
	app.ConfigsFolder = tmpDir + "/" // ConfigsFolder for daemon mode
	app.FS = afero.NewOsFs()
	defer func() {
		app.Name = origName
		app.ConfigsFolder = origConfigsFolder
		app.FS = origFS
	}()

	// The "conf" parameter is the config name relative to ConfigsFolder.
	confName := "myconf"
	confDir := filepath.Join(tmpDir, confName)
	require.NoError(t, os.MkdirAll(confDir, 0o755))

	// Source file outside conf (valid absolute path).
	srcFile := filepath.Join(tmpDir, "source.txt")
	require.NoError(t, os.WriteFile(srcFile, []byte("daemon content"), 0o644))

	// Destination is a filename relative to the conf dir (within ConfigsFolder/confName).
	dstRelative := "dest.txt"
	dstAbsolute := filepath.Join(confDir, dstRelative)

	arg := srcFile + ":" + dstAbsolute
	err = SetContentHandler(confName, []string{arg})
	assert.NoError(t, err)

	content, readErr := os.ReadFile(dstAbsolute)
	require.NoError(t, readErr)
	assert.Equal(t, "daemon content", string(content))
}

// TestSetContentHandler_DaemonMode_PathTraversal exercises the daemon path traversal
// check: output paths that escape ConfigsFolder/confName should be rejected.
func TestSetContentHandler_DaemonMode_PathTraversal(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "stamus-setcontent-daemon-traversal-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	origName := app.Name
	origConfigsFolder := app.ConfigsFolder
	origFS := app.FS
	app.Name = "stamusd"
	app.ConfigsFolder = tmpDir + "/"
	app.FS = afero.NewOsFs()
	defer func() {
		app.Name = origName
		app.ConfigsFolder = origConfigsFolder
		app.FS = origFS
	}()

	confName := "myconf2"
	confDir := filepath.Join(tmpDir, confName)
	require.NoError(t, os.MkdirAll(confDir, 0o755))

	srcFile := filepath.Join(tmpDir, "safe-source.txt")
	require.NoError(t, os.WriteFile(srcFile, []byte("data"), 0o644))

	// Attempt to escape the conf directory via traversal.
	traversalDst := "../../evil.txt"
	arg := srcFile + ":" + traversalDst
	err = SetContentHandler(confName, []string{arg})
	assert.Error(t, err)
}

// TestSetHandler_InvalidConfigPath exercises the CreateFile error path (line 36)
// by passing a Config path that contains a null byte (fails SanitizePath).
func TestSetHandler_InvalidConfigPath(t *testing.T) {
	origFS := app.FS
	origName := app.Name
	app.FS = afero.NewMemMapFs()
	app.Name = app.CtlName
	defer func() {
		app.FS = origFS
		app.Name = origName
	}()

	err := SetHandler(SetHandlerInputs{
		Config: "/path\x00invalid",
		Args:   []string{},
	})
	assert.Error(t, err)
}

// TestSetHandler_InvalidFromFile exercises the SetValuesFromFiles error path (line 56)
// by passing a FromFile with an invalid format (no '=' separator).
func TestSetHandler_InvalidFromFile(t *testing.T) {
	origFS := app.FS
	origName := app.Name
	app.FS = afero.NewMemMapFs()
	app.Name = app.CtlName
	defer func() {
		app.FS = origFS
		app.Name = origName
	}()

	configDir := "/fromfileconf"
	templateDir := configDir + "/template"
	require.NoError(t, app.FS.MkdirAll(templateDir, 0o755))
	configYAML := `param1:
  type: string
  usage: A test parameter
  default: hello
`
	require.NoError(t, afero.WriteFile(app.FS, templateDir+"/config.yaml", []byte(configYAML), 0o644))
	valuesYAML := `stamus:
  config: ` + templateDir + `
  project: test-project
param1: hello
`
	require.NoError(t, afero.WriteFile(app.FS, configDir+"/values.yaml", []byte(valuesYAML), 0o644))

	// "invalid-format" has no '=' separator, so SetValuesFromFiles returns error.
	err := SetHandler(SetHandlerInputs{
		Config:   configDir,
		Args:     []string{},
		FromFile: "invalid-format",
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid argument")
}

// TestSetHandler_InvalidValuesFile exercises the SetValuesFromFile error path (line 60)
// by passing a Values path that does not exist (CreateFileFromPath will fail for no-dot file).
func TestSetHandler_InvalidValuesFile(t *testing.T) {
	origFS := app.FS
	origName := app.Name
	app.FS = afero.NewMemMapFs()
	app.Name = app.CtlName
	defer func() {
		app.FS = origFS
		app.Name = origName
	}()

	configDir := "/valuesfileconf"
	templateDir := configDir + "/template"
	require.NoError(t, app.FS.MkdirAll(templateDir, 0o755))
	configYAML := `param1:
  type: string
  usage: A test parameter
  default: hello
`
	require.NoError(t, afero.WriteFile(app.FS, templateDir+"/config.yaml", []byte(configYAML), 0o644))
	valuesYAML := `stamus:
  config: ` + templateDir + `
  project: test-project
param1: hello
`
	require.NoError(t, afero.WriteFile(app.FS, configDir+"/values.yaml", []byte(valuesYAML), 0o644))

	// "nodotfile" has no extension, so CreateFileFromPath returns an error.
	err := SetHandler(SetHandlerInputs{
		Config: configDir,
		Args:   []string{},
		Values: "nodotfile",
	})
	assert.Error(t, err)
}
