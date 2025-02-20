package config

import (
	// Core

	"fmt"
	"os/exec"

	// Internal
	"stamus-ctl/internal/app"
	"stamus-ctl/internal/handlers/wrapper"
	"stamus-ctl/internal/stamus"
)

func Clear(conf string) error {
	// File instance
	if !app.IsCtl() {
		conf = app.GetConfigsFolder(conf)
	}
	// Down containers
	err := wrapper.HandleDown(conf, true, true)
	if err != nil {
		return err
	}
	// Delete folder
	err = deleteFolder(conf)
	if err != nil {
		return err
	}
	// Save stamus
	err = stamus.RemoveInstance(conf)
	return err
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
