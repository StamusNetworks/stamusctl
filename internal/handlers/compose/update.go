package handlers

import (
	// Common

	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	// External

	// Custom
	"stamus-ctl/internal/app"
	"stamus-ctl/internal/backup"
	"stamus-ctl/internal/logging"
	"stamus-ctl/internal/models"
	"stamus-ctl/internal/utils"
	"stamus-ctl/internal/validation"

	"github.com/spf13/viper"
	"go.uber.org/zap"
)

type UpdateHandlerParams struct {
	Config         string
	Args           []string
	Version        string
	TemplateFolder string
	Interactive    bool
}

func UpdateHandler(params UpdateHandlerParams) error {
	// Unpack params
	configPath := params.Config
	args := params.Args
	versionVal := params.Version

	// Create automatic backup before update operation
	configName := filepath.Base(configPath)
	_, err := backup.CreateBackup(configName, backup.BackupTypeAuto, logging.Logger)
	if err != nil {
		// Log warning but continue with operation
		logging.Logger.Warn("Failed to create backup before compose update",
			zap.String("config", configName),
			zap.Error(err),
		)
	}

	// Validate version string to prevent path traversal
	if err := validation.ValidateVersion(versionVal); err != nil {
		logging.Sugar.Errorf("invalid version: %v", err)
		return fmt.Errorf("invalid version: %w", err)
	}

	// Get project
	viperInstance := viper.New()
	// General configuration
	viperInstance.SetEnvPrefix(app.Name)
	viperInstance.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))
	viperInstance.AutomaticEnv()
	// Specific configuration
	viperInstance.SetConfigName("values")
	viperInstance.SetConfigType("yaml")
	viperInstance.AddConfigPath(params.Config)
	// Read the config file
	err = viperInstance.ReadInConfig()
	if err != nil {
		logging.Sugar.Error("cannot read config file: ", err)
		return fmt.Errorf("cannot read config file: %w", err)
	}
	project := viperInstance.GetString("stamus.project")
	registry := viperInstance.GetString("stamus.registry")

	// Validate project name from config file
	if err := validation.ValidateProjectName(project); err != nil {
		logging.Sugar.Errorf("invalid project name in config: %v", err)
		return fmt.Errorf("invalid project name: %w", err)
	}

	// Get registry info
	destPath := filepath.Join(app.TemplatesFolder + project + "/")
	var templatePath string
	if params.TemplateFolder == "" {
		templatePath = filepath.Join(destPath, params.Version)
	} else {
		templatePath = params.TemplateFolder
	}

	logger := logging.Sugar.With(
		"Config", params.Config,
		"Args", params.Args,
		"Version", params.Version,
		"project", project,
	)

	// Load existing config FIRST, before any template pulls or scripts run
	// This is critical to capture the OLD template defaults for smart merging
	confFile, err := models.CreateFile(configPath, "values.yaml")
	if err != nil {
		logger.Error(err)

		return err
	}
	existingConfig, err := models.LoadConfigFrom(confFile, false)
	if err != nil {
		logger.Error(err)

		return err
	}

	// Pull config AFTER loading existing config to avoid overwriting old template
	logger.Debug("pulling latest template")
	if registry != "" {
		registryInfo := models.RegistryInfo{
			Registry: registry,
		}
		err = registryInfo.PullConfigAndUnwrap(destPath, project, versionVal)
		if err != nil {
			logger.Error(err)
			if !app.Embed.IsTrue() {
				return err
			}
		}
	} else {
		err = pullLatestTemplate(destPath, project, versionVal)
		if err != nil {
			logging.Sugar.Error(err)
			if !app.Embed.IsTrue() {
				return err
			}
		}
	}

	// Execute update script AFTER loading existing config and pulling new template
	// Validate script paths to prevent arbitrary code execution
	allowedScriptDirs := []string{app.TemplatesFolder}
	prerunPath := filepath.Join(destPath, "sbin/pre-run")
	postrunPath := filepath.Join(destPath, "sbin/post-run")

	// Validate pre-run script path
	if err := validation.ValidateScriptPath(prerunPath, allowedScriptDirs); err != nil {
		logger.Warnf("Skipping pre-run script due to security validation failure: %v", err)
		return fmt.Errorf("pre-run script path validation failed: %w", err)
	}

	runOutput, err := runArbitraryScript(prerunPath, configPath)
	if err != nil {
		return err
	}

	// Save output
	outputFile, err := app.FS.Create(filepath.Join(configPath, "values.yaml"))
	if err != nil {
		logger.Error(err)
		return err
	}
	defer outputFile.Close()
	if _, err := outputFile.WriteString(runOutput.String()); err != nil {
		logger.Error(err)

		return err
	}

	// Create new config
	newConfFile, err := models.CreateFile(templatePath, "config.yaml")
	if err != nil {
		logger.Error(err)

		return err
	}
	newConfig, err := models.ConfigFromFile(newConfFile)
	if err != nil {
		logger.Error(err)

		return err
	}
	_, _, err = newConfig.ExtractParams()
	if err != nil {
		logger.Error(err)

		return err
	}
	newConfig.SetSeed(existingConfig.GetSeed())

	// Extract and set values from args and existing config
	paramsArgs := utils.ExtractArgs(args)
	newConfig.SetProject(project)
	// Use SetValuesSmartMerge to intelligently merge old values
	// Only values that differ from old defaults are preserved (user customizations)
	// Values that match old defaults are replaced with new defaults
	newConfig.GetParams().SetValuesSmartMerge(existingConfig.GetParams())
	newConfig.GetArbitrary().SetArbitrary(paramsArgs)
	err = newConfig.GetParams().SetLooseValues(paramsArgs)
	if err != nil {
		logger.Error(err)

		return err
	}
	err = newConfig.GetParams().ProcessOptionnalParams(false)
	if err != nil {
		logger.Error(err)

		return err
	}

	// Ask for missing parameters or use defaults
	if app.IsCtl() {
		if params.Interactive {
			// When --interactive flag is set, prompt for missing parameters (old behavior)
			err = newConfig.GetParams().AskMissing()
			if err != nil {
				logger.Error(err)

				return err
			}
		} else {
			// Default behavior: use defaults without prompting
			err = newConfig.GetParams().SetToDefaults()
			if err != nil {
				logger.Error(err)

				return err
			}
		}
	}

	// Save the configuration
	err = newConfig.SaveConfigTo(confFile, true, false)
	if err != nil {
		if !errors.Is(err, models.ErrorEmptyFolder) {
			logger.Error(err)

			return err
		}
	}

	// Run post-run script
	// Validate post-run script path
	if err := validation.ValidateScriptPath(postrunPath, allowedScriptDirs); err != nil {
		logger.Warnf("Skipping post-run script due to security validation failure: %v", err)
		return fmt.Errorf("post-run script path validation failed: %w", err)
	}

	_, err = runArbitraryScript(postrunPath, configPath)
	if err != nil {
		logger.Error(err)

		return err
	}

	logger.Debug("Update ran successfully")
	return nil
}

func runArbitraryScript(path string, config string) (*strings.Builder, error) {
	// Verify script file exists
	if _, err := os.Stat(path); os.IsNotExist(err) {
		// Script doesn't exist - this is OK, just skip it
		logging.Sugar.Debugf("Script does not exist, skipping: %s", path)
		return new(strings.Builder), nil
	}

	// Validate config path to prevent injection
	cleanConfig, err := validation.SanitizePath(config, "")
	if err != nil {
		return nil, fmt.Errorf("invalid config path: %w", err)
	}

	// Execute script with validated arguments
	arbitrary := exec.Command(path, "--config", cleanConfig)
	// Display output to terminal
	runOutput := new(strings.Builder)
	arbitrary.Stdout = runOutput
	arbitrary.Stderr = os.Stderr
	// Change execution rights
	err = app.FS.Chmod(path, 0o755)
	if err != nil {
		return nil, err
	}
	// Run script
	if err := arbitrary.Run(); err != nil {
		return nil, err
	}
	return runOutput, nil
}
