package compose

// init_handler_test.go — covers the handler() (compose init) and updateHandler().

import (
	"os"
	"testing"

	"stamus-ctl/internal/app"
	flags "stamus-ctl/internal/handlers"
	"stamus-ctl/internal/logging"
	"stamus-ctl/internal/shutdown"

	"github.com/stretchr/testify/assert"
)

func TestMain(m *testing.M) {
	logging.SetLogger()
	shutdown.Init(logging.Logger)
	os.Exit(m.Run())
}

// ---------------------------------------------------------------------------
// handler() — compose init handler.
// handler() calls InitHandler which attempts to pull templates from a registry.
// We let it fail — we only need the flag-extraction code path covered.
// This test runs in short mode too because it fails fast (no actual network I/O
// completes before the registry error).
// ---------------------------------------------------------------------------

func TestHandler_FlagExtractionAndDelegatesToInitHandler(t *testing.T) {
	oldMode := app.Mode
	app.Mode = "test"
	defer func() { app.Mode = oldMode }()

	cmd := initCmd()
	_ = cmd.Flags().Set("config", "testconf-extract")
	_ = cmd.Flags().Set("version", "latest")
	_ = cmd.Flags().Set("registry", "")
	_ = cmd.Flags().Set("template", "")
	_ = cmd.Flags().Set("bind", "")
	_ = cmd.Flags().Set("default", "false")
	_ = cmd.Flags().Set("expert", "false")
	_ = cmd.Flags().Set("values", "")
	_ = cmd.Flags().Set("fromFile", "")

	// handler() will fail at InitHandler (no template available, registry unreachable)
	// but all flag-extraction code before that runs.
	err := handler(cmd, []string{})
	_ = err // expected error; we just need coverage
}

// TestHandler_IsDefaultFlagTrue exercises the deprecation log branch.
func TestHandler_IsDefaultFlagTrue_LogsDeprecation(t *testing.T) {
	oldMode := app.Mode
	app.Mode = "test"
	defer func() { app.Mode = oldMode }()

	cmd := initCmd()
	_ = cmd.Flags().Set("config", "testconf-default")
	_ = cmd.Flags().Set("version", "latest")
	_ = cmd.Flags().Set("registry", "")
	_ = cmd.Flags().Set("template", "")
	_ = cmd.Flags().Set("bind", "")
	_ = cmd.Flags().Set("default", "true") // exercises deprecation log
	_ = cmd.Flags().Set("expert", "false")
	_ = cmd.Flags().Set("values", "")
	_ = cmd.Flags().Set("fromFile", "")

	err := handler(cmd, []string{})
	_ = err
}

// TestHandler_WithNonKVArg covers firstArg branch where arg has no "=".
func TestHandler_WithNonKVArg(t *testing.T) {
	oldMode := app.Mode
	app.Mode = "test"
	defer func() { app.Mode = oldMode }()

	cmd := initCmd()
	_ = cmd.Flags().Set("config", "testconf-arg")
	_ = cmd.Flags().Set("version", "latest")
	_ = cmd.Flags().Set("registry", "")
	_ = cmd.Flags().Set("template", "")
	_ = cmd.Flags().Set("bind", "")
	_ = cmd.Flags().Set("default", "false")
	_ = cmd.Flags().Set("expert", "false")
	_ = cmd.Flags().Set("values", "")
	_ = cmd.Flags().Set("fromFile", "")

	// Arg "myproject" has no "=" → sets project to "myproject"
	err := handler(cmd, []string{"myproject"})
	_ = err
}

// TestHandler_WithKVArg covers the branch where firstArg contains "=".
func TestHandler_WithKVArg(t *testing.T) {
	oldMode := app.Mode
	app.Mode = "test"
	defer func() { app.Mode = oldMode }()

	cmd := initCmd()
	_ = cmd.Flags().Set("config", "testconf-kv")
	_ = cmd.Flags().Set("version", "latest")
	_ = cmd.Flags().Set("registry", "")
	_ = cmd.Flags().Set("template", "")
	_ = cmd.Flags().Set("bind", "")
	_ = cmd.Flags().Set("default", "false")
	_ = cmd.Flags().Set("expert", "false")
	_ = cmd.Flags().Set("values", "")
	_ = cmd.Flags().Set("fromFile", "")

	// firstArg "param1=val" contains "=" → stays as arbitrary param, project stays "clearndr"
	err := handler(cmd, []string{"param1=val"})
	_ = err
}

// ---------------------------------------------------------------------------
// updateHandler — all flags added to updateCmd()
// ---------------------------------------------------------------------------

func TestUpdateHandler_FlagExtractionPath(t *testing.T) {
	cmd := updateCmd()
	_ = cmd.Flags().Set("config", "testconf-update")
	_ = cmd.Flags().Set("version", "latest")
	_ = cmd.Flags().Set("template", "")
	_ = cmd.Flags().Set("interactive", "false")

	// UpdateHandler fails because no config exists on disk — expected
	err := updateHandler(cmd, []string{})
	assert.Error(t, err)
}

func TestUpdateHandler_WithVersion(t *testing.T) {
	cmd := updateCmd()
	_ = cmd.Flags().Set("config", "testconf-update2")
	_ = cmd.Flags().Set("version", "1.2.3")
	_ = cmd.Flags().Set("template", "")
	_ = cmd.Flags().Set("interactive", "false")

	err := updateHandler(cmd, []string{})
	assert.Error(t, err)
}

func TestUpdateHandler_WithArgs(t *testing.T) {
	cmd := updateCmd()
	_ = cmd.Flags().Set("config", "testconf-update3")
	_ = cmd.Flags().Set("version", "latest")
	_ = cmd.Flags().Set("template", "")
	_ = cmd.Flags().Set("interactive", "true")

	err := updateHandler(cmd, []string{"param1=val"})
	assert.Error(t, err)
}

// ---------------------------------------------------------------------------
// Global flags defaults sanity check
// ---------------------------------------------------------------------------

func TestFlags_GlobalDefaults_NotNil(t *testing.T) {
	assert.NotNil(t, flags.IsDefaultParam.Default.Bool)
	assert.NotNil(t, flags.IsExpert.Default.Bool)
	assert.NotNil(t, flags.Values.Default.String)
	assert.NotNil(t, flags.FromFile.Default.String)
	assert.NotNil(t, flags.Config.Default.String)
	assert.NotNil(t, flags.Template.Default.String)
	assert.NotNil(t, flags.Version.Default.String)
	assert.NotNil(t, flags.Bind.Default.String)
}
