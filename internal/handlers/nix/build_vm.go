package handlers

import (
	"fmt"

	"stamus-ctl/internal/app"
	"stamus-ctl/internal/logging"
	"stamus-ctl/internal/nix"

	"github.com/spf13/afero"
)

type NixBuildVMHandlerInputs struct {
	Config string
	Run    bool
}

func NixBuildVMHandler(params NixBuildVMHandlerInputs) error {
	logger := logging.Sugar.With("Config", params.Config, "Run", params.Run)

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

	if err := nix.NixosRebuild(configPath, "build-vm"); err != nil {
		return err
	}

	logger.Info("VM built successfully — result available at ./result")

	if params.Run {
		logger.Info("Launching VM...")
		if err := nix.FindAndRunVM("./result"); err != nil {
			return err
		}
	}

	return nil
}
