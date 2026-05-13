package nix

import "github.com/spf13/cobra"

func NixCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "nix",
		Short: "Manage NixOS-based deployments",
		Long: `Manage NixOS-based deployments.

The nix command provides tools for initializing, switching, and
managing NixOS-based Stamus appliance deployments. It generates
NixOS configuration from templates and applies it via nixos-rebuild.

Examples:
  # Initialize a new NixOS deployment
  stamusctl nix init

  # Apply NixOS configuration
  stamusctl nix switch

  # Generate an installable ISO
  stamusctl nix iso

  # Convert an existing system to NixOS (advanced)
  stamusctl nix infect
`,
	}

	cmd.AddCommand(initCmd())
	cmd.AddCommand(switchCmd())
	cmd.AddCommand(isoCmd())
	cmd.AddCommand(infectCmd())

	return cmd
}
