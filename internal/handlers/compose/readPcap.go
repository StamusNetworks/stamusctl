package handlers

import (
	// Core

	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime/debug"
	"strings"

	// Internal

    "stamus-ctl/internal/app"
    "stamus-ctl/internal/docker"
    "stamus-ctl/internal/utils"
	"stamus-ctl/internal/logging"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	"github.com/spf13/afero"
	"gopkg.in/yaml.v3"
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

func createConfig(configName, pcap string, image string) (container.Config, container.HostConfig, network.NetworkingConfig, error) {
	splitted := strings.Split(pcap, "/")
	pcapName := splitted[len(splitted)-1]

	dir, err := os.Getwd()
	if err != nil {
		return container.Config{}, container.HostConfig{}, network.NetworkingConfig{}, nil
	}

	config := container.Config{
		Image:      image,
		Entrypoint: []string{"/docker-entrypoint.sh"},
		Cmd: []string{"-vvv -k none -r /replay/" + pcapName +
			" --runmode autofp -l /var/log/suricata --set sensor-name=" + pcapName},
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
	// Resolve suricata image from compose or fallback to default
	image, err := resolveSuricataImage(configName)
	if err != nil {
		logger.With("error", err).Warn("image resolution failed; using default")
	}
	config, hostConfig, networkConfig, err := createConfig(configName, pcap, image)
	if err != nil {
		logger.With("error", err).Error("container configs")
		return "", err
	}

	reg, name := splitImageRef(image)
	_, err = docker.PullImageIfNotExisted(reg, name)
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

// resolveSuricataImage attempts to read the Suricata image from the instance compose file.
// Falls back to default if not found or on error.
func resolveSuricataImage(configPath string) (string, error) {
	// Optional override via env for emergencies
	logger := logging.Sugar.With("name", "resolve-suricata-image")
	if env := os.Getenv("STAMUSCTL_SURICATA_IMAGE"); strings.TrimSpace(env) != "" {
		return env, nil
	}

	// 1) Try effective compose config (handles multi-file setups)
	if img := resolveImageViaComposeConfig(configPath); img != "" {
		logger.With("image", img).Debug("resolved via docker compose config")
		return img, nil
	}

    // 2) Fallback: parse a single compose file for services.suricata.image
    compose := utils.GetComposeFilePath(configPath)
	logger.With("compose", compose).Debug("found compose file")

	b, err := afero.ReadFile(app.FS, compose)
	if err != nil {
		return defaultSuricataImage, err
	}

	type svc struct {
		Image string `yaml:"image"`
	}
	var payload struct {
		Services map[string]svc `yaml:"services"`
	}
	if err := yaml.Unmarshal(b, &payload); err != nil {
		return defaultSuricataImage, err
	}
	if payload.Services == nil {
		return defaultSuricataImage, errors.New("compose missing services")
	}
	s, ok := payload.Services["suricata"]
	if !ok || strings.TrimSpace(s.Image) == "" {
		return defaultSuricataImage, errors.New("suricata image not found in compose")
	}
	logger.With("image", s.Image).Debug("resolved suricata image")
	return s.Image, nil
}

// resolveImageViaComposeConfig runs `docker compose -f <file> config` to get the fully
// resolved compose configuration, then extracts services.suricata.image.
func resolveImageViaComposeConfig(configPath string) string {
    composeFile := utils.GetComposeFilePath(configPath)
	// Run from the compose file directory and reference the file by basename
	fileDir := filepath.Dir(composeFile)
	fileName := filepath.Base(composeFile)
	logging.Sugar.With("dir", fileDir, "file", fileName).Debug("resolving via compose config")
	cmd := exec.Command("docker", "compose", "-f", fileName, "config")
	cmd.Dir = fileDir
	out, err := cmd.CombinedOutput()
	if err != nil {
		logging.Sugar.With("error", err).With("out", string(out)).Info("docker compose config")
		return ""
	}
	type svc struct {
		Image string `yaml:"image"`
	}
	var payload struct {
		Services map[string]svc `yaml:"services"`
	}
	if err := yaml.Unmarshal(out, &payload); err != nil {
		return ""
	}
	if s, ok := payload.Services["suricata"]; ok {
		if strings.TrimSpace(s.Image) != "" {
			return s.Image
		}
	}
	return ""
}

var defaultSuricataImage = "jasonish/suricata:master-amd64"

// splitImageRef splits an image reference into a registry-like prefix and the remainder name:tag.
// Example: "ghcr.io/org/repo:tag" -> ("ghcr.io/org/", "repo:tag"); "jasonish/suricata:tag" -> ("jasonish/", "suricata:tag")
// If there is no slash, returns ("", image).
func splitImageRef(image string) (string, string) {
	idx := strings.LastIndex(image, "/")
	if idx == -1 {
		return "", image
	}
	return image[:idx+1], image[idx+1:]
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
