package nix

import (
	"stamus-ctl/internal/logging"

	"github.com/spf13/cobra"

	flags "stamus-ctl/internal/handlers"
	handlers "stamus-ctl/internal/handlers/nix"
)

func switchCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:          "switch",
		Short:        "Apply NixOS configuration via nixos-rebuild switch",
		RunE:         switchHandler,
		SilenceUsage: true,
	}

	flags.Config.AddAsFlag(cmd, false)

	return cmd
}

func switchHandler(_ *cobra.Command, _ []string) error {
	conf, err := flags.Config.GetValue()
	if err != nil {
		logging.Sugar.Error(err)
		return err
	}

	return handlers.NixSwitchHandler(handlers.NixSwitchHandlerInputs{
		Config: conf.(string),
	})
}
