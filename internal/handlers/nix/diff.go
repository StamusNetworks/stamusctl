package handlers

import (
	"fmt"
	"os"

	"stamus-ctl/internal/app"
	"stamus-ctl/internal/logging"
	"stamus-ctl/internal/nix"

	"github.com/spf13/afero"
)

type NixDiffHandlerInputs struct {
	Config string
}

func NixDiffHandler(params NixDiffHandlerInputs) error {
	logger := logging.Sugar.With("Config", params.Config)

	if !nix.IsNixOS() {
		return fmt.Errorf("this system is not running NixOS — use 'stamusctl nix infect' to convert it first")
	}

	configPath := params.Config
	if !app.IsCtl() {
		configPath = app.GetConfigsFolder(params.Config)
	}

	exists, err := afero.DirExists(app.FS, configPath)
	if err != nil || !exists {
		return fmt.Errorf("configuration %q not found — run 'stamusctl nix init' first", params.Config)
	}

	logger.Info("Building configuration to compute diff...")
	if err := nix.NixosRebuild(configPath, "build"); err != nil {
		return err
	}

	// Clean up the ./result symlink created by nixos-rebuild build
	defer os.Remove("./result")

	if err := nix.DiffClosures("/run/current-system", "./result"); err != nil {
		return err
	}

	logger.Info("Diff complete")
	return nil
}
