package ctl

import (
	// Common
	"context"

	// External
	"github.com/docker/docker/api/types/registry"
	"github.com/docker/docker/client"
	"github.com/spf13/cobra"

	// Custom
	flags "stamus-ctl/internal/handlers"
	"stamus-ctl/internal/logging"
	"stamus-ctl/internal/models"
	"stamus-ctl/internal/stamus"
)

var noVerif = models.Parameter{
	Name:    "verif",
	Type:    "bool",
	Usage:   "Verify registry connectivity",
	Default: models.CreateVariableBool(true),
}

func loginCmd() *cobra.Command {
	// Create command
	cmd := &cobra.Command{
		Use:   "login",
		Short: "Login to a registry",
		RunE:  loginHandler,
	}
	// Flags
	flags.Registry.AddAsFlag(cmd, false)
	flags.Username.AddAsFlag(cmd, false)
	flags.Password.AddAsFlag(cmd, false)

	noVerif.AddAsFlag(cmd, false)

	return cmd
}

func loginHandler(cmd *cobra.Command, args []string) error {
	// Extract flags
	registry, err := flags.Registry.GetValue()
	if err != nil {
		return err
	}
	username, err := flags.Username.GetValue()
	if err != nil {
		return err
	}
	password, err := flags.Password.GetValue()
	if err != nil {
		return err
	}

	verif, err := noVerif.GetValue()
	if err != nil {
		return err
	}

	// Call handler
	params := models.RegistryInfo{
		Registry: registry.(string),
		Username: username.(string),
		Password: password.(string),
		Verif:    verif.(bool),
	}
	return LoginHandler(params)
}

func LoginHandler(registryInfo models.RegistryInfo) error {
	// Validate flags
	err := registryInfo.ValidateAllRegistry()
	if err != nil {
		return err
	}

	// Create a Docker client
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		logging.Sugar.Info("Error creating Docker client:", err)
		return err
	}

	// Log in to registry
	authConfig := registry.AuthConfig{
		ServerAddress: registryInfo.Registry,
		Username:      registryInfo.Username,
		Password:      registryInfo.Password,
	}

	if registryInfo.Verif {
		_, err = cli.RegistryLogin(context.Background(), authConfig)
		if err != nil {
			return err
		}
	}

	// Save credentials
	stamus.SaveLogin(registryInfo)

	logging.Sugar.Info("Login successful")

	return nil
}
