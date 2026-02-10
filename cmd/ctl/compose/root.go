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
		Short: "Manage Docker Compose deployments",
		Long: `Manage Docker Compose deployments.

The compose command provides tools for initializing, updating, and
managing Docker Compose-based deployments. It wraps standard Docker
Compose commands and adds Stamus-specific functionality.

Examples:
  # Initialize a new deployment
  stamusctl compose init

  # Update an existing deployment
  stamusctl compose update

  # Start services
  stamusctl compose up -d

  # Stop services
  stamusctl compose down

  # View service logs
  stamusctl compose logs -f
`,
	}

	// Custom commands
	cmd.AddCommand(initCmd())
	cmd.AddCommand(updateCmd())
	cmd.AddCommand(readPcapCmd())
	// Docker commands
	wrappedCmds, _ := wrappedCmd()
	cmd.AddCommand(wrappedCmds...)

	return cmd
}
