package handlers

import (
	"fmt"
	"path/filepath"
	"strings"

	"stamus-ctl/internal/app"
	"stamus-ctl/internal/logging"
	"stamus-ctl/internal/nix"
	"stamus-ctl/internal/stamus"

	"github.com/spf13/afero"
)

type NixStatusHandlerInputs struct {
	Config string
}

func NixStatusHandler(params NixStatusHandlerInputs) error {
	isNixOS := nix.IsNixOS()

	configPath := params.Config
	if !app.IsCtl() {
		configPath = app.GetConfigsFolder(params.Config)
	}

	fmt.Printf("NixOS system:     %v\n", isNixOS)
	fmt.Printf("Config path:      %s\n", configPath)

	exists, err := afero.DirExists(app.FS, configPath)
	if err != nil || !exists {
		fmt.Printf("Config status:    not found\n")
	} else {
		fmt.Printf("Config status:    present\n")

		version, err := afero.ReadFile(app.FS, filepath.Join(configPath, "version"))
		if err == nil {
			versionString := strings.TrimSpace(strings.Split(string(version), "\n")[0])
			if versionString != "" {
				fmt.Printf("Template version: %s\n", versionString)
			}
		}

		projectName := stamus.GetProjectName(configPath)
		if projectName != "" {
			fmt.Printf("Project:          %s\n", projectName)
		}
	}

	if isNixOS {
		generations, err := nix.ListGenerations()
		if err != nil {
			logging.Sugar.Warnf("Failed to list generations: %v", err)
		} else if generations != "" {
			fmt.Printf("\nNixOS Generations:\n%s", generations)
		}
	}

	return nil
}
