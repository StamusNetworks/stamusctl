package models

import (
	"reflect"
	"slices"
	"testing"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

func TestAddAsParameters(t *testing.T) {
	param1 := &Parameter{Type: "string", Variable: CreateVariableString("value1")}
	param2 := &Parameter{Type: "int", Variable: CreateVariableInt(2)}
	param3 := &Parameter{Type: "bool", Variable: CreateVariableBool(true)}

	params1 := &Parameters{
		"param1": param1,
	}

	params2 := &Parameters{
		"param2": param2,
		"param3": param3,
	}

	expected := &Parameters{
		"param1": param1,
		"param2": param2,
		"param3": param3,
	}

	params1.AddAsParameters(params2)
	if !reflect.DeepEqual(params1, expected) {
		t.Errorf("expected %v, got %v", expected, params1)
	}
}

func TestAddAsParameter(t *testing.T) {
	param1 := &Parameter{Type: "string", Variable: CreateVariableString("value1")}
	param2 := &Parameter{Type: "int", Variable: CreateVariableInt(2)}

	params := &Parameters{
		"param1": param1,
	}

	expected := &Parameters{
		"param1": param1,
		"param2": param2,
	}

	params.AddAsParameter("param2", param2)
	if !reflect.DeepEqual(params, expected) {
		t.Errorf("expected %v, got %v", expected, params)
	}
}

func TestAddAsFlags(t *testing.T) {
	param1 := &Parameter{Name: "param1", Type: "string", Variable: CreateVariableString("value1")}
	param2 := &Parameter{Name: "param2", Type: "int", Variable: CreateVariableInt(2)}

	params := &Parameters{
		"param1": param1,
		"param2": param2,
	}

	cmd := &cobra.Command{}
	params.AddAsFlags(cmd, false)

	setFlagsCount := 0
	cmd.Flags().VisitAll(func(f *pflag.Flag) {
		setFlagsCount++
	})

	if setFlagsCount != 2 {
		t.Errorf("expected 2 flags, got %d", len(cmd.Flags().Args()))
	}
}

func TestValidateAll(t *testing.T) {
	validParam := &Parameter{Type: "string", Variable: CreateVariableString("value1"), ValidateFunc: func(v Variable) bool { return true }}
	invalidParam := &Parameter{Type: "int", Variable: CreateVariableInt(2), ValidateFunc: func(v Variable) bool { return false }}

	tests := []struct {
		name    string
		params  *Parameters
		wantErr bool
	}{
		{
			name: "All valid parameters",
			params: &Parameters{
				"param1": validParam,
			},
			wantErr: false,
		},
		{
			name: "One invalid parameter",
			params: &Parameters{
				"param1": validParam,
				"param2": invalidParam,
			},
			wantErr: true,
		},
		{
			name: "All invalid parameters",
			params: &Parameters{
				"param1": invalidParam,
				"param2": invalidParam,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.params.ValidateAll()
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateAll() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestGetValues(t *testing.T) {
	param1 := &Parameter{Type: "string", Variable: CreateVariableString("value1")}
	param2 := &Parameter{Type: "int", Variable: CreateVariableInt(2)}
	param3 := &Parameter{Type: "bool", Variable: CreateVariableBool(true)}

	params := &Parameters{
		"param1": param1,
		"param2": param2,
		"param3": param3,
	}

	tests := []struct {
		name string
		keys []string
		want map[string]string
	}{
		{
			name: "No keys provided",
			keys: []string{},
			want: map[string]string{
				"param1": "value1",
				"param2": "2",
				"param3": "true",
			},
		},
		{
			name: "Single key provided",
			keys: []string{"param1"},
			want: map[string]string{
				"param1": "value1",
			},
		},
		{
			name: "Multiple keys provided",
			keys: []string{"param1", "param3"},
			want: map[string]string{
				"param1": "value1",
				"param3": "true",
			},
		},
		{
			name: "Key prefix provided",
			keys: []string{"param"},
			want: map[string]string{
				"param1": "value1",
				"param2": "2",
				"param3": "true",
			},
		},
		{
			name: "Non-matching key provided",
			keys: []string{"nonexistent"},
			want: map[string]string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := params.GetValues(tt.keys...)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetValues() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetVariablesValues(t *testing.T) {
	param1 := &Parameter{Type: "string", Variable: CreateVariableString("value1")}
	param2 := &Parameter{Type: "int", Variable: CreateVariableInt(2)}
	param3 := &Parameter{Type: "bool", Variable: CreateVariableBool(true)}

	params := &Parameters{
		"param1": param1,
		"param2": param2,
		"param3": param3,
	}

	tests := []struct {
		name string
		keys []string
		want map[string]*Variable
	}{
		{
			name: "No keys provided",
			keys: []string{},
			want: map[string]*Variable{
				"param1": &param1.Variable,
				"param2": &param2.Variable,
				"param3": &param3.Variable,
			},
		},
		{
			name: "Single key provided",
			keys: []string{"param1"},
			want: map[string]*Variable{
				"param1": &param1.Variable,
			},
		},
		{
			name: "Multiple keys provided",
			keys: []string{"param1", "param3"},
			want: map[string]*Variable{
				"param1": &param1.Variable,
				"param3": &param3.Variable,
			},
		},
		{
			name: "Key prefix provided",
			keys: []string{"param"},
			want: map[string]*Variable{
				"param1": &param1.Variable,
				"param2": &param2.Variable,
				"param3": &param3.Variable,
			},
		},
		{
			name: "Non-matching key provided",
			keys: []string{"nonexistent"},
			want: map[string]*Variable{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := params.GetVariablesValues(tt.keys...)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetVariablesValues() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetOrdered(t *testing.T) {
	param1 := &Parameter{Type: "string", Variable: CreateVariableString("value1")}
	param2 := &Parameter{Type: "int", Variable: CreateVariableInt(2)}
	param3 := &Parameter{Type: "bool", Variable: CreateVariableBool(true)}

	params := &Parameters{
		"paramB": param1,
		"paramA": param2,
		"paramC": param3,
	}

	expected := []string{"paramA", "paramB", "paramC"}

	got := params.GetOrdered()
	if !reflect.DeepEqual(got, expected) {
		t.Errorf("GetOrdered() = %v, want %v", got, expected)
	}
}

func TestSetToDefaults(t *testing.T) {
	defaultString := "default"
	defaultInt := 10
	defaultBool := true

	param1 := &Parameter{
		Type:    "string",
		Default: CreateVariableString(defaultString),
	}
	param2 := &Parameter{
		Type:    "int",
		Default: CreateVariableInt(defaultInt),
	}
	param3 := &Parameter{
		Type:    "bool",
		Default: CreateVariableBool(defaultBool),
	}

	params := &Parameters{
		"param1": param1,
		"param2": param2,
		"param3": param3,
	}

	err := params.SetToDefaults()
	if err != nil {
		t.Errorf("SetToDefaults() error = %v", err)
	}

	if *param1.Variable.String != defaultString {
		t.Errorf("expected param1 to be %v, got %v", defaultString, *param1.Variable.String)
	}
	if *param2.Variable.Int != defaultInt {
		t.Errorf("expected param2 to be %v, got %v", defaultInt, *param2.Variable.Int)
	}
	if *param3.Variable.Bool != defaultBool {
		t.Errorf("expected param3 to be %v, got %v", defaultBool, *param3.Variable.Bool)
	}
}

func TestProcessOptionnalParams(t *testing.T) {

	param1 := &Parameter{
		Name: "param1",
		Type: "optional",
	}
	param2 := &Parameter{
		Name: "param1.param2",
		Type: "optional",
	}
	param3 := &Parameter{
		Name: "param1.param3",
		Type: "string",
	}
	param4 := &Parameter{
		Name: "param1.param2.param4",
		Type: "optional",
	}

	tests := []struct {
		name   string
		params Parameters
		want   []string
	}{
		{
			name: "No optional parameters",
			params: Parameters{
				param1.Name: param1.Copy().SetVariable(CreateVariableBool(false)),
			},
			want: []string{"param1"},
		},
		{
			name: "Optional parameters",
			params: Parameters{
				param1.Name: param1.Copy().SetVariable(CreateVariableBool(false)),
				param2.Name: param2.Copy().SetVariable(CreateVariableBool(false)),
				param3.Name: param3.Copy().SetVariable(CreateVariableString("value")),
				param4.Name: param4.Copy().SetVariable(CreateVariableBool(false)),
			},
			want: []string{"param1"},
		},
		{
			name: "Deep optional parameter",
			params: Parameters{
				param1.Name: param1.Copy().SetVariable(CreateVariableBool(true)),
				param2.Name: param2.Copy().SetVariable(CreateVariableBool(false)),
				param3.Name: param3.Copy().SetVariable(CreateVariableString("value")),
				param4.Name: param4.Copy().SetVariable(CreateVariableBool(false)),
			},
			want: []string{"param1.param2", "param1.param3"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.params.ProcessOptionnalParams(false)
			if err != nil {
				t.Errorf("ProcessOptionnalParams() error = %v", err)
			}
			keys := make([]string, 0, len(tt.params))
			for k := range tt.params {
				keys = append(keys, k)
			}

			if len(keys) != len(tt.want) {
				t.Errorf("expected %v, got %v", tt.want, keys)
			}
			for _, k := range keys {
				if !slices.Contains(tt.want, k) {
					t.Errorf("expected %v, got %v", tt.want, keys)
				}
			}
		})
	}
}
