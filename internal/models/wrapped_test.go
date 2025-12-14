package models

import (
	"testing"

	"github.com/spf13/pflag"
	"github.com/stretchr/testify/assert"
)

func TestFlags_ExtractFlags(t *testing.T) {
	tests := []struct {
		name     string
		flags    *Flags
		rootArgs []string
		leafArgs []string
		want     []string
	}{
		{
			name: "Extract root flags only",
			flags: &Flags{
				Root: []string{"root-flag1", "root-flag2"},
				Leaf: []string{},
			},
			rootArgs: []string{"root-flag1", "root-flag2"},
			leafArgs: []string{},
			want:     []string{"root-flag1", "root-flag2"},
		},
		{
			name: "Extract leaf flags only",
			flags: &Flags{
				Root: []string{},
				Leaf: []string{"leaf-flag1", "leaf-flag2"},
			},
			rootArgs: []string{},
			leafArgs: []string{"leaf-flag1", "leaf-flag2"},
			want:     []string{"leaf-flag1", "leaf-flag2"},
		},
		{
			name: "Extract both root and leaf flags",
			flags: &Flags{
				Root: []string{"root-flag"},
				Leaf: []string{"leaf-flag"},
			},
			rootArgs: []string{"root-flag"},
			leafArgs: []string{"leaf-flag"},
			want:     []string{"root-flag", "leaf-flag"},
		},
		{
			name: "Non-existent flags",
			flags: &Flags{
				Root: []string{"non-existent"},
				Leaf: []string{"also-non-existent"},
			},
			rootArgs: []string{},
			leafArgs: []string{},
			want:     []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rootFlags := pflag.NewFlagSet("root", pflag.ContinueOnError)
			leafFlags := pflag.NewFlagSet("leaf", pflag.ContinueOnError)

			// Add root flags
			for _, flag := range tt.rootArgs {
				rootFlags.String(flag, "", "test flag")
			}

			// Add leaf flags
			for _, flag := range tt.leafArgs {
				leafFlags.String(flag, "", "test flag")
			}

			result := tt.flags.ExtractFlags(rootFlags, leafFlags)

			// Verify all expected flags are present
			var foundFlags []string
			result.VisitAll(func(f *pflag.Flag) {
				foundFlags = append(foundFlags, f.Name)
			})

			assert.ElementsMatch(t, tt.want, foundFlags)
		})
	}
}

func TestCreateComposeFlags(t *testing.T) {
	rootFlags := []string{"root1", "root2"}
	leafFlags := []string{"leaf1", "leaf2"}

	result := CreateComposeFlags(rootFlags, leafFlags)

	assert.NotNil(t, result)
	assert.Equal(t, rootFlags, result.Root)
	assert.Equal(t, leafFlags, result.Leaf)
}

func TestComposeFlags_Contains(t *testing.T) {
	composeFlags := ComposeFlags{
		"command1": &Flags{Root: []string{"flag1"}, Leaf: []string{"flag2"}},
		"command2": &Flags{Root: []string{"flag3"}, Leaf: []string{"flag4"}},
	}

	tests := []struct {
		name     string
		command  string
		expected bool
	}{
		{
			name:     "Command exists",
			command:  "command1",
			expected: true,
		},
		{
			name:     "Command does not exist",
			command:  "command3",
			expected: false,
		},
		{
			name:     "Empty command",
			command:  "",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := composeFlags.Contains(tt.command)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestComposeFlags_Get(t *testing.T) {
	flags1 := &Flags{Root: []string{"flag1"}, Leaf: []string{"flag2"}}
	flags2 := &Flags{Root: []string{"flag3"}, Leaf: []string{"flag4"}}

	composeFlags := ComposeFlags{
		"command1": flags1,
		"command2": flags2,
	}

	tests := []struct {
		name     string
		command  string
		expected *Flags
	}{
		{
			name:     "Get existing command",
			command:  "command1",
			expected: flags1,
		},
		{
			name:     "Get another existing command",
			command:  "command2",
			expected: flags2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := composeFlags.Get(tt.command)
			assert.Len(t, result, 1)
			assert.Equal(t, tt.expected, result[0])
		})
	}
}

func TestContains(t *testing.T) {
	tests := []struct {
		name     string
		slice    []string
		element  string
		expected bool
	}{
		{
			name:     "Element exists in slice",
			slice:    []string{"a", "b", "c"},
			element:  "b",
			expected: true,
		},
		{
			name:     "Element does not exist in slice",
			slice:    []string{"a", "b", "c"},
			element:  "d",
			expected: false,
		},
		{
			name:     "Empty slice",
			slice:    []string{},
			element:  "a",
			expected: false,
		},
		{
			name:     "Empty element in slice",
			slice:    []string{"", "a", "b"},
			element:  "",
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := contains(tt.slice, tt.element)
			assert.Equal(t, tt.expected, result)
		})
	}
}
