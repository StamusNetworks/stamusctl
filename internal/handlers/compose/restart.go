package handlers

import (
	// Common
	"context"
	"fmt"
	"sync"

	// Internal

	"stamus-ctl/internal/app"
	"stamus-ctl/internal/handlers/wrapper"
	"stamus-ctl/pkg/mocker"

	// External
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
)

func HandleConfigRestart(conf string) error {
	if app.Mode.IsTest() {
		return mocker.Mocked.Restart(conf)
	}
	return handleConfigRestart(conf)
}

// HandleConfigRestart restarts the containers defined in the container composition file
func handleConfigRestart(conf string) error {
	// Note: Don't call GetConfigsFolder here - wrapper.HandleDown and HandleUp already do it
	err := wrapper.HandleDown(conf, false, false)
	if err != nil {
		return err
	}
	return wrapper.HandleUp(conf)
}

func HandleContainersRestart(containers []string) error {
	if app.Mode.IsTest() {
		return mocker.Mocked.RestartContainers(containers)
	}
	return handleContainersRestart(containers)
}

// Given a list of container IDs, restarts them
func handleContainersRestart(containers []string) error {
	// Create docker client
	apiClient, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		return err
	}
	defer apiClient.Close()
	// Sync
	wg := sync.WaitGroup{}
	wg.Add(len(containers))
	returned := make(chan error)
	defer close(returned)
	// Restart containers
	for _, containerID := range containers {
		go func(containerID string) {
			defer wg.Done()
			err := RestartContainer(containerID)
			if err != nil {
				returned <- err
			}
		}(containerID)
	}
	// Resync
	wg.Wait()
	if len(returned) != 0 {
		var toReturn error
		for err := range returned {
			toReturn = fmt.Errorf("%s\n%s", toReturn, err)
		}
		return toReturn
	}
	return nil
}

// RestartContainer restarts a container given its ID
func RestartContainer(containerID string) error {
	// Create docker client
	apiClient, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		return err
	}
	defer apiClient.Close()
	// Restart container
	return apiClient.ContainerRestart(context.Background(), containerID, container.StopOptions{})
}
