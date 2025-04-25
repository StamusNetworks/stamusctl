package docker

import (
	"testing"

	"github.com/docker/docker/api/types/image"
	"github.com/go-playground/assert/v2"
)

func TestImageName(t *testing.T) {
	result := ImageName(image.Summary{
		RepoTags: []string{"test"},
	})

	assert.Equal(t, "test", result)

	result = ImageName(image.Summary{
		RepoTags: []string{},
	})
	assert.Equal(t, "none", result)
}

func TestGetImagesName(t *testing.T) {
	result := GetImagesName([]image.Summary{{
		RepoTags: []string{"test"},
	}})

	assert.Equal(t, []string{"test"}, result)

	result = GetImagesName([]image.Summary{
		{
			RepoTags: []string{"test"},
		},
		{
			RepoTags: []string{},
		},
	})
	assert.Equal(t, []string{"test", "none"}, result)
}

func TestGetInstalledImagesName(t *testing.T) {
	cli = &mockCli{}

	result, err := GetInstalledImagesName()

	assert.Equal(t, []string{"test", "none"}, result)
	assert.Equal(t, nil, err)

	cli = &mockCli{
		failImageList: true,
	}

	result, err = GetInstalledImagesName()

	assert.Equal(t, nil, result)
	assert.Equal(t, "mock error", err.Error())
}

func TestIsImageAlreadyInstalled(t *testing.T) {
	cli = &mockCli{}

	result, err := IsImageAlreadyInstalled("test", "")
	assert.Equal(t, true, result)
	assert.Equal(t, nil, err)

	result, err = IsImageAlreadyInstalled("docker.io/library/", "test")
	assert.Equal(t, true, result)
	assert.Equal(t, nil, err)

	result, err = IsImageAlreadyInstalled("toto", "")
	assert.Equal(t, false, result)
	assert.Equal(t, nil, err)

	cli = &mockCli{
		failImageList: true,
	}

	result, err = IsImageAlreadyInstalled("toto", "")
	assert.Equal(t, false, result)
	assert.Equal(t, "mock error", err.Error())
}

func TestGetImageIdFromName(t *testing.T) {
	cli = &mockCli{}

	result, err := GetImageIdFromName("test", "")
	assert.Equal(t, "1", result)
	assert.Equal(t, nil, err)

	result, err = GetImageIdFromName("docker.io/library/", "test")
	assert.Equal(t, "1", result)
	assert.Equal(t, nil, err)

	result, err = GetImageIdFromName("toto", "")
	assert.Equal(t, "", result)
	assert.Equal(t, "image not found", err.Error())

	cli = &mockCli{
		failImageList: true,
	}

	result, err = GetImageIdFromName("toto", "")
	assert.Equal(t, "", result)
	assert.Equal(t, "mock error", err.Error())
}
