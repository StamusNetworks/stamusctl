package models

import (
	// Common
	"fmt"
	"log"
	"os"
	"path/filepath"
	"stamus-ctl/internal/logging"
	"strings"

	// External

	"github.com/spf13/viper"
	// Custom
)

// Config is a struct that represents a configuration file
// It contains the path to the file, the arbitrary values, the parameters values and the viper instnace to interact with the file
// It can be used to get or set values, validates them etc
type Config struct {
	file          File
	project       string
	arbitrary     *Arbitrary
	parameters    *Parameters
	viperInstance *viper.Viper
}

// Create a new config instance from a file
func ConfigFromFile(file File) (*Config, error) {
	viperInstance, err := InstanciateViper(file)
	if err != nil {
		return nil, err
	}
	return NewConfig(file, viperInstance), nil
}

func NewConfig(file File, viperInstance *viper.Viper) *Config {
	return &Config{
		file:          file,
		viperInstance: viperInstance,
		arbitrary:     &Arbitrary{},
	}
}

// Create a new config instance from a path, extract the values and return the instance
// Reload is used to not keep the arbitrary values
func LoadConfigFrom(path File, reload bool) (*Config, error) {
	// Load the config
	configured, err := ConfigFromFile(path)
	if err != nil {
		return nil, err
	}
	// Extract config data
	values := configured.ExtractValues()
	file, err := GetStamusFile(values)
	if err != nil {
		return nil, err
	}
	// Get project
	project := values["stamus.project"]
	projectName := *project.String
	// Load origin config
	originConf, err := ConfigFromFile(file)
	if err != nil {
		return nil, err
	}
	_, _, err = originConf.ExtractParams()
	if err != nil {
		return nil, err
	}
	// Set arbitrary
	if !reload {
		for key, value := range values {
			originConf.arbitrary.SetArbitrary(map[string]string{key: value.AsString()})
		}
	}
	// Merge
	originConf.parameters.SetValues(values)
	originConf.SetProject(projectName)
	return originConf, nil
}

func GetStamusFile(values map[string]*Variable) (File, error) {
	stamusConfPathPointer := values["stamus.config"]
	if stamusConfPathPointer == nil {
		return File{}, fmt.Errorf("stamus.config not found")
	}
	stamusConfPath := *stamusConfPathPointer.String
	file, err := CreateFile(stamusConfPath, "config.yaml")
	if err != nil {
		return File{}, err
	}
	return file, nil
}

// Returns the parameter extracted from the config file
func (f *Config) extractParam(parameter string) (*Parameter, error) {
	// Extract parameter
	currentParam := f.extracParamOverview(parameter)
	// Get choices
	choices, err := GetChoices(f.getStringParamValue(parameter, "choices"))
	if err != nil {
		return nil, err
	}
	currentParam.Choices = choices
	if parameter == "suricata.interfaces" {
		asStrings := []string{}
		for _, choice := range choices {
			asStrings = append(asStrings, *choice.String)
		}
		currentParam.Default = CreateVariableString(strings.Join(asStrings, ","))
	}
	return &currentParam, nil
}

// Extract parameters and includes from the config file
func (f *Config) ExtractParams() (*Parameters, []string, error) {
	// To return
	var parameters Parameters = make(Parameters)
	var includes []string = []string{}
	// Extract lists
	includesList, parametersList := f.extracKeys()
	includes = append(includes, includesList...)
	// Extract parameters
	for _, parameter := range parametersList {
		param, err := f.extractParam(parameter)
		if err != nil {
			return nil, nil, err
		}
		parameters.AddAsParameter(parameter, param)
	}
	// Extract includes parameters
	for _, include := range includesList {
		// Create config instance for the include
		file, err := CreateFileFromPath(filepath.Join(f.file.Path, include))
		if err != nil {
			return nil, nil, err
		}
		conf, err := ConfigFromFile(file)
		if err != nil {
			return nil, nil, err
		}
		// Extract parameters
		fileParams, fileIncludes, err := conf.ExtractParams()
		if err != nil {
			return nil, nil, err
		}
		// Merge data
		parameters.AddAsParameters(fileParams)
		includes = append(includes, fileIncludes...)
	}
	f.parameters = &parameters
	return &parameters, includes, nil
}

// Save the config to a folder
func (f *Config) SaveConfigTo(dest File, isUpgrade, isInstall bool) error {
	// Get Data
	logger := logging.Sugar.With("dest", dest.completePath(), "isUpgrade", isUpgrade, "isInstall", isInstall)
	configData, err := f.GetData()
	if err != nil {
		return err
	}
	releaseData, err := GetReleaseData(dest, isUpgrade, isInstall)
	if err != nil {
		return err
	}
	templateData := NewTemplate(f.project, f.file.Path).AsMap()

	// Merge data
	var data = map[string]any{}
	for key, value := range configData {
		data[key] = value
	}
	for key, value := range releaseData {
		data[key] = value
	}
	for key, value := range templateData {
		data[key] = value
	}
	data = nestMap(data)

	// Process templates
	err = processTemplates(f.file.Path, dest.Path, data)
	if err != nil {
		logger.Error(err)
		return err
	}

	// Save parameters values to config file
	logger.Debug("saving config")
	err = f.saveParamsTo(dest)
	if err != nil {
		return err
	}

	// Clean destination folder
	err = f.Clean(dest)
	if err != nil {
		return err
	}

	return nil
}

func (f *Config) GetData() (map[string]any, error) {
	var data = map[string]any{}
	var configData = map[string]any{}
	for key, param := range *f.parameters {
		value, err := param.GetValue()
		if err != nil {
			logging.Sugar.Info("Error getting parameter value", key, err)
			fmt.Printf("Use %s=<your value> to set it\n", key)
			return nil, err
		}
		configData[key] = value
	}
	// Merge with arbitrary config values and create a nested map
	for key, value := range f.arbitrary.AsMap() {
		data[addValuePrefix(key)] = value
	}
	for key, value := range configData {
		data[addValuePrefix(key)] = value
	}
	return data, nil
}

func GetReleaseData(dest File, isUpgrade, isInstall bool) (map[string]any, error) {
	currentDir, err := os.Getwd()
	if err != nil {
		return nil, err
	}
	return getRelease(dest, currentDir, isUpgrade, isInstall).AsMap(), nil
}

// Set values from a file (values.yaml)
func (f *Config) SetValuesFromFile(valuesPath string) error {
	if valuesPath != "" {
		logging.Sugar.Info("Loading values from", valuesPath)
		file, err := CreateFileFromPath(valuesPath)
		if err != nil {
			return err
		}
		valuesConf, err := LoadConfigFrom(file, false)
		if err != nil {
			return err
		}
		f.GetParams().MergeValues(valuesConf.GetParams())
		f.MergeArbitrary(valuesConf.GetArbitrary().AsMap())
	}
	return nil
}

// Set specific values from files content
func (f *Config) SetValuesFromFiles(fromFiles string) error {
	if fromFiles == "" {
		return nil
	}
	// For each fromFile
	args := strings.Split(fromFiles, " ")
	args = removeEmptyStrings(args)
	values := make(map[string]*Variable)
	asMap := make(map[string]string)
	for _, arg := range args {
		// Split argument
		split := strings.Split(arg, "=")
		if len(split) != 2 {
			return fmt.Errorf("invalid argument: %s of %s. Must be parameter.subparameter=./folder/file", arg, args)
		}
		// Get file content
		content, err := os.ReadFile(split[1])
		if err != nil {
			return err
		}
		// Set value of parameter
		temp := CreateVariableString(string(content))
		values[split[0]] = &temp
		asMap[split[0]] = string(content)
	}
	f.GetParams().SetValues(values)
	f.GetArbitrary().SetArbitrary(asMap)
	return nil
}

// Cleans the config folder
func (f *Config) Clean(folder File) error {
	// Get list of all included subconfigs
	_, includes, err := f.ExtractParams()
	if err != nil {
		return err
	}
	// Delete all included subconfig files
	for _, include := range includes {
		err := os.Remove(filepath.Join(folder.Path, include))
		if err != nil {
			return err
		}
	}
	// Clean destination folder
	err = deleteEmptyFiles(folder.Path)
	if err != nil {
		return err
	}
	err = deleteEmptyFolders(folder.Path)
	if err != nil {
		return err
	}
	return nil
}

// Save parameters values to config file
func (f *Config) saveParamsTo(dest File) error {
	//Clear the file
	os.Remove(dest.completePath())
	//ReCreate the file
	file, err := os.Create(dest.completePath())
	if err != nil {
		logging.Sugar.Info("Error creating config file", err)
		return err
	}
	defer file.Close()

	// Instanciate config to dest file
	conf, err := ConfigFromFile(dest)
	if err != nil {
		logging.Sugar.Info("Error creating config instance", err)
		return err
	}
	conf.parameters = f.parameters
	//Get current config parameters values
	paramsValues := make(map[string]any)
	for key, param := range *conf.parameters {
		value, err := param.GetValue()
		if err != nil {
			logging.Sugar.Info("Error getting parameter value", key, err)
			return fmt.Errorf("error getting parameter value %s. Use %s=<your value> to set it", key, key)
		}
		paramsValues[key] = value
	}
	// Set the new values
	for key, value := range f.arbitrary.AsMap() {
		conf.viperInstance.Set(key, value)
	}
	for key, value := range paramsValues {
		conf.viperInstance.Set(key, value)
	}

	// Project
	conf.viperInstance.Set("stamus.project", f.project)
	// If latest, set stamus.config value to version
	path := removeEmptyStrings(strings.Split(f.file.Path, "/"))
	if path[len(path)-1] == "latest" {
		// Get version
		version, err := os.ReadFile(filepath.Join(f.file.Path, "version"))
		if err != nil {
			log.Println("Error reading version file", err)
			return err
		}
		// Set stamusconfig value to version
		var versionPath []string = append([]string{}, path...)
		copy(versionPath, path)
		versionPath[len(versionPath)-1] = string(version)
		conf.viperInstance.Set("stamus.config", "/"+filepath.Join(versionPath...))
	} else {
		conf.viperInstance.Set("stamus.config", f.file.Path)
	}
	// Write the new config file
	err = conf.viperInstance.WriteConfig()
	if err != nil {
		logging.Sugar.Info("Error writing config file", err)
		return err
	}

	return nil
}

func (f *Config) DeleteFolder() error {
	return os.RemoveAll(f.file.Path)
}
