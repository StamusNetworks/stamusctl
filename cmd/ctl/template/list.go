package config

import (
	handlers "stamus-ctl/internal/handlers/template"

	"github.com/spf13/cobra"
)

// Command
func listCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "Get list of available templates",
		RunE: func(cmd *cobra.Command, args []string) error {
			return handlers.ListHandler()
		},
	}
	return cmd
}
