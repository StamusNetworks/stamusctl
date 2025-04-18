package ctl

import (
	"fmt"
	"os"

	"stamus-ctl/cmd/ctl/compose"
	"stamus-ctl/cmd/ctl/config"
	tmpl "stamus-ctl/cmd/ctl/template"
	"stamus-ctl/internal/logging"
	"stamus-ctl/internal/models"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// Entry point
func Execute() {
	// Run
	if err := rootCmd().Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

// Flags
var verbose = models.Parameter{
	Name:    "verbose",
	Type:    "int",
	Default: models.CreateVariableInt(0),
	Usage:   "Verbosity level",
}

// Commands
func rootCmd() *cobra.Command {
	// Create command
	cmd := &cobra.Command{
		Use: "stamusctl",
	}
	// Common flags
	verbose.AddAsFlag(cmd, true)
	viper.BindPFlag("verbose", cmd.Flags().Lookup("verbose"))
	viper.BindEnv("verbose", "STAMUS_VERBOSE")

	logging.SetLogger()
	// SubCommands
	cmd.AddCommand(versionCmd())
	cmd.AddCommand(loginCmd())
	cmd.AddCommand(compose.ComposeCmd())
	cmd.AddCommand(config.ConfigCmd())
	cmd.AddCommand(tmpl.TemplateCmd())
	return cmd
}
