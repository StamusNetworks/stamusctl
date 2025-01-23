package compose

import (
	// Common

	// External

	"github.com/spf13/cobra"

	// Custom
	flags "stamus-ctl/internal/handlers"
	handlers "stamus-ctl/internal/handlers/compose"
)

// Commands
func updateCmd() *cobra.Command {
	// Create cmd
	cmd := &cobra.Command{
		Use:   "update",
		Short: "Update compose configuration files",
		RunE: func(cmd *cobra.Command, args []string) error {
			updateHandler(cmd, args)
			return nil
		},
	}
	// Add flags
	flags.Version.AddAsFlag(cmd, false)
	flags.Config.AddAsFlag(cmd, false)
	return cmd
}

func updateHandler(_ *cobra.Command, args []string) {
	// Validate flags
	version, err := flags.Version.GetValue()
	if err != nil {
		return
	}
	config, err := flags.Config.GetValue()
	if err != nil {
		return
	}
	// Call handler
	params := handlers.UpdateHandlerParams{
		Version:        version.(string),
		Config:         config.(string),
		Args:           args,
		TemplateFolder: "",
	}

	handlers.UpdateHandler(params)
}
