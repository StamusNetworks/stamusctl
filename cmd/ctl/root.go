package ctl

import (
	"fmt"
	"os"
	"os/signal"
	"sync/atomic"
	"syscall"
	"time"

	"stamus-ctl/cmd/ctl/backup"
	"stamus-ctl/cmd/ctl/completion"
	"stamus-ctl/cmd/ctl/compose"
	"stamus-ctl/cmd/ctl/config"
	tmpl "stamus-ctl/cmd/ctl/template"
	"stamus-ctl/internal/logging"
	"stamus-ctl/internal/models"
	"stamus-ctl/internal/shutdown"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	// interruptCount tracks consecutive SIGINT signals for force quit
	interruptCount atomic.Int32
	// lastInterruptTime tracks when the last interrupt was received
	lastInterruptTime atomic.Int64
)

// Entry point
func Execute() {
	// Initialize shutdown manager for CLI
	shutdown.Init(logging.Logger)

	// Setup signal handling for CLI
	setupCLISignalHandler()

	// Run
	if err := rootCmd().Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(shutdown.ExitError)
	}
}

// setupCLISignalHandler sets up signal handling for CLI commands.
// First SIGINT begins graceful shutdown, second SIGINT within 2 seconds forces exit.
func setupCLISignalHandler() {
	sigChan := make(chan os.Signal, 2)
	signal.Notify(sigChan, syscall.SIGTERM, syscall.SIGINT)

	go func() {
		for sig := range sigChan {
			now := time.Now().UnixNano()
			lastTime := lastInterruptTime.Swap(now)

			switch sig {
			case syscall.SIGINT:
				// Check for double-SIGINT force quit (within 2 seconds)
				if now-lastTime < int64(2*time.Second) {
					count := interruptCount.Add(1)
					if count >= 1 {
						fmt.Fprintln(os.Stderr, "\nForce quit")
						os.Exit(shutdown.ExitSIGINT)
					}
				} else {
					interruptCount.Store(0)
				}

				if !shutdown.IsShuttingDown() {
					fmt.Fprintln(os.Stderr, "\nInterrupted. Completing current operation... (press Ctrl+C again to force quit)")
					shutdown.GetManager().TriggerShutdown(shutdown.ExitSIGINT)
				}

			case syscall.SIGTERM:
				if !shutdown.IsShuttingDown() {
					fmt.Fprintln(os.Stderr, "\nTerminating...")
					shutdown.GetManager().TriggerShutdown(shutdown.ExitSIGTERM)
				}
			}
		}
	}()
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
		Use:   "stamusctl",
		Short: "Stamus Networks control tool for managing configurations and services",
		Long: `stamusctl is a command-line tool for managing Stamus Networks configurations,
Docker Compose deployments, backups, and templates.

Use "stamusctl [command] --help" for more information about a command.`,
		SuggestionsMinimumDistance: 2,
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
	cmd.AddCommand(backup.BackupCmd())
	cmd.AddCommand(completion.CompletionCmd())
	return cmd
}

// RootCmdForDoc returns the root command for documentation generation.
// This is used by the doc generator to create man pages.
func RootCmdForDoc() *cobra.Command {
	return rootCmd()
}
