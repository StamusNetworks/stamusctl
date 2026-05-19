package app

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetConfigsFolder(t *testing.T) {
	t.Run("with name", func(t *testing.T) {
		got := GetConfigsFolder("myconfig")
		assert.Equal(t, filepath.Join(ConfigsFolder, "myconfig"), got)
	})

	t.Run("empty name", func(t *testing.T) {
		got := GetConfigsFolder("")
		// filepath.Join strips trailing slashes; compare against the canonical form.
		assert.Equal(t, filepath.Join(ConfigsFolder), got)
	})
}

func TestIsCtl(t *testing.T) {
	oldName := Name
	defer func() { Name = oldName }()

	Name = CtlName
	assert.True(t, IsCtl())

	Name = "stamusd"
	assert.False(t, IsCtl())
}

func TestModeStruct(t *testing.T) {
	t.Run("test mode IsTest", func(t *testing.T) {
		m := ModeStruct("test")
		assert.True(t, m.IsTest())
	})

	t.Run("prod mode IsTest", func(t *testing.T) {
		m := ModeStruct("prod")
		assert.False(t, m.IsTest())
	})

	t.Run("prod mode IsProd", func(t *testing.T) {
		m := ModeStruct("prod")
		assert.True(t, m.IsProd())
	})

	t.Run("test mode IsProd", func(t *testing.T) {
		m := ModeStruct("test")
		assert.False(t, m.IsProd())
	})
}

func TestEmbedStruct(t *testing.T) {
	t.Run("true string IsTrue", func(t *testing.T) {
		e := EmbedStruct("true")
		assert.True(t, e.IsTrue())
	})

	t.Run("false string IsTrue", func(t *testing.T) {
		e := EmbedStruct("false")
		assert.False(t, e.IsTrue())
	})

	t.Run("empty string IsTrue", func(t *testing.T) {
		e := EmbedStruct("")
		assert.False(t, e.IsTrue())
	})
}

func TestModeStruct_set(t *testing.T) {
	var m ModeStruct
	m.set("test")
	assert.True(t, m.IsTest())
	m.set("prod")
	assert.True(t, m.IsProd())
}

func TestEmbedStruct_Set(t *testing.T) {
	var e EmbedStruct
	e.Set("true")
	assert.True(t, e.IsTrue())
	e.Set("false")
	assert.False(t, e.IsTrue())
}
