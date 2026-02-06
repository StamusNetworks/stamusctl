package models

import (
	"log"
	"os/user"
	"path/filepath"
	"strings"

	"stamus-ctl/internal/app"
	"github.com/spf13/afero"
)

// normalizeVersion strips beta/trunk/development suffixes from version strings
// Examples: "1.1.0-trunk.1" -> "1.1.0", "1.2.3-beta.2" -> "1.2.3", "1.0.0" -> "1.0.0"
func normalizeVersion(version string) string {
	// Split on common beta/development indicators
	suffixes := []string{"-trunk", "-beta", "-alpha", "-rc", "-dev", "-snapshot"}

	normalizedVersion := version
	for _, suffix := range suffixes {
		if idx := strings.Index(normalizedVersion, suffix); idx != -1 {
			normalizedVersion = normalizedVersion[:idx]
			break
		}
	}

	return normalizedVersion
}

type Release struct {
	Name      string // the name given to the release
	User      string // userid
	Group     string // groupid
	Location  string // path to the install (eg: /home/.../config, the place where the compose.yaml file will be placed)
	IsUpgrade bool   // see helm
	IsInstall bool   // see helm
	Service   string
	Seed      string
	Version   string // the version tag used in init command
}

func NewRelease(name, location, seed, version string, isUpgrade, isInstall bool) *Release {
	currentUser, err := user.Current()
	if err != nil {
		log.Fatal(err)
	}

	return &Release{
		User:      currentUser.Uid,
		Group:     currentUser.Gid,
		Name:      name,
		Location:  location,
		IsUpgrade: isUpgrade,
		IsInstall: isInstall,
		Seed:      seed,
		Version:   version,
		Service:   app.StamusAppName + ":" + normalizeVersion(app.Version),
	}
}

func getRelease(dest *File, currentDir, seed, version string, isUpgrade, isInstall bool) *Release {
	configDir := dest.Path
	if app.IsCtl() {
		configDir = filepath.Join(currentDir, dest.Path)
	}
	splitted := strings.Split(configDir, "/")
	releaseName := ""
	if len(splitted) == 0 || (len(splitted) == 1 && splitted[0] == "") {
		releaseName = "release"
	} else {
		if splitted[len(splitted)-1] == "" {
			releaseName = splitted[len(splitted)-2]
		} else {
			releaseName = splitted[len(splitted)-1]
		}
	}
	return NewRelease(releaseName, configDir, seed, version, isUpgrade, isInstall)
}

func (s *Release) AsMap() map[string]interface{} {
	prefix := "Release"
	return map[string]interface{}{
		prefix + ".name":      s.Name,
		prefix + ".user":      s.User,
		prefix + ".group":     s.Group,
		prefix + ".location":  s.Location,
		prefix + ".isUpgrade": s.IsUpgrade,
		prefix + ".isInstall": s.IsInstall,
		prefix + ".service":   s.Service,
		prefix + ".seed":      s.Seed,
		prefix + ".version":   s.Version,
	}
}

func (s *Release) SetName(name string) *Release {
	s.Name = name
	return s
}

func (s *Release) SetLocation(location string) *Release {
	s.Location = location
	return s
}

func (s *Release) SetIsUpgrade(isUpgrade bool) *Release {
	s.IsUpgrade = isUpgrade
	return s
}

func (s *Release) SetIsInstall(isInstall bool) *Release {
	s.IsInstall = isInstall
	return s
}

func (s *Release) SetService(service string) *Release {
	s.Service = service
	return s
}

type Template struct {
	templateName    string
	templateVersion string
}

func NewTemplate(name string, templatePath string) *Template {
	// Try to read version from /data/version file in the template
	versionFromFile := ""
	versionFilePath := filepath.Join(templatePath, "version")
	if versionData, err := afero.ReadFile(app.FS, versionFilePath); err == nil {
		versionFromFile = strings.TrimSpace(string(versionData))
	}

	// Fallback to extracting version from path if version file doesn't exist or is empty
	templateVersion := versionFromFile
	if templateVersion == "" {
		splitted := strings.Split(templatePath, "/")
		templateVersion = splitted[len(splitted)-1]
	}

	return &Template{
		templateName:    name,
		templateVersion: templateVersion,
	}
}

func (t *Template) AsMap() map[string]interface{} {
	prefix := "Template"
	return map[string]interface{}{
		prefix + ".name":    t.templateName,
		prefix + ".version": t.templateVersion,
	}
}
