package nix

import (
	"stamus-ctl/internal/logging"

	"github.com/spf13/cobra"

	flags "stamus-ctl/internal/handlers"
	handlers "stamus-ctl/internal/handlers/nix"
)

func isoCmd() *cobra.Command {
	var output string

	cmd := &cobra.Command{
		Use:          "iso",
		Short:        "Generate a NixOS ISO image from configuration",
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, _ []string) error {
			conf, err := flags.Config.GetValue()
			if err != nil {
				logging.Sugar.Error(err)
				return err
			}

			return handlers.NixISOHandler(handlers.NixISOHandlerInputs{
				Config: conf.(string),
				Output: output,
			})
		},
	}

	flags.Config.AddAsFlag(cmd, false)
	cmd.Flags().StringVarP(&output, "output", "o", ".", "Output directory for the ISO image")

	return cmd
}
