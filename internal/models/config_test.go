package models

import (
	"log"
	"testing"

	"stamus-ctl/internal/app"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
)

func TestSaveParamsTo(t *testing.T) {
	// Create a dummy config file
	file, err := CreateFile("./folder", "config.yaml")
	assert.NoError(t, err)
	file.InstanciateViper()
	assert.NoError(t, err)

	// Create a Config instance
	config := &Config{
		file:    file,
		project: "test_project",
		arbitrary: &Arbitrary{
			"arbite": "arbiteval",
		},
		parameters: &Parameters{
			"param1": &Parameter{
				Variable: CreateVariableString("value1"),
				Type:     "string",
			},
			"param2": &Parameter{
				Variable: CreateVariableInt(42),
				Type:     "int",
			},
		},
	}

	// Destination file
	destFile, err := CreateFile("/dest", "config.yaml")
	assert.NoError(t, err)
	destFile.InstanciateViper()
	assert.NoError(t, err)

	// Call saveParamsTo
	err = config.saveParamsTo(destFile)
	assert.NoError(t, err)

	// Verify the file was created
	exists, err := afero.Exists(app.FS, destFile.completePath())
	assert.NoError(t, err)
	assert.True(t, exists)

	// Verify the content of the file
	assert.Equal(t, "value1", destFile.GetViper().GetString("param1"))
	assert.Equal(t, "arbiteval", destFile.GetViper().GetString("arbite"))
	assert.Equal(t, 42, destFile.GetViper().GetInt("param2"))
	assert.Equal(t, "test_project", destFile.GetViper().GetString("stamus.project"))
	assert.Equal(t, "./folder", destFile.GetViper().GetString("stamus.config"))
}

func TestSaveParamsTo_LatestVersion(t *testing.T) {
	// Create a dummy config file in "latest" directory
	filePath := "/latest"
	file, err := CreateFile(filePath, "config.yaml")
	assert.NoError(t, err)

	// Create a version file
	versionPath := "/latest/version"
	err = afero.WriteFile(app.FS, versionPath, []byte("1.0.0"), 0o644)
	assert.NoError(t, err)

	// Create a Config instance
	config := &Config{
		file:    file,
		project: "test_project",
		arbitrary: &Arbitrary{
			"arbite": "arbiteval",
		},
		parameters: &Parameters{
			"param1": &Parameter{
				Variable: CreateVariableString("value1"),
				Type:     "string",
			},
			"param2": &Parameter{
				Variable: CreateVariableInt(42),
				Type:     "int",
			},
		},
	}

	// Destination file
	destFile, err := CreateFile("/dest_config.yaml", "config.yaml")
	assert.NoError(t, err)

	// Call saveParamsTo
	err = config.saveParamsTo(destFile)
	assert.NoError(t, err)

	// Verify the file was created
	exists, err := afero.Exists(app.FS, destFile.completePath())
	assert.NoError(t, err)
	assert.True(t, exists)

	// Verify the content of the file
	viperInstance, err := destFile.InstanciateViper()
	assert.NoError(t, err)
	assert.Equal(t, "value1", viperInstance.GetString("param1"))
	assert.Equal(t, 42, viperInstance.GetInt("param2"))
	assert.Equal(t, "test_project", viperInstance.GetString("stamus.project"))
	assert.Equal(t, "/1.0.0", viperInstance.GetString("stamus.config"))
}

func TestSetValuesFromFiles(t *testing.T) {
	// Create a dummy content file
	contentFile1 := "i'm some dummy content"
	err := afero.WriteFile(app.FS, "./folder/file1.txt", []byte(contentFile1), 0o644)
	assert.NoError(t, err)
	contentFile2 := "and i'm some other dummy content"
	err = afero.WriteFile(app.FS, "./folder/file2.txt", []byte(contentFile2), 0o644)
	assert.NoError(t, err)

	// Create a Config instance
	config := &Config{
		arbitrary: &Arbitrary{},
		parameters: &Parameters{
			"param1": &Parameter{
				Variable:     CreateVariableString("value1"),
				Type:         "string",
				ValidateFunc: func(variable Variable) bool { return true },
			},
		},
	}

	// Call SetValuesFromFiles
	err = config.SetValuesFromFiles("param1=./folder/file1.txt param2=./folder/file2.txt")
	assert.NoError(t, err)

	log.Println("arbite", config.arbitrary.AsMap()["param1"])
	// Verify the parameters were set correctly
	assert.Equal(t, contentFile1, config.parameters.GetValues()["param1"])
	assert.Equal(t, contentFile1, config.arbitrary.AsMap()["param1"])
	assert.Equal(t, contentFile2, config.arbitrary.AsMap()["param2"])
}

func TestSetValuesFromFile(t *testing.T) {
	// Create a dummy values file
	valuesContent := `
param1: value1
param2: 123
arbite: arbiteval
stamus:
    config: ./folder2
    project: test_project
`
	err := afero.WriteFile(app.FS, "./folder/values.yaml", []byte(valuesContent), 0o644)
	assert.NoError(t, err)
	configContent := `
param1:
    usage: "usage"
    type: "string"
    default: "default"
param2:
    usage: "usage"
    type: "int"
    default: 42
`
	err = afero.WriteFile(app.FS, "./folder2/config.yaml", []byte(configContent), 0o644)
	assert.NoError(t, err)

	// Create a Config instance
	config := &Config{
		arbitrary: &Arbitrary{},
		parameters: &Parameters{
			"param1": &Parameter{
				Variable:     CreateVariableString("old_value1"),
				Type:         "string",
				ValidateFunc: func(variable Variable) bool { return true },
			},
			"param2": &Parameter{
				Variable:     CreateVariableInt(0),
				Type:         "int",
				ValidateFunc: func(variable Variable) bool { return true },
			},
		},
	}

	// Call SetValuesFromFile
	err = config.SetValuesFromFile("./folder/values.yaml")
	assert.NoError(t, err)

	log.Println("config", config.parameters.GetValues())

	// Verify the parameters were set correctly
	assert.Equal(t, "value1", config.parameters.GetValues()["param1"])
	assert.Equal(t, "123", config.parameters.GetValues()["param2"])
	assert.Equal(t, "arbiteval", config.arbitrary.AsMap()["arbite"])
}

func TestGetData(t *testing.T) {
	// Create a Config instance
	config := &Config{
		arbitrary: &Arbitrary{
			"arbite": "arbiteval",
		},
		parameters: &Parameters{
			"param1": &Parameter{
				Variable: CreateVariableString("value1"),
				Type:     "string",
			},
			"param2": &Parameter{
				Variable: CreateVariableInt(42),
				Type:     "int",
			},
		},
	}

	// Call GetData
	data, err := config.GetData()
	assert.NoError(t, err)

	log.Println("data", data)

	// Verify the data
	assert.Equal(t, "value1", data["Values.param1"])
	assert.Equal(t, 42, data["Values.param2"])
	assert.Equal(t, "arbiteval", data["Values.arbite"])
}

func TestSaveConfigTo(t *testing.T) {
	// Destination file
	destFile, err := CreateFile("/dest", "config.yaml")
	assert.NoError(t, err)
	destFile.InstanciateViper()
	err = afero.WriteFile(app.FS, "./dest/values.yaml", []byte(""), 0o644)
	assert.NoError(t, err)
	err = afero.WriteFile(app.FS, "./folder/values.yaml", []byte(""), 0o644)
	assert.NoError(t, err)

	// Create a Config instance
	config := &Config{
		file: &File{
			Path: "folder",
			Name: "values",
			Type: "yaml",
		},
		project: "test_project",
		arbitrary: &Arbitrary{
			"arbite": "arbiteval",
		},
		parameters: &Parameters{
			"param1": &Parameter{
				Variable: CreateVariableString("value1"),
				Type:     "string",
			},
			"param2": &Parameter{
				Variable: CreateVariableInt(42),
				Type:     "int",
			},
		},
	}

	// Call SaveConfigTo
	err = config.SaveConfigTo(destFile, true, false)
	assert.NoError(t, err)

	// Verify the file was created
	exists, err := afero.Exists(app.FS, destFile.completePath())
	assert.NoError(t, err)
	assert.True(t, exists)

	log.Println("destFile", destFile.GetViper().AllKeys())

	// Verify the content of the file
	assert.Equal(t, "value1", destFile.GetViper().GetString("param1"))
	assert.Equal(t, "arbiteval", destFile.GetViper().GetString("arbite"))
	assert.Equal(t, 42, destFile.GetViper().GetInt("param2"))
	assert.Equal(t, "test_project", destFile.GetViper().GetString("stamus.project"))
	assert.Equal(t, "folder", destFile.GetViper().GetString("stamus.config"))
}

func TestExtractParams(t *testing.T) {
	// Create a dummy config file
	configContent := `
param1:
  usage: "usage"
  type: "string"
  default: "default"
param2:
  usage: "usage"
  type: "int"
  default: 42
includes:
  - include.yaml
`
	err := afero.WriteFile(app.FS, "./folder/config.yaml", []byte(configContent), 0o644)
	assert.NoError(t, err)

	// Create a dummy include file
	includeContent := `
param3:
  usage: "usage"
  type: "string"
  default: "default"
`
	err = afero.WriteFile(app.FS, "./folder/include.yaml", []byte(includeContent), 0o644)
	assert.NoError(t, err)

	// Create a Config instance
	file, err := CreateFile("./folder", "config.yaml")
	assert.NoError(t, err)
	file.InstanciateViper()
	assert.NoError(t, err)

	config := &Config{
		file: file,
	}

	// Call ExtractParams
	params, includes, err := config.ExtractParams()
	assert.NoError(t, err)

	// Verify the parameters
	assert.Equal(t, "usage", (*params.Get("param1")).Usage)
	assert.Equal(t, "usage", (*params.Get("param2")).Usage)
	assert.Equal(t, "usage", (*params.Get("param3")).Usage)
	assert.Equal(t, "string", (*params.Get("param1")).Type)
	assert.Equal(t, "int", (*params.Get("param2")).Type)
	assert.Equal(t, "string", (*params.Get("param3")).Type)
	assert.Equal(t, "default", *(*params.Get("param1")).Default.String)
	assert.Equal(t, 42, *(*params.Get("param2")).Default.Int)
	assert.Equal(t, "default", *(*params.Get("param3")).Default.String)

	// Verify the includes
	assert.Contains(t, includes, "include.yaml")
}
