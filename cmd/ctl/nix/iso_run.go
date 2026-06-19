package nix

import (
	"stamus-ctl/internal/logging"

	"github.com/spf13/cobra"

	flags "stamus-ctl/internal/handlers"
	handlers "stamus-ctl/internal/handlers/nix"
)

func isoRunCmd() *cobra.Command {
	var (
		output    string
		memory    int
		cores     int
		enableKVM bool
	)

	cmd := &cobra.Command{
		Use:          "iso-run",
		Short:        "Run a NixOS ISO image in QEMU",
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, _ []string) error {
			conf, err := flags.Config.GetValue()
			if err != nil {
				logging.Sugar.Error(err)
				return err
			}

			return handlers.NixISORunHandler(handlers.NixISORUnHandlerInputs{
				Config:    conf.(string),
				Output:    output,
				Memory:    memory,
				Cores:     cores,
				EnableKVM: enableKVM,
			})
		},
	}

	flags.Config.AddAsFlag(cmd, false)
	cmd.Flags().StringVarP(&output, "output", "o", ".",
		"Directory containing the ISO build result")
	cmd.Flags().IntVarP(&memory, "memory", "m", 4096, "Amount of RAM in megabytes")
	cmd.Flags().IntVar(&cores, "cores", 2, "Number of CPU cores")
	cmd.Flags().BoolVar(&enableKVM, "kvm", true, "Enable KVM hardware acceleration")

	return cmd
}
