package handlers

import (
	"errors"

	"stamus-ctl/internal/logging"
	"stamus-ctl/internal/nix"
)

type NixInfectHandlerInputs struct {
	Config string
}

func NixInfectHandler(params NixInfectHandlerInputs) error {
	logger := logging.Sugar.With("Config", params.Config)
	logger.Debug("infect handler called")

	if nix.IsNixOS() {
		return errors.New("this system is already running NixOS — use 'stamusctl nix switch' instead")
	}

	return nix.Infect(params.Config)
}
