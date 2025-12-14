package utils

import (
	"testing"

	"stamus-ctl/internal/app"
	"stamus-ctl/internal/models"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
)

func TestFolderExists(t *testing.T) {
	// Setup mock filesystem
	oldFS := app.FS
	app.FS = afero.NewMemMapFs()
	defer func() { app.FS = oldFS }()

	// Create test folder
	err := app.FS.MkdirAll("/test/folder", 0755)
	assert.NoError(t, err)

	tests := []struct {
		name        string
		path        string
		expectExist bool
		expectError bool
	}{
		{
			name:        "existing folder",
			path:        "/test/folder",
			expectExist: true,
			expectError: false,
		},
		{
			name:        "non-existing folder",
			path:        "/non/existing",
			expectExist: false,
			expectError: false,
		},
		{
			name:        "root folder",
			path:        "/",
			expectExist: true,
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			exists, err := FolderExists(tt.path)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, tt.expectExist, exists)
		})
	}
}

func TestValidatePath(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		expected bool
	}{
		{
			name:     "valid simple path",
			path:     "validpath",
			expected: true,
		},
		{
			name:     "empty path",
			path:     "",
			expected: false,
		},
		{
			name:     "path with backslash",
			path:     "path\\invalid",
			expected: false,
		},
		{
			name:     "path with forward slash",
			path:     "path/invalid",
			expected: false,
		},
		{
			name:     "path with colon",
			path:     "path:invalid",
			expected: false,
		},
		{
			name:     "path with asterisk",
			path:     "path*invalid",
			expected: false,
		},
		{
			name:     "path with question mark",
			path:     "path?invalid",
			expected: false,
		},
		{
			name:     "path with double quote",
			path:     "path\"invalid",
			expected: false,
		},
		{
			name:     "path with less than",
			path:     "path<invalid",
			expected: false,
		},
		{
			name:     "path with greater than",
			path:     "path>invalid",
			expected: false,
		},
		{
			name:     "path with pipe",
			path:     "path|invalid",
			expected: false,
		},
		{
			name:     "path with dollar sign",
			path:     "path$invalid",
			expected: false,
		},
		{
			name:     "path with double dots",
			path:     "path..invalid",
			expected: false,
		},
		{
			name:     "path with space",
			path:     "path invalid",
			expected: false,
		},
		{
			name:     "path with tab",
			path:     "path\tinvalid",
			expected: false,
		},
		{
			name:     "path with newline",
			path:     "path\ninvalid",
			expected: false,
		},
		{
			name:     "path with carriage return",
			path:     "path\rinvalid",
			expected: false,
		},
		{
			name:     "valid path with hyphens and underscores",
			path:     "valid-path_name",
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pathVar := models.Variable{String: &tt.path}
			result := ValidatePath(pathVar)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestListFilesInFolder(t *testing.T) {
	// Note: ListFilesInFolder uses filepath.Walk which doesn't work with afero mock filesystem
	// Testing with non-existing folder to verify error handling
	tests := []struct {
		name        string
		folderPath  string
		expectError bool
	}{
		{
			name:        "list files in non-existing folder",
			folderPath:  "/non/existing/folder/that/does/not/exist",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ListFilesInFolder(tt.folderPath)
			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, result)
			}
		})
	}
}
