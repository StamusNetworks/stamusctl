package docker

import (
	"testing"

	"github.com/go-playground/assert/v2"
)

func TestGetNetworkIdByName(t *testing.T) {
	cli = &mockCli{}

	result, err := GetNetworkIdByName("test1")
	assert.Equal(t, "1", result)
	assert.Equal(t, nil, err)

	result, err = GetNetworkIdByName("test2")
	assert.Equal(t, "2", result)
	assert.Equal(t, nil, err)

	result, err = GetNetworkIdByName("toto")
	assert.Equal(t, "", result)
	assert.Equal(t, "network not found", err.Error())

	cli = &mockCli{
		failNetworkList: true,
	}

	result, err = GetNetworkIdByName("test")
	assert.Equal(t, "", result)
	assert.Equal(t, "mock error", err.Error())
}
