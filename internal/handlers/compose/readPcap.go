package handlers

import (
	// Core

	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"os"
	"runtime/debug"
	"strings"

	// Internal

	"stamus-ctl/internal/app"
	"stamus-ctl/internal/docker"
	"stamus-ctl/internal/logging"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
)

type ReadPcapParams struct {
	Config   string
	PcapPath string
}

func initCli() *client.Client {
	docker, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	cli := docker

	if err != nil {
		debug.PrintStack()
		panic(err)
	}

	return cli
}

func createConfig(configName, pcap string) (container.Config, container.HostConfig, network.NetworkingConfig, error) {
	splitted := strings.Split(pcap, "/")
	pcapName := splitted[len(splitted)-1]

	dir, err := os.Getwd()
	if err != nil {

		return container.Config{}, container.HostConfig{}, network.NetworkingConfig{}, nil
	}

	config := container.Config{
		Image:      "jasonish/suricata:master-amd64-profiling",
		Entrypoint: []string{"/docker-entrypoint.sh"},
		Cmd:        []string{"-vvv -k none -r /replay/" + pcapName + " --runmode autofp -l /var/log/suricata --set sensor-name=" + pcapName},
	}

	hostConfig := container.HostConfig{
		AutoRemove: true,
		Mounts: []mount.Mount{
			{
				Type:   mount.TypeBind,
				Source: dir + "/" + configName + "/containers-data/suricata/etc",
				Target: "/etc/suricata",
			},
			{
				Type:   mount.TypeBind,
				Source: dir + "/" + configName + "/containers-data/suricata/rules",
				Target: "/etc/suricata/rules",
			},
			{
				Type:   mount.TypeBind,
				Source: dir + "/" + configName + "/containers-data/suricata/logs",
				Target: "/var/log/suricata",
			},
			{
				Type:   mount.TypeBind,
				Source: dir + "/" + configName + "/fpc",
				Target: "/var/log/suricata/fpc",
			},
			{
				Type:     mount.TypeBind,
				Source:   pcap,
				Target:   "/replay/" + pcapName,
				ReadOnly: true,
			},
		},
		CapAdd: []string{"net_admin", "sys_nice"},
	}

	var networkConfig network.NetworkingConfig
	return config, hostConfig, networkConfig, nil
}

func runContainer(configName, pcap string) (string, error) {
	logger := logging.Sugar.With("name", "suricata-readpcap")
	cli := initCli()
	ctx := context.Background()
	config, hostConfig, networkConfig, err := createConfig(configName, pcap)

	if err != nil {
		logger.With("error", err).Error("container configs")
		return "", err
	}

	_, err = docker.PullImageIfNotExisted("jasonish/", "suricata:master-amd64-profiling")
	if err != nil {
		logger.With("error", err).Error("image pull")
		return "", err
	}

	resp, err := cli.ContainerCreate(ctx, &config, &hostConfig, &networkConfig, nil, "")
	if err != nil {
		logger.With("error", err).Error("container create")
		return "", err
	}

	if err := cli.ContainerStart(ctx, resp.ID, container.StartOptions{}); err != nil {
		logger.With("error", err).Error("container start")
		return "", err
	}

	// statusCh, errCh := cli.ContainerWait(ctx, resp.ID, container.WaitConditionNotRunning)
	// select {
	// case err := <-errCh:
	// 	if err != nil {
	// 		logging.Sugar.Warn(err)
	// 		return "", err
	// 	}
	// case status := <-statusCh:
	// 	logging.Sugar.Info(status.StatusCode)
	// 	logging.Sugar.Info(status.Error)
	// }

	out, _ := cli.ContainerLogs(ctx, resp.ID, container.LogsOptions{
		ShowStdout: true,
		ShowStderr: true,
		Timestamps: false,
		Follow:     true,
		Tail:       "40",
	})
	hdr := make([]byte, 8)
	for {
		_, err := out.Read(hdr)
		if err != nil {
			return "", err
		}
		var w io.Writer
		switch hdr[0] {
		case 1:
			w = os.Stdout
		default:
			w = os.Stderr
		}
		count := binary.BigEndian.Uint32(hdr[4:])
		dat := make([]byte, count)
		_, _ = out.Read(dat)
		fmt.Fprint(w, string(dat))
	}
}

func PcapHandler(params ReadPcapParams) error {

	if _, err := app.FS.Stat(params.Config); os.IsNotExist(err) {
		return errors.New("instance '" + params.Config + "' don't exist")
	}
	if _, err := app.FS.Stat(params.Config + "/containers-data/suricata/etc"); os.IsNotExist(err) {
		return errors.New("instance '" + params.Config + "' seems do not have been started.")
	}

	output, _ := runContainer(params.Config, params.PcapPath)

	logging.Sugar.Info(output)

	return nil
}
