package backup

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"stamus-ctl/internal/app"
	"stamus-ctl/internal/validation"

	cp "github.com/otiai10/copy"
	"go.uber.org/zap"
)

// BackupType represents the type of backup
type BackupType string

const (
	// BackupTypeAuto is for automatic backups before destructive operations
	BackupTypeAuto BackupType = "auto"
	// BackupTypeManual is for user-initiated backups
	BackupTypeManual BackupType = "manual"
)

// BackupInfo contains metadata about a backup
type BackupInfo struct {
	ConfigName string    // Configuration name
	Timestamp  string    // Backup timestamp (YYYYMMDD_HHMMSS)
	Type       string    // "auto" or "manual"
	Path       string    // Full path to backup
	Size       int64     // Backup size in bytes
	CreatedAt  time.Time // Creation time
}

// getBackupRootPath returns the root directory for all backups
func getBackupRootPath() string {
	return filepath.Join(app.ConfigFolder, "backups")
}

// getConfigBackupPath returns the backup directory for a specific config
func getConfigBackupPath(configName string) string {
	return filepath.Join(getBackupRootPath(), configName)
}

// generateBackupName generates a timestamped backup directory name
func generateBackupName(backupType BackupType) string {
	timestamp := time.Now().Format("20060102_150405")
	return fmt.Sprintf("%s_%s", timestamp, string(backupType))
}

// parseBackupName parses a backup directory name into timestamp and type
func parseBackupName(name string) (timestamp string, backupType string, err error) {
	parts := strings.Split(name, "_")
	if len(parts) < 3 {
		return "", "", fmt.Errorf("invalid backup name format: %s", name)
	}

	// Reconstruct timestamp (YYYYMMDD_HHMMSS)
	timestamp = parts[0] + "_" + parts[1]
	backupType = parts[2]

	if backupType != string(BackupTypeAuto) && backupType != string(BackupTypeManual) {
		return "", "", fmt.Errorf("invalid backup type: %s", backupType)
	}

	return timestamp, backupType, nil
}

// CreateBackup creates a backup of a configuration
func CreateBackup(configName string, backupType BackupType, logger *zap.Logger) (string, error) {
	// Validate config name
	if err := validation.ValidateProjectName(configName); err != nil {
		return "", fmt.Errorf("invalid config name: %w", err)
	}

	// Get source config path
	srcPath := app.GetConfigsFolder(configName)

	// Check if config exists
	if _, err := os.Stat(srcPath); os.IsNotExist(err) {
		return "", fmt.Errorf("configuration does not exist: %s", configName)
	}

	// Generate backup name
	backupName := generateBackupName(backupType)

	// Create backup directory path
	backupDir := getConfigBackupPath(configName)
	backupPath := filepath.Join(backupDir, backupName)

	// Create backup directory if it doesn't exist
	if err := os.MkdirAll(backupDir, 0o700); err != nil {
		return "", fmt.Errorf("failed to create backup directory: %w", err)
	}

	// Log backup creation
	if logger != nil {
		logger.Info("Creating backup",
			zap.String("config", configName),
			zap.String("type", string(backupType)),
			zap.String("path", backupPath),
		)
	}

	// Copy configuration directory to backup location
	if err := cp.Copy(srcPath, backupPath); err != nil {
		return "", fmt.Errorf("failed to copy configuration: %w", err)
	}

	// Log success
	if logger != nil {
		logger.Info("Backup created successfully",
			zap.String("config", configName),
			zap.String("path", backupPath),
		)
	}

	return backupPath, nil
}

// ListBackups returns a list of all backups for a configuration
func ListBackups(configName string) ([]BackupInfo, error) {
	// Validate config name
	if err := validation.ValidateProjectName(configName); err != nil {
		return nil, fmt.Errorf("invalid config name: %w", err)
	}

	backupDir := getConfigBackupPath(configName)

	// Check if backup directory exists
	if _, err := os.Stat(backupDir); os.IsNotExist(err) {
		return []BackupInfo{}, nil // Return empty list if no backups exist
	}

	// Read backup directory
	entries, err := os.ReadDir(backupDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read backup directory: %w", err)
	}

	backups := []BackupInfo{}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		// Parse backup name
		timestamp, backupType, err := parseBackupName(entry.Name())
		if err != nil {
			// Skip invalid backup names
			continue
		}

		backupPath := filepath.Join(backupDir, entry.Name())

		// Get directory info
		info, err := entry.Info()
		if err != nil {
			continue
		}

		// Calculate backup size
		size, err := getDirectorySize(backupPath)
		if err != nil {
			size = 0 // Use 0 if size calculation fails
		}

		// Parse timestamp to time.Time
		createdAt, err := time.Parse("20060102_150405", timestamp)
		if err != nil {
			createdAt = info.ModTime() // Fallback to modification time
		}

		backups = append(backups, BackupInfo{
			ConfigName: configName,
			Timestamp:  timestamp,
			Type:       backupType,
			Path:       backupPath,
			Size:       size,
			CreatedAt:  createdAt,
		})
	}

	// Sort by timestamp (newest first)
	sort.Slice(backups, func(i, j int) bool {
		return backups[i].CreatedAt.After(backups[j].CreatedAt)
	})

	return backups, nil
}

// RestoreBackup restores a configuration from a backup
func RestoreBackup(configName string, timestamp string, logger *zap.Logger) error {
	// Validate config name
	if err := validation.ValidateProjectName(configName); err != nil {
		return fmt.Errorf("invalid config name: %w", err)
	}

	// Find the backup
	backups, err := ListBackups(configName)
	if err != nil {
		return fmt.Errorf("failed to list backups: %w", err)
	}

	var targetBackup *BackupInfo
	for i := range backups {
		if backups[i].Timestamp == timestamp {
			targetBackup = &backups[i]
			break
		}
	}

	if targetBackup == nil {
		return fmt.Errorf("backup not found: %s", timestamp)
	}

	// Validate backup path
	_, err = validation.SanitizePath(targetBackup.Path, getBackupRootPath())
	if err != nil {
		return fmt.Errorf("invalid backup path: %w", err)
	}

	// Check if backup exists
	if _, err := os.Stat(targetBackup.Path); os.IsNotExist(err) {
		return fmt.Errorf("backup directory does not exist: %s", targetBackup.Path)
	}

	// Get target config path
	configPath := app.GetConfigsFolder(configName)

	// Create temporary backup of current config (safety measure)
	tempBackupName := fmt.Sprintf("temp_%s", time.Now().Format("20060102_150405"))
	tempBackupPath := filepath.Join(getConfigBackupPath(configName), tempBackupName)

	currentExists := false
	if _, err := os.Stat(configPath); err == nil {
		currentExists = true

		if logger != nil {
			logger.Info("Creating temporary backup of current configuration",
				zap.String("path", tempBackupPath),
			)
		}

		// Create temporary backup
		if err := cp.Copy(configPath, tempBackupPath); err != nil {
			return fmt.Errorf("failed to create temporary backup: %w", err)
		}

		// Remove current config
		if err := os.RemoveAll(configPath); err != nil {
			return fmt.Errorf("failed to remove current configuration: %w", err)
		}
	}

	// Log restore operation
	if logger != nil {
		logger.Info("Restoring backup",
			zap.String("config", configName),
			zap.String("timestamp", timestamp),
			zap.String("from", targetBackup.Path),
			zap.String("to", configPath),
		)
	}

	// Restore backup
	if err := cp.Copy(targetBackup.Path, configPath); err != nil {
		// Attempt to restore from temporary backup on failure
		if currentExists {
			if logger != nil {
				logger.Warn("Restore failed, attempting to restore previous configuration",
					zap.Error(err),
				)
			}

			if restoreErr := cp.Copy(tempBackupPath, configPath); restoreErr != nil {
				return fmt.Errorf("restore failed and unable to recover: %w (original error: %v)", restoreErr, err)
			}
		}
		return fmt.Errorf("failed to restore backup: %w", err)
	}

	// Clean up temporary backup
	if currentExists {
		if err := os.RemoveAll(tempBackupPath); err != nil {
			if logger != nil {
				logger.Warn("Failed to clean up temporary backup",
					zap.String("path", tempBackupPath),
					zap.Error(err),
				)
			}
		}
	}

	// Log success
	if logger != nil {
		logger.Info("Backup restored successfully",
			zap.String("config", configName),
			zap.String("timestamp", timestamp),
		)
	}

	return nil
}

// DeleteBackup deletes a specific backup
func DeleteBackup(configName string, timestamp string, logger *zap.Logger) error {
	// Validate config name
	if err := validation.ValidateProjectName(configName); err != nil {
		return fmt.Errorf("invalid config name: %w", err)
	}

	// Find the backup
	backups, err := ListBackups(configName)
	if err != nil {
		return fmt.Errorf("failed to list backups: %w", err)
	}

	var targetBackup *BackupInfo
	for i := range backups {
		if backups[i].Timestamp == timestamp {
			targetBackup = &backups[i]
			break
		}
	}

	if targetBackup == nil {
		return fmt.Errorf("backup not found: %s", timestamp)
	}

	// Validate backup path (security check)
	_, err = validation.SanitizePath(targetBackup.Path, getBackupRootPath())
	if err != nil {
		return fmt.Errorf("invalid backup path: %w", err)
	}

	// Log deletion
	if logger != nil {
		logger.Info("Deleting backup",
			zap.String("config", configName),
			zap.String("timestamp", timestamp),
			zap.String("path", targetBackup.Path),
		)
	}

	// Delete backup directory
	if err := os.RemoveAll(targetBackup.Path); err != nil {
		return fmt.Errorf("failed to delete backup: %w", err)
	}

	// Log success
	if logger != nil {
		logger.Info("Backup deleted successfully",
			zap.String("config", configName),
			zap.String("timestamp", timestamp),
		)
	}

	return nil
}

// getDirectorySize calculates the total size of a directory
func getDirectorySize(path string) (int64, error) {
	var size int64

	err := filepath.Walk(path, func(_ string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			size += info.Size()
		}
		return nil
	})

	return size, err
}

// GetBackupPath returns the backup root directory path
func GetBackupPath() string {
	return getBackupRootPath()
}

// ValidateBackup validates the integrity of a backup
func ValidateBackup(backupPath string) error {
	// Check if backup directory exists
	info, err := os.Stat(backupPath)
	if os.IsNotExist(err) {
		return fmt.Errorf("backup does not exist: %s", backupPath)
	}
	if err != nil {
		return fmt.Errorf("failed to stat backup: %w", err)
	}

	// Check if it's a directory
	if !info.IsDir() {
		return fmt.Errorf("backup path is not a directory: %s", backupPath)
	}

	// Validate backup path is within backup root
	_, err = validation.SanitizePath(backupPath, getBackupRootPath())
	if err != nil {
		return fmt.Errorf("backup path validation failed: %w", err)
	}

	return nil
}
