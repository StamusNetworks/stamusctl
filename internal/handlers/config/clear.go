package config

import (
	// Core

	"fmt"
	"os"
	"path/filepath"

	// Internal
	"stamus-ctl/internal/app"
	"stamus-ctl/internal/backup"
	"stamus-ctl/internal/handlers/wrapper"
	"stamus-ctl/internal/logging"
	"stamus-ctl/internal/stamus"
	"stamus-ctl/internal/validation"

	// External
	"go.uber.org/zap"
)

func Clear(conf string) error {
	// Get config name from path
	configName := conf
	if !app.IsCtl() {
		conf = app.GetConfigsFolder(conf)
	} else {
		// Extract config name from path
		configName = filepath.Base(conf)
	}

	// Create automatic backup before destructive operation
	_, err := backup.CreateBackup(configName, backup.BackupTypeAuto, logging.Logger)
	if err != nil {
		// Log warning but continue with operation
		logging.Logger.Warn("Failed to create backup before clear operation",
			zap.String("config", configName),
			zap.Error(err),
		)
	}

	// Down containers
	err = wrapper.HandleDown(conf, true, true)
	if err != nil {
		return err
	}
	// Delete folder
	err = deleteFolder(conf)
	if err != nil {
		return err
	}
	// Save stamus
	err = stamus.RemoveInstance(conf)
	return err
}

func deleteFolder(conf string) error {
	// Validate the path to prevent directory traversal and deletion of arbitrary folders
	// Only allow deletion within the configs folder
	configsFolder := app.ConfigsFolder
	if !app.IsCtl() {
		// For daemon mode, validate against the base configs folder
		if configsFolder == "" {
			return fmt.Errorf("configs folder not configured")
		}
	}

	// Sanitize and validate the path is within the allowed directory
	sanitizedPath, err := validation.SanitizePath(conf, configsFolder)
	if err != nil {
		return fmt.Errorf("invalid config path: %w", err)
	}

	// Additional safety check: ensure path is not root or system directories
	if sanitizedPath == "/" || sanitizedPath == "/etc" || sanitizedPath == "/usr" ||
		sanitizedPath == "/var" || sanitizedPath == "/home" || sanitizedPath == "" {
		return fmt.Errorf("refusing to delete protected directory: %s", sanitizedPath)
	}

	// Use os.RemoveAll instead of shell command to prevent command injection
	err = os.RemoveAll(sanitizedPath)
	if err != nil {
		fmt.Println("Error:", err)
		return err
	}
	return nil
}
