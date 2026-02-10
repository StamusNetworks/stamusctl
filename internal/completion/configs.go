package completion

import (
	"strings"

	"stamus-ctl/internal/stamus"

	"github.com/spf13/cobra"
)

// CompleteConfigs returns completion suggestions for config names
func CompleteConfigs(toComplete string) ([]string, cobra.ShellCompDirective) {
	cache := GetCache()
	cacheKey := "configs"

	// Check cache first
	if cached, ok := cache.Get(cacheKey); ok {
		return filterByPrefix(cached, toComplete), NoFileCompDirective
	}

	// Fetch fresh data
	configs, err := stamus.GetConfigsList()
	if err != nil {
		return nil, ErrorDirective
	}

	// Cache the results
	cache.Set(cacheKey, configs, ConfigsTTL)

	return filterByPrefix(configs, toComplete), NoFileCompDirective
}

// CompleteConfigsFunc returns a ValidArgsFunction for config name completion
func CompleteConfigsFunc() func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	return func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return CompleteConfigs(toComplete)
	}
}

// filterByPrefix filters a slice of strings to only include those matching the prefix
func filterByPrefix(items []string, prefix string) []string {
	if prefix == "" {
		return items
	}

	var filtered []string
	for _, item := range items {
		if strings.HasPrefix(item, prefix) {
			filtered = append(filtered, item)
		}
	}
	return filtered
}
