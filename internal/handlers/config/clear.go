package config

import (
	// Core
	"bytes"
	"fmt"
	"log"
	"os/exec"
	"slices"
	"strings"

	// Internal
	"stamus-ctl/internal/app"
	docker "stamus-ctl/internal/docker-compose"
	"stamus-ctl/internal/handlers/wrapper"
)

func Clear(conf string) error {
	// File instance
	if !app.IsCtl() {
		conf = app.GetConfigsFolder(conf)
	}
	// Get networks
	networks, err := getNetworks(conf)
	if err != nil {
		return err
	}
	log.Println("networks", networks)
	// Down containers
	err = wrapper.HandleDown(conf, true, true)
	if err != nil {
		return err
	}
	// Clear networks
	err = clearNetworks(networks)
	if err != nil {
		return err
	}
	// Delete folder
	err = deleteFolder(conf)
	if err != nil {
		return err
	}
	return nil
}

func getNetworks(conf string) ([]string, error) {
	var outBuf, errBuf bytes.Buffer
	dockerCmpsFile := docker.GetComposeFilePath(conf)
	externalCmd := exec.Command("docker", "compose", "-f", dockerCmpsFile, "ps", "--format", "{{.Networks}}")
	externalCmd.Stdout = &outBuf
	externalCmd.Stderr = &errBuf

	err := externalCmd.Run()
	if err != nil {
		fmt.Println("Error:", err)
		fmt.Println("Stderr:", errBuf.String())
		return nil, err
	}
	// Make a list
	networks := strings.Split(outBuf.String(), "\n")
	// Remove empty strings, duplicates
	toReturn := []string{}
	for _, network := range networks {
		if network != "" && !slices.Contains(toReturn, network) {
			toReturn = append(toReturn, network)
		}
	}
	return toReturn, nil
}

func clearNetworks(networks []string) error {
	for _, network := range networks {
		externalCmd := exec.Command("docker", "network", "rm", network)
		err := externalCmd.Run()
		if err != nil {
			fmt.Println("Error:", err)
			return err
		}
	}
	return nil
}

func deleteFolder(conf string) error {
	externalCmd := exec.Command("rm", "-rf", conf)
	err := externalCmd.Run()
	if err != nil {
		fmt.Println("Error:", err)
		return err
	}
	return nil
}
