package handlers

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"stamus-ctl/internal/app"
	"stamus-ctl/internal/backup"
	"stamus-ctl/internal/handlers/common"
	"stamus-ctl/internal/logging"
	"stamus-ctl/internal/models"
	"stamus-ctl/internal/stamus"
	"stamus-ctl/internal/utils"
	"stamus-ctl/internal/validation"

	"github.com/spf13/afero"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

// execCommandFunc is used to create exec.Cmd instances.
// Tests can replace this to avoid running real scripts.
var execCommandFunc = exec.Command

type NixUpdateHandlerInputs struct {
	Config         string
	Args           []string
	Version        string
	TemplateFolder string
	Interactive    bool
}

func NixUpdateHandler(params NixUpdateHandlerInputs) error {
	configPath := params.Config
	if !app.IsCtl() {
		configPath = app.GetConfigsFolder(params.Config)
	}

	// Create automatic backup before update
	configName := filepath.Base(configPath)
	_, err := backup.CreateBackup(configName, backup.BackupTypeAuto, logging.Logger)
	if err != nil {
		logging.Logger.Warn("Failed to create backup before nix update",
			zap.String("config", configName),
			zap.Error(err),
		)
	}

	// Validate version
	if err := validation.ValidateVersion(params.Version); err != nil {
		logging.Sugar.Errorf("invalid version: %v", err)
		return fmt.Errorf("invalid version: %w", err)
	}

	// Read project and registry from existing values.yaml
	viperInstance := viper.New()
	viperInstance.SetEnvPrefix(app.Name)
	viperInstance.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))
	viperInstance.AutomaticEnv()
	viperInstance.SetConfigName("values")
	viperInstance.SetConfigType("yaml")
	viperInstance.AddConfigPath(configPath)

	err = viperInstance.ReadInConfig()
	if err != nil {
		logging.Sugar.Error("cannot read config file: ", err)
		return fmt.Errorf("cannot read config file: %w", err)
	}
	project := viperInstance.GetString("stamus.project")
	registry := viperInstance.GetString("stamus.registry")

	// Validate project name
	if err := validation.ValidateProjectName(project); err != nil {
		logging.Sugar.Errorf("invalid project name in config: %v", err)
		return fmt.Errorf("invalid project name: %w", err)
	}

	// Resolve template path
	destPath := filepath.Join(app.TemplatesFolder, project+"/")
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

	// Load existing config FIRST, before any template pulls
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

	// Pull new template version
	logger.Debug("pulling latest template")
	if registry != "" {
		registryInfo := models.RegistryInfo{
			Registry: registry,
		}
		err = registryInfo.PullConfigAndUnwrap(destPath, project, params.Version)
		if err != nil {
			logger.Error(err)
			if !app.Embed.IsTrue() {
				return err
			}
		}
	} else {
		err = common.PullLatestTemplate(destPath, project, params.Version)
		if err != nil {
			logging.Sugar.Error(err)
			if !app.Embed.IsTrue() {
				return err
			}
		}
	}

	// Run pre-run script if present
	allowedScriptDirs := []string{app.TemplatesFolder}
	prerunPath := filepath.Join(destPath, "sbin/pre-run")

	if err := validation.ValidateScriptPath(prerunPath, allowedScriptDirs); err != nil {
		logger.Warnf("Skipping pre-run script due to security validation failure: %v", err)
	} else {
		runOutput, err := runScript(prerunPath, configPath)
		if err != nil {
			return err
		}

		// Save pre-run output to values.yaml
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
	}

	// Create new config from template
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

	// Smart merge: preserve user customizations, update defaults
	paramsArgs := utils.ExtractArgs(params.Args)
	newConfig.SetProject(project)
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

	// Handle interactive/default parameter setting
	if app.IsCtl() {
		if params.Interactive {
			err = newConfig.GetParams().AskMissing()
			if err != nil {
				logger.Error(err)
				return err
			}
		} else {
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
	postrunPath := filepath.Join(destPath, "sbin/post-run")
	if err := validation.ValidateScriptPath(postrunPath, allowedScriptDirs); err != nil {
		logger.Warnf("Skipping post-run script due to security validation failure: %v", err)
	} else {
		_, err = runScript(postrunPath, configPath)
		if err != nil {
			logger.Error(err)
			return err
		}
	}

	// Update instance registry with new version
	version, _ := afero.ReadFile(app.FS, filepath.Join(configPath, "version"))
	versionString := strings.Split(string(version), "\n")[0]
	if versionString == "" {
		versionString = params.Version
	}
	err = stamus.AddInstance(configPath, project, versionString)
	if err != nil {
		logger.Error(err)
		return err
	}

	logger.Debug("Update ran successfully")
	return nil
}

func runScript(path string, config string) (*strings.Builder, error) {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		logging.Sugar.Debugf("Script does not exist, skipping: %s", path)
		return new(strings.Builder), nil
	}

	cleanConfig, err := validation.SanitizePath(config, "")
	if err != nil {
		return nil, fmt.Errorf("invalid config path: %w", err)
	}

	arbitrary := execCommandFunc(path, "--config", cleanConfig)
	runOutput := new(strings.Builder)
	arbitrary.Stdout = runOutput
	arbitrary.Stderr = os.Stderr

	err = app.FS.Chmod(path, 0o755)
	if err != nil {
		return nil, err
	}

	if err := arbitrary.Run(); err != nil {
		return nil, err
	}
	return runOutput, nil
}
