package daemon

import (
	"fmt"
	"os"
	"runtime/debug"

	"stamus-ctl/cmd/daemon/run"
	"stamus-ctl/internal/logging"
	"stamus-ctl/internal/models"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// @securityDefinitions.basic  BasicAuth

// @title           Swagger Stamusd API
// @version         1.0
// @description     Stamus daemon server.

// @BasePath /api/v1

// Entry point
func Execute() {
	// Setup
	viper.Set("verbose", 3)
	logging.SetLogger()

	// Run
	if err := rootCmd().Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		debug.PrintStack()
		panic(err)
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
		Use: "stamusd",
	}
	// Common flags
	verbose.AddAsFlag(cmd, true)
	// SubCommands
	cmd.AddCommand(versionCmd())
	cmd.AddCommand(run.RunCmd())
	return cmd
}
