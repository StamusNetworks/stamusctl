package config

import (
	"os"
	"path/filepath"
	"strings"

	"stamus-ctl/internal/app"
	"stamus-ctl/internal/models"
	"stamus-ctl/internal/stamus"
	"stamus-ctl/internal/utils"
)

func GetVersion(config string) string {
	// File
	if !app.IsCtl() {
		config = app.GetConfigsFolder(config)
	}
	filePath := filepath.Join(config, "version")

	// Read the file content
	content, err := os.ReadFile(filePath)
	if err != nil {
		return "Could not read the version file"
	}
	return string(content)
}

// Get the grouped config values
// Essentially, this function reads the config values file and groups the values
func GetGroupedConfig(conf string, args []string, reload bool) (map[string]interface{}, error) {
	// File instance
	if !app.IsCtl() {
		conf = app.GetConfigsFolder(conf)
	}
	inputAsFile, err := models.CreateFile(conf, "values.yaml")
	if err != nil {
		return nil, err
	}
	// Load the config
	config, err := models.LoadConfigFrom(inputAsFile, reload)
	if err != nil {
		return nil, err
	}
	// Process optionnal parameters
	err = config.GetParams().ProcessOptionnalParams(false)
	if err != nil {
		return nil, err
	}
	// Group values
	groupedValues := utils.GroupValues(config.GetParams(), args)
	// Return
	return groupedValues, nil
}

// Get the grouped content
// Essentially, this function reads the config folder content and groups the folders and files
func GetGroupedContent(conf string, args []string) (map[string]interface{}, error) {
	// Get path
	if !app.IsCtl() {
		conf = app.GetConfigsFolder(conf)
	}
	// Get files
	files, err := utils.ListFilesInFolder(conf)
	if err != nil {
		return nil, err
	}
	// Filter files
	for _, arg := range args {
		for file := range files {
			if !strings.Contains(file, arg) {
				delete(files, file)
			}
		}
	}
	// Group files
	groupedFiles := utils.GroupStuff(files)
	// Return
	return groupedFiles, nil
}

// Get the list of configs, with the current one marked
func GetConfigsList() ([]string, error) {
	// Get list
	configsList, err := stamus.GetConfigsList()
	if err != nil {
		return nil, err
	}
	return configsList, nil
}

// Get the list of keys
func GetParamsList(conf string) (*models.Parameters, error) {
	// File instance
	if !app.IsCtl() {
		conf = app.GetConfigsFolder(conf)
	}
	inputAsFile, err := models.CreateFile(conf, "values.yaml")
	if err != nil {
		return nil, err
	}
	// Load the config
	config, err := models.LoadConfigFrom(inputAsFile, true)
	if err != nil {
		return nil, err
	}
	return config.GetParams(), nil
}
