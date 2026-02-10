package completion

import (
	"fmt"

	"stamus-ctl/internal/app"
	"stamus-ctl/internal/backup"

	"github.com/spf13/cobra"
)

// CompleteBackups returns completion suggestions for backup timestamps
func CompleteBackups(configName, toComplete string) ([]string, cobra.ShellCompDirective) {
	cache := GetCache()
	cacheKey := fmt.Sprintf("backups:%s", configName)

	// Check cache first for raw timestamps
	if cached, ok := cache.Get(cacheKey); ok {
		return filterByPrefix(cached, toComplete), NoFileCompDirective
	}

	// Fetch fresh data
	backups, err := backup.ListBackups(configName)
	if err != nil {
		return nil, ErrorDirective
	}

	// Build completion list with descriptions
	var completions []string
	var timestamps []string
	for _, b := range backups {
		// Format: "timestamp\tType: <type>, Size: <size>"
		sizeStr := formatSize(b.Size)
		description := fmt.Sprintf("Type: %s, Size: %s, Date: %s",
			b.Type,
			sizeStr,
			b.CreatedAt.Format("2006-01-02 15:04:05"),
		)
		completions = append(completions, fmt.Sprintf("%s\t%s", b.Timestamp, description))
		timestamps = append(timestamps, b.Timestamp)
	}

	// Cache just the timestamps for faster lookup
	cache.Set(cacheKey, timestamps, BackupsTTL)

	return filterCompletionsWithDescByPrefix(completions, toComplete), NoFileCompDirective
}

// CompleteBackupsFunc returns a ValidArgsFunction for backup timestamp completion
// It extracts the config name from the --config flag or uses the default
func CompleteBackupsFunc() func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	return func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		// If we already have an argument, don't complete more
		if len(args) > 0 {
			return nil, NoFileCompDirective
		}

		// Get config name from flag, or use default
		configName := app.DefaultConfigName
		if configFlag := cmd.Flag("config"); configFlag != nil && configFlag.Value.String() != "" {
			configName = configFlag.Value.String()
		}

		return CompleteBackups(configName, toComplete)
	}
}

// formatSize formats a size in bytes to a human-readable string
func formatSize(bytes int64) string {
	const (
		KB = 1024
		MB = 1024 * KB
		GB = 1024 * MB
	)

	switch {
	case bytes >= GB:
		return fmt.Sprintf("%.1f GB", float64(bytes)/float64(GB))
	case bytes >= MB:
		return fmt.Sprintf("%.1f MB", float64(bytes)/float64(MB))
	case bytes >= KB:
		return fmt.Sprintf("%.1f KB", float64(bytes)/float64(KB))
	default:
		return fmt.Sprintf("%d B", bytes)
	}
}
