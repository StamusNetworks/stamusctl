package backup

import (
	"github.com/spf13/cobra"
)

// BackupCmd returns the backup command with all subcommands
func BackupCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "backup",
		Short: "Manage configuration backups",
		Long: `Manage configuration backups.

The backup command allows you to create, list, restore, and delete
backups of your configurations. Backups are stored with timestamps
and can be either manual (user-created) or automatic (created by
the system before destructive operations).

Examples:
  # Create a manual backup
  stamusctl backup create

  # List all backups
  stamusctl backup list

  # Restore from a backup
  stamusctl backup restore 20240115_143022

  # Delete an old backup
  stamusctl backup delete 20240115_143022
`,
	}

	// Add subcommands
	cmd.AddCommand(createCmd())
	cmd.AddCommand(listCmd())
	cmd.AddCommand(restoreCmd())
	cmd.AddCommand(deleteCmd())

	return cmd
}
