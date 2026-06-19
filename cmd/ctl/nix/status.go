package nix

import (
	"stamus-ctl/internal/logging"

	"github.com/spf13/cobra"

	flags "stamus-ctl/internal/handlers"
	handlers "stamus-ctl/internal/handlers/nix"
)

func statusCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:          "status",
		Short:        "Show NixOS configuration and system status",
		RunE:         statusHandler,
		SilenceUsage: true,
	}

	flags.Config.AddAsFlag(cmd, false)

	return cmd
}

func statusHandler(_ *cobra.Command, _ []string) error {
	conf, err := flags.Config.GetValue()
	if err != nil {
		logging.Sugar.Error(err)
		return err
	}

	return handlers.NixStatusHandler(handlers.NixStatusHandlerInputs{
		Config: conf.(string),
	})
}
