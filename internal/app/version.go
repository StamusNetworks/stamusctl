package app

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/afero"
)

var (
	Arch    = ""
	Commit  = ""
	Version = "0.1"
)

func init() {
	// Read version file in dev mode
	execPath, _ := os.Executable()
	execName := filepath.Base(execPath)
	if execName == "cmd" {
		content, err := afero.ReadFile(FS, "VERSION")
		if err != nil {
			fmt.Println("Error reading version file:", err)
		} else {
			Version = string(content)
		}
	}
}
