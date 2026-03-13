package stamus

import (
	"context"
	"io"
	"os"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/client"
)

// FileOpener abstracts file system operations for config management.
type FileOpener interface {
	MkdirAll(path string, perm os.FileMode) error
	OpenFile(name string, flag int, perm os.FileMode) (*os.File, error)
	ReadAll(r io.Reader) ([]byte, error)
}

// ContainerLister abstracts Docker container queries.
type ContainerLister interface {
	GetContainersByProject(projectName string) ([]types.Container, error)
}

// ConfigManager holds dependencies for config and instance operations.
type ConfigManager struct {
	fs         FileOpener
	containers ContainerLister
}

// NewConfigManager creates a ConfigManager with the given dependencies.
func NewConfigManager(fs FileOpener, containers ContainerLister) *ConfigManager {
	return &ConfigManager{fs: fs, containers: containers}
}

// osFileOpener implements FileOpener with real OS calls.
type osFileOpener struct{}

func (o *osFileOpener) MkdirAll(path string, perm os.FileMode) error {
	return os.MkdirAll(path, perm)
}

func (o *osFileOpener) OpenFile(name string, flag int, perm os.FileMode) (*os.File, error) {
	return os.OpenFile(name, flag, perm)
}

func (o *osFileOpener) ReadAll(r io.Reader) ([]byte, error) {
	return io.ReadAll(r)
}

// dockerContainerLister implements ContainerLister with real Docker API.
type dockerContainerLister struct{}

func (d *dockerContainerLister) GetContainersByProject(projectName string) ([]types.Container, error) {
	apiClient, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		return nil, err
	}
	defer apiClient.Close()

	return apiClient.ContainerList(context.Background(), container.ListOptions{
		All: true,
		Filters: filters.NewArgs(
			filters.Arg("label", "com.docker.compose.project="+projectName),
		),
	})
}

// DefaultManager is the package-level ConfigManager used by backward-compatible wrapper functions.
var DefaultManager = NewConfigManager(&osFileOpener{}, &dockerContainerLister{})
