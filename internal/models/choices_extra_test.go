package models

import (
	"testing"

	"stamus-ctl/internal/app"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	modeTest = "test"
	modeProd = "prod"
)

// TestGetInterfacesHost_CachedPath covers the early-return cache branch in
// getInterfacesHost by pre-populating interfacesCache.
func TestGetInterfacesHost_CachedPath(t *testing.T) {
	// Save and restore cache and mode.
	saved := interfacesCache
	savedMode := app.Mode
	defer func() {
		interfacesCache = saved
		app.Mode = savedMode
	}()

	app.Mode = app.ModeStruct(modeTest) // non-prod → getInterfacesHost is called

	// Pre-populate the cache.
	cached := []Variable{CreateVariableString("lo"), CreateVariableString("eth0")}
	interfacesCache = cached

	// Calling getInterfacesHost should return the cached value without reading /sys.
	result, err := getInterfacesHost()
	require.NoError(t, err)
	assert.Len(t, result, 2)
	assert.Equal(t, "lo", *result[0].String)
	assert.Equal(t, "eth0", *result[1].String)
}

// TestGetChoices_Interfaces_DetectionDisabled verifies that enabling
// DisableInterfaceDetection (as `nix init` does) skips host/container probing
// and returns no choices, regardless of mode or cache state.
func TestGetChoices_Interfaces_DetectionDisabled(t *testing.T) {
	saved := interfacesCache
	savedMode := app.Mode
	savedDisable := DisableInterfaceDetection
	defer func() {
		interfacesCache = saved
		app.Mode = savedMode
		DisableInterfaceDetection = savedDisable
	}()

	// Even with a populated cache and prod mode, detection must be skipped.
	app.Mode = app.ModeStruct(modeProd)
	interfacesCache = []Variable{CreateVariableString("eth0")}
	DisableInterfaceDetection = true

	choices, err := GetChoices("interfaces")
	require.NoError(t, err)
	assert.Empty(t, choices)
}

// TestGetChoices_Interfaces_CachedPath covers the cache hit branch in
// getInterfacesBusybox (prod mode) and getInterfacesHost (non-prod mode) when
// the cache is already populated.
func TestGetChoices_Interfaces_CachedPath(t *testing.T) {
	saved := interfacesCache
	savedMode := app.Mode
	defer func() {
		interfacesCache = saved
		app.Mode = savedMode
	}()

	app.Mode = app.ModeStruct(modeTest)
	interfacesCache = []Variable{CreateVariableString("lo")}

	choices, err := GetChoices("interfaces")
	require.NoError(t, err)
	assert.NotEmpty(t, choices)
}

// TestGetChoices_BusyboxCachedPath covers the interfacesCache != nil early return
// inside getInterfacesBusybox by setting prod mode and pre-populating the cache.
func TestGetChoices_BusyboxCachedPath(t *testing.T) {
	saved := interfacesCache
	savedMode := app.Mode
	defer func() {
		interfacesCache = saved
		app.Mode = savedMode
	}()

	app.Mode = app.ModeStruct(modeProd) // → getInterfacesBusybox
	interfacesCache = []Variable{CreateVariableString("eth0")}

	choices, err := GetChoices("interfaces")
	require.NoError(t, err)
	assert.NotEmpty(t, choices)
}
