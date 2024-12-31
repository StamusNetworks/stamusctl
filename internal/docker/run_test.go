package docker

import (
	"testing"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/api/types/network"
	"github.com/go-playground/assert/v2"
)

func TestCreateConfig(t *testing.T) {
	config, host, net := createConfig("test", []string{"ls"}, []string{"host_v:container_v"}, "net1")

	assert.Equal(t, container.Config{
		Image: "test",
		Cmd:   []string{"ls"},
	}, config)
	assert.Equal(t, container.HostConfig{
		Mounts: []mount.Mount{
			{
				Type:   mount.TypeBind,
				Source: "host_v",
				Target: "container_v",
			},
		},
	}, host)
	assert.Equal(t, network.NetworkingConfig{}, net)
}

func TestCreateConfigHostNet(t *testing.T) {
	config, host, net := createConfig("test", []string{"ls"}, []string{"host_v:container_v"}, "host")

	assert.Equal(t, container.Config{
		Image: "test",
		Cmd:   []string{"ls"},
	}, config)
	assert.Equal(t, container.HostConfig{
		Mounts: []mount.Mount{
			{
				Type:   mount.TypeBind,
				Source: "host_v",
				Target: "container_v",
			},
		},
		NetworkMode: "host",
	}, host)
	assert.Equal(t, network.NetworkingConfig{}, net)
}

func TestRunContainer(t *testing.T) {
	cli = &mockCli{}

	output, err := RunContainer("test", []string{"ls"}, []string{"host_v:container_v"}, "net1")

	assert.Equal(t, "mock", output)
	assert.Equal(t, nil, err)
}

func TestRunContainerFailCreate(t *testing.T) {
	cli = &mockCli{
		failContainerCreate: true,
	}

	output, err := RunContainer("test", []string{"ls"}, []string{"host_v:container_v"}, "net1")

	assert.Equal(t, "", output)
	assert.Equal(t, "mock error", err.Error())
}

func TestRunContainerFailStart(t *testing.T) {
	cli = &mockCli{
		failContainerStart: true,
	}

	output, err := RunContainer("test", []string{"ls"}, []string{"host_v:container_v"}, "net1")

	assert.Equal(t, "", output)
	assert.Equal(t, "mock error", err.Error())
}

func TestRunContainerFailWait(t *testing.T) {
	cli = &mockCli{
		failContainerWait: true,
	}

	output, err := RunContainer("test", []string{"ls"}, []string{"host_v:container_v"}, "net1")

	assert.Equal(t, "", output)
	assert.Equal(t, "mock error wait", err.Error())
}

func TestRunContainerFailLogs(t *testing.T) {
	cli = &mockCli{
		failContainerLogs: true,
	}

	output, err := RunContainer("test", []string{"ls"}, []string{"host_v:container_v"}, "net1")

	assert.Equal(t, "", output)
	assert.Equal(t, "mock error", err.Error())
}

func TestRunContainerFailRemove(t *testing.T) {
	cli = &mockCli{
		failContainerRemove: true,
	}

	output, err := RunContainer("test", []string{"ls"}, []string{"host_v:container_v"}, "net1")

	assert.Equal(t, "", output)
	assert.Equal(t, "mock error", err.Error())
}
