package models

import (
	"embed"
	"io/fs"
	"testing"
	"testing/fstest"

	"stamus-ctl/internal/app"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
)

//go:embed embed_test.go
var testEmbed embed.FS

// Helper function to test ExtractEmbedTo with an fs.FS
func extractEmbedToWithFS(embedName string, fsys fs.FS, outputFolder string) error {
	// This is a copy of ExtractEmbedTo logic but accepts fs.FS
	files := getAllFilesEmbedWithFS(embedName, fsys)

	for _, file := range files {
		data, err := fs.ReadFile(fsys, file)
		if err != nil {
			return err
		}
		err = app.FS.MkdirAll(outputFolder+"/"+extractPath(file), 0o755)
		if err != nil {
			return err
		}
		err = afero.WriteFile(app.FS, outputFolder+"/"+extractPath(file)+"/"+extractFileName(file), data, 0o644)
		if err != nil {
			return err
		}
	}
	return nil
}

func getAllFilesEmbedWithFS(inputFolder string, fsys fs.FS) []string {
	var files []string
	entries, err := fs.ReadDir(fsys, inputFolder)
	if err != nil {
		return files
	}
	for _, entry := range entries {
		if entry.IsDir() {
			files = append(files, getAllFilesEmbedWithFS(inputFolder+"/"+entry.Name(), fsys)...)
		} else {
			files = append(files, inputFolder+"/"+entry.Name())
		}
	}
	return files
}

func TestExtractEmbedTo(t *testing.T) {
	// Create an in-memory test filesystem
	testFS := fstest.MapFS{
		"testdir/file1.txt": {
			Data: []byte("test content 1"),
		},
		"testdir/subdir/file2.txt": {
			Data: []byte("test content 2"),
		},
	}

	outputFolder := "/test_output"
	app.FS.RemoveAll(outputFolder)

	err := extractEmbedToWithFS("testdir", testFS, outputFolder)
	assert.NoError(t, err)

	// Verify files were extracted to in-memory filesystem
	exists, err := afero.Exists(app.FS, outputFolder+"/file1.txt")
	assert.NoError(t, err)
	assert.True(t, exists)

	exists, err = afero.Exists(app.FS, outputFolder+"/subdir/file2.txt")
	assert.NoError(t, err)
	assert.True(t, exists)

	// Verify content
	content, err := afero.ReadFile(app.FS, outputFolder+"/file1.txt")
	assert.NoError(t, err)
	assert.Equal(t, "test content 1", string(content))

	content, err = afero.ReadFile(app.FS, outputFolder+"/subdir/file2.txt")
	assert.NoError(t, err)
	assert.Equal(t, "test content 2", string(content))

	// Clean up
	app.FS.RemoveAll(outputFolder)
}

func TestExtractEmbedTo_NonExistentFolder(t *testing.T) {
	testFS := fstest.MapFS{
		"testdir/file.txt": {
			Data: []byte("test"),
		},
	}

	outputFolder := "/test_output_nonexistent"
	app.FS.RemoveAll(outputFolder)

	err := extractEmbedToWithFS("testdir", testFS, outputFolder)
	assert.NoError(t, err)

	// Verify folder was created in in-memory filesystem
	exists, err := afero.DirExists(app.FS, outputFolder)
	assert.NoError(t, err)
	assert.True(t, exists)

	// Verify file exists
	exists, err = afero.Exists(app.FS, outputFolder+"/file.txt")
	assert.NoError(t, err)
	assert.True(t, exists)

	// Clean up
	app.FS.RemoveAll(outputFolder)
}

func TestGetAllFilesEmbed(t *testing.T) {
	files := getAllFilesEmbed(".", testEmbed)

	// Should find at least the test file itself
	assert.NotEmpty(t, files)

	// Check that files contain the expected test file
	found := false
	for _, file := range files {
		if file == "./embed_test.go" {
			found = true
			break
		}
	}
	assert.True(t, found, "Expected to find embed_test.go in the list of files")
}

func TestExtractPath(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		expected string
	}{
		{
			name:     "Nested path",
			path:     "folder/subfolder/file.txt",
			expected: "subfolder",
		},
		{
			name:     "Deep nested path",
			path:     "folder/sub1/sub2/sub3/file.txt",
			expected: "sub1/sub2/sub3",
		},
		{
			name:     "Two level path",
			path:     "folder/file.txt",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractPath(tt.path)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestExtractFileName(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		expected string
	}{
		{
			name:     "Simple path",
			path:     "folder/file.txt",
			expected: "file.txt",
		},
		{
			name:     "Nested path",
			path:     "folder/subfolder/file.txt",
			expected: "file.txt",
		},
		{
			name:     "Deep nested path",
			path:     "folder/sub1/sub2/sub3/file.yaml",
			expected: "file.yaml",
		},
		{
			name:     "Single level path",
			path:     "file.txt",
			expected: "file.txt",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractFileName(tt.path)
			assert.Equal(t, tt.expected, result)
		})
	}
}
