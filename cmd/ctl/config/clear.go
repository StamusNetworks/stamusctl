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
		Use:   "clear [flags...]",
		Short: "Clears containers, volumes, networks and files",
		Args:  cobra.ArbitraryArgs,
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
