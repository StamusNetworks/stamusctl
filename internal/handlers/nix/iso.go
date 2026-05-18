package handlers

import (
	"fmt"

	"stamus-ctl/internal/app"
	"stamus-ctl/internal/logging"
	"stamus-ctl/internal/nix"
	"stamus-ctl/internal/validation"

	"github.com/spf13/afero"
)

type NixISOHandlerInputs struct {
	Config string
	Output string
}

func NixISOHandler(params NixISOHandlerInputs) error {
	logger := logging.Sugar.With("Config", params.Config, "Output", params.Output)

	configPath := params.Config
	if !app.IsCtl() {
		configPath = app.GetConfigsFolder(params.Config)
	}

	exists, err := afero.DirExists(app.FS, configPath)
	if err != nil || !exists {
		return fmt.Errorf("configuration %q not found — run 'stamusctl nix init' first", params.Config)
	}

	sanitizedOutput, err := validation.SanitizePath(params.Output, "")
	if err != nil {
		return fmt.Errorf("invalid output path: %w", err)
	}

	if err := nix.BuildISO(configPath, sanitizedOutput); err != nil {
		return err
	}

	logger.Infof("ISO image generated at %s/result", sanitizedOutput)
	return nil
}
