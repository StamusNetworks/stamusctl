package backup

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"stamus-ctl/internal/completion"
	flags "stamus-ctl/internal/handlers"
	handlers "stamus-ctl/internal/handlers/backup"
)

var (
	deleteForce bool
)

// deleteCmd returns the backup delete command
func deleteCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete [timestamp]",
		Short: "Delete a specific backup",
		Long: `Delete a specific backup permanently.

Permanently removes the backup with the specified timestamp. This action
cannot be undone.

Examples:
  # Delete a backup (interactive confirmation)
  stamusctl backup delete 20240115_143022

  # Delete without confirmation
  stamusctl backup delete -f 20240115_143022

  # Delete a backup for a specific config
  stamusctl backup delete -c myconfig 20240115_143022

  # List available backups first
  stamusctl backup list
`,
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: completion.CompleteBackupsFunc(),
		RunE: func(cmd *cobra.Command, args []string) error {
			return deleteHandler(args[0])
		},
	}

	// Add flags
	flags.Config.AddAsFlag(cmd, false)
	cmd.Flags().BoolVarP(&deleteForce, "force", "f", false, "Skip confirmation prompt")

	return cmd
}

func deleteHandler(timestamp string) error {
	// Get config flag value
	conf, err := flags.Config.GetValue()
	if err != nil {
		return err
	}

	configName := conf.(string)

	// Prompt for confirmation unless --force is used
	if !deleteForce {
		fmt.Printf("WARNING: This will permanently delete the backup '%s' for configuration '%s'.\n", timestamp, configName)
		fmt.Print("This action cannot be undone. Do you want to continue? (yes/no): ")

		reader := bufio.NewReader(os.Stdin)
		response, err := reader.ReadString('\n')
		if err != nil {
			return fmt.Errorf("failed to read confirmation: %w", err)
		}

		response = strings.TrimSpace(strings.ToLower(response))
		if response != "yes" && response != "y" {
			fmt.Println("Delete cancelled")
			return nil
		}
	}

	// Call the handler
	err = handlers.HandleDelete(handlers.DeleteHandlerInputs{
		Config:    configName,
		Timestamp: timestamp,
	})
	if err != nil {
		return err
	}

	// Print success message
	fmt.Printf("Backup '%s' for configuration '%s' deleted successfully\n", timestamp, configName)
	return nil
}
