package stamus

import (
	"bytes"
	"encoding/json"
	"os/exec"
	"strings"
	"unicode"

	"stamus-ctl/internal/app"
	compose "stamus-ctl/internal/docker-compose"
	"stamus-ctl/internal/models"

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
			// Remove instance by its folder, not compose file path
			RemoveInstance(string(folder))
			continue
		}
		// Determine seed and check by docker ps first (match by name containing seed)
		seed := getInstanceSeed(string(folder))
		var outBuf, errBuf bytes.Buffer
		if seed != "" {
			cmd := exec.Command("docker", "ps", "--format", "json", "--filter", "name="+seed)
			cmd.Stdout = &outBuf
			cmd.Stderr = &errBuf
			if err := cmd.Run(); err != nil {
				return nil, err
			}
		}
		// If ps via seed indicates running, mark up. Otherwise, fallback to compose-scoped ps.
		if isComposeUpFromPs(outBuf.String()) {
			instancesInfos[folder] = Infos{
				Project: infos.Project,
				Version: infos.Version,
				IsUp:    true,
			}
		} else {
			// Fallback: docker compose ps scoped to this instance's compose file
			var cOut, cErr bytes.Buffer
			cfile := compose.GetComposeFilePath(string(folder))
			ccmd := exec.Command("docker", "compose", "-f", cfile, "ps", "--format", "json")
			ccmd.Stdout = &cOut
			ccmd.Stderr = &cErr
			if err := ccmd.Run(); err != nil {
				// If compose fails, treat as down
				instancesInfos[folder] = Infos{Project: infos.Project, Version: infos.Version, IsUp: false}
			} else if isComposeUpFromPs(cOut.String()) {
				instancesInfos[folder] = Infos{Project: infos.Project, Version: infos.Version, IsUp: true}
			} else {
				instancesInfos[folder] = Infos{Project: infos.Project, Version: infos.Version, IsUp: false}
			}
		}
	}
	return instancesInfos, nil
}

// isComposeUpFromPs determines if any service is running based on `docker compose ps` output.
// It supports JSON output from Compose V2 and falls back to substring checks.
func isComposeUpFromPs(output string) bool {
	s := strings.TrimSpace(output)
	if s == "" {
		return false
	}
	// Try JSON array from `--format json`
	if strings.HasPrefix(s, "[") {
		var items []map[string]any
		if err := json.Unmarshal([]byte(s), &items); err == nil {
			for _, it := range items {
				// Check common fields
				if state, ok := it["State"].(string); ok {
					if strings.EqualFold(state, "running") || strings.EqualFold(state, "healthy") {
						return true
					}
				}
				if status, ok := it["Status"].(string); ok {
					ls := strings.ToLower(status)
					if strings.Contains(ls, "running") || strings.Contains(ls,
						"healthy") || strings.Contains(ls, "up") {
						return true
					}
				}
			}
			return false
		}
		// If JSON parsing fails, fall back to string checks
	}
	// Try JSON-per-line (one object per line)
	lines := strings.Split(s, "\n")
	parsedAny := false
	for _, line := range lines {
		ln := strings.TrimSpace(line)
		if ln == "" {
			continue
		}
		if strings.HasPrefix(ln, "{") && strings.HasSuffix(ln, "}") {
			var it map[string]any
			if err := json.Unmarshal([]byte(ln), &it); err == nil {
				parsedAny = true
				if state, ok := it["State"].(string); ok {
					if strings.EqualFold(state, "running") || strings.EqualFold(state, "healthy") {
						return true
					}
				}
				if status, ok := it["Status"].(string); ok {
					ls := strings.ToLower(status)
					if strings.Contains(ls, "running") || strings.Contains(ls,
						"healthy") || strings.Contains(ls, "up") {
						return true
					}
				}
			}
		}
	}
	if parsedAny {
		return false
	}
	// Fallback: scan lines and consider a status token equal to "up"/"running"/"healthy"
	for _, line := range lines {
		ln := strings.TrimFunc(line, unicode.IsSpace)
		if ln == "" {
			continue
		}
		lnl := strings.ToLower(ln)
		// Ignore header rows
		if strings.Contains(lnl, "status") && strings.Contains(lnl, "name") {
			continue
		}
		// Tokenize and look for specific status words
		fields := strings.Fields(lnl)
		for _, f := range fields {
			if f == "up" || f == "running" || f == "healthy" {
				return true
			}
		}
	}
	return false
}

// getInstanceSeed reads the instance seed from its values.yaml (stamus.seed)
func getInstanceSeed(folder string) string {
	file, err := models.CreateFile(folder, "values.yaml")
	if err != nil {
		return ""
	}
	conf, err := models.LoadConfigFrom(file, true)
	if err != nil {
		return ""
	}
	return conf.GetSeed()
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
