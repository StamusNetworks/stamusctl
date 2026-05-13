package nix

import (
	"stamus-ctl/internal/logging"

	"github.com/spf13/cobra"

	flags "stamus-ctl/internal/handlers"
	handlers "stamus-ctl/internal/handlers/nix"
)

func infectCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:          "infect",
		Short:        "Convert current system to NixOS (destructive, advanced)",
		RunE:         infectHandler,
		SilenceUsage: true,
	}

	flags.Config.AddAsFlag(cmd, false)

	return cmd
}

func infectHandler(_ *cobra.Command, _ []string) error {
	conf, err := flags.Config.GetValue()
	if err != nil {
		logging.Sugar.Error(err)
		return err
	}

	return handlers.NixInfectHandler(handlers.NixInfectHandlerInputs{
		Config: conf.(string),
	})
}
