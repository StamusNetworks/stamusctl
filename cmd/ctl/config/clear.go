package config

import (
	// Core

	// External

	"github.com/spf13/cobra"

	// Internal

	flags "stamus-ctl/internal/handlers"
	handlers "stamus-ctl/internal/handlers/config"
)

// Command
func clearCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "clear",
		Short: "Clear containers, volumes, networks and files",
		Long: `Clear containers, volumes, networks and files.

Removes all Docker resources (containers, volumes, networks) and
configuration files associated with a deployment. This is a
destructive operation.

Examples:
  # Clear the default configuration
  stamusctl config clear

  # Clear a specific configuration
  stamusctl config clear -c myconfig
`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return clearHandler()
		},
	}
	// Flags
	flags.Config.AddAsFlag(cmd, false)
	return cmd
}

func clearHandler() error {
	// Get properties
	conf, err := flags.Config.GetValue()
	if err != nil {
		return err
	}
	// Clear
	return handlers.Clear(conf.(string))
}
