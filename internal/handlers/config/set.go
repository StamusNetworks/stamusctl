package config

import (
	// Core
	"fmt"
	"path/filepath"
	"strings"

	// Internal

	"stamus-ctl/internal/app"
	wrapper "stamus-ctl/internal/handlers/wrapper"
	"stamus-ctl/internal/models"
	"stamus-ctl/internal/utils"
	// External
)

type SetHandlerInputs struct {
	Values   string   // Path to the values.yaml file
	Reload   bool     // Reload the configuration, don't keep arbitrary parameters
	Apply    bool     // Apply the new configuration, reload the services
	Args     []string // Cmd arguments
	FromFile string   // Path to the file containing the values
	Config   string   // Config name
}

// func SetHandler(configPath string, args []string, reload bool, apply bool) error {
func SetHandler(params SetHandlerInputs) error {
	// Load the config
	file, err := models.CreateFile(params.Config, "values.yaml")
	if err != nil {
		return err
	}
	config, err := models.LoadConfigFrom(file, params.Reload)
	if err != nil {
		return err
	}
	// Extract and set parameters from args
	paramsArgs := utils.ExtractArgs(params.Args)
	err = config.GetParams().SetLooseValues(paramsArgs)
	if err != nil {
		return err
	}
	config.GetArbitrary().SetArbitrary(paramsArgs)
	err = config.GetParams().ProcessOptionnalParams(false)
	if err != nil {
		return err
	}
	// Set values from file
	err = config.SetValuesFromFiles(params.FromFile)
	if err != nil {
		return err
	}
	err = config.SetValuesFromFile(params.Values)
	if err != nil {
		return err
	}
	// Validate
	err = config.GetParams().ValidateAll()
	if err != nil {
		return err
	}

	// Save the configuration
	outputAsFile, err := models.CreateFile(params.Config, "values.yaml")
	if err != nil {
		return err
	}
	err = config.SaveConfigTo(outputAsFile, false, false)
	if err != nil {
		return err
	}
	// Apply the configuration
	if params.Apply {
		err = wrapper.HandleUp(params.Config)
		if err != nil {
			return err
		}
	}
	return nil
}

// For each argument, copy the input path to the output path
func SetContentHandler(conf string, args []string) error {
	// For each argument
	for _, arg := range args {
		if arg == "" {
			continue
		}
		// Split argument
		split := strings.Split(arg, ":")
		if len(split) != 2 {
			return fmt.Errorf("invalid argument: %s", arg)
		}
		// Get paths
		inputPath := split[0]
		outputPath := split[1]
		// Deamon specific, concatenate the config path
		if !app.IsCtl() {
			configPath := app.GetConfigsFolder(conf)
			outputPath = filepath.Join(configPath, outputPath)
		}
		// Call handler
		err := utils.Copy(inputPath, filepath.Join(conf, outputPath))
		if err != nil {
			return err
		}
	}
	return nil
}
