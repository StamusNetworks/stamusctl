package backup

import (
	"fmt"

	"stamus-ctl/internal/backup"
	"stamus-ctl/internal/logging"

	"go.uber.org/zap"
)

type ListHandlerInputs struct {
	Config string // Config name to list backups for (optional - empty means all)
}

// HandleList lists all backups for a configuration
func HandleList(params ListHandlerInputs) ([]backup.BackupInfo, error) {
	logger := logging.Logger.With(
		zap.String("config", params.Config),
	)

	logger.Debug("Listing backups")

	// List backups
	backups, err := backup.ListBackups(params.Config)
	if err != nil {
		logger.Error("Failed to list backups", zap.Error(err))
		return nil, fmt.Errorf("failed to list backups: %w", err)
	}

	logger.Info("Successfully listed backups",
		zap.Int("count", len(backups)),
	)

	return backups, nil
}
