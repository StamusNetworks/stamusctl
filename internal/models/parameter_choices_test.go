package models

import "testing"

func TestValidateChoices_CommaSeparatedStrings(t *testing.T) {
	param := &Parameter{
		Name:         "suricata.interfaces",
		Type:         "string",
		ValidateFunc: func(v Variable) bool { return true },
		Choices: []Variable{
			CreateVariableString("lo"),
			CreateVariableString("eth0"),
			CreateVariableString("eth1"),
		},
	}

	// Single valid
	param.Variable = CreateVariableString("lo")
	if !param.IsValid() {
		t.Fatalf("expected single valid choice to pass")
	}

	// Multiple valid (duplicates allowed)
	param.Variable = CreateVariableString("lo,lo")
	if !param.IsValid() {
		t.Fatalf("expected comma-separated duplicates to pass")
	}

	// Multiple valid mixed
	param.Variable = CreateVariableString("eth0, lo")
	if !param.IsValid() {
		t.Fatalf("expected comma-separated trimmed values to pass")
	}

	// Contains invalid
	param.Variable = CreateVariableString("eth0,eno1")
	if param.IsValid() {
		t.Fatalf("expected invalid value in list to fail")
	}
}
