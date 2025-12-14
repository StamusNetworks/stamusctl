package utils

import (
	"testing"

	"stamus-ctl/internal/app"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
)

func TestCopy(t *testing.T) {
	// Setup mock filesystem
	oldFS := app.FS
	app.FS = afero.NewMemMapFs()
	defer func() { app.FS = oldFS }()

	tests := []struct {
		name        string
		inputPath   string
		outputPath  string
		expectError bool
	}{
		{
			name:        "copy non-existing file",
			inputPath:   "/non-existing.txt",
			outputPath:  "/destination.txt",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Copy(tt.inputPath, tt.outputPath)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
