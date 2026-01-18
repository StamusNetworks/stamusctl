package backup

import (
	"os"
	"testing"
	"time"

	"stamus-ctl/internal/app"
	"stamus-ctl/internal/backup"
	"stamus-ctl/internal/logging"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTestFS(t *testing.T) (cleanup func()) {
	// Save original filesystem and config folder
	oldFS := app.FS
	oldConfigFolder := app.ConfigFolder

	// Create temporary directory for testing
	tmpDir, err := os.MkdirTemp("", "handler-test-*")
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
	testFile := configPath + "/test.txt"
	err = os.WriteFile(testFile, []byte("test content"), 0600)
	require.NoError(t, err)

	return configPath
}

func TestHandleCreate(t *testing.T) {
	tests := []struct {
		name        string
		config      string
		setup       func(t *testing.T)
		expectError bool
		errorMsg    string
	}{
		{
			name:   "create valid backup",
			config: "test-config",
			setup: func(t *testing.T) {
				createTestConfig(t, "test-config")
			},
			expectError: false,
		},
		{
			name:   "invalid config name",
			config: "../invalid",
			setup:  func(t *testing.T) {},
			expectError: true,
			errorMsg:    "invalid config name",
		},
		{
			name:   "non-existent config",
			config: "nonexistent",
			setup:  func(t *testing.T) {},
			expectError: true,
			errorMsg:    "configuration does not exist",
		},
		{
			name:   "empty config name",
			config: "",
			setup:  func(t *testing.T) {},
			expectError: true,
			errorMsg:    "invalid config name",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cleanup := setupTestFS(t)
			defer cleanup()

			tt.setup(t)

			backupPath, err := HandleCreate(CreateHandlerInputs{
				Config: tt.config,
			})

			if tt.expectError {
				assert.Error(t, err)
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg)
				}
				assert.Empty(t, backupPath)
			} else {
				assert.NoError(t, err)
				assert.NotEmpty(t, backupPath)

				// Verify backup was created
				_, err := os.Stat(backupPath)
				assert.NoError(t, err)
			}
		})
	}
}

func TestHandleList(t *testing.T) {
	tests := []struct {
		name        string
		config      string
		setup       func(t *testing.T)
		wantCount   int
		expectError bool
		errorMsg    string
	}{
		{
			name:   "no backups",
			config: "test-config",
			setup:  func(t *testing.T) {},
			wantCount: 0,
			expectError: false,
		},
		{
			name:   "single backup",
			config: "test-config-single",
			setup: func(t *testing.T) {
				createTestConfig(t, "test-config-single")
				_, err := backup.CreateBackup("test-config-single", backup.BackupTypeManual, logging.Logger)
				require.NoError(t, err)
			},
			wantCount:   1,
			expectError: false,
		},
		{
			name:   "multiple backups",
			config: "test-config-multi",
			setup: func(t *testing.T) {
				createTestConfig(t, "test-config-multi")
				for i := 0; i < 2; i++ {
					_, err := backup.CreateBackup("test-config-multi", backup.BackupTypeManual, logging.Logger)
					require.NoError(t, err)
					if i < 1 {
						time.Sleep(1 * time.Second)
					}
				}
			},
			wantCount:   2,
			expectError: false,
		},
		{
			name:   "invalid config name",
			config: "../invalid",
			setup:  func(t *testing.T) {},
			expectError: true,
			errorMsg:    "invalid config name",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cleanup := setupTestFS(t)
			defer cleanup()

			tt.setup(t)

			backups, err := HandleList(ListHandlerInputs{
				Config: tt.config,
			})

			if tt.expectError {
				assert.Error(t, err)
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg)
				}
			} else {
				assert.NoError(t, err)
				assert.Len(t, backups, tt.wantCount)
			}
		})
	}
}

func TestHandleRestore(t *testing.T) {
	tests := []struct {
		name        string
		config      string
		setup       func(t *testing.T) string // Returns timestamp to restore
		expectError bool
		errorMsg    string
	}{
		{
			name:   "restore valid backup",
			config: "test-restore",
			setup: func(t *testing.T) string {
				// Create config and backup
				configPath := createTestConfig(t, "test-restore")
				_, err := backup.CreateBackup("test-restore", backup.BackupTypeManual, logging.Logger)
				require.NoError(t, err)

				// Modify the config
				testFile := configPath + "/test.txt"
				err = os.WriteFile(testFile, []byte("modified"), 0600)
				require.NoError(t, err)

				// Get timestamp
				backups, err := backup.ListBackups("test-restore")
				require.NoError(t, err)
				require.Len(t, backups, 1)

				return backups[0].Timestamp
			},
			expectError: false,
		},
		{
			name:   "invalid config name",
			config: "../invalid",
			setup: func(t *testing.T) string {
				return "20240115_143022"
			},
			expectError: true,
			errorMsg:    "invalid config name",
		},
		{
			name:   "non-existent backup",
			config: "test-restore-missing",
			setup: func(t *testing.T) string {
				createTestConfig(t, "test-restore-missing")
				return "20240115_143022"
			},
			expectError: true,
			errorMsg:    "backup not found",
		},
		{
			name:   "empty timestamp",
			config: "test-restore-empty-ts",
			setup: func(t *testing.T) string {
				createTestConfig(t, "test-restore-empty-ts")
				return ""
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

			err := HandleRestore(RestoreHandlerInputs{
				Config:    tt.config,
				Timestamp: timestamp,
			})

			if tt.expectError {
				assert.Error(t, err)
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg)
				}
			} else {
				assert.NoError(t, err)

				// Verify config was restored
				configPath := app.GetConfigsFolder(tt.config)
				testFile := configPath + "/test.txt"
				content, err := os.ReadFile(testFile)
				assert.NoError(t, err)
				assert.Equal(t, "test content", string(content))
			}
		})
	}
}

func TestHandleDelete(t *testing.T) {
	tests := []struct {
		name        string
		config      string
		setup       func(t *testing.T) string // Returns timestamp to delete
		expectError bool
		errorMsg    string
	}{
		{
			name:   "delete valid backup",
			config: "test-delete",
			setup: func(t *testing.T) string {
				createTestConfig(t, "test-delete")
				_, err := backup.CreateBackup("test-delete", backup.BackupTypeManual, logging.Logger)
				require.NoError(t, err)

				backups, err := backup.ListBackups("test-delete")
				require.NoError(t, err)
				require.Len(t, backups, 1)

				return backups[0].Timestamp
			},
			expectError: false,
		},
		{
			name:   "delete one of multiple backups",
			config: "test-delete-multi",
			setup: func(t *testing.T) string {
				createTestConfig(t, "test-delete-multi")

				// Create two backups
				_, err := backup.CreateBackup("test-delete-multi", backup.BackupTypeManual, logging.Logger)
				require.NoError(t, err)
				time.Sleep(1 * time.Second)
				_, err = backup.CreateBackup("test-delete-multi", backup.BackupTypeManual, logging.Logger)
				require.NoError(t, err)

				backups, err := backup.ListBackups("test-delete-multi")
				require.NoError(t, err)
				require.Len(t, backups, 2)

				// Return first backup timestamp
				return backups[0].Timestamp
			},
			expectError: false,
		},
		{
			name:   "invalid config name",
			config: "../invalid",
			setup: func(t *testing.T) string {
				return "20240115_143022"
			},
			expectError: true,
			errorMsg:    "invalid config name",
		},
		{
			name:   "non-existent backup",
			config: "test-delete-missing",
			setup: func(t *testing.T) string {
				createTestConfig(t, "test-delete-missing")
				return "20240115_143022"
			},
			expectError: true,
			errorMsg:    "backup not found",
		},
		{
			name:   "empty timestamp",
			config: "test-delete-empty-ts",
			setup: func(t *testing.T) string {
				createTestConfig(t, "test-delete-empty-ts")
				_, err := backup.CreateBackup("test-delete-empty-ts", backup.BackupTypeManual, logging.Logger)
				require.NoError(t, err)
				return ""
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

			// Get initial backup count
			var beforeCount int
			if !tt.expectError && tt.config != "../invalid" {
				backups, err := backup.ListBackups(tt.config)
				require.NoError(t, err)
				beforeCount = len(backups)
			}

			err := HandleDelete(DeleteHandlerInputs{
				Config:    tt.config,
				Timestamp: timestamp,
			})

			if tt.expectError {
				assert.Error(t, err)
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg)
				}
			} else {
				assert.NoError(t, err)

				// Verify backup was deleted
				backups, err := backup.ListBackups(tt.config)
				assert.NoError(t, err)
				assert.Equal(t, beforeCount-1, len(backups))

				// Verify deleted backup is not in list
				for _, b := range backups {
					assert.NotEqual(t, timestamp, b.Timestamp)
				}
			}
		})
	}
}
