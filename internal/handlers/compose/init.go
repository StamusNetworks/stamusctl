package handlers

import (
	"path/filepath"
	"stamus-ctl/internal/app"
	"stamus-ctl/internal/logging"
	"stamus-ctl/internal/models"
	"stamus-ctl/internal/stamus"

	confHandler "stamus-ctl/internal/handlers/config"
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
	Bind             []string
}

func InitHandler(isCli bool, params InitHandlerInputs) error {
	logger := logging.Sugar.With(
		"IsDefault", params.IsDefault,
		"BackupFolderPath", params.BackupFolderPath,
		"Project", params.Project,
		"Version", params.Version,
		"Arbitrary", params.Arbitrary,
		"Values", params.Values,
		"Config", params.Config,
		"FromFile", params.FromFile,
		"TemplateFolder", params.TemplateFolder,
		"Bind", params.Bind,
	)
	// Get registry info
	destPath := filepath.Join(app.TemplatesFolder, params.Project)

	// Pull latest template
	logger.Debug("pulling latest template")
	err := pullLatestTemplate(destPath, params.Project, params.Version)
	if err != nil {
		if !app.Embed.IsTrue() {
			return err
		}
		logging.Sugar.Error(err)
	}
	// Instanciate config
	var templatePath string
	if params.TemplateFolder == "" {
		templatePath = filepath.Join(destPath, params.Version)
	} else {
		templatePath = params.TemplateFolder
	}

	logger.Debug("instanciation config")
	config, err := instanciateConfig(templatePath, params.BackupFolderPath)
	if err != nil {
		logger.Error(err)
		return err
	}

	// Read the folder configuration
	_, _, err = config.ExtractParams()
	if err != nil {
		logger.Error(err)
		return err
	}

	logger.Debug("Setting parameters from files")
	// Set parameters
	err = config.SetValuesFromFiles(params.FromFile)
	if err != nil {
		logger.Error(err)
		return err
	}

	logger.Debug("set values from file")
	err = config.SetValuesFromFile(params.Values)
	if err != nil {
		logger.Error(err)
		return err
	}

	logger.Debug("Setting parameters")
	err = setParameters(isCli, config, params)
	if err != nil {
		logger.Error(err)
		return err
	}

	logger.Debug("Set project")
	config.SetProject(params.Project)

	// Validate parameters
	logger.Debug("Validate params")
	err = config.GetParams().ValidateAll()
	if err != nil {
		logger.Error(err)
		return err
	}

	// Save the configuration
	logger.Debug("Create values.yaml")
	outputFile, err := models.CreateFile(params.Config, "values.yaml")
	if err != nil {
		logger.Error(err)
		return err
	}

	logger.Debug("Save config to: ", outputFile)
	if err = config.SaveConfigTo(outputFile, false, true); err != nil {
		logger.Error("Error saving config to", outputFile, err)
	}

	// Bind files
	logger.Debug("Set content handler")
	err = confHandler.SetContentHandler(params.Config, params.Bind)
	if err != nil {
		logger.Error(err)
		return err
	}
	return nil
}

// Pull latest template from saved registries
func pullLatestTemplate(destPath string, project, version string) error {
	// Get registries infos
	stamusConf, err := stamus.GetStamusConfig()
	if err != nil {
		return err
	}
	// Pull latest config
	if len(stamusConf.Registries.AsList()) != 0 {
		//Logged in
		for _, registryInfo := range stamusConf.Registries.AsList() {
			err = registryInfo.PullConfig(destPath, project, version)
			if err == nil {
				return nil
			} else {
				logging.Sugar.Debug(err)
			}
		}
	}
	// Not logged in
	infos := models.RegistryInfo{
		Registry: app.DefaultRegistry,
	}
	err = infos.PullConfig(destPath, project, version)
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
	confFile, err := models.CreateFile(folderPath, "config.yaml")
	if err != nil {
		return nil, err
	}
	config, err := models.ConfigFromFile(confFile)
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
