package stamus

import (
	"stamus-ctl/internal/models"
)

type Registry string
type User string
type Token string

type Registries map[Registry]Logins
type Logins map[User]Token


func SaveLogin(registryInfo models.RegistryInfo) error {
	// Get config content
	Config, err := GetStamusConfig()
	if err != nil {
		return err
	}

	// Save in struct
	Config.SetRegistry(
		Registry(registryInfo.Registry),
		User(registryInfo.Username),
		Token(registryInfo.Password),
	)

	// Save config
	Config.setStamusConfig()

	return nil
}

func (c *Config) SetRegistry(registry Registry, user User, token Token) {
	// Create Registries if it does not exist in Config
	if c.Registries == nil {
		c.Registries = make(Registries)
	}
	// Create Registry if it does not exist in Registries
	if c.Registries[Registry(registry)] == nil {
		c.Registries[Registry(registry)] = make(Logins)
	}

	c.Registries[Registry(registry)][User(user)] = Token(token)
}