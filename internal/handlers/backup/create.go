package backup

import (
	"fmt"

	"stamus-ctl/internal/backup"
	"stamus-ctl/internal/logging"

	"go.uber.org/zap"
)

type CreateHandlerInputs struct {
	Config string // Config name to backup
}

// HandleCreate creates a manual backup of a configuration
func HandleCreate(params CreateHandlerInputs) (string, error) {
	logger := logging.Logger.With(
		zap.String("config", params.Config),
		zap.String("type", "manual"),
	)

	logger.Info("Creating manual backup")

	// Create the backup
	backupPath, err := backup.CreateBackup(params.Config, backup.BackupTypeManual, logging.Logger)
	if err != nil {
		logger.Error("Failed to create backup", zap.Error(err))
		return "", fmt.Errorf("failed to create backup: %w", err)
	}

	logger.Info("Backup created successfully",
		zap.String("backup_path", backupPath),
	)

	return backupPath, nil
}
