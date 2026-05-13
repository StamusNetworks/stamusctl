package nix

import (
	"strings"

	"stamus-ctl/internal/app"
	"stamus-ctl/internal/logging"
	"stamus-ctl/internal/utils"

	"github.com/spf13/cobra"

	flags "stamus-ctl/internal/handlers"
	handlers "stamus-ctl/internal/handlers/nix"
)

const InitHelp = `Initialize a NixOS configuration.

Creates a new NixOS configuration by downloading templates and setting up
the required files. The generated configuration can then be applied with
'stamusctl nix switch'.

Examples:
  # Initialize with default values
  stamusctl nix init

  # Initialize with specific parameters
  stamusctl nix init suricata.interfaces=eth0

  # Initialize a specific version
  stamusctl nix init --version 1.2.3

  # Initialize with a values file
  stamusctl nix init -v my-values.yaml
`

func initCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:          "init",
		Short:        "Init NixOS config",
		Long:         InitHelp,
		RunE:         initHandler,
		SilenceUsage: true,
	}

	flags.IsDefaultParam.AddAsFlag(cmd, false)
	flags.IsExpert.AddAsFlag(cmd, false)
	flags.Values.AddAsFlag(cmd, false)
	flags.FromFile.AddAsFlag(cmd, false)
	flags.Config.AddAsFlag(cmd, false)
	flags.Template.AddAsFlag(cmd, false)
	flags.Bind.AddAsFlag(cmd, false)
	flags.Version.AddAsFlag(cmd, false)
	flags.Registry.AddAsFlag(cmd, false)

	cmd.AddCommand(ClearNDRCmd())

	return cmd
}

func ClearNDRCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "clearndr",
		Short: "Init ClearNDR NixOS configuration",
		Long:  InitHelp,
		RunE: func(cmd *cobra.Command, args []string) error {
			args = append([]string{"clearndr"}, args...)
			return initHandler(cmd, args)
		},
		SilenceUsage: true,
	}

	flags.IsDefaultParam.AddAsFlag(cmd, false)
	flags.IsExpert.AddAsFlag(cmd, false)
	flags.Values.AddAsFlag(cmd, false)
	flags.FromFile.AddAsFlag(cmd, false)
	flags.Config.AddAsFlag(cmd, false)
	flags.Template.AddAsFlag(cmd, false)
	flags.Version.AddAsFlag(cmd, false)
	flags.Bind.AddAsFlag(cmd, false)
	flags.Registry.AddAsFlag(cmd, false)

	return cmd
}

func initHandler(_ *cobra.Command, args []string) error {
	isDefault, err := flags.IsDefaultParam.GetValue()
	if err != nil {
		logging.Sugar.Error(err)
		return err
	}

	if isDefault.(bool) {
		logging.Sugar.Info("--default flag is deprecated. It is true by default.")
	}

	isExpert, err := flags.IsExpert.GetValue()
	if err != nil {
		logging.Sugar.Error(err)
		return err
	}

	values, err := flags.Values.GetValue()
	if err != nil {
		logging.Sugar.Error(err)
		return err
	}

	fromFile, err := flags.FromFile.GetValue()
	if err != nil {
		logging.Sugar.Error(err)
		return err
	}

	config, err := flags.Config.GetValue()
	if err != nil {
		logging.Sugar.Error(err)
		return err
	}

	templateFolder, err := flags.Template.GetValue()
	if err != nil {
		logging.Sugar.Error(err)
		return err
	}

	version, err := flags.Version.GetValue()
	if err != nil {
		logging.Sugar.Error(err)
		return err
	}

	registry, err := flags.Registry.GetValue()
	if err != nil {
		logging.Sugar.Error(err)
		return err
	}

	bind, err := flags.Bind.GetValue()
	if err != nil {
		logging.Sugar.Error(err)
		return err
	}

	toBind := strings.Split(bind.(string), ",")

	project := "clearndr"
	if len(args) > 0 {
		firstArg := args[0]
		if !strings.Contains(firstArg, "=") {
			args = args[1:]
			project = firstArg
		}
	}

	initParams := handlers.NixInitHandlerInputs{
		IsDefault:        !isExpert.(bool),
		BackupFolderPath: app.DefaultClearNDRPath,
		Arbitrary:        utils.ExtractArgs(args),
		Project:          project,
		Version:          version.(string),
		Values:           values.(string),
		Config:           config.(string),
		FromFile:         fromFile.(string),
		TemplateFolder:   templateFolder.(string),
		Registry:         registry.(string),
		Bind:             toBind,
	}

	return handlers.NixInitHandler(true, initParams)
}
