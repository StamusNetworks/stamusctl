package backup

import (
	"fmt"

	"stamus-ctl/internal/backup"
	"stamus-ctl/internal/logging"

	"go.uber.org/zap"
)

type DeleteHandlerInputs struct {
	Config    string // Config name
	Timestamp string // Backup timestamp to delete
}

// HandleDelete deletes a specific backup
func HandleDelete(params DeleteHandlerInputs) error {
	logger := logging.Logger.With(
		zap.String("config", params.Config),
		zap.String("timestamp", params.Timestamp),
	)

	logger.Info("Deleting backup")

	// Delete the backup
	err := backup.DeleteBackup(params.Config, params.Timestamp, logging.Logger)
	if err != nil {
		logger.Error("Failed to delete backup", zap.Error(err))
		return fmt.Errorf("failed to delete backup: %w", err)
	}

	logger.Info("Backup deleted successfully")

	return nil
}
