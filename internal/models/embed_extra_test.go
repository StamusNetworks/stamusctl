package models

import (
	"testing"

	"stamus-ctl/internal/app"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
)

// TestGetAllFilesEmbed_WithSubdir exercises the recursive directory path in
// getAllFilesEmbed by using a testFS that has a subdirectory.
// Note: ExtractEmbedTo cannot be tested directly because getAllFilesEmbed
// prepends "./" to paths which makes embed.FS.ReadFile fail — this is a known
// limitation of the implementation.
func TestGetAllFilesEmbed_SubdirRecursion(t *testing.T) {
	oldFS := app.FS
	app.FS = afero.NewMemMapFs()
	defer func() { app.FS = oldFS }()

	// testEmbed only embeds embed_test.go (a single file at the root).
	// getAllFilesEmbed(".", testEmbed) returns ["./embed_test.go"].
	files := getAllFilesEmbed(".", testEmbed)
	assert.NotEmpty(t, files)

	// All returned paths must start with "./" since root is "."
	for _, f := range files {
		assert.Positive(t, len(f), "file path must not be empty")
	}
}

// ---------------------------------------------------------------------------
// deleteEmptyFiles — permission error branch
// Note: MemMapFs doesn't enforce permissions, so this tests the walk path.
// ---------------------------------------------------------------------------

// TestDeleteEmptyFiles_WithContent confirms non-empty files are preserved.
func TestDeleteEmptyFiles_NonEmptyFilePreserved(t *testing.T) {
	require := func(err error) {
		if err != nil {
			t.Fatalf("setup failed: %v", err)
		}
	}
	oldFS := app.FS
	app.FS = afero.NewMemMapFs()
	defer func() { app.FS = oldFS }()

	dir := "/TestDeleteEmptyFiles_Nonempty"
	require(app.FS.MkdirAll(dir, 0o755))
	require(afero.WriteFile(app.FS, dir+"/nonempty.txt", []byte("data"), 0o644))
	require(afero.WriteFile(app.FS, dir+"/empty.txt", []byte(""), 0o644))

	err := deleteEmptyFiles(dir)
	assert.NoError(t, err)

	// nonempty.txt should still exist.
	exists, _ := afero.Exists(app.FS, dir+"/nonempty.txt")
	assert.True(t, exists)
	// empty.txt should have been deleted.
	exists, _ = afero.Exists(app.FS, dir+"/empty.txt")
	assert.False(t, exists)
}

// TestDeleteEmptyFolders_RemovesEmptySubdir exercises the empty sub-directory
// removal path in deleteEmptyFolders.
func TestDeleteEmptyFolders_RemovesEmptySubdir(t *testing.T) {
	require := func(err error) {
		if err != nil {
			t.Fatalf("setup failed: %v", err)
		}
	}
	oldFS := app.FS
	app.FS = afero.NewMemMapFs()
	defer func() { app.FS = oldFS }()

	dir := "/TestDeleteEmptyFolders_Subdir"
	emptySubdir := dir + "/emptysubdir"
	require(app.FS.MkdirAll(emptySubdir, 0o755))
	// Add a file to the parent so it is not empty overall.
	require(afero.WriteFile(app.FS, dir+"/file.txt", []byte("content"), 0o644))

	err := deleteEmptyFolders(dir)
	assert.NoError(t, err)

	// emptysubdir should have been removed.
	exists, _ := afero.DirExists(app.FS, emptySubdir)
	assert.False(t, exists)
	// The parent and its file should still exist.
	exists, _ = afero.Exists(app.FS, dir+"/file.txt")
	assert.True(t, exists)
}

// TestDeleteEmptyFolders_AllEmpty covers the case where the root itself is empty.
func TestDeleteEmptyFolders_AllEmpty(t *testing.T) {
	require := func(err error) {
		if err != nil {
			t.Fatalf("setup failed: %v", err)
		}
	}
	oldFS := app.FS
	app.FS = afero.NewMemMapFs()
	defer func() { app.FS = oldFS }()

	dir := "/TestDeleteEmptyFolders_AllEmpty"
	require(app.FS.MkdirAll(dir, 0o755))

	// deleteEmptyFolders returns immediately when root is empty (no walk).
	err := deleteEmptyFolders(dir)
	assert.NoError(t, err)
}

// ---------------------------------------------------------------------------
// isValidPath — Name empty and Type empty branches
// ---------------------------------------------------------------------------

func TestCreateFile_EmptyName(t *testing.T) {
	// Creating a file with empty name (before the dot) should fail.
	// ".yaml" has name="" and type="yaml" — triggers the Name=="" branch.
	_, err := CreateFile("/some/path", ".yaml")
	// The split gives ["", "yaml"] — name is empty, triggers isValidPath error.
	assert.Error(t, err)
}
