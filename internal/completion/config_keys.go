package completion

import (
	"fmt"
	"strings"

	"stamus-ctl/internal/app"
	config "stamus-ctl/internal/handlers/config"

	"github.com/spf13/cobra"
)

// CompleteConfigKeys returns completion suggestions for config parameter keys
func CompleteConfigKeys(configName, toComplete string) ([]string, cobra.ShellCompDirective) {
	cache := GetCache()
	cacheKey := fmt.Sprintf("config_keys:%s", configName)

	// Check cache first
	if cached, ok := cache.Get(cacheKey); ok {
		return filterByPrefix(cached, toComplete), NoFileCompDirective
	}

	// Fetch fresh data
	params, err := config.GetParamsList(configName)
	if err != nil {
		return nil, ErrorDirective
	}

	// Build completion list with descriptions
	var completions []string
	for _, key := range params.GetOrdered() {
		param := params.Get(key)
		if param == nil {
			continue
		}

		// Format: "key\tType: <type>, Current: <value>"
		description := fmt.Sprintf("Type: %s", param.Type)
		if !param.Variable.IsNil() {
			currentValue := param.Variable.AsString()
			// Truncate long values
			if len(currentValue) > 30 {
				currentValue = currentValue[:27] + "..."
			}
			description = fmt.Sprintf("%s, Current: %s", description, currentValue)
		}
		completions = append(completions, fmt.Sprintf("%s\t%s", key, description))
	}

	// Cache just the keys (without descriptions) for faster lookup
	var keys []string
	for _, key := range params.GetOrdered() {
		keys = append(keys, key)
	}
	cache.Set(cacheKey, keys, ConfigKeysTTL)

	return filterCompletionsWithDescByPrefix(completions, toComplete), NoFileCompDirective
}

// CompleteConfigKeysFunc returns a ValidArgsFunction for config key completion
// It extracts the config name from the --config flag or uses the default
func CompleteConfigKeysFunc() func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	return func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		// Get config name from flag, or use default
		configName := app.DefaultConfigName
		if configFlag := cmd.Flag("config"); configFlag != nil && configFlag.Value.String() != "" {
			configName = configFlag.Value.String()
		}

		return CompleteConfigKeys(configName, toComplete)
	}
}

// CompleteConfigKeysForSet returns completions for set command (key=value format)
func CompleteConfigKeysForSet(configName, toComplete string) ([]string, cobra.ShellCompDirective) {
	// If toComplete contains '=', don't complete (user is entering value)
	if strings.Contains(toComplete, "=") {
		return nil, NoFileCompDirective
	}

	return CompleteConfigKeys(configName, toComplete)
}

// CompleteConfigKeysForSetFunc returns a ValidArgsFunction for the set command
func CompleteConfigKeysForSetFunc() func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	return func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		// Get config name from flag, or use default
		configName := app.DefaultConfigName
		if configFlag := cmd.Flag("config"); configFlag != nil && configFlag.Value.String() != "" {
			configName = configFlag.Value.String()
		}

		return CompleteConfigKeysForSet(configName, toComplete)
	}
}

// filterCompletionsWithDescByPrefix filters completions (with tab-separated descriptions) by prefix
func filterCompletionsWithDescByPrefix(completions []string, prefix string) []string {
	if prefix == "" {
		return completions
	}

	var filtered []string
	for _, comp := range completions {
		// Extract key part (before the tab separator)
		key := comp
		if idx := strings.Index(comp, "\t"); idx != -1 {
			key = comp[:idx]
		}
		if strings.HasPrefix(key, prefix) {
			filtered = append(filtered, comp)
		}
	}
	return filtered
}
