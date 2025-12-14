package models

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRelease_SetName(t *testing.T) {
	release := &Release{Name: "old-name"}

	result := release.SetName("new-name")

	assert.Equal(t, "new-name", release.Name)
	assert.Equal(t, release, result, "Should return the same release instance")
}

func TestRelease_SetLocation(t *testing.T) {
	release := &Release{Location: "/old/location"}

	result := release.SetLocation("/new/location")

	assert.Equal(t, "/new/location", release.Location)
	assert.Equal(t, release, result, "Should return the same release instance")
}

func TestRelease_SetIsUpgrade(t *testing.T) {
	tests := []struct {
		name     string
		initial  bool
		setValue bool
	}{
		{
			name:     "Set to true",
			initial:  false,
			setValue: true,
		},
		{
			name:     "Set to false",
			initial:  true,
			setValue: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			release := &Release{IsUpgrade: tt.initial}

			result := release.SetIsUpgrade(tt.setValue)

			assert.Equal(t, tt.setValue, release.IsUpgrade)
			assert.Equal(t, release, result, "Should return the same release instance")
		})
	}
}

func TestRelease_SetIsInstall(t *testing.T) {
	tests := []struct {
		name     string
		initial  bool
		setValue bool
	}{
		{
			name:     "Set to true",
			initial:  false,
			setValue: true,
		},
		{
			name:     "Set to false",
			initial:  true,
			setValue: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			release := &Release{IsInstall: tt.initial}

			result := release.SetIsInstall(tt.setValue)

			assert.Equal(t, tt.setValue, release.IsInstall)
			assert.Equal(t, release, result, "Should return the same release instance")
		})
	}
}

func TestRelease_SetService(t *testing.T) {
	release := &Release{Service: "old-service"}

	result := release.SetService("new-service")

	assert.Equal(t, "new-service", release.Service)
	assert.Equal(t, release, result, "Should return the same release instance")
}

func TestRelease_AsMap(t *testing.T) {
	release := &Release{
		Name:      "test-release",
		User:      "1000",
		Group:     "1000",
		Location:  "/test/location",
		IsUpgrade: true,
		IsInstall: false,
		Service:   "test-service",
		Seed:      "test-seed",
	}

	result := release.AsMap()

	assert.Equal(t, "test-release", result["Release.name"])
	assert.Equal(t, "1000", result["Release.user"])
	assert.Equal(t, "1000", result["Release.group"])
	assert.Equal(t, "/test/location", result["Release.location"])
	assert.Equal(t, true, result["Release.isUpgrade"])
	assert.Equal(t, false, result["Release.isInstall"])
	assert.Equal(t, "test-service", result["Release.service"])
	assert.Equal(t, "test-seed", result["Release.seed"])
}

func TestNewRelease(t *testing.T) {
	name := "test-release"
	location := "/test/location"
	seed := "test-seed"
	isUpgrade := true
	isInstall := false

	release := NewRelease(name, location, seed, isUpgrade, isInstall)

	assert.NotNil(t, release)
	assert.Equal(t, name, release.Name)
	assert.Equal(t, location, release.Location)
	assert.Equal(t, seed, release.Seed)
	assert.Equal(t, isUpgrade, release.IsUpgrade)
	assert.Equal(t, isInstall, release.IsInstall)
	assert.NotEmpty(t, release.User)
	assert.NotEmpty(t, release.Group)
	assert.NotEmpty(t, release.Service)
}

func TestNewTemplate(t *testing.T) {
	tests := []struct {
		name             string
		templateName     string
		templatePath     string
		expectedName     string
		expectedVersion  string
	}{
		{
			name:            "Simple path",
			templateName:    "test-template",
			templatePath:    "/path/to/1.0.0",
			expectedName:    "test-template",
			expectedVersion: "1.0.0",
		},
		{
			name:            "Nested path",
			templateName:    "another-template",
			templatePath:    "/deep/nested/path/2.5.3",
			expectedName:    "another-template",
			expectedVersion: "2.5.3",
		},
		{
			name:            "Single level path",
			templateName:    "simple",
			templatePath:    "latest",
			expectedName:    "simple",
			expectedVersion: "latest",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			template := NewTemplate(tt.templateName, tt.templatePath)

			assert.NotNil(t, template)
			assert.Equal(t, tt.expectedName, template.templateName)
			assert.Equal(t, tt.expectedVersion, template.templateVersion)
		})
	}
}

func TestTemplate_AsMap(t *testing.T) {
	template := &Template{
		templateName:    "test-template",
		templateVersion: "1.2.3",
	}

	result := template.AsMap()

	assert.Equal(t, "test-template", result["Template.name"])
	assert.Equal(t, "1.2.3", result["Template.version"])
}

func TestRelease_SetterChaining(t *testing.T) {
	// Test that setters can be chained
	release := &Release{}

	result := release.
		SetName("chained-release").
		SetLocation("/chained/location").
		SetIsUpgrade(true).
		SetIsInstall(false).
		SetService("chained-service")

	assert.Equal(t, "chained-release", release.Name)
	assert.Equal(t, "/chained/location", release.Location)
	assert.Equal(t, true, release.IsUpgrade)
	assert.Equal(t, false, release.IsInstall)
	assert.Equal(t, "chained-service", release.Service)
	assert.Equal(t, release, result)
}
