package utils

import (
	"os"
	"testing"

	"stamus-ctl/internal/app"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

func TestCopy_Success(t *testing.T) {
	// Use real OsFs so both app.FS.Stat and cp.Copy use the same filesystem.
	oldFS := app.FS
	app.FS = afero.NewOsFs()
	defer func() { app.FS = oldFS }()

	// Create a temp source file.
	srcFile, err := os.CreateTemp("", "copy-src-*.txt")
	if err != nil {
		t.Fatalf("failed to create temp source file: %v", err)
	}
	srcPath := srcFile.Name()
	srcFile.WriteString("hello copy")
	srcFile.Close()
	defer os.Remove(srcPath)

	dstPath := srcPath + ".dst"
	defer os.Remove(dstPath)

	err = Copy(srcPath, dstPath)
	assert.NoError(t, err)
}

// TestCopy_CopyFails exercises the cp.Copy error path (line 22-24).
// The file is created in afero MemMapFs so app.FS.Stat succeeds,
// but cp.Copy uses the real OS and the file doesn't exist there — so it fails.
func TestCopy_CopyFails(t *testing.T) {
	oldFS := app.FS
	app.FS = afero.NewMemMapFs()
	defer func() { app.FS = oldFS }()

	// Create a file in the in-memory FS (so Stat succeeds).
	require.NoError(t, afero.WriteFile(app.FS, "/ghost-file.txt", []byte("data"), 0o644))

	// cp.Copy will fail because /ghost-file.txt doesn't exist on the real OS.
	err := Copy("/ghost-file.txt", "/ghost-dst.txt")
	assert.Error(t, err)
}
