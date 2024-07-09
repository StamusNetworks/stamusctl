package compose

import (
	// Common
	// External
	"github.com/spf13/cobra"
)

// Commands
func ComposeCmd() *cobra.Command {
	// Create command
	cmd := &cobra.Command{
		Use:   "compose",
		Short: "Create container compose file",
	}

	// Custom commands
	cmd.AddCommand(initCmd())
	cmd.AddCommand(configCmd())
	cmd.AddCommand(updateCmd())
	// Docker commands
	wrappedCmds, _ := wrappedCmd()
	cmd.AddCommand(wrappedCmds...)

	return cmd
}
