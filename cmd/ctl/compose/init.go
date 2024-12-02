package compose

import (
	// External

	"strings"

	"github.com/spf13/cobra"

	// Internal
	"stamus-ctl/internal/app"
	"stamus-ctl/internal/embeds"
	flags "stamus-ctl/internal/handlers"
	handlers "stamus-ctl/internal/handlers/compose"
	"stamus-ctl/internal/logging"
	"stamus-ctl/internal/utils"
)

// Const
const (
	InitHelp = `Init compose config file
	You can set parameters in the config file or pass them as arguments.
	(ex: parameter.subparameter=value)
	`
)

// Commands
func initCmd() *cobra.Command {
	// Setup
	embeds.InitClearNDRFolder(app.DefaultClearNDRPath)
	// Command
	cmd := &cobra.Command{
		Use:   "init",
		Short: "Init compose config file",
		Long:  InitHelp,
		RunE: func(cmd *cobra.Command, args []string) error {
			return handler(cmd, args)
		},
	}
	// Flags
	flags.IsDefaultParam.AddAsFlag(cmd, false)
	flags.IsExpert.AddAsFlag(cmd, false)
	flags.Values.AddAsFlag(cmd, false)
	flags.FromFile.AddAsFlag(cmd, false)
	flags.Config.AddAsFlag(cmd, false)
	flags.Template.AddAsFlag(cmd, false)
	flags.Bind.AddAsFlag(cmd, false)
	flags.Version.AddAsFlag(cmd, false)

	// Commands
	cmd.AddCommand(ClearNDRCmd())
	return cmd
}

func ClearNDRCmd() *cobra.Command {
	// Setup
	embeds.InitClearNDRFolder(app.DefaultClearNDRPath)
	// Command
	cmd := &cobra.Command{
		Use:   "clearndr",
		Short: "Init ClearNDR container compose file",
		Long:  InitHelp,
		RunE: func(cmd *cobra.Command, args []string) error {
			args = append([]string{"clearndr"}, args...)
			return handler(cmd, args)
		},
	}
	// Flags
	flags.IsDefaultParam.AddAsFlag(cmd, false)
	flags.IsExpert.AddAsFlag(cmd, false)
	flags.Values.AddAsFlag(cmd, false)
	flags.FromFile.AddAsFlag(cmd, false)
	flags.Config.AddAsFlag(cmd, false)
	flags.Template.AddAsFlag(cmd, false)
	flags.Version.AddAsFlag(cmd, false)
	flags.Bind.AddAsFlag(cmd, false)
	return cmd
}

func handler(_ *cobra.Command, args []string) error {
	// Get flags
	isDefault, err := flags.IsDefaultParam.GetValue()
	if err != nil {
		return err
	}

	if isDefault.(bool) {
		logging.Sugar.Info("--default flag is deprecated. It is true by default.")
	}

	isExpert, err := flags.IsExpert.GetValue()
	if err != nil {
		return err
	}

	values, err := flags.Values.GetValue()
	if err != nil {
		return err
	}
	fromFile, err := flags.FromFile.GetValue()
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
	version, err := flags.Version.GetValue()
	if err != nil {
		return err
	}
	bind, err := flags.Bind.GetValue()
	if err != nil {
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

	// Call handler
	initParams := handlers.InitHandlerInputs{
		IsDefault:        !isExpert.(bool),
		BackupFolderPath: app.DefaultClearNDRPath,
		Arbitrary:        utils.ExtractArgs(args),
		Project:          project,
		Version:          version.(string),
		Values:           values.(string),
		Config:           config.(string),
		FromFile:         fromFile.(string),
		TemplateFolder:   templateFolder.(string),
		Bind:             toBind,
	}
	return handlers.InitHandler(true, initParams)
}
