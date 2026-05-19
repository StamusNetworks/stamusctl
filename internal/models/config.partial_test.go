package models

import (
	"testing"

	"stamus-ctl/internal/app"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
)

// TestPartialValues_InitContext tests that missing parameters use defaults when loading partial values file
func TestPartialValues_InitContext(t *testing.T) {
	// Create config template with defaults
	configContent := `
param1:
  usage: "Parameter 1"
  type: "string"
  default: "default_value1"
param2:
  usage: "Parameter 2"
  type: "int"
  default: 42
param3:
  usage: "Parameter 3"
  type: "string"
  default: "default_value3"
`
	err := afero.WriteFile(app.FS, "./test_init/config.yaml", []byte(configContent), 0o644)
	assert.NoError(t, err)

	// Create partial values file - only specify param1
	partialValuesContent := `
param1: custom_value1
stamus:
  config: ./test_init
  project: test_project
`
	err = afero.WriteFile(app.FS, "./test_init/partial_values.yaml", []byte(partialValuesContent), 0o644)
	assert.NoError(t, err)

	// Load config from template (simulating init context)
	confFile, err := CreateFile("./test_init", "config.yaml")
	assert.NoError(t, err)
	config, err := ConfigFromFile(confFile)
	assert.NoError(t, err)

	// Extract parameters
	_, _, err = config.ExtractParams()
	assert.NoError(t, err)

	// Load partial values file
	err = config.SetValuesFromFile("./test_init/partial_values.yaml")
	assert.NoError(t, err)

	// Set remaining parameters to defaults (simulating init behavior)
	err = config.GetParams().SetToDefaults()
	assert.NoError(t, err)

	// Verify results
	params := config.GetParams()

	// param1 should use value from partial file
	param1Val := params.Get("param1").Variable.String
	assert.NotNil(t, param1Val)
	assert.Equal(t, "custom_value1", *param1Val, "param1 should use value from partial file")

	// param2 should use default (not in partial file)
	param2Val := params.Get("param2").Variable.Int
	assert.NotNil(t, param2Val)
	assert.Equal(t, 42, *param2Val, "param2 should use default value")

	// param3 should use default (not in partial file)
	param3Val := params.Get("param3").Variable.String
	assert.NotNil(t, param3Val)
	assert.Equal(t, "default_value3", *param3Val, "param3 should use default value")
}

// TestPartialValues_UpdateContext tests that missing parameters preserve existing values
func TestPartialValues_UpdateContext(t *testing.T) {
	// Create config template
	configContent := `
param1:
  usage: "Parameter 1"
  type: "string"
  default: "default_value1"
param2:
  usage: "Parameter 2"
  type: "int"
  default: 42
param3:
  usage: "Parameter 3"
  type: "string"
  default: "default_value3"
`
	err := afero.WriteFile(app.FS, "./test_update/config.yaml", []byte(configContent), 0o644)
	assert.NoError(t, err)

	// Create existing config with current values
	existingValuesContent := `
param1: existing_value1
param2: 100
param3: existing_value3
stamus:
  config: ./test_update
  project: test_project
`
	err = afero.WriteFile(app.FS, "./test_update/existing_values.yaml", []byte(existingValuesContent), 0o644)
	assert.NoError(t, err)

	// Load existing config
	existingFile, err := CreateFile("./test_update", "existing_values.yaml")
	assert.NoError(t, err)
	existingConfig, err := LoadConfigFrom(existingFile, false)
	assert.NoError(t, err)

	// Create new config from template (simulating update context)
	newConfFile, err := CreateFile("./test_update", "config.yaml")
	assert.NoError(t, err)
	newConfig, err := ConfigFromFile(newConfFile)
	assert.NoError(t, err)
	_, _, err = newConfig.ExtractParams()
	assert.NoError(t, err)

	// Copy existing values to new config
	newConfig.GetParams().MergeValues(existingConfig.GetParams())

	// Debug: Check values after initial merge
	t.Logf("After initial merge - param1: %v, param2: %v, param3: %v",
		newConfig.GetParams().Get("param1").Variable,
		newConfig.GetParams().Get("param2").Variable,
		newConfig.GetParams().Get("param3").Variable)

	// Create partial values file - only update param1
	partialValuesContent := `
param1: updated_value1
stamus:
  config: ./test_update
  project: test_project
`
	err = afero.WriteFile(app.FS, "./test_update/partial_update.yaml", []byte(partialValuesContent), 0o644)
	assert.NoError(t, err)

	// Load partial values file
	err = newConfig.SetValuesFromFile("./test_update/partial_update.yaml")
	assert.NoError(t, err)

	// Debug: Check values after SetValuesFromFile
	t.Logf("After SetValuesFromFile - param1: %v, param2: %v, param3: %v",
		newConfig.GetParams().Get("param1").Variable,
		newConfig.GetParams().Get("param2").Variable,
		newConfig.GetParams().Get("param3").Variable)

	// Verify results
	params := newConfig.GetParams()

	// param1 should use new value from partial file
	param1 := params.Get("param1")
	assert.NotNil(t, param1, "param1 should exist")
	param1Val := param1.Variable.String
	assert.NotNil(t, param1Val, "param1 value should not be nil")
	assert.Equal(t, "updated_value1", *param1Val, "param1 should use updated value from partial file")

	// param2 should preserve existing value (not in partial file)
	param2 := params.Get("param2")
	assert.NotNil(t, param2, "param2 should exist")
	param2Val := param2.Variable.Int
	assert.NotNil(t, param2Val, "param2 value should not be nil")
	assert.Equal(t, 100, *param2Val, "param2 should preserve existing value")

	// param3 should preserve existing value (not in partial file)
	param3 := params.Get("param3")
	assert.NotNil(t, param3, "param3 should exist")
	param3Val := param3.Variable.String
	assert.NotNil(t, param3Val, "param3 value should not be nil")
	assert.Equal(t, "existing_value3", *param3Val, "param3 should preserve existing value")
}

// TestPartialValues_ArbitraryPreservation tests that arbitrary values are merged, not replaced
func TestPartialValues_ArbitraryPreservation(t *testing.T) {
	// Create config template
	configContent := `
param1:
  usage: "Parameter 1"
  type: "string"
  default: "default_value1"
`
	err := afero.WriteFile(app.FS, "./test_arbitrary/config.yaml", []byte(configContent), 0o644)
	assert.NoError(t, err)

	// Create existing config with arbitrary values
	existingValuesContent := `
param1: existing_value1
stamus:
  config: ./test_arbitrary
  project: test_project
custom:
  existing_key: existing_value
  shared_key: old_value
`
	err = afero.WriteFile(app.FS, "./test_arbitrary/existing_values.yaml", []byte(existingValuesContent), 0o644)
	assert.NoError(t, err)

	// Load existing config
	existingFile, err := CreateFile("./test_arbitrary", "existing_values.yaml")
	assert.NoError(t, err)
	existingConfig, err := LoadConfigFrom(existingFile, false)
	assert.NoError(t, err)

	// Create new config from template
	newConfFile, err := CreateFile("./test_arbitrary", "config.yaml")
	assert.NoError(t, err)
	newConfig, err := ConfigFromFile(newConfFile)
	assert.NoError(t, err)
	_, _, err = newConfig.ExtractParams()
	assert.NoError(t, err)

	// Copy existing values and arbitrary data
	newConfig.GetParams().MergeValues(existingConfig.GetParams())
	newConfig.MergeArbitrary(existingConfig.GetArbitrary().AsMap())

	// Create partial values file with new arbitrary values
	partialValuesContent := `
param1: updated_value1
stamus:
  config: ./test_arbitrary
  project: test_project
custom:
  new_key: new_value
  shared_key: updated_value
another:
  arbitrary_key: arbitrary_value
`
	err = afero.WriteFile(app.FS, "./test_arbitrary/partial_update.yaml", []byte(partialValuesContent), 0o644)
	assert.NoError(t, err)

	// Load partial values file
	err = newConfig.SetValuesFromFile("./test_arbitrary/partial_update.yaml")
	assert.NoError(t, err)

	// Verify results
	arbitrary := newConfig.GetArbitrary().AsMap()

	// Existing arbitrary key should be preserved
	assert.Equal(t, "existing_value", arbitrary["custom.existing_key"],
		"existing arbitrary key should be preserved")

	// New arbitrary key should be added
	assert.Equal(t, "new_value", arbitrary["custom.new_key"], "new arbitrary key should be added")

	// Shared key should be updated
	assert.Equal(t, "updated_value", arbitrary["custom.shared_key"],
		"shared arbitrary key should be updated")

	// Another arbitrary key should be added
	assert.Equal(t, "arbitrary_value", arbitrary["another.arbitrary_key"],
		"another arbitrary key should be added")
}

// TestPartialValues_EmptyFile tests that empty partial values file doesn't break anything
func TestPartialValues_EmptyFile(t *testing.T) {
	// Create config template
	configContent := `
param1:
  usage: "Parameter 1"
  type: "string"
  default: "default_value1"
`
	err := afero.WriteFile(app.FS, "./test_empty/config.yaml", []byte(configContent), 0o644)
	assert.NoError(t, err)

	// Create empty partial values file (only stamus config required)
	emptyValuesContent := `
stamus:
  config: ./test_empty
  project: test_project
`
	err = afero.WriteFile(app.FS, "./test_empty/empty_values.yaml", []byte(emptyValuesContent), 0o644)
	assert.NoError(t, err)

	// Load config from template
	confFile, err := CreateFile("./test_empty", "config.yaml")
	assert.NoError(t, err)
	config, err := ConfigFromFile(confFile)
	assert.NoError(t, err)

	// Extract parameters
	_, _, err = config.ExtractParams()
	assert.NoError(t, err)

	// Load empty partial values file
	err = config.SetValuesFromFile("./test_empty/empty_values.yaml")
	assert.NoError(t, err)

	// Set defaults
	err = config.GetParams().SetToDefaults()
	assert.NoError(t, err)

	// Verify param1 uses default
	params := config.GetParams()
	param1Val := params.Get("param1").Variable.String
	assert.NotNil(t, param1Val)
	assert.Equal(t, "default_value1", *param1Val, "param1 should use default value when not in empty partial file")
}

// TestPartialValues_OnlyArbitraryKeys tests partial file with only arbitrary keys
func TestPartialValues_OnlyArbitraryKeys(t *testing.T) {
	// Create config template
	configContent := `
param1:
  usage: "Parameter 1"
  type: "string"
  default: "default_value1"
`
	err := afero.WriteFile(app.FS, "./test_arbitrary_only/config.yaml", []byte(configContent), 0o644)
	assert.NoError(t, err)

	// Create partial values file with only arbitrary keys
	arbitraryOnlyContent := `
stamus:
  config: ./test_arbitrary_only
  project: test_project
custom:
  arbitrary_key1: value1
  arbitrary_key2: value2
`
	err = afero.WriteFile(app.FS, "./test_arbitrary_only/arbitrary_only.yaml", []byte(arbitraryOnlyContent), 0o644)
	assert.NoError(t, err)

	// Load config from template
	confFile, err := CreateFile("./test_arbitrary_only", "config.yaml")
	assert.NoError(t, err)
	config, err := ConfigFromFile(confFile)
	assert.NoError(t, err)

	// Extract parameters
	_, _, err = config.ExtractParams()
	assert.NoError(t, err)

	// Load partial values file with only arbitrary keys
	err = config.SetValuesFromFile("./test_arbitrary_only/arbitrary_only.yaml")
	assert.NoError(t, err)

	// Set defaults for parameters
	err = config.GetParams().SetToDefaults()
	assert.NoError(t, err)

	// Verify param1 uses default
	params := config.GetParams()
	param1Val := params.Get("param1").Variable.String
	assert.NotNil(t, param1Val)
	assert.Equal(t, "default_value1", *param1Val, "param1 should use default value")

	// Verify arbitrary keys are set
	arbitrary := config.GetArbitrary().AsMap()
	assert.Equal(t, "value1", arbitrary["custom.arbitrary_key1"], "arbitrary key 1 should be set")
	assert.Equal(t, "value2", arbitrary["custom.arbitrary_key2"], "arbitrary key 2 should be set")
}

// TestMergeValues tests that MergeValues only updates existing parameters
func TestMergeValues(t *testing.T) {
	// Create source parameters
	source := &Parameters{
		"param1": &Parameter{
			Type:         "string",
			Variable:     CreateVariableString("new_value1"),
			ValidateFunc: func(v Variable) bool { return true },
		},
		"param2": &Parameter{
			Type:         "int",
			Variable:     CreateVariableInt(200),
			ValidateFunc: func(v Variable) bool { return true },
		},
		"param3": &Parameter{
			Type:         "string",
			Variable:     CreateVariableString("new_value3"),
			ValidateFunc: func(v Variable) bool { return true },
		},
	}

	// Create target parameters (missing param3)
	target := &Parameters{
		"param1": &Parameter{
			Type:         "string",
			Variable:     CreateVariableString("old_value1"),
			ValidateFunc: func(v Variable) bool { return true },
		},
		"param2": &Parameter{
			Type:         "int",
			Variable:     CreateVariableInt(100),
			ValidateFunc: func(v Variable) bool { return true },
		},
	}

	// Merge values
	target.MergeValues(source)

	// Verify param1 was updated
	assert.Equal(t, "new_value1", *target.Get("param1").Variable.String, "param1 should be updated")

	// Verify param2 was updated
	assert.Equal(t, 200, *target.Get("param2").Variable.Int, "param2 should be updated")

	// Verify param3 was NOT added (target doesn't have it)
	assert.Nil(t, target.Get("param3"), "param3 should not be added to target")
}

// TestMergeArbitrary tests that MergeArbitrary merges values correctly
func TestMergeArbitrary(t *testing.T) {
	// Create config with existing arbitrary values
	config := &Config{
		arbitrary: &Arbitrary{
			"existing.key1": "value1",
			"existing.key2": "value2",
		},
	}

	// Merge new arbitrary values
	newArbitrary := map[string]any{
		"existing.key2": "updated_value2", // Should update
		"new.key3":      "value3",         // Should add
		"new.key4":      "value4",         // Should add
	}
	config.MergeArbitrary(newArbitrary)

	// Verify results
	arbitrary := config.GetArbitrary().AsMap()

	// Existing key1 should be preserved
	assert.Equal(t, "value1", arbitrary["existing.key1"], "existing.key1 should be preserved")

	// Existing key2 should be updated
	assert.Equal(t, "updated_value2", arbitrary["existing.key2"], "existing.key2 should be updated")

	// New keys should be added
	assert.Equal(t, "value3", arbitrary["new.key3"], "new.key3 should be added")
	assert.Equal(t, "value4", arbitrary["new.key4"], "new.key4 should be added")
}
