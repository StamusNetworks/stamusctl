package stamus

import (
	"encoding/json"
	"os"

	"stamus-ctl/internal/app"
	"stamus-ctl/internal/models"

	"github.com/spf13/afero"
)

type Config struct {
	Registries Registries `json:"registries"`
	Instances  Instances  `json:"instances"`
}

func (r *Registries) AsList() []models.RegistryInfo {
	// Create RegistryInfo
	registryInfos := []models.RegistryInfo{}
	for registry, logins := range *r {
		for user, token := range logins {
			registryInfos = append(registryInfos, models.RegistryInfo{
				Registry: string(registry),
				Username: string(user),
				Password: string(token),
			})
		}
	}
	return registryInfos
}

func (c *Config) Save() error {
	// Save config
	return c.setStamusConfig()
}

func (conf *Config) setStamusConfig() error {
	// Open or create
	file, err := getOrCreateStamusConfigFile()
	if err != nil {
		return err
	}

	// Write the new content
	bytes, err := json.Marshal(conf)
	if err != nil {
		return err
	}

	// Delete content
	err = file.Truncate(0)
	if err != nil {
		return err
	}

	_, err = file.WriteAt(bytes, 0)
	if err != nil {
		return err
	}
	return nil
}

func GetConfigsList() ([]string, error) {
	// Get list of configs in app.ConfigsFolder
	entries, err := afero.ReadDir(app.FS, app.ConfigsFolder)
	if err != nil {
		// Create folder if it does not exist
		if os.IsNotExist(err) {
			err = app.FS.MkdirAll(app.ConfigsFolder, 0o755)
			if err != nil {
				return nil, err
			}
			return GetConfigsList()
		}
		return nil, err
	}
	// Get the list of configs
	configs := []string{}
	for _, e := range entries {
		if e.IsDir() {
			configs = append(configs, e.Name())
		}
	}
	return configs, nil
}
