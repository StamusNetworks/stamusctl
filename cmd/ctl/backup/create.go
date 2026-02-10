package backup

import (
	"fmt"

	"github.com/spf13/cobra"

	flags "stamus-ctl/internal/handlers"
	handlers "stamus-ctl/internal/handlers/backup"
)

// createCmd returns the backup create command
func createCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a manual backup of a configuration",
		Long: `Create a manual backup of a configuration.

Creates a timestamped backup of the specified configuration. Manual
backups are stored alongside automatic backups and can be restored
or deleted at any time.

Examples:
  # Create a backup of the default config
  stamusctl backup create

  # Create a backup of a specific config
  stamusctl backup create -c myconfig
`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return createHandler()
		},
	}

	// Add flags
	flags.Config.AddAsFlag(cmd, false)

	return cmd
}

func createHandler() error {
	// Get config flag value
	conf, err := flags.Config.GetValue()
	if err != nil {
		return err
	}

	// Call the handler
	backupPath, err := handlers.HandleCreate(handlers.CreateHandlerInputs{
		Config: conf.(string),
	})
	if err != nil {
		return err
	}

	// Print success message
	fmt.Printf("Backup created successfully: %s\n", backupPath)
	return nil
}
