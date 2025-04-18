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
func TemplateCmd() *cobra.Command {
	// Create command
	cmd := &cobra.Command{
		Use:   "template",
		Short: "Interact with compose config file",
	}
	// Add Commands
	cmd.AddCommand(keysCmd())
	return cmd
}
