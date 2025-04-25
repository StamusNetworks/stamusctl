package handlers

import (
	// Common

	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	// External

	// Custom
	"stamus-ctl/internal/app"
	"stamus-ctl/internal/logging"
	"stamus-ctl/internal/models"
	"stamus-ctl/internal/utils"

	"github.com/spf13/viper"
)

type UpdateHandlerParams struct {
	Config         string
	Args           []string
	Version        string
	TemplateFolder string
}

func UpdateHandler(params UpdateHandlerParams) error {
	// Unpack params
	configPath := params.Config
	args := params.Args
	versionVal := params.Version

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
	err := viperInstance.ReadInConfig()
	if err != nil {
		logging.Sugar.Error("cannot read config file: ", err)
		return fmt.Errorf("cannot read config file: %w", err)
	}
	project := viperInstance.GetString("stamus.project")
	registry := viperInstance.GetString("stamus.registry")

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

	// Pull config
	logger.Debug("pulling latest template")
	if registry != "" {
		registryInfo := models.RegistryInfo{
			Registry: registry,
		}
		err = registryInfo.PullConfig(destPath, project, versionVal)
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

	// Execute update script
	prerunPath := filepath.Join(destPath, "sbin/pre-run")
	postrunPath := filepath.Join(destPath, "sbin/post-run")
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

	// Load existing config
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
	newConfig.GetParams().SetValues(existingConfig.GetParams().GetVariablesValues())
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

	// Ask for missing parameters
	if app.IsCtl() {
		err = newConfig.GetParams().AskMissing()
		if err != nil {
			logger.Error(err)

			return err
		}
	}

	// Save the configuration
	err = newConfig.SaveConfigTo(confFile, true, false)
	if err != nil {
		logger.Error(err)

		return err
	}

	// Run post-run script
	_, err = runArbitraryScript(postrunPath, configPath)
	if err != nil {
		logger.Error(err)

		return err
	}

	logger.Debug("Update ran successfully")
	return nil
}

func runArbitraryScript(path string, config string) (*strings.Builder, error) {
	// Execute arbitrary script
	arbitrary := exec.Command(path, "--config", config)
	// Display output to terminal
	runOutput := new(strings.Builder)
	arbitrary.Stdout = runOutput
	arbitrary.Stderr = os.Stderr
	// Change execution rights
	err := app.FS.Chmod(path, 0o755)
	if err != nil {
		return nil, err
	}
	// Run arbitrary script
	if err := arbitrary.Run(); err != nil {
		return nil, err
	}
	return runOutput, nil
}
