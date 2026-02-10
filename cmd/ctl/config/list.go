package config

import (
	handlers "stamus-ctl/internal/handlers/config"

	"github.com/spf13/cobra"
)

// Command
func listCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List all available configurations",
		Long: `List all available configurations.

Displays all configurations that have been created with 'compose init'.
Each configuration represents a separate deployment environment.

Examples:
  # List all configurations
  stamusctl config list
`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return handlers.ListHandler()
		},
	}
	return cmd
}
