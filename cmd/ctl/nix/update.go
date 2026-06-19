package nix

import (
	"github.com/spf13/cobra"

	flags "stamus-ctl/internal/handlers"
	handlers "stamus-ctl/internal/handlers/nix"
)

func updateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:          "update",
		Short:        "Update NixOS configuration templates",
		RunE:         updateHandler,
		SilenceUsage: true,
	}

	flags.Version.AddAsFlag(cmd, false)
	flags.Config.AddAsFlag(cmd, false)
	flags.Template.AddAsFlag(cmd, false)
	flags.IsInteractive.AddAsFlag(cmd, false)

	return cmd
}

func updateHandler(_ *cobra.Command, args []string) error {
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

	return handlers.NixUpdateHandler(handlers.NixUpdateHandlerInputs{
		Version:        version.(string),
		Config:         config.(string),
		TemplateFolder: templateFolder.(string),
		Args:           args,
		Interactive:    interactive.(bool),
	})
}
