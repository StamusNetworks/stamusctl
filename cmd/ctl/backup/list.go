package backup

import (
	"fmt"
	"os"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/spf13/cobra"

	flags "stamus-ctl/internal/handlers"
	handlers "stamus-ctl/internal/handlers/backup"
)

// listCmd returns the backup list command
func listCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List all backups for a configuration",
		Long: `List all backups for a configuration.

Displays all available backups for the specified configuration, sorted
by date (newest first). Shows timestamp, type (manual or auto), and size.

Examples:
  # List backups for the default config
  stamusctl backup list

  # List backups for a specific config
  stamusctl backup list -c myconfig
`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return listHandler()
		},
	}

	// Add flags
	flags.Config.AddAsFlag(cmd, false)

	return cmd
}

func listHandler() error {
	// Get config flag value
	conf, err := flags.Config.GetValue()
	if err != nil {
		return err
	}

	// Call the handler
	backups, err := handlers.HandleList(handlers.ListHandlerInputs{
		Config: conf.(string),
	})
	if err != nil {
		return err
	}

	// Display results
	if len(backups) == 0 {
		fmt.Println("No backups found")
		return nil
	}

	fmt.Printf("Found %d backup(s) for config '%s':\n\n", len(backups), conf.(string))

	// Prepare table data
	rows := []table.Row{}
	for _, backup := range backups {
		sizeStr := formatSize(backup.Size)
		rows = append(rows, table.Row{backup.Timestamp, backup.Type, sizeStr})
	}

	// Create and configure table
	t := table.NewWriter()
	t.SetStyle(table.StyleRounded)
	t.SetOutputMirror(os.Stdout)
	t.AppendHeader(table.Row{"Timestamp", "Type", "Size"})
	t.AppendRows(rows)
	t.Render()

	return nil
}

// formatSize formats size in bytes to human-readable format
func formatSize(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}
