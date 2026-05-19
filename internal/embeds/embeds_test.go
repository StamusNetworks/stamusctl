package embeds

import (
	"testing"

	"stamus-ctl/internal/app"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestInitClearNDRFolder_FolderExists exercises the path where the folder already exists.
// In this case, no extraction occurs.
func TestInitClearNDRFolder_FolderExists(t *testing.T) {
	// Save and restore FS.
	oldFS := app.FS
	app.FS = afero.NewMemMapFs()
	defer func() { app.FS = oldFS }()

	// Create the folder so FolderExists returns true.
	path := "/test-clearndr"
	require.NoError(t, app.FS.MkdirAll(path, 0o755))

	// Should NOT panic (folder exists → no extraction needed).
	assert.NotPanics(t, func() {
		InitClearNDRFolder(path)
	})
}

// TestInitClearNDRFolder_FolderMissingEmbedFalse exercises the path where the folder
// does not exist AND embed mode is false (the else branch where extraction is skipped).
func TestInitClearNDRFolder_FolderMissingEmbedFalse(t *testing.T) {
	oldFS := app.FS
	app.FS = afero.NewMemMapFs()
	oldEmbed := app.Embed
	app.Embed = app.EmbedStruct("false")
	defer func() {
		app.FS = oldFS
		app.Embed = oldEmbed
	}()

	// Folder doesn't exist + embed mode off → function does nothing.
	assert.NotPanics(t, func() {
		InitClearNDRFolder("/nonexistent-clearndr-path")
	})
}

// TestInitClearNDRFolder_FolderMissingEmbedTrue exercises the extraction path
// where the folder does not exist AND embed mode is true.
// ExtractEmbedTo writes to app.FS (in-memory), so no real files are created.
func TestInitClearNDRFolder_FolderMissingEmbedTrue(t *testing.T) {
	oldFS := app.FS
	memFS := afero.NewMemMapFs()
	app.FS = memFS
	oldEmbed := app.Embed
	app.Embed = app.EmbedStruct("true")
	oldTemplatesFolder := app.TemplatesFolder
	app.TemplatesFolder = "/tmp-templates/"
	defer func() {
		app.FS = oldFS
		app.Embed = oldEmbed
		app.TemplatesFolder = oldTemplatesFolder
	}()

	// Path does not exist in memFS → clearndrConfigExist == false → extraction runs.
	assert.NotPanics(t, func() {
		InitClearNDRFolder("/nonexistent-embed-path")
	})
}
