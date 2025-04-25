package docker

import (
	"context"
	"io"
	"runtime/debug"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
)

type IClient interface {
	ImageList(ctx context.Context, options image.ListOptions) ([]image.Summary, error)
	ImageRemove(ctx context.Context, imageID string, options image.RemoveOptions) ([]image.DeleteResponse, error)
	NetworkList(ctx context.Context, options network.ListOptions) ([]network.Summary, error)
	ImagePull(ctx context.Context, refStr string, options image.PullOptions) (io.ReadCloser, error)
	ContainerCreate(
		ctx context.Context,
		config *container.Config,
		hostConfig *container.HostConfig,
		networkingConfig *network.NetworkingConfig,
		platform *ocispec.Platform,
		containerName string,
	) (container.CreateResponse, error)
	ContainerStart(ctx context.Context, containerID string, options container.StartOptions) error
	ContainerWait(ctx context.Context, containerID string, condition container.WaitCondition) (
		<-chan container.WaitResponse, <-chan error)
	ContainerLogs(ctx context.Context, container string, options container.LogsOptions) (io.ReadCloser, error)
	ContainerRemove(ctx context.Context, containerID string, options container.RemoveOptions) error
}

var (
	ctx = context.Background()
	cli IClient
)

func init() {
	docker, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	cli = docker

	if err != nil {
		debug.PrintStack()
		panic(err)
	}
}
