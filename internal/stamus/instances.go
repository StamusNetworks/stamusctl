package stamus

import (
	"strings"

	"stamus-ctl/internal/app"
	"stamus-ctl/internal/utils"

	"github.com/docker/docker/api/types"
	"github.com/spf13/afero"
)

// Status represents the operational status of an instance
type Status string

const (
	StatusUp        Status = "up"        // All containers running
	StatusPartial   Status = "partial"   // Some containers running
	StatusDown      Status = "down"      // No containers running
	StatusUnhealthy Status = "unhealthy" // Has unhealthy containers
)

// ContainerStatus holds container count information
type ContainerStatus struct {
	Running   int
	Total     int
	Unhealthy int
}

// calculateStatus determines the status based on container states
func calculateStatus(containers []types.Container) (Status, ContainerStatus) {
	cs := ContainerStatus{Total: len(containers)}

	if cs.Total == 0 {
		return StatusDown, cs
	}

	for _, c := range containers {
		if c.State == "running" {
			cs.Running++
		}
		if strings.Contains(strings.ToLower(c.Status), "unhealthy") {
			cs.Unhealthy++
		}
	}

	switch {
	case cs.Unhealthy > 0:
		return StatusUnhealthy, cs
	case cs.Running == cs.Total:
		return StatusUp, cs
	case cs.Running == 0:
		return StatusDown, cs
	default:
		return StatusPartial, cs
	}
}

type (
	Folder string
	Infos  struct {
		IsUp       bool            // Backward compatibility: true if status is up or partial
		Status     Status          // Detailed status
		Containers ContainerStatus // Container counts
		Project    string          `json:"project"`
		Version    string          `json:"version"`
	}
)
type Instances map[Folder]Infos

func (cm *ConfigManager) GetInstances() (Instances, error) {
	// Get config content
	Config, err := cm.GetConfig()
	if err != nil {
		return nil, err
	}
	// Get instances infos
	var instancesInfos Instances = make(Instances)
	for folder, infos := range Config.Instances {
		file := utils.GetComposeFilePath(string(folder))
		// File exists
		exists, _ := afero.Exists(app.FS, file)
		if !exists {
			cm.RemoveInstance(string(folder))
			continue
		}
		// Get containers by project name using Docker API
		containers, err := cm.containers.GetContainersByProject(infos.Project)
		if err != nil {
			// Graceful degradation: if Docker API fails, mark as down
			instancesInfos[folder] = Infos{
				Project:    infos.Project,
				Version:    infos.Version,
				IsUp:       false,
				Status:     StatusDown,
				Containers: ContainerStatus{},
			}
			continue
		}
		// Calculate status
		status, containerStatus := calculateStatus(containers)
		// IsUp for backward compatibility: true if any containers are running
		isUp := status == StatusUp || status == StatusPartial || status == StatusUnhealthy
		instancesInfos[folder] = Infos{
			Project:    infos.Project,
			Version:    infos.Version,
			IsUp:       isUp,
			Status:     status,
			Containers: containerStatus,
		}
	}
	return instancesInfos, nil
}

func (cm *ConfigManager) AddInstance(folder string, project string, version string) error {
	// Get config content
	Config, err := cm.GetConfig()
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

func (cm *ConfigManager) RemoveInstance(folder string) error {
	// Get config content
	Config, err := cm.GetConfig()
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

func (cm *ConfigManager) GetProjectName(folder string) string {
	config, err := cm.GetConfig()
	if err != nil {
		return ""
	}
	if config.Instances == nil {
		return ""
	}
	if infos, ok := config.Instances[Folder(folder)]; ok {
		return infos.Project
	}
	return ""
}

func removeString(slice []Folder, s Folder) []Folder {
	for i, v := range slice {
		if v == s {
			return append(slice[:i], slice[i+1:]...)
		}
	}
	return slice
}

// Backward-compatible package-level functions

func GetInstances() (Instances, error) {
	return DefaultManager.GetInstances()
}

func AddInstance(folder string, project string, version string) error {
	return DefaultManager.AddInstance(folder, project, version)
}

func RemoveInstance(folder string) error {
	return DefaultManager.RemoveInstance(folder)
}

func GetProjectName(folder string) string {
	return DefaultManager.GetProjectName(folder)
}
