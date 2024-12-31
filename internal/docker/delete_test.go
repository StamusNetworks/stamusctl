package docker

import (
	"testing"

	"github.com/go-playground/assert/v2"
)

func TestDeleteDockerImageByName(t *testing.T) {
	cli = &mockCli{}

	response, err := DeleteDockerImageByName("test", "")

	assert.Equal(t, true, response)
	assert.Equal(t, nil, err)

	response, err = DeleteDockerImageByName("toto", "")

	assert.Equal(t, false, response)
	assert.Equal(t, "image not found", err.Error())

	cli = &mockCli{
		failImageList: true,
	}
	response, err = DeleteDockerImageByName("test", "")

	assert.Equal(t, false, response)
	assert.Equal(t, "mock error", err.Error())

	cli = &mockCli{
		failImageRemove: true,
	}
	response, err = DeleteDockerImageByName("test", "")

	assert.Equal(t, false, response)
	assert.Equal(t, "mock error", err.Error())
}
