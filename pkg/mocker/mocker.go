package mocker

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"path/filepath"
	"time"

	"stamus-ctl/internal/app"
	"stamus-ctl/pkg"

	"github.com/docker/docker/api/types"
	"github.com/spf13/afero"
	"github.com/spf13/viper"
)

type mocked map[string]types.Container

func createMocked() mocked {
	return make(map[string]types.Container)
}

var Mocked mocked = createMocked()

func (m *mocked) Up(path string) error {
	// Get services
	services, err := getServices(path)
	if err != nil {
		return err
	}
	// Create mocked services
	for _, service := range services {
		(*m)[service] = types.Container{
			ID:    randomContainerId(),
			Names: []string{"/" + service},
		}
	}
	return nil
}

func (m *mocked) Down(path string) error {
	getServices(path)
	*m = createMocked()
	return nil
}

func (m *mocked) Restart(path string) error {
	m.Down(path)
	return m.Up(path)
}

func (m *mocked) RestartContainers(containers []string) error {
	for _, container := range containers {
		if _, ok := (*m)[container]; ok {
			(*m)[container] = createContainer(container)
		}
	}
	return nil
}

func (m *mocked) Ps() ([]types.Container, error) {
	var containers []types.Container
	for _, container := range *m {
		containers = append(containers, container)
	}
	return containers, nil
}

func (m *mocked) Logs() (pkg.LogsResponse, error) {
	var logs []pkg.ContainerLogs
	for _, container := range *m {
		logs = append(logs, pkg.ContainerLogs{
			Container: container,
			Logs:      []string{"log1", "log2"},
		})
	}
	return pkg.LogsResponse{Containers: logs}, nil
}

func getServices(path string) ([]string, error) {
	// get file content
	filePath := filepath.Join(path, "docker-compose.yaml")
	content, err := afero.ReadFile(app.FS, filePath)
	if err != nil {
		return []string{}, err
	}
	// Instantiate viper with file content
	v := viper.New()
	v.SetConfigType("yaml")
	v.ReadConfig(bytes.NewBuffer(content))
	// Get services
	allServices := v.GetStringMap("services")
	// Filter first class
	var services []string
	for service := range allServices {
		services = append(services, service)
	}
	return services, nil
}

func randomContainerId() string {
	// Docker container IDs are 64 characters of hexadecimal
	bytes := make([]byte, 32) // 32 bytes * 2 characters per byte = 64 characters
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}

func createContainer(service string) types.Container {
	return types.Container{
		ID:         randomContainerId(),
		Names:      []string{"/" + service},
		Image:      service,
		ImageID:    randomContainerId(),
		Command:    "bash -c 'while true; do echo hello; sleep 1; done'",
		Created:    time.Now().Unix(),
		Ports:      []types.Port{{PrivatePort: 80, PublicPort: 8080, Type: "tcp"}},
		SizeRw:     1000,
		SizeRootFs: 2000,
		Labels: map[string]string{
			"com.docker.compose.service": service,
			"com.docker.compose.project": "mocked",
		},
		State:  "running",
		Status: "Up 1 second",
		HostConfig: struct {
			NetworkMode string            `json:",omitempty"`
			Annotations map[string]string `json:",omitempty"`
		}{
			NetworkMode: "default",
		},
		// NetworkSettings: SummaryNetworkSettings{},
		Mounts: []types.MountPoint{
			{
				Type:        "bind",
				Name:        "mocked",
				Source:      "/mocked",
				Destination: "/mocked",
				Driver:      "local",
				Mode:        "rw",
				RW:          true,
				Propagation: "rprivate",
			},
		},
	}
}
