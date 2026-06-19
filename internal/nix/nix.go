package nix

import (
	// Core
	"bytes"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	// Internal
	"stamus-ctl/internal/app"
	"stamus-ctl/internal/logging"

	// External
	"github.com/spf13/afero"
)

const nixosMarker = "/etc/NIXOS"

// execCommand is the function used to create exec.Cmd instances.
// Tests can replace this to avoid running real commands.
var execCommand = exec.Command

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

	cmd := execCommand("nixos-rebuild", args...) //nolint:gosec // action is validated against allowlist above
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

	cmd := execCommand("nix-build", args...) //nolint:gosec // configPath and outputDir are caller-controlled
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("nix-build iso failed: %w", err)
	}
	return nil
}

// RunShellTest executes a shell test script with bash, passing the config
// path as the STAMUSCTL_CONFIG_PATH environment variable.
func RunShellTest(scriptPath string, configPath string) error {
	logging.Sugar.Infow("running shell test", "script", scriptPath, "config", configPath)

	cmd := execCommand("bash", scriptPath) //nolint:gosec // scriptPath validated by caller
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Env = append(os.Environ(), "STAMUSCTL_CONFIG_PATH="+configPath)

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("shell test %s failed: %w", filepath.Base(scriptPath), err)
	}
	return nil
}

// RunNixTest builds a Nix test expression using nix-build.
// The config path is passed as --arg so Nix expressions can access it.
// This supports NixOS integration tests (e.g. pkgs.testers.runNixOSTest)
// as well as any derivation-producing test expression.
func RunNixTest(scriptPath string, configPath string) error {
	logging.Sugar.Infow("running nix test", "script", scriptPath, "config", configPath)

	args := []string{
		scriptPath,
		"--arg", "configPath", fmt.Sprintf("%q", configPath),
		"--no-out-link",
	}

	cmd := execCommand("nix-build", args...) //nolint:gosec // scriptPath validated by caller
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Env = append(os.Environ(), "STAMUSCTL_CONFIG_PATH="+configPath)

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("nix test %s failed: %w", filepath.Base(scriptPath), err)
	}
	return nil
}

// RunISO launches a QEMU virtual machine that boots from the given ISO file.
// memory is the amount of RAM in megabytes, cores is the number of CPU cores,
// and enableKVM enables hardware acceleration when true.
func RunISO(isoPath string, memory int, cores int, enableKVM bool) error {
	args := []string{
		"-cdrom", isoPath,
		"-m", fmt.Sprintf("%d", memory),
		"-smp", fmt.Sprintf("%d", cores),
		"-boot", "d",
	}
	if enableKVM {
		args = append(args, "-enable-kvm")
	}

	logging.Sugar.Infow("launching qemu", "iso", isoPath, "memory", memory, "cores", cores, "kvm", enableKVM)

	cmd := execCommand("qemu-system-x86_64", args...) //nolint:gosec // arguments are caller-controlled
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("qemu failed: %w", err)
	}
	return nil
}

// DiffClosures runs `nix store diff-closures` between two store paths to show
// package-level differences. currentSystem is typically "/run/current-system"
// and newSystem is the result of a `nixos-rebuild build`.
func DiffClosures(currentSystem string, newSystem string) error {
	args := []string{"store", "diff-closures", currentSystem, newSystem}

	logging.Sugar.Infow("running nix store diff-closures", "current", currentSystem, "new", newSystem)

	cmd := execCommand("nix", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("nix store diff-closures failed: %w", err)
	}
	return nil
}

// ListGenerations runs `nixos-rebuild list-generations` and returns its output
// as a string.
func ListGenerations() (string, error) {
	args := []string{"list-generations"}

	logging.Sugar.Infow("listing NixOS generations")

	cmd := execCommand("nixos-rebuild", args...)
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("nixos-rebuild list-generations failed: %w", err)
	}
	return out.String(), nil
}

// FindAndRunVM locates the run-*-vm script inside resultDir/bin/ and executes it.
func FindAndRunVM(resultDir string) error {
	binDir := filepath.Join(resultDir, "bin")
	pattern := filepath.Join(binDir, "run-*-vm")

	matches, err := filepath.Glob(pattern)
	if err != nil {
		return fmt.Errorf("failed to glob VM scripts in %s: %w", binDir, err)
	}
	if len(matches) == 0 {
		return fmt.Errorf("no run-*-vm script found in %s", binDir)
	}

	scriptPath := matches[0]
	logging.Sugar.Infow("running built VM", "script", scriptPath)

	cmd := execCommand(scriptPath)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("VM script %s failed: %w", filepath.Base(scriptPath), err)
	}
	return nil
}

// Infect is reserved for a future feature that converts a running Linux system
// to NixOS via nix-infect. It is not yet implemented.
func Infect(_ string) error {
	return errors.New("nix infect is not yet implemented")
}
