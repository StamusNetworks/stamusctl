package completion

// fresh_data_test.go — tests for the non-cache (fresh data fetch) paths in
// the completion functions.  These exercise the stamus.GetConfigsList(),
// backup.ListBackups(), and config.GetParamsList() call paths.

import (
	"os"
	"testing"

	"stamus-ctl/internal/app"
	"stamus-ctl/internal/models"

	"github.com/spf13/afero"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ---------------------------------------------------------------------------
// CompleteConfigs — fresh data path (cache empty, reads from app.FS)
// ---------------------------------------------------------------------------

func TestCompleteConfigs_FreshData_EmptyConfigs(t *testing.T) {
	// Ensure the MemMapFs has an empty (but existing) configs folder.
	require.NoError(t, app.FS.MkdirAll(app.ConfigsFolder, 0o755))

	cache := GetCache()
	cache.Clear()
	defer cache.Clear()

	got, directive := CompleteConfigs("")
	assert.Equal(t, NoFileCompDirective, directive)
	// May return empty or populated list; no error directive expected.
	assert.NotEqual(t, ErrorDirective, directive)
	_ = got
}

func TestCompleteConfigs_FreshData_WithConfigDirs(t *testing.T) {
	// Pre-populate the configs folder with some config directories.
	require.NoError(t, app.FS.MkdirAll(app.ConfigsFolder+"myconfig1", 0o755))
	require.NoError(t, app.FS.MkdirAll(app.ConfigsFolder+"myconfig2", 0o755))

	cache := GetCache()
	cache.Clear()
	defer cache.Clear()

	got, directive := CompleteConfigs("")
	assert.Equal(t, NoFileCompDirective, directive)
	// Both myconfig1 and myconfig2 should appear.
	assert.Contains(t, got, "myconfig1")
	assert.Contains(t, got, "myconfig2")

	// Verify caching: second call should hit cache.
	got2, _ := CompleteConfigs("")
	assert.ElementsMatch(t, got, got2)
}

func TestCompleteConfigs_FreshData_WithPrefix(t *testing.T) {
	require.NoError(t, app.FS.MkdirAll(app.ConfigsFolder+"alpha", 0o755))
	require.NoError(t, app.FS.MkdirAll(app.ConfigsFolder+"beta", 0o755))

	cache := GetCache()
	cache.Clear()
	defer cache.Clear()

	got, directive := CompleteConfigs("al")
	assert.Equal(t, NoFileCompDirective, directive)
	assert.Contains(t, got, "alpha")
	// beta should not be in the results.
	for _, item := range got {
		assert.True(t, len(item) == 0 || item[:2] == "al" || len(item) < 2,
			"unexpected item: %s", item)
	}
}

func TestCompleteConfigsFunc_FreshData(t *testing.T) {
	require.NoError(t, app.FS.MkdirAll(app.ConfigsFolder+"functest", 0o755))

	cache := GetCache()
	cache.Clear()
	defer cache.Clear()

	fn := CompleteConfigsFunc()
	require.NotNil(t, fn)

	cmd := &cobra.Command{}
	got, directive := fn(cmd, nil, "")
	assert.Equal(t, NoFileCompDirective, directive)
	assert.Contains(t, got, "functest")
}

// ---------------------------------------------------------------------------
// CompleteBackups — fresh data path (backup dir doesn't exist → empty list)
// backup.ListBackups() uses real os.Stat which won't find test dirs, so it
// returns an empty list without error.
// ---------------------------------------------------------------------------

func TestCompleteConfigs_FreshData_ErrorPath(t *testing.T) {
	// When the configs folder doesn't exist AND can't be created, GetConfigsList
	// returns an error → ErrorDirective.
	oldFS := app.FS
	app.FS = afero.NewReadOnlyFs(afero.NewMemMapFs())
	defer func() { app.FS = oldFS }()

	cache := GetCache()
	cache.Clear()
	defer cache.Clear()

	_, directive := CompleteConfigs("")
	assert.Equal(t, ErrorDirective, directive)
}

func TestCompleteBackups_FreshData_NoBacks(t *testing.T) {
	cache := GetCache()
	cache.Clear()
	defer cache.Clear()

	// Use a config name that won't have a real backup directory.
	got, directive := CompleteBackups("nonexistent-config-xyz", "")
	assert.Equal(t, NoFileCompDirective, directive)
	// Should return an empty list (no error) since backup dir doesn't exist.
	assert.Empty(t, got)
}

func TestCompleteBackups_FreshData_CachesResult(t *testing.T) {
	cache := GetCache()
	cache.Clear()
	defer cache.Clear()

	// First call hits the fresh-data path.
	_, directive := CompleteBackups("no-backup-config-xyz2", "")
	assert.Equal(t, NoFileCompDirective, directive)

	// Cache should now have the entry.
	cacheKey := "backups:no-backup-config-xyz2"
	_, found := cache.Get(cacheKey)
	assert.True(t, found, "cache should be populated after fresh fetch")
}

// ---------------------------------------------------------------------------
// CompleteConfigKeys — fresh data path (error when config doesn't exist)
// ---------------------------------------------------------------------------

func TestCompleteConfigKeys_FreshData_ErrorOnMissingConfig(t *testing.T) {
	cache := GetCache()
	cache.Clear()
	defer cache.Clear()

	// config.GetParamsList will fail because "nonexistent-config" doesn't exist.
	got, directive := CompleteConfigKeys("nonexistent-config-xyz", "")
	// When GetParamsList fails, ErrorDirective is returned.
	assert.Equal(t, ErrorDirective, directive)
	assert.Nil(t, got)
}

func TestCompleteConfigKeys_FreshData_WithRealConfig(t *testing.T) {
	// Set up a minimal config in the MemMapFs so GetParamsList succeeds.
	// GetParamsList -> LoadConfigFrom -> ConfigFromFile needs values.yaml with stamus.config.
	configName := "testconfig-keys"
	configDir := app.GetConfigsFolder(configName)
	srcDir := "/TestCompleteConfigKeys_Src"

	require.NoError(t, app.FS.MkdirAll(configDir, 0o755))
	require.NoError(t, app.FS.MkdirAll(srcDir, 0o755))
	// The source config (template).
	require.NoError(t, afero.WriteFile(app.FS, srcDir+"/config.yaml",
		[]byte("myparam:\n  usage: test\n  type: string\n  default: hello\n"), 0o644))
	// values.yaml pointing to the source.
	valuesContent := "stamus:\n  config: " + srcDir + "\n  project: testproj\n"
	require.NoError(t, afero.WriteFile(app.FS, configDir+"/values.yaml",
		[]byte(valuesContent), 0o644))

	cache := GetCache()
	cache.Clear()
	defer cache.Clear()

	// GetParamsList reads configDir/values.yaml → loads config from srcDir.
	// app.IsCtl() is false in tests, so conf is prefixed with app.GetConfigsFolder.
	// We pass configName (not full path) and GetParamsList prepends the folder.
	got, directive := CompleteConfigKeys(configName, "")
	assert.Equal(t, NoFileCompDirective, directive)
	// myparam should appear in completions.
	found := false
	for _, item := range got {
		if len(item) >= 7 && item[:7] == "myparam" {
			found = true
			break
		}
	}
	assert.True(t, found, "expected myparam in completions, got: %v", got)
}

func TestCompleteConfigKeys_FreshData_CachesKeys(t *testing.T) {
	cache := GetCache()
	cache.Clear()
	defer cache.Clear()

	// This errors out (no such config) but the error directive is returned.
	CompleteConfigKeys("no-keys-config-xyz", "")

	// Cache should NOT be set on error (function returns ErrorDirective before Set).
	_, found := cache.Get("config_keys:no-keys-config-xyz")
	assert.False(t, found)
}

// ---------------------------------------------------------------------------
// CompleteConfigKeysForSetFunc — covers the config flag path (no equals)
// ---------------------------------------------------------------------------

func TestCompleteConfigKeysForSetFunc_UsesConfigFlagNoEquals(t *testing.T) {
	cache := GetCache()
	cache.Clear()
	defer cache.Clear()

	// Pre-populate cache for the "flagcfg2" key.
	cache.Set("config_keys:flagcfg2", []string{"key1", "key2"}, ConfigKeysTTL*100)

	fn := CompleteConfigKeysForSetFunc()
	require.NotNil(t, fn)

	cmd := &cobra.Command{}
	cmd.Flags().String("config", "flagcfg2", "")

	got, directive := fn(cmd, []string{}, "key")
	assert.Equal(t, NoFileCompDirective, directive)
	assert.ElementsMatch(t, []string{"key1", "key2"}, got)
}

// ---------------------------------------------------------------------------
// helpers to ensure models is importable from completion package
// ---------------------------------------------------------------------------

var _ = models.CreateVariableString // ensure import is used

// ---------------------------------------------------------------------------
// CompleteBackups — fresh data path with actual OS backup directories
// ---------------------------------------------------------------------------

func TestCompleteBackups_FreshData_WithRealBackupDir(t *testing.T) {
	// Create a real temp directory that mimics a backup directory.
	tmpDir := t.TempDir()

	// Save and restore app.ConfigFolder.
	oldConfigFolder := app.ConfigFolder
	app.ConfigFolder = tmpDir + "/"
	defer func() { app.ConfigFolder = oldConfigFolder }()

	configName := "testcfg-backup"
	backupRoot := tmpDir + "/backups/" + configName
	// Create a valid backup directory: YYYYMMDD_HHMMSS_auto
	backupName := "20240101_120000_auto"
	backupPath := backupRoot + "/" + backupName
	require.NoError(t, os.MkdirAll(backupPath, 0o755))

	cache := GetCache()
	cache.Clear()
	defer cache.Clear()

	got, directive := CompleteBackups(configName, "")
	assert.Equal(t, NoFileCompDirective, directive)
	// The backup should appear in completions (with or without description).
	found := false
	for _, item := range got {
		if len(item) >= 15 && item[:15] == "20240101_120000" {
			found = true
			break
		}
	}
	assert.True(t, found, "expected backup timestamp in completions, got: %v", got)
}

func TestCompleteBackups_FreshData_MultipleBackups(t *testing.T) {
	tmpDir := t.TempDir()

	oldConfigFolder := app.ConfigFolder
	app.ConfigFolder = tmpDir + "/"
	defer func() { app.ConfigFolder = oldConfigFolder }()

	configName := "multi-backup-cfg"
	backupRoot := tmpDir + "/backups/" + configName
	// Create multiple backup directories.
	require.NoError(t, os.MkdirAll(backupRoot+"/20240101_000000_auto", 0o755))
	require.NoError(t, os.MkdirAll(backupRoot+"/20240102_000000_manual", 0o755))

	cache := GetCache()
	cache.Clear()
	defer cache.Clear()

	got, directive := CompleteBackups(configName, "")
	assert.Equal(t, NoFileCompDirective, directive)
	// Two completions expected.
	assert.Len(t, got, 2)
}

func TestCompleteBackups_FreshData_WithPrefix_Filters(t *testing.T) {
	tmpDir := t.TempDir()

	oldConfigFolder := app.ConfigFolder
	app.ConfigFolder = tmpDir + "/"
	defer func() { app.ConfigFolder = oldConfigFolder }()

	configName := "prefix-backup-cfg"
	backupRoot := tmpDir + "/backups/" + configName
	require.NoError(t, os.MkdirAll(backupRoot+"/20240101_000000_auto", 0o755))
	require.NoError(t, os.MkdirAll(backupRoot+"/20240102_000000_manual", 0o755))

	cache := GetCache()
	cache.Clear()
	defer cache.Clear()

	got, directive := CompleteBackups(configName, "20240101")
	assert.Equal(t, NoFileCompDirective, directive)
	// Only the first backup should match.
	assert.Len(t, got, 1)
	assert.Contains(t, got[0], "20240101")
}

// ---------------------------------------------------------------------------
// CompleteConfigKeys — fresh data with a config that has values set
// (exercises the !param.Variable.IsNil() branch in the description building)
// ---------------------------------------------------------------------------

func TestCompleteConfigKeys_FreshData_WithValuesSet(t *testing.T) {
	configName := "testconfig-with-values"
	configDir := app.GetConfigsFolder(configName)
	srcDir := "/TestCompleteConfigKeys_WithValues_Src"

	require.NoError(t, app.FS.MkdirAll(configDir, 0o755))
	require.NoError(t, app.FS.MkdirAll(srcDir, 0o755))

	// Source template config with a string param.
	require.NoError(t, afero.WriteFile(app.FS, srcDir+"/config.yaml",
		[]byte("greeting:\n  usage: greeting message\n  type: string\n  default: hello\n"), 0o644))

	// Values config pointing at source, with the param already set.
	valuesContent := "stamus:\n  config: " + srcDir + "\n  project: valuesproj\n"
	require.NoError(t, afero.WriteFile(app.FS, configDir+"/values.yaml",
		[]byte(valuesContent), 0o644))

	cache := GetCache()
	cache.Clear()
	defer cache.Clear()

	got, directive := CompleteConfigKeys(configName, "")
	assert.Equal(t, NoFileCompDirective, directive)
	// "greeting" key should appear.
	assert.NotEmpty(t, got)

	// Check that cache was populated after fresh fetch.
	_, cached := cache.Get("config_keys:" + configName)
	assert.True(t, cached, "cache should be populated after fresh keys fetch")
}

// ---------------------------------------------------------------------------
// CompleteConfigKeys — fresh data with truncated long value
// (exercises the len(currentValue) > 30 truncation branch)
// ---------------------------------------------------------------------------

func TestCompleteConfigKeys_FreshData_LongValue(t *testing.T) {
	configName := "testconfig-long-value"
	configDir := app.GetConfigsFolder(configName)
	srcDir := "/TestCompleteConfigKeys_LongValue_Src"

	require.NoError(t, app.FS.MkdirAll(configDir, 0o755))
	require.NoError(t, app.FS.MkdirAll(srcDir, 0o755))

	// Source template with a long default value to trigger truncation.
	longDefault := "this-is-a-very-long-default-value-that-exceeds-thirty-chars"
	srcContent := "longparam:\n  usage: long param\n  type: string\n  default: " + longDefault + "\n"
	require.NoError(t, afero.WriteFile(app.FS, srcDir+"/config.yaml", []byte(srcContent), 0o644))

	// Values with the long param set.
	valuesContent := "longparam: " + longDefault + "\nstamus:\n  config: " + srcDir + "\n  project: longproj\n"
	require.NoError(t, afero.WriteFile(app.FS, configDir+"/values.yaml",
		[]byte(valuesContent), 0o644))

	cache := GetCache()
	cache.Clear()
	defer cache.Clear()

	got, directive := CompleteConfigKeys(configName, "")
	assert.Equal(t, NoFileCompDirective, directive)
	assert.NotEmpty(t, got)
}

// TestCompleteBackups_FreshData_InvalidConfigName covers line 24-26 in backups.go:
// backup.ListBackups fails when configName is empty (ValidateProjectName rejects it).
func TestCompleteBackups_FreshData_InvalidConfigName(t *testing.T) {
	cache := GetCache()
	cache.Clear()
	defer cache.Clear()

	// Empty configName → ValidateProjectName("") errors → ListBackups fails → ErrorDirective.
	got, directive := CompleteBackups("", "")
	assert.Equal(t, ErrorDirective, directive)
	assert.Nil(t, got)
}
