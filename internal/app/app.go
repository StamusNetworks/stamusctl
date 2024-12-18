package app

import (
	// Common
	"os"
	"path/filepath"
	"runtime"
	"runtime/debug"

	// External
	"github.com/adrg/xdg"
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
)

// Constants
const (
	binaryNameEnv = "STAMUSCTL_NAME"
	CtlName       = "stamusctl"
)

func CatchException() {
	if err := recover(); err != nil {
		switch err.(type) {
		case *runtime.Error:
			debug.PrintStack()
			panic(err)
		default:
		}
	}
}

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
		Embed.set(val)
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

}

func GetConfigsFolder(name string) string {
	return filepath.Join(ConfigsFolder, name)
}

func IsCtl() bool {
	return CtlName == "stamusctl"
}
