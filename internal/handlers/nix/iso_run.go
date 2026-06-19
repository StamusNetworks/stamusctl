package handlers

import (
	"fmt"
	"path/filepath"

	"stamus-ctl/internal/app"
	"stamus-ctl/internal/logging"
	"stamus-ctl/internal/nix"

	"github.com/spf13/afero"
)

type NixISORUnHandlerInputs struct {
	Config    string
	Output    string
	Memory    int
	Cores     int
	EnableKVM bool
}

func NixISORunHandler(params NixISORUnHandlerInputs) error {
	logger := logging.Sugar.With("Config", params.Config, "Output", params.Output)

	// Find the ISO file under <output>/result/iso/*.iso
	resultDir := filepath.Join(params.Output, "result", "iso")

	exists, err := afero.DirExists(app.FS, resultDir)
	if err != nil || !exists {
		return fmt.Errorf("ISO result directory %q not found — run 'stamusctl nix iso' first", resultDir)
	}

	matches, err := afero.Glob(app.FS, filepath.Join(resultDir, "*.iso"))
	if err != nil || len(matches) == 0 {
		return fmt.Errorf("no .iso file found in %q — run 'stamusctl nix iso' first", resultDir)
	}

	isoPath := matches[0]
	logger.Infof("booting ISO %s in QEMU", isoPath)

	return nix.RunISO(isoPath, params.Memory, params.Cores, params.EnableKVM)
}
