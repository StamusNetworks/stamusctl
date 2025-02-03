package app

import (
	// Common

	"log"
	"os"
	"path/filepath"
	"strings"

	// External
	"github.com/adrg/xdg"
	"github.com/spf13/afero"
)

// Variables
var (
	Name                = ""
	Mode                = ModeStruct("prod")
	Embed               = EmbedStruct("false")
	ConfigFolder        = "/"
	ConfigsFolder       = "/"
	TemplatesFolder     = "/"
	DefaultClearNDRPath = "/"
	LatestClearNDRPath  = "/"
	DefaultConfigName   = "config"
	StamusAppName       = ""
	DefaultRegistry     = "ghcr.io/stamusnetworks/stamusctl-templates"
	FS                  = afero.NewOsFs()
)

// Constants
const (
	binaryNameEnv = "STAMUSCTL_NAME"
	CtlName       = "stamusctl"
)

func init() {
	// Binary name
	Name = filepath.Base(os.Args[0])
	if val := os.Getenv(binaryNameEnv); val != "" {
		Name = val
	}

	// Mode
	if val := os.Getenv("BUILD_MODE"); val != "" {
		Mode.set(val)
	}
	if val := os.Getenv("EMBED_MODE"); val != "" {
		Embed.Set(val)
	}
	if val := os.Getenv("STAMUS_APP_NAME"); val != "" {
		StamusAppName = val
	}

	// Folders
	if val := os.Getenv("STAMUS_CONFIG_FOLDER"); val != "" {
		ConfigFolder = val
	} else {
		ConfigFolder = xdg.ConfigHome + "/stamus/"
	}
	if val := os.Getenv("STAMUS_TEMPLATES_FOLDER"); val != "" {
		TemplatesFolder = val
	} else {
		TemplatesFolder = xdg.UserDirs.Templates + "/stamus/"
	}

	// Derived paths
	DefaultClearNDRPath = TemplatesFolder + "clearndr/embedded/"
	LatestClearNDRPath = TemplatesFolder + "clearndr/latest/"
	ConfigsFolder = ConfigFolder + "configs/"

	// Test mode
	if isUnderTest() && !isMemFSDisabled() {
		FS = afero.NewMemMapFs()
		log.Println("Using in-memory filesystem for tests")
	}
}

func GetConfigsFolder(name string) string {
	return filepath.Join(ConfigsFolder, name)
}

func IsCtl() bool {
	return CtlName == "stamusctl"
}
func isUnderTest() bool {
	for _, arg := range os.Args {
		if strings.HasPrefix(arg, "-test.") {
			return true
		}
	}
	return false
}
func isMemFSDisabled() bool {
	return os.Getenv("DISABLE_MEM_FS") == "true"
}
