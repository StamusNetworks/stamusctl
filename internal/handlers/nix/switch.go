package handlers

import (
	"fmt"
	"path/filepath"

	"stamus-ctl/internal/app"
	"stamus-ctl/internal/backup"
	"stamus-ctl/internal/logging"
	"stamus-ctl/internal/nix"

	"github.com/spf13/afero"
	"go.uber.org/zap"
)

type NixSwitchHandlerInputs struct {
	Config string
}

func NixSwitchHandler(params NixSwitchHandlerInputs) error {
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

	configName := filepath.Base(configPath)
	_, err = backup.CreateBackup(configName, backup.BackupTypeAuto, logging.Logger)
	if err != nil {
		logging.Logger.Warn("Failed to create backup before nix switch",
			zap.String("config", configName),
			zap.Error(err),
		)
	}

	if err := nix.NixosRebuild(configPath, "switch"); err != nil {
		return err
	}

	logger.Info("NixOS configuration applied successfully")
	return nil
}
