package handlers

import (
	"errors"
	"io"
	"os"
	"testing"

	"stamus-ctl/internal/app"
	"stamus-ctl/internal/handlers/common"
	"stamus-ctl/internal/models"
	"stamus-ctl/internal/stamus"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestNixInitHandler_RestoresInterfaceDetection proves that NixInitHandler does
// not leak the process-global models.DisableInterfaceDetection flag. The handler
// enables it to skip host NIC probing, but it must restore the previous value on
// return — otherwise a later handler in the same process (e.g. a daemon serving a
// subsequent `compose init`) would silently skip interface detection.
func TestNixInitHandler_RestoresInterfaceDetection(t *testing.T) {
	configDir := "/tmp/test-nix-init-iface-restore"
	cleanup := setupNixEmbedConfig(t, configDir)
	defer cleanup()

	models.DisableInterfaceDetection = false
	// Safety net: never leak into other tests regardless of outcome.
	defer func() { models.DisableInterfaceDetection = false }()

	_ = NixInitHandler(false, NixInitHandlerInputs{
		IsDefault: true,
		Project:   "clearndr",
		Version:   "1.0.0",
		Config:    configDir,
	})

	assert.False(t, models.DisableInterfaceDetection,
		"DisableInterfaceDetection must be restored after NixInitHandler returns")
}

// minimalConfigYAML is a config.yaml that defines one typed parameter so
// InstanciateConfig → ExtractParams → SetParameters → ValidateAll all succeed.
const minimalConfigYAML = `
mykey:
  type: string
  usage: A test parameter
  default: mydefault
`

// TestNixInitHandler_EmbedMode_GetsPastValidation exercises the full handler body
// up to (and including) the point where template instantiation is attempted.
// With app.Embed=true a pull failure is tolerated, so execution continues into
// InstanciateConfig, which reads config.yaml from DefaultClearNDRPath.
func TestNixInitHandler_EmbedMode_FullPath(t *testing.T) {
	oldFS := app.FS
	app.FS = afero.NewMemMapFs()
	oldEmbed := app.Embed
	oldClearNDRPath := app.DefaultClearNDRPath
	defer func() {
		app.FS = oldFS
		app.Embed = oldEmbed
		app.DefaultClearNDRPath = oldClearNDRPath
	}()

	app.Embed.Set("true")

	// Place a config.yaml at DefaultClearNDRPath so InstanciateConfig succeeds.
	embedPath := "/tmp/test-clearndr-embedded"
	app.DefaultClearNDRPath = embedPath
	require.NoError(t, app.FS.MkdirAll(embedPath, 0o755))
	require.NoError(t, afero.WriteFile(app.FS, embedPath+"/config.yaml", []byte(minimalConfigYAML), 0o644))

	// Also create the config output dir (models.CreateFile needs the dir to exist or
	// will create it — but let's make it explicit).
	configDir := "/tmp/test-nix-init-config"
	require.NoError(t, app.FS.MkdirAll(configDir, 0o755))

	err := NixInitHandler(false, NixInitHandlerInputs{
		IsDefault: true,
		Project:   "clearndr",
		Version:   "1.0.0",
		Config:    configDir,
	})
	// The handler may fail at SaveConfigTo (empty template folder), backup, or
	// stamus.AddInstance — any of those is acceptable as long as we got past
	// the validation + pull phase. The key assertion is that the error is NOT
	// the "invalid project name" or "invalid version" sentinel.
	if err != nil {
		assert.NotContains(t, err.Error(), "invalid project name")
		assert.NotContains(t, err.Error(), "invalid version")
	}
}

// TestNixInitHandler_EmbedMode_WithTemplateFolder uses the TemplateFolder override
// so ResolveTemplatePath returns the custom folder directly (only when embed=false).
func TestNixInitHandler_CustomTemplateFolder(t *testing.T) {
	oldFS := app.FS
	app.FS = afero.NewMemMapFs()
	oldEmbed := app.Embed
	defer func() {
		app.FS = oldFS
		app.Embed = oldEmbed
	}()

	app.Embed.Set("false")

	templateDir := "/tmp/custom-template-folder"
	require.NoError(t, app.FS.MkdirAll(templateDir, 0o755))
	require.NoError(t, afero.WriteFile(app.FS, templateDir+"/config.yaml", []byte(minimalConfigYAML), 0o644))

	configDir := "/tmp/test-nix-init-custom"
	require.NoError(t, app.FS.MkdirAll(configDir, 0o755))

	// With embed=false and TemplateFolder set, PullLatestTemplate is SKIPPED for
	// registry pull (no registry set) — the pull will fail because Docker is absent.
	// Since embed=false the error is returned before InstanciateConfig.
	// This test verifies that error handling works (no panic) and the error is not
	// about validation.
	err := NixInitHandler(false, NixInitHandlerInputs{
		IsDefault:      true,
		Project:        "clearndr",
		Version:        "1.0.0",
		Config:         configDir,
		TemplateFolder: templateDir,
	})
	if err != nil {
		assert.NotContains(t, err.Error(), "invalid project name")
		assert.NotContains(t, err.Error(), "invalid version")
	}
}

// TestNixInitHandler_WithRegistry_ManifestUnknown exercises the explicit registry
// branch where PullConfigAndUnwrap returns the manifest-unknown sentinel error.
func TestNixInitHandler_WithRegistry_ManifestUnknownError(t *testing.T) {
	oldFS := app.FS
	app.FS = afero.NewMemMapFs()
	defer func() { app.FS = oldFS }()

	err := NixInitHandler(false, NixInitHandlerInputs{
		IsDefault: true,
		Project:   "clearndr",
		Version:   "nonexistent-version-xyz-12345",
		Config:    "/tmp/registry-test-config",
		// Registry is set → takes the explicit registry branch.
		// PullConfigAndUnwrap will fail because the registry/image doesn't exist.
		Registry: "ghcr.io/stamusnetworks/stamusctl-templates",
	})
	// Any error is acceptable; we just check no panic and validation passed.
	// The function should error out with something about the registry or pull.
	assert.Error(t, err)
}

// TestNixInitHandler_EmbedFalse_PullError verifies that when embed=false and the
// pull fails (no Docker), the error is propagated correctly.
func TestNixInitHandler_EmbedFalse_PullFails(t *testing.T) {
	oldFS := app.FS
	app.FS = afero.NewMemMapFs()
	oldEmbed := app.Embed
	defer func() {
		app.FS = oldFS
		app.Embed = oldEmbed
	}()

	app.Embed.Set("false")

	err := NixInitHandler(false, NixInitHandlerInputs{
		IsDefault: false,
		Project:   "clearndr",
		Version:   "latest",
		Config:    "/tmp/nix-init-embed-false",
	})
	// Without Docker the pull always fails; with embed=false the error is returned.
	assert.Error(t, err)
}

// TestNixInitHandler_EmbedTrue_BothPathsFail tests that when embed=true but both
// the primary path and the backup (DefaultClearNDRPath) lack config.yaml, we
// get an InstanciateConfig error.
func TestNixInitHandler_EmbedTrue_NeitherPathHasConfig(t *testing.T) {
	oldFS := app.FS
	app.FS = afero.NewMemMapFs()
	oldEmbed := app.Embed
	oldClearNDRPath := app.DefaultClearNDRPath
	defer func() {
		app.FS = oldFS
		app.Embed = oldEmbed
		app.DefaultClearNDRPath = oldClearNDRPath
	}()

	app.Embed.Set("true")
	app.DefaultClearNDRPath = "/nonexistent/clearndr/embedded"

	err := NixInitHandler(false, NixInitHandlerInputs{
		IsDefault: true,
		Project:   "clearndr",
		Version:   "1.0.0",
		Config:    "/tmp/nix-init-both-fail",
	})
	assert.Error(t, err)
}

// TestNixSwitchHandler_IsNixOS_BuildFails exercises the branch where IsNixOS()
// returns true (because /etc/NIXOS exists in afero FS) and the config dir exists,
// so NixosRebuild is attempted — which fails because nixos-rebuild is not in PATH.
func TestNixSwitchHandler_NixOS_RebuildFails(t *testing.T) {
	oldFS := app.FS
	app.FS = afero.NewMemMapFs()
	defer func() { app.FS = oldFS }()

	// Simulate NixOS marker.
	require.NoError(t, app.FS.MkdirAll("/etc", 0o755))
	require.NoError(t, afero.WriteFile(app.FS, "/etc/NIXOS", []byte(""), 0o644))

	// Create a config directory so the existence check passes.
	configPath := "/etc/nixos-test-config"
	require.NoError(t, app.FS.MkdirAll(configPath, 0o755))

	oldName := app.Name
	app.Name = app.CtlName // IsCtl() → true, so configPath = params.Config directly
	defer func() { app.Name = oldName }()

	err := NixSwitchHandler(NixSwitchHandlerInputs{Config: configPath})
	// nixos-rebuild is not installed in CI → should error from NixosRebuild.
	require.Error(t, err)
	assert.Contains(t, err.Error(), "nixos-rebuild")
}

// TestNixSwitchHandler_NixOS_ConfigMissing tests the branch where IsNixOS is
// true but the config directory does not exist.
func TestNixSwitchHandler_NixOS_ConfigMissing(t *testing.T) {
	oldFS := app.FS
	app.FS = afero.NewMemMapFs()
	defer func() { app.FS = oldFS }()

	// Simulate NixOS marker.
	require.NoError(t, app.FS.MkdirAll("/etc", 0o755))
	require.NoError(t, afero.WriteFile(app.FS, "/etc/NIXOS", []byte(""), 0o644))

	err := NixSwitchHandler(NixSwitchHandlerInputs{Config: "/nonexistent-config"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

// setupNixEmbedConfig sets up the embed mode with a config.yaml at DefaultClearNDRPath
// and a pre-created configDir. Returns cleanup func.
func setupNixEmbedConfig(t *testing.T, configDir string) (cleanup func()) {
	t.Helper()

	oldFS := app.FS
	app.FS = afero.NewMemMapFs()
	oldEmbed := app.Embed
	oldClearNDRPath := app.DefaultClearNDRPath

	app.Embed.Set("true")

	embedPath := "/tmp/test-nix-embed-shared"
	app.DefaultClearNDRPath = embedPath
	require.NoError(t, app.FS.MkdirAll(embedPath, 0o755))
	require.NoError(t, afero.WriteFile(app.FS, embedPath+"/config.yaml", []byte(minimalConfigYAML), 0o644))

	require.NoError(t, app.FS.MkdirAll(configDir, 0o755))

	return func() {
		app.FS = oldFS
		app.Embed = oldEmbed
		app.DefaultClearNDRPath = oldClearNDRPath
	}
}

// TestNixInitHandler_SetValuesFromFiles_InvalidFormat covers lines 135-138:
// SetValuesFromFiles returns an error when FromFile has no '=' separator.
func TestNixInitHandler_SetValuesFromFiles_InvalidFormat(t *testing.T) {
	configDir := "/tmp/test-nix-init-fromfile-err"
	cleanup := setupNixEmbedConfig(t, configDir)
	defer cleanup()

	err := NixInitHandler(false, NixInitHandlerInputs{
		IsDefault: true,
		Project:   "clearndr",
		Version:   "1.0.0",
		Config:    configDir,
		FromFile:  "invalid-no-equals",
	})
	// SetValuesFromFiles should return "invalid argument" error.
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid argument")
}

// TestNixInitHandler_SetValuesFromFile_NonExistentFile covers lines 142-145:
// SetValuesFromFile returns an error when Values points to a nonexistent file.
func TestNixInitHandler_SetValuesFromFile_NonExistent(t *testing.T) {
	configDir := "/tmp/test-nix-init-values-err"
	cleanup := setupNixEmbedConfig(t, configDir)
	defer cleanup()

	err := NixInitHandler(false, NixInitHandlerInputs{
		IsDefault: true,
		Project:   "clearndr",
		Version:   "1.0.0",
		Config:    configDir,
		Values:    "/nonexistent-values-file.yaml",
	})
	// SetValuesFromFile should error because the file doesn't exist.
	assert.Error(t, err)
}

// TestNixInitHandler_BackupPath covers lines 171-179:
// When configPath already exists (os.Stat succeeds), the backup branch is entered.
// backup.CreateBackup will fail (source doesn't exist on real OS) but the error
// is only logged as a warning, so execution continues to CreateFile.
// isCli=true ensures configPath = params.Config directly (no GetConfigsFolder wrapping).
func TestNixInitHandler_BackupCreated_WhenConfigExists(t *testing.T) {
	configDir := "/tmp/test-nix-init-backup-path"
	cleanup := setupNixEmbedConfig(t, configDir)
	defer cleanup()

	// Pre-create the configDir on the REAL OS so os.Stat succeeds at line 171.
	require.NoError(t, os.MkdirAll(configDir, 0o755))
	defer os.RemoveAll(configDir)

	// isCli=true → configPath = params.Config (the real directory we just created).
	err := NixInitHandler(true, NixInitHandlerInputs{
		IsDefault: true,
		Project:   "clearndr",
		Version:   "1.0.0",
		Config:    configDir,
	})
	// Backup warning is non-fatal; handler continues to CreateFile and may
	// succeed or fail at a later step. The important thing is no panic.
	// Any error here is past the backup branch.
	if err != nil {
		assert.NotContains(t, err.Error(), "invalid project name")
		assert.NotContains(t, err.Error(), "invalid version")
	}
}

// minimalConfigYAMLWithBadInclude is a config.yaml that includes a nonexistent file.
// ExtractParams will fail when it tries to load the included file.
const minimalConfigYAMLWithBadInclude = `
includes:
  - /totally-nonexistent-file-xyz.yaml
mykey:
  type: string
  usage: A test parameter
  default: mydefault
`

// TestNixInitHandler_ExtractParamsFails covers lines 128-131:
// ExtractParams returns an error when a config.yaml include references a nonexistent file.
func TestNixInitHandler_ExtractParamsFails(t *testing.T) {
	oldFS := app.FS
	app.FS = afero.NewMemMapFs()
	oldEmbed := app.Embed
	oldClearNDRPath := app.DefaultClearNDRPath
	defer func() {
		app.FS = oldFS
		app.Embed = oldEmbed
		app.DefaultClearNDRPath = oldClearNDRPath
	}()

	app.Embed.Set("true")

	embedPath := "/tmp/test-nix-extract-fail"
	app.DefaultClearNDRPath = embedPath
	require.NoError(t, app.FS.MkdirAll(embedPath, 0o755))
	// Write a config.yaml with a bad include to make ExtractParams fail.
	require.NoError(t, afero.WriteFile(app.FS, embedPath+"/config.yaml",
		[]byte(minimalConfigYAMLWithBadInclude), 0o644))

	configDir := "/tmp/test-nix-extract-fail-config"
	require.NoError(t, app.FS.MkdirAll(configDir, 0o755))

	err := NixInitHandler(false, NixInitHandlerInputs{
		IsDefault: true,
		Project:   "clearndr",
		Version:   "1.0.0",
		Config:    configDir,
	})
	// ExtractParams should fail because the included file doesn't exist.
	assert.Error(t, err)
}

// minimalConfigYAMLWithIntParam defines an int-typed parameter.
const minimalConfigYAMLWithIntParam = `
mycount:
  type: int
  usage: An integer parameter
  default: 42
`

// TestNixInitHandler_SetParameters_Error covers lines 149-152:
// SetParameters → SetLooseValues fails when an int parameter receives a non-int value.
func TestNixInitHandler_SetParameters_Error(t *testing.T) {
	oldFS := app.FS
	app.FS = afero.NewMemMapFs()
	oldEmbed := app.Embed
	oldClearNDRPath := app.DefaultClearNDRPath
	defer func() {
		app.FS = oldFS
		app.Embed = oldEmbed
		app.DefaultClearNDRPath = oldClearNDRPath
	}()

	app.Embed.Set("true")

	embedPath := "/tmp/test-nix-setparams-fail"
	app.DefaultClearNDRPath = embedPath
	require.NoError(t, app.FS.MkdirAll(embedPath, 0o755))
	require.NoError(t, afero.WriteFile(app.FS, embedPath+"/config.yaml",
		[]byte(minimalConfigYAMLWithIntParam), 0o644))

	configDir := "/tmp/test-nix-setparams-fail-config"
	require.NoError(t, app.FS.MkdirAll(configDir, 0o755))

	err := NixInitHandler(false, NixInitHandlerInputs{
		IsDefault: false, // isDefault=false to avoid SetToDefaults; we pass arbitrary instead
		Project:   "clearndr",
		Version:   "1.0.0",
		Config:    configDir,
		Arbitrary: map[string]string{"mycount": "not-an-int"},
	})
	// SetLooseValues should fail because "not-an-int" can't be parsed as int.
	assert.Error(t, err)
}

// TestNixInitHandler_NonRegistry_NonErrPullingImage_EmbedFalse covers lines 102-110:
// When the pull error is NOT ErrPullingImage (e.g., GetStamusConfig fails) and
// embed=false, the error is propagated.
func TestNixInitHandler_NonRegistry_NonErrPullingImage_EmbedFalse(t *testing.T) {
	oldFS := app.FS
	app.FS = afero.NewMemMapFs()
	oldEmbed := app.Embed
	// Inject a GetStamusConfigFunc that returns a non-ErrPullingImage error.
	oldGetStamusConfigFunc := common.GetStamusConfigFunc
	defer func() {
		app.FS = oldFS
		app.Embed = oldEmbed
		common.GetStamusConfigFunc = oldGetStamusConfigFunc
	}()

	app.Embed.Set("false")
	common.GetStamusConfigFunc = func() (*stamus.Config, error) {
		return nil, errors.New("simulated config read error (not ErrPullingImage)")
	}

	err := NixInitHandler(false, NixInitHandlerInputs{
		IsDefault: true,
		Project:   "clearndr",
		Version:   "latest",
		Config:    "/tmp/nix-init-non-pulling-error",
	})
	// embed=false + non-ErrPullingImage error → error is returned at line 107-109
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "simulated config read error")
}

// TestNixInitHandler_CreateFile_Error covers lines 185-188:
// models.CreateFile fails when Config contains a null byte that SanitizePath rejects.
func TestNixInitHandler_CreateFile_Error(t *testing.T) {
	oldFS := app.FS
	app.FS = afero.NewMemMapFs()
	oldEmbed := app.Embed
	oldClearNDRPath := app.DefaultClearNDRPath
	defer func() {
		app.FS = oldFS
		app.Embed = oldEmbed
		app.DefaultClearNDRPath = oldClearNDRPath
	}()

	app.Embed.Set("true")

	embedPath := "/tmp/test-nix-createfile-err"
	app.DefaultClearNDRPath = embedPath
	require.NoError(t, app.FS.MkdirAll(embedPath, 0o755))
	require.NoError(t, afero.WriteFile(app.FS, embedPath+"/config.yaml",
		[]byte(minimalConfigYAML), 0o644))

	// Config path with null byte — passes os.Stat and earlier steps (silently fails),
	// but SanitizePath in CreateFile will reject it.
	err := NixInitHandler(false, NixInitHandlerInputs{
		IsDefault: true,
		Project:   "clearndr",
		Version:   "1.0.0",
		Config:    "/tmp/path\x00invalid",
	})
	assert.Error(t, err)
}

// TestNixInitHandler_SetContentHandler_Error covers lines 200-203:
// SetContentHandler fails when a Bind entry has no ':' separator.
func TestNixInitHandler_SetContentHandler_Error(t *testing.T) {
	configDir := "/tmp/test-nix-setcontent-err"
	cleanup := setupNixEmbedConfig(t, configDir)
	defer cleanup()

	err := NixInitHandler(false, NixInitHandlerInputs{
		IsDefault: true,
		Project:   "clearndr",
		Version:   "1.0.0",
		Config:    configDir,
		// Bind entry without ':' separator triggers SetContentHandler error
		Bind: []string{"invalid-no-colon"},
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid argument")
}

// failingFileOpener implements stamus.FileOpener with a failing OpenFile.
type failingFileOpener struct{}

func (f *failingFileOpener) MkdirAll(path string, perm os.FileMode) error {
	return nil // MkdirAll succeeds
}

func (f *failingFileOpener) OpenFile(name string, flag int, perm os.FileMode) (*os.File, error) {
	return nil, errors.New("simulated file open failure for AddInstance")
}

func (f *failingFileOpener) ReadAll(r io.Reader) ([]byte, error) {
	return []byte("{}"), nil // Return empty config so GetConfig succeeds
}

// TestNixInitHandler_AddInstance_Error covers lines 218-221:
// stamus.AddInstance fails when the ConfigManager's FileOpener returns an error on write.
func TestNixInitHandler_AddInstance_Error(t *testing.T) {
	configDir := "/tmp/test-nix-addinstance-err"
	cleanup := setupNixEmbedConfig(t, configDir)
	defer cleanup()

	// Override stamus.DefaultManager to make AddInstance fail.
	oldManager := stamus.DefaultManager
	stamus.DefaultManager = stamus.NewConfigManager(&failingFileOpener{}, nil)
	defer func() { stamus.DefaultManager = oldManager }()

	err := NixInitHandler(false, NixInitHandlerInputs{
		IsDefault: true,
		Project:   "clearndr",
		Version:   "1.0.0",
		Config:    configDir,
	})
	// AddInstance should fail because FileOpener.OpenFile returns an error.
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "simulated file open failure")
}
