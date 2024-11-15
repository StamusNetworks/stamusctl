package models

import (
	"fmt"
	"slices"
	"stamus-ctl/internal/logging"
	"strconv"
	"strings"

	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"
)

// Parameter is a struct that stores the information of a parameter
type Parameter struct {
	Name         string
	Shorthand    string
	Usage        string
	Type         string
	Variable     Variable
	Default      Variable
	Choices      []Variable
	ValidateFunc func(Variable) bool
	Hidden       bool
}

// Variable struct is used to store values of different types
type Variable struct {
	String *string
	Bool   *bool
	Int    *int
}

// Return the value of a string to any type
func asLooseTyped(value string) any {
	if value == "true" || value == "false" {
		return value == "true"
	}
	asInt, err := strconv.Atoi(value)
	if err != nil {
		return value
	}
	return asInt
}

func (v *Variable) IsNil() bool {
	return v.String == nil && v.Bool == nil && v.Int == nil
}

// Return the value of the variable as a string
func (v *Variable) AsString() string {
	if v.String != nil {
		return *v.String
	}
	if v.Bool != nil {
		return strconv.FormatBool(*v.Bool)
	}
	if v.Int != nil {
		return strconv.Itoa(*v.Int)
	}
	return ""
}

// Return the value of the variable as a any type
func (p *Parameter) GetValue() (any, error) {
	if p.Variable.IsNil() && p.Default.IsNil() {
		return nil, fmt.Errorf("Variable has not been set")
	}
	if !p.Variable.IsNil() {
		switch p.Type {
		case "string":
			return *p.Variable.String, nil
		case "bool", "optional":
			return *p.Variable.Bool, nil
		case "int":
			return *p.Variable.Int, nil
		}
	}
	if !p.Default.IsNil() {
		switch p.Type {
		case "string":
			return *p.Default.String, nil
		case "bool", "optional":
			return *p.Default.Bool, nil
		case "int":
			return *p.Default.Int, nil
		}
	}
	return nil, fmt.Errorf("Invalid type")
}

// Adds the parameter as a flag to the command
func (p *Parameter) AddAsFlag(cmd *cobra.Command, persistent bool) {
	// Set flag validation function
	if p.ValidateFunc == nil {
		p.ValidateFunc = func(v Variable) bool {
			return true
		}
	}
	// Add the flag based on the type
	switch p.Type {
	case "string":
		p.AddStringFlag(cmd, persistent)
	case "bool":
		p.AddBoolFlag(cmd, persistent)
	case "int":
		p.AddIntFlag(cmd, persistent)
	}
	// Hide the flag if needed
	if p.Hidden {
		cmd.Flags().MarkHidden(p.Name)
	}
}

// Adds the parameter as a string flag to the command
func (p *Parameter) AddStringFlag(cmd *cobra.Command, persistent bool) {
	if p.Default.String == nil {
		p.Default = CreateVariableString("")
	}
	p.Variable = p.Default

	if p.Shorthand == "" {
		if persistent {
			cmd.PersistentFlags().StringVar(p.Variable.String, p.Name, *p.Default.String, p.Usage)
		} else {
			cmd.Flags().StringVar(p.Variable.String, p.Name, *p.Default.String, p.Usage)
		}
	} else {
		if persistent {
			cmd.PersistentFlags().StringVarP(p.Variable.String, p.Name, p.Shorthand, *p.Default.String, p.Usage)
		} else {
			cmd.Flags().StringVarP(p.Variable.String, p.Name, p.Shorthand, *p.Default.String, p.Usage)
		}
	}
}

// Adds the parameter as a bool flag to the command
func (p *Parameter) AddBoolFlag(cmd *cobra.Command, persistent bool) {
	if p.Default.Bool == nil {
		p.Default = CreateVariableBool(false)
	}
	p.Variable = p.Default
	if p.Shorthand == "" {
		if persistent {
			cmd.PersistentFlags().BoolVar(p.Variable.Bool, p.Name, *p.Default.Bool, p.Usage)
		} else {
			cmd.Flags().BoolVar(p.Variable.Bool, p.Name, *p.Default.Bool, p.Usage)
		}
	} else {
		if persistent {
			cmd.PersistentFlags().BoolVarP(p.Variable.Bool, p.Name, p.Shorthand, *p.Default.Bool, p.Usage)
		} else {
			cmd.Flags().BoolVarP(p.Variable.Bool, p.Name, p.Shorthand, *p.Default.Bool, p.Usage)
		}
	}
}

// Adds the parameter as an int flag to the command
func (p *Parameter) AddIntFlag(cmd *cobra.Command, persistent bool) {
	if p.Default.Int == nil {
		p.Default = CreateVariableInt(0)
	}
	p.Variable = p.Default
	if p.Shorthand == "" {
		if persistent {
			cmd.PersistentFlags().IntVar(p.Variable.Int, p.Name, *p.Default.Int, p.Usage)
		} else {
			cmd.Flags().IntVar(p.Variable.Int, p.Name, *p.Default.Int, p.Usage)
		}
	} else {
		if persistent {
			cmd.PersistentFlags().IntVarP(p.Variable.Int, p.Name, p.Shorthand, *p.Default.Int, p.Usage)
		} else {
			cmd.Flags().IntVarP(p.Variable.Int, p.Name, p.Shorthand, *p.Default.Int, p.Usage)
		}
	}
}

// Validates the variable with the given function
// If choices are provided, the variable must be in the list of choices
func (p *Parameter) IsValid() bool {
	return !p.Variable.IsNil() && p.ValidateFunc(p.Variable) && p.validateChoices()
}

// Validates the choices of the parameter
func (p *Parameter) validateChoices() bool {
	if len(p.Choices) > 0 {
		switch p.Type {
		case "string":
			// Convert choices to strings
			asStrings := []string{}
			for _, choice := range p.Choices {
				asStrings = append(asStrings, *choice.String)
			}
			// Add as list
			def := strings.Join(asStrings, ",")
			asStrings = append(asStrings, def)
			// Check
			isOk := slices.Contains(asStrings, *p.Variable.String)
			if !isOk {
				logging.Sugar.Info("Error: Must be one of:", asStrings)
			}
			return isOk
		case "int":
			asInts := []int{}
			for _, choice := range p.Choices {
				asInts = append(asInts, *choice.Int)
			}
			isOk := slices.Contains(asInts, *p.Variable.Int)
			if !isOk {
				logging.Sugar.Info("Error: Must be one of:", asInts)
			}
			return isOk
		}
	}
	return true
}

// Set the variable to the default value
func (p *Parameter) SetToDefault() {
	if p.Variable.IsNil() {
		p.Variable = p.Default
	}
}

// Set the variable to the value provided
func (p *Parameter) SetLooseValue(key string, value string) error {
	switch p.Type {
	case "string":
		p.Variable = CreateVariableString(value)
	case "bool", "optional":
		if value == "true" || value == "false" {
			p.Variable = CreateVariableBool(value == "true")
		} else {
			logging.Sugar.Info("Invalid value for", key)
		}
	case "int":
		// Convert string to int
		asInt, err := strconv.Atoi(value)
		if err != nil {
			logging.Sugar.Info("Error converting string to int:", err)
			return err
		}
		p.Variable = CreateVariableInt(asInt)
	}
	return nil
}

// Ask the user for the value of the parameter
func (p *Parameter) AskUser() error {
	switch p.Type {
	case "string":
		// If choices are provided, use select prompt
		if p.Choices != nil && len(p.Choices) > 0 {
			choices := []string{}
			for _, choice := range p.Choices {
				choices = append(choices, *choice.String)
			}
			result, err := selectPrompt(p, choices)
			if err != nil {
				return err
			}
			p.Variable = CreateVariableString(result)
			return nil
		}
		// Otherwise use text prompt
		var defaultValue string
		if p.Default.String != nil {
			defaultValue = *p.Default.String
		}
		result, err := textPrompt(p, defaultValue)
		if err != nil {
			return err
		}
		p.Variable = CreateVariableString(result)
	case "bool", "optional":
		result, err := selectPrompt(p, []string{"true", "false"})
		if err != nil {
			return err
		}
		p.Variable = CreateVariableBool(result == "true")
	case "int":
		var defaultValue string
		if p.Default.Int != nil {
			defaultValue = strconv.Itoa(*p.Default.Int)
		}
		result, err := textPrompt(p, defaultValue)
		if err != nil {
			return err
		}
		asInt, _ := strconv.Atoi(result)
		p.Variable = CreateVariableInt(asInt)
	}
	return nil
}

func CreateVariableString(value string) Variable {
	return Variable{String: &value}
}
func CreateVariableBool(value bool) Variable {
	return Variable{Bool: &value}
}
func CreateVariableInt(value int) Variable {
	return Variable{Int: &value}
}

// Prompt the user for a string value
func textPrompt(param *Parameter, defaultValue string) (string, error) {
	prompt := promptui.Prompt{
		Label:    param.Usage,
		Default:  defaultValue,
		Validate: validateParamFunc(param),
	}
	result, err := prompt.Run()
	if err != nil {
		return "", fmt.Errorf("Prompt cancelled")
	}
	return result, nil
}

func validateParamFunc(param *Parameter) func(input string) error {
	return func(input string) error {
		current := param
		err := current.SetLooseValue(param.Name, input)
		if err != nil {
			return err
		}
		if !current.IsValid() {
			return fmt.Errorf("Invalid value")
		}
		return nil
	}
}

// Prompt the user for a selection
func selectPrompt(p *Parameter, choices []string) (string, error) {
	prompt := promptui.Select{
		Label: p.Usage,
		Items: choices,
		Templates: &promptui.SelectTemplates{
			Selected: fmt.Sprintf("%s {{ . | green }} %s ", promptui.IconGood, p.Usage),
		},
	}
	_, result, err := prompt.Run()

	if err != nil {
		return "", fmt.Errorf("Prompt cancelled")
	}
	return result, nil
}
