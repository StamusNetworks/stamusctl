package completion

import (
	"testing"
	"time"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ---------------------------------------------------------------------------
// formatSize
// ---------------------------------------------------------------------------

func TestFormatSize(t *testing.T) {
	tests := []struct {
		bytes int64
		want  string
	}{
		{0, "0 B"},
		{512, "512 B"},
		{1023, "1023 B"},
		{1024, "1.0 KB"},
		{2048, "2.0 KB"},
		{1024 * 1024, "1.0 MB"},
		{int64(1.5 * 1024 * 1024), "1.5 MB"},
		{1024 * 1024 * 1024, "1.0 GB"},
		{int64(2.5 * 1024 * 1024 * 1024), "2.5 GB"},
	}

	for _, tc := range tests {
		t.Run(tc.want, func(t *testing.T) {
			got := formatSize(tc.bytes)
			assert.Equal(t, tc.want, got)
		})
	}
}

// ---------------------------------------------------------------------------
// filterCompletionsWithDescByPrefix
// ---------------------------------------------------------------------------

func TestFilterCompletionsWithDescByPrefix(t *testing.T) {
	completions := []string{
		"apple\ttype: string, current: foo",
		"apricot\ttype: int, current: 42",
		"banana\ttype: bool",
		"cherry\ttype: string",
	}

	tests := []struct {
		name    string
		prefix  string
		wantLen int
	}{
		{"empty prefix returns all", "", 4},
		{"prefix ap matches two entries", "ap", 2},
		{"prefix ban matches one", "ban", 1},
		{"prefix xyz matches none", "xyz", 0},
		// A prefix that contains a tab is not a valid key prefix — no match
		{"prefix with tab no match", "apple\t", 0},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := filterCompletionsWithDescByPrefix(completions, tc.prefix)
			assert.Len(t, got, tc.wantLen)
		})
	}
}

func TestFilterCompletionsWithDescByPrefix_NoTabSeparator(t *testing.T) {
	// Items without tab still match by full value as the key
	completions := []string{"alpha", "beta", "gamma"}

	got := filterCompletionsWithDescByPrefix(completions, "al")
	assert.Equal(t, []string{"alpha"}, got)

	got = filterCompletionsWithDescByPrefix(completions, "")
	assert.Len(t, got, 3)
}

func TestFilterCompletionsWithDescByPrefix_PreservesDesc(t *testing.T) {
	completions := []string{"key1\tsome description"}

	got := filterCompletionsWithDescByPrefix(completions, "key")
	require.Len(t, got, 1)
	assert.Equal(t, "key1\tsome description", got[0])
}

// ---------------------------------------------------------------------------
// CompleteConfigs — using cache injection to avoid real FS calls
// ---------------------------------------------------------------------------

func TestCompleteConfigs_FromCache(t *testing.T) {
	cache := GetCache()
	cache.Clear()
	defer cache.Clear()

	cache.Set("configs", []string{"config-a", "config-b", "config-c"}, time.Minute)

	got, directive := CompleteConfigs("")
	assert.Equal(t, NoFileCompDirective, directive)
	assert.ElementsMatch(t, []string{"config-a", "config-b", "config-c"}, got)
}

func TestCompleteConfigs_FromCache_WithPrefix(t *testing.T) {
	cache := GetCache()
	cache.Clear()
	defer cache.Clear()

	cache.Set("configs", []string{"config-a", "config-b", "other"}, time.Minute)

	got, directive := CompleteConfigs("config")
	assert.Equal(t, NoFileCompDirective, directive)
	assert.ElementsMatch(t, []string{"config-a", "config-b"}, got)
}

func TestCompleteConfigsFunc_ReturnsWrappedFunc(t *testing.T) {
	cache := GetCache()
	cache.Clear()
	defer cache.Clear()

	cache.Set("configs", []string{"cfg1"}, time.Minute)

	fn := CompleteConfigsFunc()
	require.NotNil(t, fn)

	cmd := &cobra.Command{}
	got, directive := fn(cmd, nil, "")
	assert.Equal(t, NoFileCompDirective, directive)
	assert.Contains(t, got, "cfg1")
}

// ---------------------------------------------------------------------------
// CompleteBackups — cache-based to avoid real backup FS
// ---------------------------------------------------------------------------

func TestCompleteBackups_FromCache(t *testing.T) {
	cache := GetCache()
	cache.Clear()
	defer cache.Clear()

	cache.Set("backups:myconfig", []string{"20240101", "20240202"}, time.Minute)

	got, directive := CompleteBackups("myconfig", "")
	assert.Equal(t, NoFileCompDirective, directive)
	assert.ElementsMatch(t, []string{"20240101", "20240202"}, got)
}

func TestCompleteBackups_FromCache_WithPrefix(t *testing.T) {
	cache := GetCache()
	cache.Clear()
	defer cache.Clear()

	cache.Set("backups:myconfig", []string{"20240101", "20240202", "19990101"}, time.Minute)

	got, directive := CompleteBackups("myconfig", "2024")
	assert.Equal(t, NoFileCompDirective, directive)
	assert.ElementsMatch(t, []string{"20240101", "20240202"}, got)
}

func TestCompleteBackupsFunc_NoArgs(t *testing.T) {
	cache := GetCache()
	cache.Clear()
	defer cache.Clear()

	cache.Set("backups:config", []string{"ts1"}, time.Minute)

	fn := CompleteBackupsFunc()
	require.NotNil(t, fn)

	cmd := &cobra.Command{}
	cmd.Flags().String("config", "", "config name")

	got, directive := fn(cmd, []string{}, "")
	assert.Equal(t, NoFileCompDirective, directive)
	assert.Contains(t, got, "ts1")
}

func TestCompleteBackupsFunc_HasArgs_StopsCompletion(t *testing.T) {
	fn := CompleteBackupsFunc()
	cmd := &cobra.Command{}

	got, directive := fn(cmd, []string{"arg1"}, "")
	assert.Nil(t, got)
	assert.Equal(t, NoFileCompDirective, directive)
}

func TestCompleteBackupsFunc_UsesConfigFlag(t *testing.T) {
	cache := GetCache()
	cache.Clear()
	defer cache.Clear()

	cache.Set("backups:myconfig", []string{"ts-myconfig"}, time.Minute)

	fn := CompleteBackupsFunc()
	cmd := &cobra.Command{}
	cmd.Flags().String("config", "myconfig", "")

	got, _ := fn(cmd, []string{}, "")
	assert.Contains(t, got, "ts-myconfig")
}

// ---------------------------------------------------------------------------
// CompleteConfigKeys — cache-based
// ---------------------------------------------------------------------------

func TestCompleteConfigKeys_FromCache(t *testing.T) {
	cache := GetCache()
	cache.Clear()
	defer cache.Clear()

	cache.Set("config_keys:myconfig", []string{"key1", "key2"}, time.Minute)

	got, directive := CompleteConfigKeys("myconfig", "")
	assert.Equal(t, NoFileCompDirective, directive)
	assert.ElementsMatch(t, []string{"key1", "key2"}, got)
}

func TestCompleteConfigKeysFunc_UsesConfigFlag(t *testing.T) {
	cache := GetCache()
	cache.Clear()
	defer cache.Clear()

	cache.Set("config_keys:flagcfg", []string{"flagkey"}, time.Minute)

	fn := CompleteConfigKeysFunc()
	require.NotNil(t, fn)

	cmd := &cobra.Command{}
	cmd.Flags().String("config", "flagcfg", "")

	got, directive := fn(cmd, nil, "")
	assert.Equal(t, NoFileCompDirective, directive)
	assert.Contains(t, got, "flagkey")
}

// ---------------------------------------------------------------------------
// CompleteConfigKeysForSet
// ---------------------------------------------------------------------------

func TestCompleteConfigKeysForSet_WithEquals(t *testing.T) {
	// When toComplete contains '=', returns nil immediately
	got, directive := CompleteConfigKeysForSet("anyconfig", "key=value")
	assert.Nil(t, got)
	assert.Equal(t, NoFileCompDirective, directive)
}

func TestCompleteConfigKeysForSet_WithoutEquals_UsesCache(t *testing.T) {
	cache := GetCache()
	cache.Clear()
	defer cache.Clear()

	cache.Set("config_keys:myconfig", []string{"key1", "key2"}, time.Minute)

	got, directive := CompleteConfigKeysForSet("myconfig", "key")
	assert.Equal(t, NoFileCompDirective, directive)
	assert.ElementsMatch(t, []string{"key1", "key2"}, got)
}

func TestCompleteConfigKeysForSetFunc_WithEqualsInToComplete(t *testing.T) {
	fn := CompleteConfigKeysForSetFunc()
	require.NotNil(t, fn)

	cmd := &cobra.Command{}
	cmd.Flags().String("config", "", "")

	got, directive := fn(cmd, []string{}, "something=val")
	assert.Nil(t, got)
	assert.Equal(t, NoFileCompDirective, directive)
}

// ---------------------------------------------------------------------------
// RegisterConfigFlagCompletion
// ---------------------------------------------------------------------------

func TestRegisterConfigFlagCompletion(t *testing.T) {
	cache := GetCache()
	cache.Clear()
	defer cache.Clear()

	cache.Set("configs", []string{"myreg"}, time.Minute)

	cmd := &cobra.Command{Use: "test"}
	cmd.Flags().String("config", "", "the config flag")

	// Should not panic
	RegisterConfigFlagCompletion(cmd)

	completionFn, ok := cmd.GetFlagCompletionFunc("config")
	require.True(t, ok)
	require.NotNil(t, completionFn)

	got, directive := completionFn(cmd, nil, "")
	assert.Equal(t, NoFileCompDirective, directive)
	assert.Contains(t, got, "myreg")
}

// ---------------------------------------------------------------------------
// RegisterConfigFlagCompletionRecursive
// ---------------------------------------------------------------------------

func TestRegisterConfigFlagCompletionRecursive_Root(t *testing.T) {
	cache := GetCache()
	cache.Clear()
	defer cache.Clear()

	cache.Set("configs", []string{"recursive-cfg"}, time.Minute)

	root := &cobra.Command{Use: "root"}
	root.PersistentFlags().String("config", "", "config")

	// Should register on root without panic
	RegisterConfigFlagCompletionRecursive(root)

	fn, ok := root.GetFlagCompletionFunc("config")
	require.True(t, ok)
	assert.NotNil(t, fn)
}

func TestRegisterConfigFlagCompletionRecursive_SubCommand(t *testing.T) {
	cache := GetCache()
	cache.Clear()
	defer cache.Clear()

	cache.Set("configs", []string{"sub-cfg"}, time.Minute)

	root := &cobra.Command{Use: "root"}
	sub := &cobra.Command{Use: "sub"}
	sub.Flags().String("config", "", "config")
	root.AddCommand(sub)

	RegisterConfigFlagCompletionRecursive(root)

	fn, ok := sub.GetFlagCompletionFunc("config")
	require.True(t, ok)
	assert.NotNil(t, fn)
}

func TestRegisterConfigFlagCompletionRecursive_NoConfigFlag(t *testing.T) {
	// Command with no --config flag — should not panic, just skip
	cmd := &cobra.Command{Use: "nocfg"}
	cmd.Flags().String("other", "", "other flag")

	// Should not panic
	RegisterConfigFlagCompletionRecursive(cmd)
}
