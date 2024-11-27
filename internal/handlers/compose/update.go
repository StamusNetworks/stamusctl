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
	"stamus-ctl/internal/stamus"
	"stamus-ctl/internal/utils"

	"github.com/spf13/viper"
)

type UpdateHandlerParams struct {
	Config  string
	Args    []string
	Version string
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
		return fmt.Errorf("cannot read config file: %w", err)
	}
	project := viperInstance.GetString("stamus.project")

	// Get registry info
	destPath := filepath.Join(app.TemplatesFolder + project + "/")
	latestPath := filepath.Join(destPath, "latest/")

	// Get registries infos
	stamusConf, err := stamus.GetStamusConfig()
	if err != nil {
		return err
	}

	// Pull config
	logging.Sugar.Info("Getting configuration")
	for _, registryInfo := range stamusConf.Registries.AsList() {
		err = registryInfo.PullConfig(destPath, project, versionVal)
		if err == nil {
			break
		}
	}
	if err != nil {
		return err
	}

	// Execute update script
	prerunPath := filepath.Join(destPath, "sbin/pre-run")
	postrunPath := filepath.Join(destPath, "sbin/post-run")
	runOutput, err := runArbitraryScript(prerunPath, configPath)
	if err != nil {
		return err
	}

	// Save output
	outputFile, err := os.Create(filepath.Join(configPath, "values.yaml"))
	if err != nil {
		return err
	}
	defer outputFile.Close()
	if _, err := outputFile.WriteString(runOutput.String()); err != nil {
		return err
	}

	// Load existing config
	confFile, err := models.CreateFileInstance(configPath, "values.yaml")
	if err != nil {
		return err
	}
	existingConfig, err := models.LoadConfigFrom(confFile, false)
	if err != nil {
		return err
	}

	// Create new config
	newConfFile, err := models.CreateFileInstance(latestPath, "config.yaml")
	if err != nil {
		return err
	}
	newConfig, err := models.NewConfigFrom(newConfFile)
	if err != nil {
		return err
	}
	_, _, err = newConfig.ExtractParams()
	if err != nil {
		return err
	}

	// Extract and set values from args and existing config
	paramsArgs := utils.ExtractArgs(args)
	newConfig.GetParams().SetValues(existingConfig.GetParams().GetVariablesValues())
	newConfig.GetParams().SetLooseValues(paramsArgs)
	newConfig.GetArbitrary().SetArbitrary(paramsArgs)
	newConfig.GetParams().ProcessOptionnalParams(false)
	newConfig.SetProject(project)

	// Ask for missing parameters
	if app.IsCtl() {
		err = newConfig.GetParams().AskMissing()
		if err != nil {
			return err
		}
	}

	// Save the configuration
	err = newConfig.SaveConfigTo(confFile, false, true)
	if err != nil {
		return err
	}

	// Run post-run script
	_, err = runArbitraryScript(postrunPath, configPath)
	if err != nil {
		return err
	}
	logging.Sugar.Info("")

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
	err := os.Chmod(path, 0755)
	if err != nil {
		return nil, err
	}
	// Run arbitrary script
	if err := arbitrary.Run(); err != nil {
		return nil, err
	}
	return runOutput, nil
}
