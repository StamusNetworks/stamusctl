package docker

import (
	"testing"

	"github.com/go-playground/assert/v2"
)

func TestPullImageIfNotExisted(t *testing.T) {
	cli = &mockCli{}

	result, err := PullImageIfNotExisted("test", "")
	assert.Equal(t, true, result)
	assert.Equal(t, nil, err)

	result, err = PullImageIfNotExisted("toto", "")
	assert.Equal(t, false, result)
	assert.Equal(t, nil, err)

	cli = &mockCli{
		failImageList: true,
	}
	result, err = PullImageIfNotExisted("test", "")
	assert.Equal(t, true, result)
	assert.Equal(t, "mock error", err.Error())

	cli = &mockCli{
		failImagePull: true,
	}
	result, err = PullImageIfNotExisted("test", "")
	assert.Equal(t, true, result)
	assert.Equal(t, nil, err)

	cli = &mockCli{
		failImagePull: true,
	}
	result, err = PullImageIfNotExisted("toto", "")
	assert.Equal(t, false, result)
	assert.Equal(t, "mock error", err.Error())
}
