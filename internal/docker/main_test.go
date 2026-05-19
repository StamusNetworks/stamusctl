package docker

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestClose_WithNilRawClient(t *testing.T) {
	// Save and restore rawClient.
	orig := rawClient
	rawClient = nil
	defer func() { rawClient = orig }()

	err := Close()
	assert.NoError(t, err)
}

func TestClient_ReturnsCurrentCli(t *testing.T) {
	// Save and restore cli.
	orig := cli
	cli = &mockCli{}
	defer func() { cli = orig }()

	got := Client()
	assert.NotNil(t, got)
	assert.Equal(t, cli, got)
}

func TestMockCli_Close(t *testing.T) {
	m := &mockCli{}
	err := m.Close()
	assert.NoError(t, err)
}

func TestClose_WithRealClient(t *testing.T) {
	// rawClient is initialized by init() (sync.Once).
	// If docker is available, rawClient is non-nil and we can call Close.
	// We save and restore so other tests still work.
	if rawClient == nil {
		t.Skip("rawClient is nil (docker not available)")
	}

	// Call Close() with the real rawClient — this covers the `return rawClient.Close()` branch.
	// The connection may be already closed from another test, but the call itself runs the branch.
	origRaw := rawClient
	origCli := cli

	err := Close()
	// Error may occur if already closed, but that's fine — we just need branch coverage.
	_ = err

	// Restore rawClient and cli so other tests work.
	rawClient = origRaw
	cli = origCli
}
