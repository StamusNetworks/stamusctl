package config

import (
	// Common

	// External

	"github.com/spf13/cobra"
	// Custom
	"stamus-ctl/internal/app"
	"stamus-ctl/internal/embeds"
)

// Init
func init() {
	// Setup
	embeds.InitClearNDRFolder(app.DefaultClearNDRPath)
}

// Commands
func ConfigCmd() *cobra.Command {
	// Create command
	cmd := &cobra.Command{
		Use:   "config",
		Short: "Manage configuration parameters",
		Long: `Manage configuration parameters.

The config command allows you to view, modify, and manage configuration
parameters for your deployments.

Examples:
  # List all configurations
  stamusctl config list

  # Get all configuration values
  stamusctl config get

  # Get a specific value
  stamusctl config get scirius.token

  # Set a configuration value
  stamusctl config set scirius.token=MyToken

  # View configuration version
  stamusctl config version
`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return getHandler(cmd, args)
		},
	}
	// Add Commands
	cmd.AddCommand(getCmd())
	cmd.AddCommand(setCmd())
	cmd.AddCommand(versionCmd())
	cmd.AddCommand(clearCmd())
	cmd.AddCommand(listCmd())
	return cmd
}
