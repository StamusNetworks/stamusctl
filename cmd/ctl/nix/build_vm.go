package nix

import (
	"stamus-ctl/internal/logging"

	"github.com/spf13/cobra"

	flags "stamus-ctl/internal/handlers"
	handlers "stamus-ctl/internal/handlers/nix"
)

func buildVMCmd() *cobra.Command {
	var run bool

	cmd := &cobra.Command{
		Use:          "build-vm",
		Short:        "Build a throwaway QEMU VM from NixOS configuration",
		SilenceUsage: true,
		RunE: func(_ *cobra.Command, _ []string) error {
			conf, err := flags.Config.GetValue()
			if err != nil {
				logging.Sugar.Error(err)
				return err
			}

			return handlers.NixBuildVMHandler(handlers.NixBuildVMHandlerInputs{
				Config: conf.(string),
				Run:    run,
			})
		},
	}

	flags.Config.AddAsFlag(cmd, false)
	cmd.Flags().BoolVar(&run, "run", false, "Automatically launch the VM after building")

	return cmd
}
