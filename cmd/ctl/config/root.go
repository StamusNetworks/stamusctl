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
		Short: "Interact with compose config file",
		RunE: func(cmd *cobra.Command, args []string) error {
			return getHandler(cmd, args)
		},
	}
	// Add Commands
	cmd.AddCommand(getCmd())
	cmd.AddCommand(setCmd())
	cmd.AddCommand(versionCmd())
	return cmd
}
