package handlers

import (
	"errors"
	"os"
	"path/filepath"
	"strings"

	"stamus-ctl/internal/app"
	"stamus-ctl/internal/backup"
	"stamus-ctl/internal/embeds"
	"stamus-ctl/internal/handlers/common"
	"stamus-ctl/internal/logging"
	"stamus-ctl/internal/models"
	"stamus-ctl/internal/stamus"
	"stamus-ctl/internal/validation"

	confHandler "stamus-ctl/internal/handlers/config"

	"github.com/spf13/afero"
	"go.uber.org/zap"
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
	Registry         string
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

	// Validate project name to prevent path traversal
	if err := validation.ValidateProjectName(params.Project); err != nil {
		logger.Errorf("invalid project name: %v", err)
		return err
	}

	// Validate version string to prevent path traversal
	if err := validation.ValidateVersion(params.Version); err != nil {
		logger.Errorf("invalid version: %v", err)
		return err
	}

	// Setup
	embeds.InitClearNDRFolder(app.DefaultClearNDRPath)
	// Get registry info
	destPath := filepath.Join(app.TemplatesFolder, params.Project)

	// Pull latest template
	logger.Debug("pulling latest template")
	if params.Registry != "" {
		registryInfo := models.RegistryInfo{
			Registry: params.Registry,
		}
		err := registryInfo.PullConfigAndUnwrap(destPath, params.Project, params.Version)
		if err != nil && err.Error() == "Error response from daemon: manifest unknown" {
			logger.Fatal(params.Project + ":" + params.Version + " template not found in registry " +
				params.Registry + ". Please check the registry or use a different version.")
		}
		if err != nil {
			if errors.Is(err, models.ErrPullingImage) {
				logger.Warn("error pulling template. Will try using from local images. If you want to pull from remote, check your internet connection and retry.")
			} else {
				if !app.Embed.IsTrue() {
					logger.Error("error pulling template: ", err)

					return err
				}
			}
		}
	} else {
		err := common.PullLatestTemplate(destPath, params.Project, params.Version)
		if err != nil && err.Error() == "Error response from daemon: manifest unknown" {
			logger.Fatal(params.Project + ":" + params.Version + " template not found in registry " +
				params.Registry + ". Please check the registry or use a different version.")
		}
		if err != nil {
			if errors.Is(err, models.ErrPullingImage) {
				logger.Warn("error pulling template. Will try using from local images. If you want to pull from remote, check your internet connection and retry.")
			} else {
				if !app.Embed.IsTrue() {
					logger.Error("error pulling template: ", err)

					return err
				}
			}
		}
	}

	// Instantiate config
	templatePath := common.ResolveTemplatePath(destPath, params.TemplateFolder, params.Version)

	logger.Debug("instanciation config")
	config, err := common.InstanciateConfig(templatePath, params.BackupFolderPath)
	if err != nil {
		logger.Error(err)
		return err
	}

	// Read the folder configuration
	logger.Debug("extracting params")
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
	err = common.SetParameters(isCli, config, params.Arbitrary, params.IsDefault)
	if err != nil {
		logger.Error(err)
		return err
	}

	logger.Debug("Set project")
	config.SetProject(params.Project)
	config.SetRegistry(params.Registry)

	// Validate parameters
	logger.Debug("Validate params")
	err = config.GetParams().ValidateAll()
	if err != nil {
		logger.Error(err)
		return err
	}

	// Check if config already exists and create backup if so
	configPath := params.Config
	if !isCli {
		configPath = app.GetConfigsFolder(params.Config)
	}
	if _, err := os.Stat(configPath); err == nil {
		// Config exists, create backup before overwriting
		configName := filepath.Base(params.Config)
		_, backupErr := backup.CreateBackup(configName, backup.BackupTypeAuto, logging.Logger)
		if backupErr != nil {
			// Log warning but continue with operation
			logging.Logger.Warn("Failed to create backup before compose init",
				zap.String("config", configName),
				zap.Error(backupErr),
			)
		}
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
		if !errors.Is(err, models.ErrorEmptyFolder) {
			return err
		}
	}

	// Bind files
	logger.Debug("Set content handler")
	err = confHandler.SetContentHandler(params.Config, params.Bind)
	if err != nil {
		logger.Error(err)
		return err
	}

	// Save instance
	logger.Debug("Save instance")
	configPath = params.Config
	if isCli {
		currentPath, _ := os.Getwd()
		configPath = filepath.Join(currentPath, params.Config)
	}
	version, _ := afero.ReadFile(app.FS, filepath.Join(configPath, "version"))
	versionString := strings.Split(string(version), "\n")[0]
	if versionString == "" {
		versionString = params.Version
	}
	err = stamus.AddInstance(configPath, params.Project, versionString)
	if err != nil {
		logger.Error(err)
		return err
	}

	logger.Debug("Init finished")
	return nil
}
