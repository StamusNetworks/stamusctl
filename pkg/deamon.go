package pkg

import "github.com/docker/docker/api/types"

// Common
type Config struct {
	Value string `json:"config"` // Config name, default is config
}
type ErrorResponse struct {
	Error string `json:"error"`
}
type SuccessResponse struct {
	Message string `json:"message"`
}
type Containers struct {
	Containers []string `json:"containers"`
}

// Init
type InitRequest struct {
	Config     string            `json:"config"`      // Config name, default is config
	IsDefault  bool              `json:"default"`     // Use default settings, default is false
	Project    string            `json:"project"`     // Project name, default is "clearndr"
	Values     map[string]string `json:"values"`      // Values to set, key is the name of the value, value is the value
	Version    string            `json:"version"`     // Target version, default is latest
	ValuesPath string            `json:"values_path"` // Path to a values.yaml file
	FromFile   map[string]string `json:"from_file"`   // Values keys and paths to files containing the content used as value
	Registry   string            `json:"registry"`    // Registry to use, default is docker.io/library
}

// Config
type SetRequest struct {
	Reload     bool              `json:"reload"`      // Reload the configuration, don't keep arbitrary parameters
	Values     map[string]string `json:"values"`      // Values to set, key is the name of the value, value is the value
	Apply      bool              `json:"apply"`       // Apply the new configuration, relaunch it, default is false
	ValuesPath string            `json:"values_path"` // Path to a values.yaml file
	FromFile   map[string]string `json:"from_file"`   // Values keys and paths to files containing the content used as value
	Config     string            `json:"config"`      // Config name, default is config
}
type GetRequest struct {
	Values  []string `json:"values"`  // Values to retrieve, default is all
	Content bool     `json:"content"` // Get content or values, default is false
	Config  string   `json:"config"`  // Config name, default is config
}
type GetListResponse struct {
	Configs []string `json:"configs"` // List of available configurations on the system
}

// Update
type UpdateRequest struct {
	Version string            `json:"version"` // Version to update to, default is latest
	Values  map[string]string `json:"values"`  // Values to set, key is the name of the value, value is the value
}

// Log
type LogsRequest struct {
	Containers []string `json:"containers"` // Containers ids to show logs from, default is all
	Timestamps bool     `json:"timestamps"` // Show timestamps, default is false
	Tail       string   `json:"tail"`       // Number of lines to show from the end, default is all
	Since      string   `json:"since"`      // Show logs since (e.g. 2013-01-02T13:23:37Z) or relative (e.g. 42m for 42 minutes)
	Until      string   `json:"until"`      // Show logs until(e.g. 2013-01-02T13:23:37Z) or relative (e.g. 42m for 42 minutes)
}
type ContainerLogs struct {
	types.Container
	Logs []string `json:"logs"` // Logs from the container
}
type LogsResponse struct {
	Containers []ContainerLogs `json:"containers"`
}

// Wrapper
type RestartRequest struct {
	Containers []string `json:"containers"` // Container ID to restart,
}
type PsResponse struct {
	Containers []types.Container `json:"containers"`
}
