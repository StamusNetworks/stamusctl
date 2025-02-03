package models

import (
	"fmt"
	"path/filepath"
	"stamus-ctl/internal/app"
	"testing"

	"github.com/spf13/afero"
)

func TestInstanciateViperFunc(t *testing.T) {
	// Create a file
	file, err := app.FS.Create("/tmp/config.yaml")
	if err != nil {
		t.Fatalf("failed to create file: %v", err)
	}
	_, err = file.Write([]byte("key: value"))
	if err != nil {
		t.Fatalf("failed to write to file: %v", err)
	}
	file.Close()

	// Define test cases
	tests := []struct {
		Path          string
		Name          string
		Type          string
		Expected      map[string]string
		ExpectedError bool
	}{
		{"/tmp", "config", "yaml", map[string]string{"key": "value"}, false},
		{"/path", "config", "yaml", nil, true},
	}

	for _, test := range tests {
		t.Run(fmt.Sprintf("Path=%s, Name=%s, Type=%s", test.Path, test.Name, test.Type), func(t *testing.T) {
			// Call the function
			viperInstance, err := InstanciateViper(NewFile(
				test.Path,
				test.Name,
				test.Type,
			))

			// Check for unexpected error
			if err != nil && !test.ExpectedError {
				t.Fatalf("unexpected error: %v", err)
			} else if err == nil && test.ExpectedError {
				// Check if viper instance is not nil
				if viperInstance == nil {
					t.Fatalf("expected viper instance, got nil")
				}

				// Check if the config file was created
				exists, err := afero.Exists(app.FS, filepath.Join(test.Path, fmt.Sprintf("%s.%s", test.Name, test.Type)))
				if err != nil {
					t.Fatalf("failed to check if file exists: %v", err)
				}
				if !exists {
					t.Fatalf("expected config file to be created")
				}

				// Check if the config file has the expected content
				for key, value := range test.Expected {
					if viperInstance.GetString(key) != value {
						t.Fatalf("expected %s = %s, got %s", key, value, viperInstance.GetString(key))
					}
				}
			}
		})
	}
}
