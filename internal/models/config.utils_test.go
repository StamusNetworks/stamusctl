package models

import (
	"testing"

	"stamus-ctl/internal/app"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestNestMap(t *testing.T) {
	result := nestMap(map[string]interface{}{
		"Release.truc":   "test",
		"Release.machin": "test",
	})
	assert.Equal(t, map[string]interface{}{
		"Release": map[string]interface{}{
			"truc":   "test",
			"machin": "test",
		},
	}, result)
}

func TestNestMapWithMoreNesting(t *testing.T) {
	result := nestMap(map[string]interface{}{
		"Release.deep.truc":   "test",
		"Release.deep.machin": "test",
	})
	assert.Equal(t, map[string]interface{}{
		"Release": map[string]interface{}{
			"deep": map[string]interface{}{
				"truc":   "test",
				"machin": "test",
			},
		},
	}, result)
}

func TestRemoveEmptyStrings(t *testing.T) {
	tests := []struct {
		input    []string
		expected []string
	}{
		{input: []string{"a", "", "b", "", "c"}, expected: []string{"a", "b", "c"}},
		{input: []string{"", "", ""}, expected: []string{}},
		{input: []string{"a", "b", "c"}, expected: []string{"a", "b", "c"}},
		{input: []string{}, expected: []string{}},
	}

	for _, test := range tests {
		result := removeEmptyStrings(test.input)
		if len(result) != len(test.expected) {
			t.Errorf("expected %v, got %v", test.expected, result)
		}
		for i, v := range result {
			if v != test.expected[i] {
				t.Errorf("expected %v, got %v", test.expected, result)
			}
		}
	}
}

func TestIsDirEmpty(t *testing.T) {
	// Test when directory is empty
	err := app.FS.Mkdir("/TestIsDirEmpty/emptydir", 0o755)
	assert.NoError(t, err)

	isEmpty, err := isDirEmpty("/TestIsDirEmpty/emptydir")
	assert.NoError(t, err)
	assert.True(t, isEmpty)

	// Test when directory is not empty
	err = afero.WriteFile(app.FS, "/TestIsDirEmpty/emptydir/file.txt", []byte("content"), 0o644)
	assert.NoError(t, err)

	isEmpty, err = isDirEmpty("/TestIsDirEmpty/emptydir")
	assert.NoError(t, err)
	assert.False(t, isEmpty)

	// Test when directory does not exist
	isEmpty, err = isDirEmpty("/TestIsDirEmpty/nonexistentdir")
	assert.Error(t, err)
	assert.False(t, isEmpty)
}

func TestRemoveDirIfEmpty(t *testing.T) {
	// Test when directory is empty
	err := app.FS.Mkdir("/TestRemoveDirIfEmpty/emptydir", 0o755)
	assert.NoError(t, err)

	err = removeDirIfEmpty("/TestRemoveDirIfEmpty/emptydir")
	assert.NoError(t, err)

	exists, err := afero.DirExists(app.FS, "/TestRemoveDirIfEmpty/emptydir")
	assert.NoError(t, err)
	assert.False(t, exists)

	// Test when directory is not empty
	err = app.FS.Mkdir("/TestRemoveDirIfEmpty/notemptydir", 0o755)
	assert.NoError(t, err)

	err = afero.WriteFile(app.FS, "/TestRemoveDirIfEmpty/notemptydir/file.txt", []byte("content"), 0o644)
	assert.NoError(t, err)

	err = removeDirIfEmpty("/TestRemoveDirIfEmpty/notemptydir")
	assert.NoError(t, err)

	exists, err = afero.DirExists(app.FS, "/TestRemoveDirIfEmpty/notemptydir")
	assert.NoError(t, err)
	assert.True(t, exists)

	// Test when directory does not exist
	err = removeDirIfEmpty("/TestRemoveDirIfEmpty/nonexistentdir")
	assert.Error(t, err)
}

func TestDeleteEmptyFolders(t *testing.T) {
	// Test when directory is empty
	err := app.FS.Mkdir("/TestDeleteEmptyFolders/emptydir", 0o755)
	assert.NoError(t, err)

	err = deleteEmptyFolders("/TestDeleteEmptyFolders/emptydir")
	assert.NoError(t, err)

	exists, err := afero.DirExists(app.FS, "/TestDeleteEmptyFolders/emptydir")
	assert.NoError(t, err)
	assert.True(t, exists)

	// Test when directory is not empty
	err = app.FS.Mkdir("/TestDeleteEmptyFolders/notemptydir", 0o755)
	assert.NoError(t, err)

	err = afero.WriteFile(app.FS, "/TestDeleteEmptyFolders/notemptydir/file.txt", []byte("content"), 0o644)
	assert.NoError(t, err)

	err = deleteEmptyFolders("/TestDeleteEmptyFolders/notemptydir")
	assert.NoError(t, err)

	exists, err = afero.DirExists(app.FS, "/TestDeleteEmptyFolders/notemptydir")
	assert.NoError(t, err)
	assert.True(t, exists)

	// Test nested empty directories
	err = app.FS.MkdirAll("/TestDeleteEmptyFolders/nested/emptydir/dir", 0o755)
	assert.NoError(t, err)

	err = deleteEmptyFolders("/TestDeleteEmptyFolders/nested")
	assert.NoError(t, err)

	exists, err = afero.DirExists(app.FS, "/TestDeleteEmptyFolders/nested/emptydir/dir")
	assert.NoError(t, err)
	assert.False(t, exists)

	exists, err = afero.DirExists(app.FS, "/TestDeleteEmptyFolders/nested/emptydir")
	assert.NoError(t, err)
	assert.True(t, exists)

	exists, err = afero.DirExists(app.FS, "/TestDeleteEmptyFolders/nested")
	assert.NoError(t, err)
	assert.True(t, exists)

	// Test nested directories with files
	err = app.FS.MkdirAll("/TestDeleteEmptyFolders/nested/notemptydir", 0o755)
	assert.NoError(t, err)

	err = afero.WriteFile(app.FS, "/TestDeleteEmptyFolders/nested/notemptydir/file.txt", []byte("content"), 0o644)
	assert.NoError(t, err)

	err = deleteEmptyFolders("/TestDeleteEmptyFolders/nested")
	assert.NoError(t, err)

	exists, err = afero.DirExists(app.FS, "/TestDeleteEmptyFolders/nested/notemptydir")
	assert.NoError(t, err)
	assert.True(t, exists)

	exists, err = afero.DirExists(app.FS, "/TestDeleteEmptyFolders/nested")
	assert.NoError(t, err)
	assert.True(t, exists)
}

func TestDeleteEmptyFiles(t *testing.T) {
	// Test when directory has no empty files
	err := app.FS.Mkdir("/TestDeleteEmptyFiles/dir", 0o755)
	assert.NoError(t, err)

	err = afero.WriteFile(app.FS, "/TestDeleteEmptyFiles/dir/file.txt", []byte("content"), 0o644)
	assert.NoError(t, err)

	err = deleteEmptyFiles("/TestDeleteEmptyFiles/dir")
	assert.NoError(t, err)

	exists, err := afero.Exists(app.FS, "/TestDeleteEmptyFiles/dir/file.txt")
	assert.NoError(t, err)
	assert.True(t, exists)

	// Test when directory has empty files
	err = afero.WriteFile(app.FS, "/TestDeleteEmptyFiles/dir/emptyfile.txt", []byte(""), 0o644)
	assert.NoError(t, err)

	err = deleteEmptyFiles("/TestDeleteEmptyFiles/dir")
	assert.NoError(t, err)

	exists, err = afero.Exists(app.FS, "/TestDeleteEmptyFiles/dir/emptyfile.txt")
	assert.NoError(t, err)
	assert.False(t, exists)

	// Test when directory is empty
	err = app.FS.Mkdir("/TestDeleteEmptyFiles/emptydir", 0o755)
	assert.NoError(t, err)

	err = deleteEmptyFiles("/TestDeleteEmptyFiles/emptydir")
	assert.NoError(t, err)

	exists, err = afero.DirExists(app.FS, "/TestDeleteEmptyFiles/emptydir")
	assert.NoError(t, err)
	assert.True(t, exists)

	// Test when directory does not exist
	err = deleteEmptyFiles("/TestDeleteEmptyFiles/nonexistentdir")
	assert.Error(t, err)
}

func TestGetAllFiles(t *testing.T) {
	// Setup test directories and files
	err := app.FS.Mkdir("/TestGetAllFiles/testdir", 0o755)
	assert.NoError(t, err)

	err = afero.WriteFile(app.FS, "/TestGetAllFiles/testdir/file1.txt", []byte("content"), 0o644)
	assert.NoError(t, err)

	err = afero.WriteFile(app.FS, "/TestGetAllFiles/testdir/file2.tpl", []byte("content"), 0o644)
	assert.NoError(t, err)

	err = afero.WriteFile(app.FS, "/TestGetAllFiles/testdir/file3.tpl", []byte("content"), 0o644)
	assert.NoError(t, err)

	err = app.FS.Mkdir("/TestGetAllFiles/testdir/subdir", 0o755)
	assert.NoError(t, err)

	err = afero.WriteFile(app.FS, "/TestGetAllFiles/testdir/subdir/file4.tpl", []byte("content"), 0o644)
	assert.NoError(t, err)

	// Test getting all .tpl files
	files, err := getAllFiles("/TestGetAllFiles/testdir", ".tpl")
	assert.NoError(t, err)
	assert.ElementsMatch(t, []string{
		"/TestGetAllFiles/testdir/file2.tpl",
		"/TestGetAllFiles/testdir/file3.tpl",
		"/TestGetAllFiles/testdir/subdir/file4.tpl",
	}, files)

	// Test getting all .txt files
	files, err = getAllFiles("/TestGetAllFiles/testdir", ".txt")
	assert.NoError(t, err)
	assert.ElementsMatch(t, []string{
		"/TestGetAllFiles/testdir/file1.txt",
	}, files)

	// Test getting files with non-existent extension
	files, err = getAllFiles("/TestGetAllFiles/testdir", ".nonexistent")
	assert.NoError(t, err)
	assert.Empty(t, files)

	// Test getting files from non-existent directory
	files, err = getAllFiles("/TestGetAllFiles/nonexistentdir", ".tpl")
	assert.Error(t, err)
	assert.Empty(t, files)
}

func TestProcessTemplate(t *testing.T) {
	logger := zap.NewExample().Sugar()

	// Setup test directories and files
	err := app.FS.Mkdir("/TestProcessTemplate/input", 0o755)
	assert.NoError(t, err)

	err = afero.WriteFile(app.FS, "/TestProcessTemplate/input/template.txt", []byte("Hello, {{ .Name }}!"), 0o644)
	assert.NoError(t, err)

	err = app.FS.Mkdir("/TestProcessTemplate/output", 0o755)
	assert.NoError(t, err)

	data := map[string]interface{}{
		"Name": "World",
	}

	tpls, err := getAllFiles("/TestProcessTemplate/input", ".tpl")
	assert.NoError(t, err)

	info, err := app.FS.Stat("/TestProcessTemplate/input/template.txt")
	assert.NoError(t, err)

	err = processTemplate(data, tpls, "/TestProcessTemplate/input/template.txt",
		"/TestProcessTemplate/input", "/TestProcessTemplate/output", info, logger)
	assert.NoError(t, err)

	// Verify the output file
	exists, err := afero.Exists(app.FS, "/TestProcessTemplate/output/template.txt")
	assert.NoError(t, err)
	assert.True(t, exists)

	content, err := afero.ReadFile(app.FS, "/TestProcessTemplate/output/template.txt")
	assert.NoError(t, err)
	assert.Equal(t, "Hello, World!\n", string(content))

	// Test processing a directory
	err = app.FS.Mkdir("/TestProcessTemplate/input2/dir", 0o755)
	assert.NoError(t, err)

	info, err = app.FS.Stat("/TestProcessTemplate/input2/dir")
	assert.NoError(t, err)

	err = processTemplate(data, tpls, "/TestProcessTemplate/input2/dir",
		"/TestProcessTemplate/input2", "/TestProcessTemplate/output", info, logger)
	assert.NoError(t, err)

	// Verify the output directory
	exists, err = afero.DirExists(app.FS, "/TestProcessTemplate/output/dir")
	assert.NoError(t, err)
	assert.True(t, exists)
}
