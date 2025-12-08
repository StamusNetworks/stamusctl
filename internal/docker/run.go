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

	var resp container.CreateResponse
	err := WithRetrySimple(func() error {
		var createErr error
		resp, createErr = cli.ContainerCreate(ctx, &config, &hostConfig, &networkConfig, nil, "")
		return createErr
	}, "container-create")
	if err != nil {
		logger.With("error", err).Error("container create")
		return "", err
	}

	err = WithRetrySimple(func() error {
		return cli.ContainerStart(ctx, resp.ID, container.StartOptions{})
	}, "container-start")
	if err != nil {
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

	var out bytes.Buffer
	err = WithRetrySimple(func() error {
		reader, logsErr := cli.ContainerLogs(ctx, resp.ID, container.LogsOptions{
			ShowStdout: true,
		})
		if logsErr != nil {
			return logsErr
		}
		defer reader.Close()

		out.Reset()
		_, readErr := out.ReadFrom(reader)
		return readErr
	}, "container-logs")
	if err != nil {
		logger.With("error", err).Error("container logs")
		return "", err
	}

	output := out.String()

	logger.Debugw("run output", "output", output)

	err = WithRetrySimple(func() error {
		return cli.ContainerRemove(ctx, resp.ID, container.RemoveOptions{RemoveVolumes: true, Force: true})
	}, "container-remove")
	if err != nil {
		logger.With("error", err).Error("container remove")
		return "", err
	}

	return output, nil
}
