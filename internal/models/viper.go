package models

import (
	"bytes"
	"fmt"
	"strings"

	"stamus-ctl/internal/app"
	"stamus-ctl/internal/validation"

	"github.com/spf13/afero"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"
)

func InstanciateViper(file *File) (*viper.Viper, error) {
	// Create a new viper instance
	viperInstance := viper.New()
	// General configuration
	viperInstance.SetFs(app.FS)
	viperInstance.SetEnvPrefix(app.Name)
	viperInstance.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))
	viperInstance.AutomaticEnv()
	// Specific configuration
	viperInstance.SetConfigName(file.Name)
	viperInstance.SetConfigType(file.Type)
	viperInstance.AddConfigPath(file.Path)
	// Get file content
	completePath := fmt.Sprintf("%s/%s.%s", file.Path, file.Name, file.Type)
	content, err := afero.ReadFile(app.FS, completePath)
	if err != nil {
		return nil, fmt.Errorf("cannot read file: %w", err)
	}

	// Validate YAML structure depth before parsing with viper
	if file.Type == "yaml" || file.Type == "yml" {
		var rawData map[string]interface{}
		if err := yaml.Unmarshal(content, &rawData); err != nil {
			return nil, fmt.Errorf("invalid YAML syntax: %w", err)
		}

		// Validate depth to prevent DoS attacks with deeply nested YAML
		if err := validation.ValidateMapDepth(rawData, validation.MaxYAMLDepth); err != nil {
			return nil, fmt.Errorf("YAML structure validation failed: %w", err)
		}
	}

	// Read the config file
	err = viperInstance.ReadConfig(bytes.NewBuffer(content))
	if err != nil {
		return nil, fmt.Errorf("cannot read config file: %w", err)
	}
	return viperInstance, nil
}
