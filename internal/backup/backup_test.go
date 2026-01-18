package backup

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"stamus-ctl/internal/app"
	"stamus-ctl/internal/logging"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)
 
func TestParseBackupName(t *testing.T) {
	tests := []struct {
		name          string
		backupName    string
		wantTimestamp string
		wantType      string
		wantErr       bool
	}{
		{
			name:          "valid auto backup",
			backupName:    "20240115_143022_auto",
			wantTimestamp: "20240115_143022",
			wantType:      "auto",
			wantErr:       false,
		},
		{
			name:          "valid manual backup",
			backupName:    "20240115_143022_manual",
			wantTimestamp: "20240115_143022",
			wantType:      "manual",
			wantErr:       false,
		},
		{
			name:       "invalid format - too few parts",
			backupName: "20240115_143022",
			wantErr:    true,
		},
		{
			name:       "invalid format - only one part",
			backupName: "invalid",
			wantErr:    true,
		},
		{
			name:       "invalid backup type",
			backupName: "20240115_143022_invalid",
			wantErr:    true,
		},
		{
			name:       "empty string",
			backupName: "",
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			timestamp, backupType, err := parseBackupName(tt.backupName)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantTimestamp, timestamp)
				assert.Equal(t, tt.wantType, backupType)
			}
		})
	}
}

func TestGenerateBackupName(t *testing.T) {
	tests := []struct {
		name       string
		backupType BackupType
	}{
		{
			name:       "auto backup",
			backupType: BackupTypeAuto,
		},
		{
			name:       "manual backup",
			backupType: BackupTypeManual,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			name := generateBackupName(tt.backupType)

			// Verify format
			assert.True(t, strings.HasSuffix(name, "_"+string(tt.backupType)))

			// Verify it can be parsed back
			timestamp, backupType, err := parseBackupName(name)
			assert.NoError(t, err)
			assert.Equal(t, string(tt.backupType), backupType)
			assert.NotEmpty(t, timestamp)
		})
	}
}

func TestGetBackupPath(t *testing.T) {
	path := GetBackupPath()
	assert.NotEmpty(t, path)
	assert.True(t, strings.HasSuffix(path, "backups"))
}

func setupTestFS(t *testing.T) (cleanup func()) {
	// Save original filesystem and config folder
	oldFS := app.FS
	oldConfigFolder := app.ConfigFolder

	// Create temporary directory for testing
	tmpDir, err := os.MkdirTemp("", "backup-test-*")
	require.NoError(t, err)

	// Set up test filesystem
	app.FS = afero.NewOsFs()
	app.ConfigFolder = tmpDir

	// Initialize logger if not already done
	if logging.Logger == nil {
		logging.SetLogger()
	}

	return func() {
		app.FS = oldFS
		app.ConfigFolder = oldConfigFolder
		os.RemoveAll(tmpDir)
	}
}

func createTestConfig(t *testing.T, configName string) string {
	configPath := app.GetConfigsFolder(configName)
	err := os.MkdirAll(configPath, 0700)
	require.NoError(t, err)

	// Create a test file in the config
	testFile := filepath.Join(configPath, "test.txt")
	err = os.WriteFile(testFile, []byte("test content"), 0600)
	require.NoError(t, err)

	return configPath
}

func TestCreateBackup(t *testing.T) {
	tests := []struct {
		name        string
		configName  string
		backupType  BackupType
		setup       func(t *testing.T)
		expectError bool
		errorMsg    string
	}{
		{
			name:       "valid manual backup",
			configName: "test-config",
			backupType: BackupTypeManual,
			setup: func(t *testing.T) {
				createTestConfig(t, "test-config")
			},
			expectError: false,
		},
		{
			name:       "valid auto backup",
			configName: "test-config-auto",
			backupType: BackupTypeAuto,
			setup: func(t *testing.T) {
				createTestConfig(t, "test-config-auto")
			},
			expectError: false,
		},
		{
			name:        "invalid config name",
			configName:  "../invalid",
			backupType:  BackupTypeManual,
			setup:       func(t *testing.T) {},
			expectError: true,
			errorMsg:    "invalid config name",
		},
		{
			name:        "non-existent config",
			configName:  "nonexistent",
			backupType:  BackupTypeManual,
			setup:       func(t *testing.T) {},
			expectError: true,
			errorMsg:    "configuration does not exist",
		},
		{
			name:       "empty config name",
			configName: "",
			backupType: BackupTypeManual,
			setup:      func(t *testing.T) {},
			expectError: true,
			errorMsg:    "invalid config name",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cleanup := setupTestFS(t)
			defer cleanup()

			tt.setup(t)

			backupPath, err := CreateBackup(tt.configName, tt.backupType, logging.Logger)

			if tt.expectError {
				assert.Error(t, err)
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg)
				}
			} else {
				assert.NoError(t, err)
				assert.NotEmpty(t, backupPath)

				// Verify backup was created
				_, err := os.Stat(backupPath)
				assert.NoError(t, err)

				// Verify backup contains the test file
				testFile := filepath.Join(backupPath, "test.txt")
				content, err := os.ReadFile(testFile)
				assert.NoError(t, err)
				assert.Equal(t, "test content", string(content))
			}
		})
	}
}

func TestListBackups(t *testing.T) {
	tests := []struct {
		name        string
		configName  string
		setup       func(t *testing.T)
		wantCount   int
		expectError bool
		errorMsg    string
	}{
		{
			name:       "no backups",
			configName: "test-config",
			setup:      func(t *testing.T) {},
			wantCount:  0,
			expectError: false,
		},
		{
			name:       "single backup",
			configName: "test-config-single",
			setup: func(t *testing.T) {
				createTestConfig(t, "test-config-single")
				_, err := CreateBackup("test-config-single", BackupTypeManual, logging.Logger)
				require.NoError(t, err)
			},
			wantCount:   1,
			expectError: false,
		},
		{
			name:       "multiple backups",
			configName: "test-config-multi",
			setup: func(t *testing.T) {
				createTestConfig(t, "test-config-multi")

				// Create multiple backups with 1 second delay to ensure different timestamps
				for i := 0; i < 3; i++ {
					_, err := CreateBackup("test-config-multi", BackupTypeManual, logging.Logger)
					require.NoError(t, err)
					if i < 2 {
						time.Sleep(1 * time.Second)
					}
				}
			},
			wantCount:   3,
			expectError: false,
		},
		{
			name:       "mixed backup types",
			configName: "test-config-mixed",
			setup: func(t *testing.T) {
				createTestConfig(t, "test-config-mixed")
				_, err := CreateBackup("test-config-mixed", BackupTypeAuto, logging.Logger)
				require.NoError(t, err)
				time.Sleep(1 * time.Second)
				_, err = CreateBackup("test-config-mixed", BackupTypeManual, logging.Logger)
				require.NoError(t, err)
			},
			wantCount:   2,
			expectError: false,
		},
		{
			name:       "invalid backup names ignored",
			configName: "test-config-invalid",
			setup: func(t *testing.T) {
				createTestConfig(t, "test-config-invalid")
				_, err := CreateBackup("test-config-invalid", BackupTypeManual, logging.Logger)
				require.NoError(t, err)

				// Create invalid backup directory
				backupDir := getConfigBackupPath("test-config-invalid")
				invalidDir := filepath.Join(backupDir, "invalid_backup")
				err = os.MkdirAll(invalidDir, 0700)
				require.NoError(t, err)
			},
			wantCount:   1, // Only the valid backup should be counted
			expectError: false,
		},
		{
			name:        "invalid config name",
			configName:  "../invalid",
			setup:       func(t *testing.T) {},
			wantCount:   0,
			expectError: true,
			errorMsg:    "invalid config name",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cleanup := setupTestFS(t)
			defer cleanup()

			tt.setup(t)

			backups, err := ListBackups(tt.configName)

			if tt.expectError {
				assert.Error(t, err)
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg)
				}
			} else {
				assert.NoError(t, err)
				assert.Len(t, backups, tt.wantCount)

				// Verify backups are sorted by timestamp (newest first)
				if len(backups) > 1 {
					for i := 1; i < len(backups); i++ {
						assert.True(t, backups[i-1].CreatedAt.After(backups[i].CreatedAt) ||
							backups[i-1].CreatedAt.Equal(backups[i].CreatedAt))
					}
				}

				// Verify backup info fields are populated
				for _, backup := range backups {
					assert.Equal(t, tt.configName, backup.ConfigName)
					assert.NotEmpty(t, backup.Timestamp)
					assert.NotEmpty(t, backup.Type)
					assert.NotEmpty(t, backup.Path)
					assert.True(t, backup.Size >= 0)
					assert.False(t, backup.CreatedAt.IsZero())
				}
			}
		})
	}
}

func TestRestoreBackup(t *testing.T) {
	tests := []struct {
		name        string
		configName  string
		setup       func(t *testing.T) string // Returns timestamp to restore
		expectError bool
		errorMsg    string
	}{
		{
			name:       "restore valid backup",
			configName: "test-restore",
			setup: func(t *testing.T) string {
				// Create initial config
				configPath := createTestConfig(t, "test-restore")

				// Create backup
				_, err := CreateBackup("test-restore", BackupTypeManual, logging.Logger)
				require.NoError(t, err)

				// Modify the config
				testFile := filepath.Join(configPath, "test.txt")
				err = os.WriteFile(testFile, []byte("modified content"), 0600)
				require.NoError(t, err)

				// Extract timestamp from backup path
				backups, err := ListBackups("test-restore")
				require.NoError(t, err)
				require.Len(t, backups, 1)

				return backups[0].Timestamp
			},
			expectError: false,
		},
		{
			name:       "restore to empty config location",
			configName: "test-restore-empty",
			setup: func(t *testing.T) string {
				// Create config and backup
				createTestConfig(t, "test-restore-empty")
				_, err := CreateBackup("test-restore-empty", BackupTypeManual, logging.Logger)
				require.NoError(t, err)

				// Remove the config
				configPath := app.GetConfigsFolder("test-restore-empty")
				err = os.RemoveAll(configPath)
				require.NoError(t, err)

				// Get timestamp
				backups, err := ListBackups("test-restore-empty")
				require.NoError(t, err)
				require.Len(t, backups, 1)

				return backups[0].Timestamp
			},
			expectError: false,
		},
		{
			name:       "invalid config name",
			configName: "../invalid",
			setup: func(t *testing.T) string {
				return "20240115_143022"
			},
			expectError: true,
			errorMsg:    "invalid config name",
		},
		{
			name:       "non-existent backup",
			configName: "test-restore-missing",
			setup: func(t *testing.T) string {
				createTestConfig(t, "test-restore-missing")
				return "20240115_143022" // Non-existent timestamp
			},
			expectError: true,
			errorMsg:    "backup not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cleanup := setupTestFS(t)
			defer cleanup()

			timestamp := tt.setup(t)

			err := RestoreBackup(tt.configName, timestamp, logging.Logger)

			if tt.expectError {
				assert.Error(t, err)
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg)
				}
			} else {
				assert.NoError(t, err)

				// Verify config was restored
				configPath := app.GetConfigsFolder(tt.configName)
				testFile := filepath.Join(configPath, "test.txt")
				content, err := os.ReadFile(testFile)
				assert.NoError(t, err)
				assert.Equal(t, "test content", string(content))
			}
		})
	}
}

func TestDeleteBackup(t *testing.T) {
	tests := []struct {
		name        string
		configName  string
		setup       func(t *testing.T) string // Returns timestamp to delete
		expectError bool
		errorMsg    string
	}{
		{
			name:       "delete valid backup",
			configName: "test-delete",
			setup: func(t *testing.T) string {
				createTestConfig(t, "test-delete")
				_, err := CreateBackup("test-delete", BackupTypeManual, logging.Logger)
				require.NoError(t, err)

				backups, err := ListBackups("test-delete")
				require.NoError(t, err)
				require.Len(t, backups, 1)

				return backups[0].Timestamp
			},
			expectError: false,
		},
		{
			name:       "delete one of multiple backups",
			configName: "test-delete-multi",
			setup: func(t *testing.T) string {
				createTestConfig(t, "test-delete-multi")

				// Create multiple backups with 1 second delay to ensure different timestamps
				for i := 0; i < 3; i++ {
					_, err := CreateBackup("test-delete-multi", BackupTypeManual, logging.Logger)
					require.NoError(t, err)
					if i < 2 {
						time.Sleep(1 * time.Second)
					}
				}

				backups, err := ListBackups("test-delete-multi")
				require.NoError(t, err)
				require.Len(t, backups, 3)

				// Return middle backup timestamp
				return backups[1].Timestamp
			},
			expectError: false,
		},
		{
			name:       "invalid config name",
			configName: "../invalid",
			setup: func(t *testing.T) string {
				return "20240115_143022"
			},
			expectError: true,
			errorMsg:    "invalid config name",
		},
		{
			name:       "non-existent backup",
			configName: "test-delete-missing",
			setup: func(t *testing.T) string {
				createTestConfig(t, "test-delete-missing")
				return "20240115_143022"
			},
			expectError: true,
			errorMsg:    "backup not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cleanup := setupTestFS(t)
			defer cleanup()

			timestamp := tt.setup(t)

			// Get backup count before deletion
			var beforeCount int
			if !tt.expectError && tt.configName != "../invalid" {
				backups, err := ListBackups(tt.configName)
				require.NoError(t, err)
				beforeCount = len(backups)
			}

			err := DeleteBackup(tt.configName, timestamp, logging.Logger)

			if tt.expectError {
				assert.Error(t, err)
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg)
				}
			} else {
				assert.NoError(t, err)

				// Verify backup was deleted
				backups, err := ListBackups(tt.configName)
				assert.NoError(t, err)
				assert.Equal(t, beforeCount-1, len(backups))

				// Verify deleted backup is not in list
				for _, backup := range backups {
					assert.NotEqual(t, timestamp, backup.Timestamp)
				}
			}
		})
	}
}

func TestGetDirectorySize(t *testing.T) {
	cleanup := setupTestFS(t)
	defer cleanup()

	// Create test directory with files
	testDir := filepath.Join(app.ConfigFolder, "size-test")
	err := os.MkdirAll(testDir, 0700)
	require.NoError(t, err)

	// Create files of known sizes
	file1 := filepath.Join(testDir, "file1.txt")
	err = os.WriteFile(file1, []byte("12345"), 0600) // 5 bytes
	require.NoError(t, err)

	file2 := filepath.Join(testDir, "file2.txt")
	err = os.WriteFile(file2, []byte("1234567890"), 0600) // 10 bytes
	require.NoError(t, err)

	// Create subdirectory with file
	subDir := filepath.Join(testDir, "subdir")
	err = os.MkdirAll(subDir, 0700)
	require.NoError(t, err)

	file3 := filepath.Join(subDir, "file3.txt")
	err = os.WriteFile(file3, []byte("123"), 0600) // 3 bytes
	require.NoError(t, err)

	size, err := getDirectorySize(testDir)
	assert.NoError(t, err)
	assert.Equal(t, int64(18), size) // 5 + 10 + 3 = 18 bytes
}

func TestValidateBackup(t *testing.T) {
	cleanup := setupTestFS(t)
	defer cleanup()

	tests := []struct {
		name        string
		setup       func(t *testing.T) string // Returns path to validate
		expectError bool
		errorMsg    string
	}{
		{
			name: "valid backup",
			setup: func(t *testing.T) string {
				createTestConfig(t, "test-validate")
				backupPath, err := CreateBackup("test-validate", BackupTypeManual, logging.Logger)
				require.NoError(t, err)
				return backupPath
			},
			expectError: false,
		},
		{
			name: "non-existent backup",
			setup: func(t *testing.T) string {
				return filepath.Join(app.ConfigFolder, "backups", "test", "nonexistent")
			},
			expectError: true,
			errorMsg:    "backup does not exist",
		},
		{
			name: "backup is a file not directory",
			setup: func(t *testing.T) string {
				backupRoot := filepath.Join(app.ConfigFolder, "backups")
				err := os.MkdirAll(backupRoot, 0700)
				require.NoError(t, err)

				filePath := filepath.Join(backupRoot, "file.txt")
				err = os.WriteFile(filePath, []byte("test"), 0600)
				require.NoError(t, err)

				return filePath
			},
			expectError: true,
			errorMsg:    "not a directory",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			backupPath := tt.setup(t)

			err := ValidateBackup(backupPath)

			if tt.expectError {
				assert.Error(t, err)
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestCreateBackupWithNilLogger(t *testing.T) {
	cleanup := setupTestFS(t)
	defer cleanup()

	createTestConfig(t, "test-nil-logger")

	// Should not panic with nil logger
	backupPath, err := CreateBackup("test-nil-logger", BackupTypeManual, nil)
	assert.NoError(t, err)
	assert.NotEmpty(t, backupPath)
}

func TestRestoreBackupWithNilLogger(t *testing.T) {
	cleanup := setupTestFS(t)
	defer cleanup()

	createTestConfig(t, "test-restore-nil-logger")
	_, err := CreateBackup("test-restore-nil-logger", BackupTypeManual, nil)
	require.NoError(t, err)

	backups, err := ListBackups("test-restore-nil-logger")
	require.NoError(t, err)
	require.Len(t, backups, 1)

	// Should not panic with nil logger
	err = RestoreBackup("test-restore-nil-logger", backups[0].Timestamp, nil)
	assert.NoError(t, err)
}

func TestDeleteBackupWithNilLogger(t *testing.T) {
	cleanup := setupTestFS(t)
	defer cleanup()

	createTestConfig(t, "test-delete-nil-logger")
	_, err := CreateBackup("test-delete-nil-logger", BackupTypeManual, nil)
	require.NoError(t, err)

	backups, err := ListBackups("test-delete-nil-logger")
	require.NoError(t, err)
	require.Len(t, backups, 1)

	// Should not panic with nil logger
	err = DeleteBackup("test-delete-nil-logger", backups[0].Timestamp, nil)
	assert.NoError(t, err)
}
