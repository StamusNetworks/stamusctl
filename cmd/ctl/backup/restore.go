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

var restoreForce bool

// restoreCmd returns the backup restore command
func restoreCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "restore [timestamp]",
		Short: "Restore a configuration from a backup",
		Long: `Restore a configuration from a backup.

Restores the specified configuration from a backup timestamp. A safety
backup of the current state is created before restoring.

Examples:
  # Restore from a specific backup (interactive confirmation)
  stamusctl backup restore 20240115_143022

  # Restore without confirmation
  stamusctl backup restore -f 20240115_143022

  # Restore a specific config from backup
  stamusctl backup restore -c myconfig 20240115_143022

  # List available backups first
  stamusctl backup list
`,
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: completion.CompleteBackupsFunc(),
		RunE: func(cmd *cobra.Command, args []string) error {
			return restoreHandler(args[0])
		},
	}

	// Add flags
	flags.Config.AddAsFlag(cmd, false)
	cmd.Flags().BoolVarP(&restoreForce, "force", "f", false, "Skip confirmation prompt")

	return cmd
}

func restoreHandler(timestamp string) error {
	// Get config flag value
	conf, err := flags.Config.GetValue()
	if err != nil {
		return err
	}

	configName := conf.(string)

	// Prompt for confirmation unless --force is used
	if !restoreForce {
		fmt.Printf("WARNING: This will replace the current configuration '%s' with the backup from %s.\n", configName, timestamp)
		fmt.Print("A safety backup of the current state will be created first.\n")
		fmt.Print("Do you want to continue? (yes/no): ")

		reader := bufio.NewReader(os.Stdin)
		response, err := reader.ReadString('\n')
		if err != nil {
			return fmt.Errorf("failed to read confirmation: %w", err)
		}

		response = strings.TrimSpace(strings.ToLower(response))
		if response != "yes" && response != "y" {
			fmt.Println("Restore cancelled")
			return nil
		}
	}

	// Call the handler
	err = handlers.HandleRestore(handlers.RestoreHandlerInputs{
		Config:    configName,
		Timestamp: timestamp,
	})
	if err != nil {
		return err
	}

	// Print success message
	fmt.Printf("Configuration '%s' restored successfully from backup %s\n", configName, timestamp)
	return nil
}
