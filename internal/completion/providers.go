package completion

import (
	"github.com/spf13/cobra"
)

// CompletionProvider is the interface for all completion providers
type CompletionProvider interface {
	// Complete returns completions matching the given prefix
	Complete(toComplete string) ([]string, cobra.ShellCompDirective)
}

// ContextualProvider provides completions that depend on command context
type ContextualProvider interface {
	// CompleteWithContext returns completions using context from the command
	CompleteWithContext(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective)
}

// NoFileCompDirective indicates that file completion should be disabled
const NoFileCompDirective = cobra.ShellCompDirectiveNoFileComp

// ErrorDirective indicates that an error occurred during completion
const ErrorDirective = cobra.ShellCompDirectiveError
