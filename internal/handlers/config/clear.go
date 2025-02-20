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
	"stamus-ctl/internal/stamus"
)

func Clear(conf string) error {
	// File instance
	if !app.IsCtl() {
		conf = app.GetConfigsFolder(conf)
	}
	// Get networks
	log.Println("Getting networks")
	networks, err := getNetworks(conf)
	log.Println("Got networks", networks)
	if err != nil {
		return err
	}
	// Down containers
	log.Println("Handling down")
	err = wrapper.HandleDown(conf, true, true)
	log.Println("Handled down with err", err)
	if err != nil {
		return err
	}
	// Clear networks
	log.Println("Clearing networks", networks)
	err = clearNetworks(networks)
	log.Println("Cleared networks with err", err)
	if err != nil {
		return err
	}
	// Delete folder
	log.Println("Deleting folder", conf)
	err = deleteFolder(conf)
	log.Println("Deleted folder with err", err)
	if err != nil {
		return err
	}
	// Save stamus
	log.Println("Removing instance", conf)
	err = stamus.RemoveInstance(conf)
	log.Println("Removed instance with err", err)
	return err
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
