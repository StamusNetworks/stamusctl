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

  # Preview what would change before switching
  stamusctl nix diff

  # Update templates to a newer version
  stamusctl nix update --version 1.2.3

  # Build a throwaway VM to test configuration
  stamusctl nix build-vm --run

  # Show configuration and system status
  stamusctl nix status

  # Generate an installable ISO
  stamusctl nix iso

  # Run the ISO in QEMU
  stamusctl nix iso-run

  # Convert an existing system to NixOS (advanced)
  stamusctl nix infect
`,
	}

	cmd.AddCommand(initCmd())
	cmd.AddCommand(switchCmd())
	cmd.AddCommand(testCmd())
	cmd.AddCommand(isoCmd())
	cmd.AddCommand(isoRunCmd())
	cmd.AddCommand(infectCmd())
	cmd.AddCommand(diffCmd())
	cmd.AddCommand(updateCmd())
	cmd.AddCommand(buildVMCmd())
	cmd.AddCommand(statusCmd())

	return cmd
}
