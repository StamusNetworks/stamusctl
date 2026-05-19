package common

import (
	"path/filepath"

	"stamus-ctl/internal/app"
	"stamus-ctl/internal/logging"
	"stamus-ctl/internal/models"
	"stamus-ctl/internal/stamus"
)

// GetStamusConfigFunc allows test injection of stamus.GetStamusConfig.
var GetStamusConfigFunc = stamus.GetStamusConfig

// PullLatestTemplate pulls the latest template from saved registries.
// If logged into registries, tries each in turn. Falls back to the default registry.
func PullLatestTemplate(destPath string, project, version string) error {
	stamusConf, err := GetStamusConfigFunc()
	if err != nil {
		return err
	}
	if len(stamusConf.Registries.AsList()) != 0 {
		for _, registryInfo := range stamusConf.Registries.AsList() {
			err = registryInfo.PullConfigAndUnwrap(destPath, project, version)
			if err == nil {
				return nil
			} else {
				logging.Sugar.Debug(err)
			}
		}
	}
	infos := models.RegistryInfo{
		Registry: app.DefaultRegistry,
	}
	return infos.PullConfigAndUnwrap(destPath, project, version)
}

// InstanciateConfig tries to instantiate a config from folderPath, falling back
// to backupFolderPath when app.Embed is enabled.
func InstanciateConfig(folderPath string, backupFolderPath string) (*models.Config, error) {
	config, err := InstanciateConfigFromPath(folderPath)
	if err == nil {
		return config, nil
	}
	if app.Embed.IsTrue() {
		config, err = InstanciateConfigFromPath(backupFolderPath)
		if err == nil {
			return config, nil
		}
	}
	return nil, err
}

// InstanciateConfigFromPath instantiates a Config from a config.yaml located at folderPath.
func InstanciateConfigFromPath(folderPath string) (*models.Config, error) {
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

// SetParameters sets config parameters from arbitrary values, defaults, and interactive prompts.
func SetParameters(isCli bool, config *models.Config, arbitrary map[string]string, isDefault bool) error {
	err := config.GetParams().SetLooseValues(arbitrary)
	config.GetArbitrary().SetArbitrary(arbitrary)
	if err != nil {
		return err
	}
	if isDefault {
		err = config.GetParams().SetToDefaults()
		if err != nil {
			return err
		}
	}
	if isCli {
		err = config.GetParams().AskMissing()
		if err != nil {
			return err
		}
	}
	return nil
}

// ResolveTemplatePath determines the template path based on params.
// templateFolder overrides the default; embed mode always returns DefaultClearNDRPath.
func ResolveTemplatePath(destPath, templateFolder, version string) string {
	if app.Embed.IsTrue() {
		return app.DefaultClearNDRPath
	}
	if templateFolder != "" {
		return templateFolder
	}
	return filepath.Join(destPath, version)
}
