package nix

import (
	"testing"

	// Internal
	"stamus-ctl/internal/app"

	// External
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
)

func TestIsNixOS_True(t *testing.T) {
	oldFS := app.FS
	app.FS = afero.NewMemMapFs()
	defer func() { app.FS = oldFS }()

	_, err := app.FS.Create(nixosMarker)
	assert.NoError(t, err)

	assert.True(t, IsNixOS())
}

func TestIsNixOS_False(t *testing.T) {
	oldFS := app.FS
	app.FS = afero.NewMemMapFs()
	defer func() { app.FS = oldFS }()

	assert.False(t, IsNixOS())
}
