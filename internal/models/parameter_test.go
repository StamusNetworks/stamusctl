package models

import (
	"strconv"
	"testing"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

func TestVariable_AsString(t *testing.T) {
	tests := []struct {
		name     string
		variable Variable
		expected string
	}{
		{
			name:     "String value",
			variable: Variable{String: strPtr("test")},
			expected: "test",
		},
		{
			name:     "Bool value true",
			variable: Variable{Bool: boolPtr(true)},
			expected: "true",
		},
		{
			name:     "Bool value false",
			variable: Variable{Bool: boolPtr(false)},
			expected: "false",
		},
		{
			name:     "Int value",
			variable: Variable{Int: intPtr(123)},
			expected: "123",
		},
		{
			name:     "Nil value",
			variable: Variable{},
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.variable.AsString()
			if result != tt.expected {
				t.Errorf("expected %s, got %s", tt.expected, result)
			}
		})
	}
}

func strPtr(s string) *string {
	return &s
}
func boolPtr(b bool) *bool {
	return &b
}
func intPtr(i int) *int {
	return &i
}

func TestParameter_GetValue(t *testing.T) {
	tests := []struct {
		name        string
		parameter   Parameter
		expected    any
		expectError bool
	}{
		{
			name: "String value set",
			parameter: Parameter{
				Type:     "string",
				Variable: CreateVariableString("test"),
			},
			expected:    "test",
			expectError: false,
		},
		{
			name: "Bool value set to true",
			parameter: Parameter{
				Type:     "bool",
				Variable: CreateVariableBool(true),
			},
			expected:    true,
			expectError: false,
		},
		{
			name: "Bool value set to false",
			parameter: Parameter{
				Type:     "bool",
				Variable: CreateVariableBool(false),
			},
			expected:    false,
			expectError: false,
		},
		{
			name: "Int value set",
			parameter: Parameter{
				Type:     "int",
				Variable: CreateVariableInt(123),
			},
			expected:    123,
			expectError: false,
		},
		{
			name: "Default string value",
			parameter: Parameter{
				Type:    "string",
				Default: CreateVariableString("default"),
			},
			expected:    "default",
			expectError: false,
		},
		{
			name: "Default bool value",
			parameter: Parameter{
				Type:    "bool",
				Default: CreateVariableBool(true),
			},
			expected:    true,
			expectError: false,
		},
		{
			name: "Default int value",
			parameter: Parameter{
				Type:    "int",
				Default: CreateVariableInt(456),
			},
			expected:    456,
			expectError: false,
		},
		{
			name: "Variable not set",
			parameter: Parameter{
				Type: "string",
			},
			expected:    nil,
			expectError: true,
		},
		{
			name: "Invalid type",
			parameter: Parameter{
				Type:     "invalid",
				Variable: CreateVariableString("test"),
			},
			expected:    nil,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := tt.parameter.GetValue()
			if (err != nil) != tt.expectError {
				t.Errorf("expected error: %v, got: %v", tt.expectError, err)
			}
			if result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestParameter_getValue(t *testing.T) {
	tests := []struct {
		name      string
		parameter Parameter
		expected  any
	}{
		{
			name: "String value set",
			parameter: Parameter{
				Type:     "string",
				Variable: CreateVariableString("test"),
			},
			expected: "test",
		},
		{
			name: "Bool value set to true",
			parameter: Parameter{
				Type:     "bool",
				Variable: CreateVariableBool(true),
			},
			expected: true,
		},
		{
			name: "Bool value set to false",
			parameter: Parameter{
				Type:     "bool",
				Variable: CreateVariableBool(false),
			},
			expected: false,
		},
		{
			name: "Int value set",
			parameter: Parameter{
				Type:     "int",
				Variable: CreateVariableInt(123),
			},
			expected: 123,
		},
		{
			name: "Default string value",
			parameter: Parameter{
				Type:    "string",
				Default: CreateVariableString("default"),
			},
			expected: "default",
		},
		{
			name: "Default bool value",
			parameter: Parameter{
				Type:    "bool",
				Default: CreateVariableBool(true),
			},
			expected: true,
		},
		{
			name: "Default int value",
			parameter: Parameter{
				Type:    "int",
				Default: CreateVariableInt(456),
			},
			expected: 456,
		},
		{
			name: "Variable not set",
			parameter: Parameter{
				Type: "string",
			},
			expected: nil,
		},
		{
			name: "Invalid type",
			parameter: Parameter{
				Type:     "invalid",
				Variable: CreateVariableString("test"),
			},
			expected: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.parameter.getValue()
			if result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestParameter_AddAsFlag(t *testing.T) {
	tests := []struct {
		name       string
		parameter  Parameter
		persistent bool
	}{
		{
			name: "Add string flag",
			parameter: Parameter{
				Name:     "test-string",
				Type:     "string",
				Usage:    "test string flag",
				Variable: CreateVariableString("default"),
			},
			persistent: false,
		},
		{
			name: "Add bool flag",
			parameter: Parameter{
				Name:     "test-bool",
				Type:     "bool",
				Usage:    "test bool flag",
				Variable: CreateVariableBool(true),
			},
			persistent: false,
		},
		{
			name: "Add int flag",
			parameter: Parameter{
				Name:     "test-int",
				Type:     "int",
				Usage:    "test int flag",
				Variable: CreateVariableInt(123),
			},
			persistent: false,
		},
		{
			name: "Add hidden flag",
			parameter: Parameter{
				Name:     "test-hidden",
				Type:     "string",
				Usage:    "test hidden flag",
				Variable: CreateVariableString("default"),
				Hidden:   true,
			},
			persistent: false,
		},
		{
			name: "Add persistent flag",
			parameter: Parameter{
				Name:     "test-persistent",
				Type:     "string",
				Usage:    "test persistent flag",
				Variable: CreateVariableString("default"),
			},
			persistent: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := &cobra.Command{}
			tt.parameter.AddAsFlag(cmd, tt.persistent)

			var flag *pflag.Flag
			if tt.persistent {
				flag = cmd.PersistentFlags().Lookup(tt.parameter.Name)
			} else {
				flag = cmd.Flags().Lookup(tt.parameter.Name)
			}

			if flag == nil {
				t.Errorf("expected flag %s to be added", tt.parameter.Name)
			}

			if tt.parameter.Hidden {
				if !flag.Hidden {
					t.Errorf("expected flag %s to be hidden", tt.parameter.Name)
				}
			}
		})
	}
}
func TestParameter_AddStringFlag(t *testing.T) {
	tests := []struct {
		name       string
		parameter  Parameter
		persistent bool
	}{
		{
			name: "Add string flag without shorthand",
			parameter: Parameter{
				Name:    "test-string",
				Type:    "string",
				Usage:   "test string flag",
				Default: CreateVariableString("default"),
			},
			persistent: false,
		},
		{
			name: "Add string flag with shorthand",
			parameter: Parameter{
				Name:      "test-string",
				Type:      "string",
				Usage:     "test string flag",
				Default:   CreateVariableString("default"),
				Shorthand: "t",
			},
			persistent: false,
		},
		{
			name: "Add persistent string flag without shorthand",
			parameter: Parameter{
				Name:    "test-string",
				Type:    "string",
				Usage:   "test string flag",
				Default: CreateVariableString("default"),
			},
			persistent: true,
		},
		{
			name: "Add persistent string flag with shorthand",
			parameter: Parameter{
				Name:      "test-string",
				Type:      "string",
				Usage:     "test string flag",
				Default:   CreateVariableString("default"),
				Shorthand: "t",
			},
			persistent: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := &cobra.Command{}
			tt.parameter.AddStringFlag(cmd, tt.persistent)

			var flag *pflag.Flag
			if tt.persistent {
				flag = cmd.PersistentFlags().Lookup(tt.parameter.Name)
			} else {
				flag = cmd.Flags().Lookup(tt.parameter.Name)
			}

			if flag == nil {
				t.Errorf("expected flag %s to be added", tt.parameter.Name)
			}

			if flag.Value.String() != *tt.parameter.Default.String {
				t.Errorf("expected flag value %s, got %s", *tt.parameter.Default.String, flag.Value.String())
			}

			if tt.parameter.Shorthand != "" && flag.Shorthand != tt.parameter.Shorthand {
				t.Errorf("expected flag shorthand %s, got %s", tt.parameter.Shorthand, flag.Shorthand)
			}
		})
	}
}
func TestParameter_AddBoolFlag(t *testing.T) {
	tests := []struct {
		name       string
		parameter  Parameter
		persistent bool
	}{
		{
			name: "Add bool flag without shorthand",
			parameter: Parameter{
				Name:    "test-bool",
				Type:    "bool",
				Usage:   "test bool flag",
				Default: CreateVariableBool(true),
			},
			persistent: false,
		},
		{
			name: "Add bool flag with shorthand",
			parameter: Parameter{
				Name:      "test-bool",
				Type:      "bool",
				Usage:     "test bool flag",
				Default:   CreateVariableBool(true),
				Shorthand: "b",
			},
			persistent: false,
		},
		{
			name: "Add persistent bool flag without shorthand",
			parameter: Parameter{
				Name:    "test-bool",
				Type:    "bool",
				Usage:   "test bool flag",
				Default: CreateVariableBool(true),
			},
			persistent: true,
		},
		{
			name: "Add persistent bool flag with shorthand",
			parameter: Parameter{
				Name:      "test-bool",
				Type:      "bool",
				Usage:     "test bool flag",
				Default:   CreateVariableBool(true),
				Shorthand: "b",
			},
			persistent: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := &cobra.Command{}
			tt.parameter.AddBoolFlag(cmd, tt.persistent)

			var flag *pflag.Flag
			if tt.persistent {
				flag = cmd.PersistentFlags().Lookup(tt.parameter.Name)
			} else {
				flag = cmd.Flags().Lookup(tt.parameter.Name)
			}

			if flag == nil {
				t.Errorf("expected flag %s to be added", tt.parameter.Name)
			}

			if flag.Value.String() != strconv.FormatBool(*tt.parameter.Default.Bool) {
				t.Errorf("expected flag value %v, got %v", *tt.parameter.Default.Bool, flag.Value.String())
			}

			if tt.parameter.Shorthand != "" && flag.Shorthand != tt.parameter.Shorthand {
				t.Errorf("expected flag shorthand %s, got %s", tt.parameter.Shorthand, flag.Shorthand)
			}
		})
	}
}
func TestParameter_AddIntFlag(t *testing.T) {
	tests := []struct {
		name       string
		parameter  Parameter
		persistent bool
	}{
		{
			name: "Add int flag without shorthand",
			parameter: Parameter{
				Name:    "test-int",
				Type:    "int",
				Usage:   "test int flag",
				Default: CreateVariableInt(123),
			},
			persistent: false,
		},
		{
			name: "Add int flag with shorthand",
			parameter: Parameter{
				Name:      "test-int",
				Type:      "int",
				Usage:     "test int flag",
				Default:   CreateVariableInt(123),
				Shorthand: "i",
			},
			persistent: false,
		},
		{
			name: "Add persistent int flag without shorthand",
			parameter: Parameter{
				Name:    "test-int",
				Type:    "int",
				Usage:   "test int flag",
				Default: CreateVariableInt(123),
			},
			persistent: true,
		},
		{
			name: "Add persistent int flag with shorthand",
			parameter: Parameter{
				Name:      "test-int",
				Type:      "int",
				Usage:     "test int flag",
				Default:   CreateVariableInt(123),
				Shorthand: "i",
			},
			persistent: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := &cobra.Command{}
			tt.parameter.AddIntFlag(cmd, tt.persistent)

			var flag *pflag.Flag
			if tt.persistent {
				flag = cmd.PersistentFlags().Lookup(tt.parameter.Name)
			} else {
				flag = cmd.Flags().Lookup(tt.parameter.Name)
			}

			if flag == nil {
				t.Errorf("expected flag %s to be added", tt.parameter.Name)
			}

			if flag.Value.String() != strconv.Itoa(*tt.parameter.Default.Int) {
				t.Errorf("expected flag value %d, got %s", *tt.parameter.Default.Int, flag.Value.String())
			}

			if tt.parameter.Shorthand != "" && flag.Shorthand != tt.parameter.Shorthand {
				t.Errorf("expected flag shorthand %s, got %s", tt.parameter.Shorthand, flag.Shorthand)
			}
		})
	}
}
func TestParameter_validateChoices(t *testing.T) {
	tests := []struct {
		name      string
		parameter Parameter
		expected  bool
	}{
		{
			name: "Valid string choice",
			parameter: Parameter{
				Type:     "string",
				Variable: CreateVariableString("choice1"),
				Choices:  []Variable{CreateVariableString("choice1"), CreateVariableString("choice2")},
			},
			expected: true,
		},
		{
			name: "Invalid string choice",
			parameter: Parameter{
				Type:     "string",
				Variable: CreateVariableString("invalid"),
				Choices:  []Variable{CreateVariableString("choice1"), CreateVariableString("choice2")},
			},
			expected: false,
		},
		{
			name: "Valid int choice",
			parameter: Parameter{
				Type:     "int",
				Variable: CreateVariableInt(1),
				Choices:  []Variable{CreateVariableInt(1), CreateVariableInt(2)},
			},
			expected: true,
		},
		{
			name: "Invalid int choice",
			parameter: Parameter{
				Type:     "int",
				Variable: CreateVariableInt(3),
				Choices:  []Variable{CreateVariableInt(1), CreateVariableInt(2)},
			},
			expected: false,
		},
		{
			name: "No choices provided",
			parameter: Parameter{
				Type:     "string",
				Variable: CreateVariableString("any"),
			},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.parameter.validateChoices()
			if result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func isEqualStrings(variable, expected Variable) bool {
	return variable.String == nil || *variable.String == *expected.String
}
func isEqualBools(variable, expected Variable) bool {
	return variable.Bool == nil || *variable.Bool == *expected.Bool
}
func isEqualInts(variable, expected Variable) bool {
	return variable.Int == nil || *variable.Int == *expected.Int
}
func isEqual(variable, expected Variable) bool {
	return isEqualStrings(variable, expected) && isEqualBools(variable, expected) && isEqualInts(variable, expected)
}

func TestParameter_SetLooseValue(t *testing.T) {
	tests := []struct {
		name        string
		parameter   Parameter
		key         string
		value       string
		expected    Variable
		expectError bool
	}{
		{
			name: "Set string value",
			parameter: Parameter{
				Type: "string",
			},
			key:         "test-string",
			value:       "test",
			expected:    CreateVariableString("test"),
			expectError: false,
		},
		{
			name: "Set bool value true",
			parameter: Parameter{
				Type: "bool",
			},
			key:         "test-bool",
			value:       "true",
			expected:    CreateVariableBool(true),
			expectError: false,
		},
		{
			name: "Set bool value false",
			parameter: Parameter{
				Type: "bool",
			},
			key:         "test-bool",
			value:       "false",
			expected:    CreateVariableBool(false),
			expectError: false,
		},
		{
			name: "Set invalid bool value",
			parameter: Parameter{
				Type: "bool",
			},
			key:         "test-bool",
			value:       "invalid",
			expected:    Variable{},
			expectError: false,
		},
		{
			name: "Set int value",
			parameter: Parameter{
				Type: "int",
			},
			key:         "test-int",
			value:       "123",
			expected:    CreateVariableInt(123),
			expectError: false,
		},
		{
			name: "Set invalid int value",
			parameter: Parameter{
				Type: "int",
			},
			key:         "test-int",
			value:       "invalid",
			expected:    Variable{},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.parameter.SetLooseValue(tt.value)
			if (err != nil) != tt.expectError {
				t.Errorf("expected error: %v, got: %v", tt.expectError, err)
			}
			if !tt.expectError && !isEqual(tt.parameter.Variable, tt.expected) {
				t.Errorf("expected %v, got %v", tt.expected, tt.parameter.Variable)
			}
		})
	}
}

func TestValidateParamFunc(t *testing.T) {
	tests := []struct {
		name        string
		parameter   Parameter
		input       string
		expectError bool
	}{
		{
			name: "Valid string input",
			parameter: Parameter{
				Name: "test-string",
				Type: "string",
				ValidateFunc: func(v Variable) bool {
					return true
				},
			},
			input:       "valid",
			expectError: false,
		},
		{
			name: "Invalid string input",
			parameter: Parameter{
				Name: "test-string",
				Type: "string",
				ValidateFunc: func(v Variable) bool {
					return *v.String == "valid"
				},
				Variable: CreateVariableString("valid"),
			},
			input:       "invalid",
			expectError: true,
		},
		{
			name: "Valid bool input",
			parameter: Parameter{
				Name: "test-bool",
				Type: "bool",
				ValidateFunc: func(v Variable) bool {
					return true
				},
			},
			input:       "true",
			expectError: false,
		},
		{
			name: "Invalid bool input",
			parameter: Parameter{
				Name: "test-bool",
				Type: "bool",
				ValidateFunc: func(v Variable) bool {
					return true
				},
			},
			input:       "invalid",
			expectError: true,
		},
		{
			name: "Valid int input",
			parameter: Parameter{
				Name: "test-int",
				Type: "int",
				ValidateFunc: func(v Variable) bool {
					return true
				},
			},
			input:       "123",
			expectError: false,
		},
		{
			name: "Invalid int input",
			parameter: Parameter{
				Name: "test-int",
				Type: "int",
				ValidateFunc: func(v Variable) bool {
					return true
				},
			},
			input:       "invalid",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			validateFunc := validateParamFunc(&tt.parameter)
			err := validateFunc(tt.input)
			if (err != nil) != tt.expectError {
				t.Errorf("expected error: %v, got: %v", tt.expectError, err)
			}
		})
	}
}
