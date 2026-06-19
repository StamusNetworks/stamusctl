package nix

import (
	"stamus-ctl/internal/logging"

	"github.com/spf13/cobra"

	flags "stamus-ctl/internal/handlers"
	handlers "stamus-ctl/internal/handlers/nix"
)

func diffCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:          "diff",
		Short:        "Show package changes between current system and pending configuration",
		RunE:         diffHandler,
		SilenceUsage: true,
	}

	flags.Config.AddAsFlag(cmd, false)

	return cmd
}

func diffHandler(_ *cobra.Command, _ []string) error {
	conf, err := flags.Config.GetValue()
	if err != nil {
		logging.Sugar.Error(err)
		return err
	}

	return handlers.NixDiffHandler(handlers.NixDiffHandlerInputs{
		Config: conf.(string),
	})
}
