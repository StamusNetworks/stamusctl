package nix

import (
	// Core
	"errors"
	"fmt"
	"os"
	"os/exec"

	// Internal
	"stamus-ctl/internal/app"
	"stamus-ctl/internal/logging"

	// External
	"github.com/spf13/afero"
)

const nixosMarker = "/etc/NIXOS"

// IsNixOS reports whether the current system is NixOS by checking for the
// presence of /etc/NIXOS.
func IsNixOS() bool {
	exists, err := afero.Exists(app.FS, nixosMarker)
	if err != nil {
		logging.Sugar.Infow("could not check for NixOS marker", "path", nixosMarker, "error", err)
		return false
	}
	return exists
}

// NixosRebuild runs `nixos-rebuild <action>` with the configuration located at
// <configPath>/configuration.nix. action must be one of: switch, boot, test,
// build, build-vm.
func NixosRebuild(configPath string, action string) error {
	validActions := map[string]bool{
		"switch":   true,
		"boot":     true,
		"test":     true,
		"build":    true,
		"build-vm": true,
	}
	if !validActions[action] {
		return fmt.Errorf("invalid nixos-rebuild action %q: must be one of switch, boot, test, build, build-vm", action)
	}

	nixosConfig := configPath + "/configuration.nix"
	args := []string{
		action,
		"-I", "nixos-config=" + nixosConfig,
	}

	logging.Sugar.Infow("running nixos-rebuild", "action", action, "config", nixosConfig)

	cmd := exec.Command("nixos-rebuild", args...) //nolint:gosec // action is validated against allowlist above
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("nixos-rebuild %s failed: %w", action, err)
	}
	return nil
}

// BuildISO runs `nix-build` to produce a NixOS ISO image from the configuration
// at <configPath>/iso.nix, placing the result symlink under <outputDir>/result.
func BuildISO(configPath string, outputDir string) error {
	isoConfig := configPath + "/iso.nix"
	outLink := outputDir + "/result"
	args := []string{
		"<nixpkgs/nixos>",
		"-A", "config.system.build.isoImage",
		"-I", "nixos-config=" + isoConfig,
		"--out-link", outLink,
	}

	logging.Sugar.Infow("running nix-build iso", "config", isoConfig, "output", outLink)

	cmd := exec.Command("nix-build", args...) //nolint:gosec // configPath and outputDir are caller-controlled
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("nix-build iso failed: %w", err)
	}
	return nil
}

// Infect is reserved for a future feature that converts a running Linux system
// to NixOS via nix-infect. It is not yet implemented.
func Infect(_ string) error {
	return errors.New("nix infect is not yet implemented")
}
