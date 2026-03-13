package stamus

import (
	"encoding/json"
	"os"
	"path/filepath"

	"stamus-ctl/internal/app"
)

func (cm *ConfigManager) getOrCreateConfigFile() (*os.File, error) {
	// Create ~/stamus directory
	err := cm.fs.MkdirAll(app.ConfigFolder, 0o755)
	if err != nil {
		return nil, err
	}

	// Open or create ~/stamus/config.json
	f, err := cm.fs.OpenFile(filepath.Join(app.ConfigFolder, "config.json"), os.O_RDWR|os.O_CREATE, 0o755)
	if err != nil {
		return nil, err
	}

	return f, nil
}

func (cm *ConfigManager) tryGetConfigFile() (*os.File, error) {
	// Open ~/stamus/config.json
	f, err := cm.fs.OpenFile(filepath.Join(app.ConfigFolder, "config.json"), os.O_RDONLY, 0o755)
	if err != nil {
		return nil, err
	}

	return f, nil
}

func (cm *ConfigManager) GetConfig() (*Config, error) {
	// Open or create ~/stamus/config.json
	file, err := cm.tryGetConfigFile()
	if err != nil {
		return &Config{}, nil
	}
	// Read the file contents
	bytes, err := cm.fs.ReadAll(file)
	if err != nil {
		return &Config{}, nil
	}
	// Unmarshal the file contents
	config := &Config{}
	if len(bytes) != 0 {
		err = json.Unmarshal(bytes, &config)
		if err != nil {
			return &Config{}, nil
		}
	}

	return config, nil
}

// Backward-compatible package-level functions

func getOrCreateStamusConfigFile() (*os.File, error) {
	return DefaultManager.getOrCreateConfigFile()
}

func GetStamusConfig() (*Config, error) {
	return DefaultManager.GetConfig()
}
