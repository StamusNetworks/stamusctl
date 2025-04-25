package docker

import (
	"bytes"
	"strings"

	"stamus-ctl/internal/logging"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/api/types/network"
)

func createConfig(name string, cmd []string, volumes []string, net string) (container.Config, container.HostConfig, network.NetworkingConfig) {
	config := container.Config{
		Image: name,
		Cmd:   cmd,
	}

	hostConfig := container.HostConfig{}
	var mounts []mount.Mount
	for _, volume := range volumes {
		split := strings.Split(volume, ":")
		mount := mount.Mount{
			Type:   mount.TypeBind,
			Source: split[0],
			Target: split[1],
		}
		mounts = append(mounts, mount)
	}
	hostConfig.Mounts = mounts
	if net == "host" {
		hostConfig.NetworkMode = "host"
	}

	var networkConfig network.NetworkingConfig
	return config, hostConfig, networkConfig
}

func RunContainer(name string, cmd []string, volumes []string, net string) (string, error) {
	logger := logging.Sugar.With("name", name, "cmd", cmd, "volumes", volumes, "net", net)
	config, hostConfig, networkConfig := createConfig(name, cmd, volumes, net)

	resp, err := cli.ContainerCreate(ctx, &config, &hostConfig, &networkConfig, nil, "")
	if err != nil {
		logger.With("error", err).Error("container create")
		return "", err
	}

	if err := cli.ContainerStart(ctx, resp.ID, container.StartOptions{}); err != nil {
		logger.With("error", err).Error("container start")
		return "", err
	}

	statusCh, errCh := cli.ContainerWait(ctx, resp.ID, container.WaitConditionNotRunning)
	select {
	case err := <-errCh:
		if err != nil {
			logger.With("error", err).Error("container wait")
			return "", err
		}
	case <-statusCh:
	}

	out, err := cli.ContainerLogs(ctx, resp.ID, container.LogsOptions{ShowStdout: true})
	if err != nil {
		logger.With("error", err).Error("container logs")
		return "", err
	}

	buf := new(bytes.Buffer)
	buf.ReadFrom(out)

	output := buf.String()

	logger.Debugw("run output", "output", output)

	err = cli.ContainerRemove(ctx, resp.ID, container.RemoveOptions{RemoveVolumes: true, Force: true})
	if err != nil {
		logger.With("error", err).Error("container logs")
		return "", err
	}

	return output, nil
}
