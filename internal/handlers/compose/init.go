package handlers

import (
	"fmt"
	"path/filepath"
	"stamus-ctl/internal/app"
	"stamus-ctl/internal/models"
	"stamus-ctl/internal/stamus"
)

type InitHandlerInputs struct {
	IsDefault        bool
	BackupFolderPath string
	Project          string
	Version          string
	Arbitrary        map[string]string
	Values           string
	Config           string
	FromFile         string
	TemplateFolder   string
}

func InitHandler(isCli bool, params InitHandlerInputs) error {
	// Get registry info
	image := "/" + params.Project + ":" + params.Version
	destPath := filepath.Join(app.TemplatesFolder, params.Project)

	// Pull latest template
	err := pullLatestTemplate(destPath, image)
	if err != nil && !app.Embed.IsTrue() {
		return err
	}
	// Instanciate config
	var templatePath string
	if params.TemplateFolder == "" {
		templatePath = filepath.Join(destPath, params.Version)
	} else {
		templatePath = params.TemplateFolder
	}
	config, err := instanciateConfig(templatePath, params.BackupFolderPath)
	if err != nil {
		return err
	}
	// Read the folder configuration
	_, _, err = config.ExtractParams(true)
	if err != nil {
		return err
	}
	// Set parameters
	err = config.SetValuesFromFiles(params.FromFile)
	if err != nil {
		return err
	}
	err = config.SetValuesFromFile(params.Values)
	if err != nil {
		return err
	}
	err = setParameters(isCli, config, params)
	if err != nil {
		return err
	}
	config.SetProject(params.Project)

	// Validate parameters
	err = config.GetParams().ValidateAll()
	if err != nil {
		return err
	}
	// Save the configuration
	var path string
	if params.Config == "" {
		stamusConfig, err := stamus.GetCurrent()
		if err != nil {
			return err
		}
		path = app.GetConfigsFolder(stamusConfig)
	} else {
		path = params.Config
	}
	outputFile, err := models.CreateFileInstance(path, "values.yaml")
	if err != nil {
		return err
	}
	err = config.SaveConfigTo(outputFile, false, true)
	if err != nil {
		return err
	}
	return nil
}

// Pull latest template from saved registries
func pullLatestTemplate(destPath string, image string) error {
	// Get registries infos
	stamusConf, err := stamus.GetStamusConfig()
	if err != nil {
		return err
	}
	// Verify registries not empty
	if len(stamusConf.Registries.AsList()) == 0 {
		return fmt.Errorf("no config registries credentials found")
	}
	// Pull latest config
	for _, registryInfo := range stamusConf.Registries.AsList() {
		err = registryInfo.PullConfig(destPath, image)
		if err == nil {
			break
		}
	}
	return err
}

// Instanciate config from folder or backup folders
func instanciateConfig(folderPath string, backupFolderPath string) (*models.Config, error) {
	// Try to instanciate from folder
	config, err := instanciateConfigFromPath(folderPath)
	if err == nil {
		return config, nil
	}
	if app.Embed.IsTrue() {
		// Try to instanciate from backup folder
		config, err = instanciateConfigFromPath(backupFolderPath)
		if err == nil {
			return config, nil
		}
	}
	// Return error
	return nil, err
}

// Instanciate config from path
func instanciateConfigFromPath(folderPath string) (*models.Config, error) {
	confFile, err := models.CreateFileInstance(folderPath, "config.yaml")
	if err != nil {
		return nil, err
	}
	config, err := models.NewConfigFrom(confFile)
	if err != nil {
		return nil, err
	}
	return config, nil
}

// Set parameters, from args, and defaults / asks rest
func setParameters(isCli bool, config *models.Config, params InitHandlerInputs) error {
	// Extract and set values from args
	err := config.GetParams().SetLooseValues(params.Arbitrary)
	config.GetArbitrary().SetArbitrary(params.Arbitrary)
	if err != nil {
		return err
	}
	// Set from default
	if params.IsDefault {
		err = config.GetParams().SetToDefaults()
		if err != nil {
			return err
		}
	}
	// Ask for missing parameters
	if isCli {
		err = config.GetParams().AskMissing()
		if err != nil {
			return err
		}
	}
	return nil
}
