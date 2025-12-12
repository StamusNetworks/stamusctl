package config

import (
	// Core
	"errors"
	"fmt"
	"strings"

	// Internal

	"stamus-ctl/internal/app"
	wrapper "stamus-ctl/internal/handlers/wrapper"
	"stamus-ctl/internal/models"
	"stamus-ctl/internal/utils"
	"stamus-ctl/internal/validation"
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
		if !errors.Is(err, models.ErrorEmptyFolder) {
			return err
		}
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
	// Validate base config path
	sanitizedConf, err := validation.SanitizePath(conf, "")
	if err != nil {
		return fmt.Errorf("invalid config path: %w", err)
	}

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

		// Validate input path exists and is readable
		sanitizedInput, err := validation.SanitizePath(inputPath, "")
		if err != nil {
			return fmt.Errorf("invalid input path '%s': %w", inputPath, err)
		}

		// Daemon specific, concatenate the config path
		var finalOutputPath string
		if !app.IsCtl() {
			configPath := app.GetConfigsFolder(conf)
			// Validate output path to prevent directory traversal
			sanitizedOutput, err := validation.SanitizePath(outputPath, configPath)
			if err != nil {
				return fmt.Errorf("invalid output path '%s': %w", outputPath, err)
			}
			finalOutputPath = sanitizedOutput
		} else {
			// For CLI mode, validate output path relative to conf
			sanitizedOutput, err := validation.SanitizePath(outputPath, sanitizedConf)
			if err != nil {
				return fmt.Errorf("invalid output path '%s': %w", outputPath, err)
			}
			finalOutputPath = sanitizedOutput
		}

		// Call handler with validated paths
		err = utils.Copy(sanitizedInput, finalOutputPath)
		if err != nil {
			return err
		}
	}
	return nil
}
