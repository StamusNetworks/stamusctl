package docker

import (
	"context"
	"errors"
	"io"
	"strings"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/api/types/network"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
)

type mockCli struct {
	failImageList       bool
	failImageRemove     bool
	failNetworkList     bool
	failImagePull       bool
	failContainerCreate bool
	failContainerStart  bool
	failContainerWait   bool
	failContainerLogs   bool
	failContainerRemove bool
}

func (m *mockCli) ImageList(ctx context.Context, options image.ListOptions) ([]image.Summary, error) {
	if m.failImageList {
		return nil, errors.New("mock error")
	}

	return []image.Summary{
		{
			RepoTags: []string{"test"},
			ID:       "1",
		},
		{
			RepoTags: []string{},
			ID:       "2",
		},
	}, nil
}

func (m *mockCli) ImageRemove(ctx context.Context, imageID string, options image.RemoveOptions) ([]image.DeleteResponse, error) {
	if m.failImageRemove {
		return nil, errors.New("mock error")
	}
	return []image.DeleteResponse{
		{
			Deleted:  imageID,
			Untagged: imageID,
		},
	}, nil
}

func (m *mockCli) NetworkList(ctx context.Context, options network.ListOptions) ([]network.Summary, error) {
	if m.failNetworkList {
		return nil, errors.New("mock error")
	}
	return []network.Summary{
		{
			Name: "test1",
			ID:   "1",
		},
		{
			Name: "test2",
			ID:   "2",
		},
	}, nil
}

func (m *mockCli) ImagePull(ctx context.Context, refStr string, options image.PullOptions) (io.ReadCloser, error) {
	if m.failImagePull {
		return nil, errors.New("mock error")
	}

	stringReader := strings.NewReader("mock")
	return io.NopCloser(stringReader), nil
}

func (m *mockCli) ContainerCreate(ctx context.Context, config *container.Config,
	hostConfig *container.HostConfig, networkingConfig *network.NetworkingConfig,
	platform *ocispec.Platform, containerName string,
) (container.CreateResponse, error) {
	if m.failContainerCreate {
		return container.CreateResponse{}, errors.New("mock error")
	}

	return container.CreateResponse{
		ID: "1",
	}, nil
}

func (m *mockCli) ContainerStart(ctx context.Context, containerID string, options container.StartOptions) error {
	if m.failContainerStart {
		return errors.New("mock error")
	}

	return nil
}

func (m *mockCli) ContainerWait(ctx context.Context, containerID string,
	condition container.WaitCondition,
) (<-chan container.WaitResponse, <-chan error) {
	response := make(chan container.WaitResponse)
	err := make(chan error)

	if m.failContainerWait {
		go func() {
			for {
				err <- errors.New("mock error wait")
			}
		}()
		return response, err
	}

	close(response)
	close(err)
	return response, err
}

func (m *mockCli) ContainerLogs(ctx context.Context, container string, options container.LogsOptions) (io.ReadCloser, error) {
	if m.failContainerLogs {
		return nil, errors.New("mock error")
	}

	stringReader := strings.NewReader("mock")
	return io.NopCloser(stringReader), nil
}

func (m *mockCli) ContainerRemove(ctx context.Context, containerID string, options container.RemoveOptions) error {
	if m.failContainerRemove {
		return errors.New("mock error")
	}

	return nil
}
