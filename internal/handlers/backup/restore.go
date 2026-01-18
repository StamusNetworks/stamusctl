package backup

import (
	"fmt"

	"stamus-ctl/internal/backup"
	"stamus-ctl/internal/logging"

	"go.uber.org/zap"
)

type RestoreHandlerInputs struct {
	Config    string // Config name to restore
	Timestamp string // Backup timestamp to restore from
}

// HandleRestore restores a configuration from a backup
func HandleRestore(params RestoreHandlerInputs) error {
	logger := logging.Logger.With(
		zap.String("config", params.Config),
		zap.String("timestamp", params.Timestamp),
	)

	logger.Info("Restoring backup")

	// Restore the backup
	err := backup.RestoreBackup(params.Config, params.Timestamp, logging.Logger)
	if err != nil {
		logger.Error("Failed to restore backup", zap.Error(err))
		return fmt.Errorf("failed to restore backup: %w", err)
	}

	logger.Info("Backup restored successfully")

	return nil
}
