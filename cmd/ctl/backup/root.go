package backup

import (
	"github.com/spf13/cobra"
)

// BackupCmd returns the backup command with all subcommands
func BackupCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "backup",
		Short: "Manage configuration backups",
		Long:  "Create, list, restore, and delete configuration backups",
	}

	// Add subcommands
	cmd.AddCommand(createCmd())
	cmd.AddCommand(listCmd())
	cmd.AddCommand(restoreCmd())
	cmd.AddCommand(deleteCmd())

	return cmd
}
