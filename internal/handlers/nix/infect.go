package handlers

import (
	"fmt"

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
		return fmt.Errorf("this system is already running NixOS — use 'stamusctl nix switch' instead")
	}

	fmt.Println(`WARNING: This operation will convert your system to NixOS.
This is a destructive, one-way operation that replaces the current OS.
It is recommended for fresh VMs or cloud instances only.

This feature is not yet implemented.`)

	return nix.Infect(params.Config)
}
