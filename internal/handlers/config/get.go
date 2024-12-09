package config

import (
	"os"
	"path/filepath"
	"stamus-ctl/internal/app"
	"stamus-ctl/internal/models"
	"stamus-ctl/internal/stamus"
	"strings"
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
	inputAsFile, err := models.CreateFileInstance(conf, "values.yaml")
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
	groupedValues := groupValues(config, args)
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
	files, err := listFilesInFolder(conf)
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
	groupedFiles := groupStuff(files)
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

// Group values
// Utility function to group values from the config to nested maps
func groupValues(config *models.Config, args []string) map[string]interface{} {
	values := config.GetParams().GetValues(args...)
	groupedValues := make(map[string]interface{})
	for _, param := range config.GetParams().GetOrdered() {
		if value, ok := values[param]; ok {
			parts := strings.Split(param, ".")
			addToGroup(parts, value, groupedValues)
		}
	}
	return groupedValues
}

func groupStuff(stuff map[string]string) map[string]interface{} {
	groupedStuff := make(map[string]interface{})
	for key, value := range stuff {
		parts := strings.Split(key, "/")
		addToGroup(parts, value, groupedStuff)
	}
	return groupedStuff
}

func addToGroup(parts []string, value string, group map[string]interface{}) {
	if len(parts) == 1 {
		group[parts[0]] = value
	} else {
		if _, ok := group[parts[0]]; !ok {
			group[parts[0]] = make(map[string]interface{})
		}
		addToGroup(parts[1:], value, group[parts[0]].(map[string]interface{}))
	}
}

// Get files as map in a folder
func listFilesInFolder(folderPath string) (map[string]string, error) {
	filesMap := make(map[string]string)
	// Walk the folder
	err := filepath.Walk(folderPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() {
			relativePath, err := filepath.Rel(folderPath, path)
			if err != nil {
				return err
			}
			filesMap[relativePath] = info.Name()
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	return filesMap, nil
}
