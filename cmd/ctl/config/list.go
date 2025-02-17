package config

import (
	handlers "stamus-ctl/internal/handlers/config"

	"github.com/spf13/cobra"
)

// Command
func listCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "Get list of configurations",
		RunE: func(cmd *cobra.Command, args []string) error {
			return handlers.ListHandler()
		},
	}
	return cmd
}
