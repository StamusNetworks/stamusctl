package handlers

import (
	// Common
	"context"
	"io"
	"strings"

	// Internal
	"stamus-ctl/internal/app"
	"stamus-ctl/pkg"
	"stamus-ctl/pkg/mocker"

	// External
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
)

func HandlePs() ([]types.Container, error) {
	if app.Mode.IsTest() {
		return mocker.Mocked.Ps()
	}
	return handlePs()
}

func handlePs() ([]types.Container, error) {
	// Create docker client
	apiClient, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		return nil, err
	}
	defer apiClient.Close()
	// Get containers
	containers, err := apiClient.ContainerList(context.Background(), container.ListOptions{All: true})
	if err != nil {
		return nil, err
	}
	return containers, nil
}

func HandleLogs(logParams pkg.LogsRequest) (pkg.LogsResponse, error) {
	if app.Mode.IsTest() {
		return mocker.Mocked.Logs()
	}
	return handleLogs(logParams)
}

func handleLogs(logParams pkg.LogsRequest) (pkg.LogsResponse, error) {
	// Get containers
	containers, err := HandlePs()
	if err != nil {
		return pkg.LogsResponse{}, err
	}
	// Filter containers
	if logParams.Containers != nil && len(logParams.Containers) > 0 {
		filteredContainers := []types.Container{}
		for _, cont := range containers {
			for _, nameOrID := range logParams.Containers {
				// Match by ID
				if cont.ID == nameOrID {
					filteredContainers = append(filteredContainers, cont)
					break
				}
				// Match by name (container names have leading slash, e.g. "/nginx")
				for _, name := range cont.Names {
					if name == nameOrID || name == "/"+nameOrID {
						filteredContainers = append(filteredContainers, cont)
						break
					}
				}
			}
		}
		containers = filteredContainers
	}
	// Create docker client
	apiClient, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		return pkg.LogsResponse{}, err
	}
	defer apiClient.Close()
	// Get logs
	response := pkg.LogsResponse{}
	for _, current := range containers {
		reader, err := apiClient.ContainerLogs(context.Background(), current.ID, container.LogsOptions{
			Since:      logParams.Since,
			Until:      logParams.Until,
			Tail:       logParams.Tail,
			Timestamps: logParams.Timestamps,
			ShowStdout: true,
			ShowStderr: true,
			Details:    true,
		})
		if err != nil {
			return pkg.LogsResponse{}, err
		}
		defer reader.Close()
		// Read logs
		buffer, err := io.ReadAll(reader)
		if err != nil {
			return pkg.LogsResponse{}, err
		}
		// Clean logs
		rawLogs := string(buffer)
		rawLogs = strings.ReplaceAll(rawLogs, "\u0000", "")
		rawLogs = strings.ReplaceAll(rawLogs, "\u0001", "")
		rawLogs = strings.ReplaceAll(rawLogs, "\u0002", "")
		logs := strings.Split(rawLogs, "\n")
		// Create response
		response.Containers = append(response.Containers, pkg.ContainerLogs{
			Container: current,
			Logs:      logs,
		})
	}

	return response, nil
}
