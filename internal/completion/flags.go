package completion

import (
	"github.com/spf13/cobra"
)

// RegisterConfigFlagCompletion registers completion for the --config flag on a command
func RegisterConfigFlagCompletion(cmd *cobra.Command) {
	cmd.RegisterFlagCompletionFunc("config", func(cmd *cobra.Command, args []string,
		toComplete string,
	) ([]string, cobra.ShellCompDirective) {
		return CompleteConfigs(toComplete)
	})
}

// RegisterConfigFlagCompletionRecursive registers completion for the --config flag
// on a command and all its subcommands
func RegisterConfigFlagCompletionRecursive(cmd *cobra.Command) {
	// Register on this command if it has the config flag
	if cmd.Flags().Lookup("config") != nil || cmd.PersistentFlags().Lookup("config") != nil {
		RegisterConfigFlagCompletion(cmd)
	}

	// Recursively register on subcommands
	for _, subCmd := range cmd.Commands() {
		RegisterConfigFlagCompletionRecursive(subCmd)
	}
}
