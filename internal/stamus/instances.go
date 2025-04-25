package stamus

import (
	"bytes"
	"os/exec"
	"strings"

	"stamus-ctl/internal/app"
	compose "stamus-ctl/internal/docker-compose"

	"github.com/spf13/afero"
)

type (
	Folder string
	Infos  struct {
		IsUp    bool
		Project string `json:"project"`
		Version string `json:"version"`
	}
)
type Instances map[Folder]Infos

func GetInstances() (Instances, error) {
	// Get config content
	Config, err := GetStamusConfig()
	if err != nil {
		return nil, err
	}
	// Get instances infos
	var instancesInfos Instances = make(Instances)
	for folder, infos := range Config.Instances {
		file := compose.GetComposeFilePath(string(folder))
		// File exists
		exists, _ := afero.Exists(app.FS, file)
		if !exists {
			RemoveInstance(file)
			continue
		}
		// Check if up
		var outBuf, errBuf bytes.Buffer
		cmd := exec.Command("docker", "compose", "-f", file, "ps")
		cmd.Stdout = &outBuf
		cmd.Stderr = &errBuf
		err := cmd.Run()
		if err != nil {
			return nil, err
		}
		// Prepare data
		if strings.Contains(outBuf.String(), "Up") {
			instancesInfos[folder] = Infos{
				Project: infos.Project,
				Version: infos.Version,
				IsUp:    true,
			}
		} else {
			instancesInfos[folder] = Infos{
				Project: infos.Project,
				Version: infos.Version,
				IsUp:    false,
			}
		}
	}
	return instancesInfos, nil
}

func AddInstance(folder string, project string, version string) error {
	// Get config content
	Config, err := GetStamusConfig()
	if err != nil {
		return err
	}
	// Modify
	if Config.Instances == nil {
		Config.Instances = make(Instances)
	}
	Config.Instances[Folder(folder)] = Infos{
		Project: project,
		Version: version,
	}
	// Save config
	return Config.setStamusConfig()
}

func RemoveInstance(folder string) error {
	// Get config content
	Config, err := GetStamusConfig()
	if err != nil {
		return err
	}
	// Modify
	if Config.Instances == nil {
		return nil
	}
	delete(Config.Instances, Folder(folder))
	// Save config
	return Config.setStamusConfig()
}

func removeString(slice []Folder, s Folder) []Folder {
	for i, v := range slice {
		if v == s {
			// Remove the element by appending slice before and after the found element
			return append(slice[:i], slice[i+1:]...)
		}
	}
	return slice
}
