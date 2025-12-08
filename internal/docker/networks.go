package docker

import (
	"errors"

	"stamus-ctl/internal/logging"

	"github.com/docker/docker/api/types/network"
)

func GetNetworkIdByName(name string) (string, error) {
	var networks []network.Summary
	err := WithRetrySimple(func() error {
		var listErr error
		networks, listErr = cli.NetworkList(ctx, network.ListOptions{})
		return listErr
	}, "list-networks")
	if err != nil {
		return "", err
	}

	for _, network := range networks {
		if network.Name == name {
			logging.Sugar.Debugw("network found", "network.ID", network.ID,
				"network.Name", network.Name, "name", name)
			return network.ID, nil
		}
	}

	logging.Sugar.Debugw("network not found", "networks", networks, "name", name)
	return "", errors.New("network not found")
}
