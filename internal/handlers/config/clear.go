package config

import (
	// Core

	"fmt"
	"os"

	// Internal
	"stamus-ctl/internal/app"
	"stamus-ctl/internal/handlers/wrapper"
	"stamus-ctl/internal/stamus"
	"stamus-ctl/internal/validation"
)

func Clear(conf string) error {
	// File instance
	if !app.IsCtl() {
		conf = app.GetConfigsFolder(conf)
	}
	// Down containers
	err := wrapper.HandleDown(conf, true, true)
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
