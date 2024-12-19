package models

import (
	"testing"

	"github.com/go-playground/assert/v2"
)

func TestArbitrary(t *testing.T) {
	a := NewArbitrary()

	a.Set("key", "value")
	assert.Equal(t, "value", a.AsMap()["key"])
}

func TestSetArbitrary(t *testing.T) {
	// Write a function that test SetArbitrary
	a := NewArbitrary()
	b := map[string]string{
		"key1": "value1",
		"key2": "value2",
	}
	a.SetArbitrary(b)
	assert.Equal(t, "value1", a.AsMap()["key1"])
	assert.Equal(t, "value2", a.AsMap()["key2"])
}
