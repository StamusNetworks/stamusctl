package compose

import (
	flags "stamus-ctl/internal/handlers"
	handlers "stamus-ctl/internal/handlers/compose"

	// Common

	// External

	"github.com/spf13/cobra"
	// Custom
)

// Commands
func updateCmd() *cobra.Command {
	// Create cmd
	cmd := &cobra.Command{
		Use:   "update",
		Short: "Update compose configuration files",
		RunE:  updateHandler,
	}
	// Add flags
	flags.Version.AddAsFlag(cmd, false)
	flags.Config.AddAsFlag(cmd, false)
	flags.Template.AddAsFlag(cmd, false)
	flags.IsInteractive.AddAsFlag(cmd, false)

	return cmd
}

func updateHandler(_ *cobra.Command, args []string) error {
	// Validate flags
	version, err := flags.Version.GetValue()
	if err != nil {
		return err
	}
	config, err := flags.Config.GetValue()
	if err != nil {
		return err
	}
	templateFolder, err := flags.Template.GetValue()
	if err != nil {
		return err
	}
	interactive, err := flags.IsInteractive.GetValue()
	if err != nil {
		return err
	}

	// Call handler
	params := handlers.UpdateHandlerParams{
		Version:        version.(string),
		Config:         config.(string),
		TemplateFolder: templateFolder.(string),
		Args:           args,
		Interactive:    interactive.(bool),
	}

	return handlers.UpdateHandler(params)
}
