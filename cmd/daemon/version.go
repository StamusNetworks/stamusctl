package daemon

import (
	"fmt"

	"stamus-ctl/internal/app"

	"github.com/spf13/cobra"
)

func printVersion() {
	fmt.Printf("version: %s\narch: %s\ncommit: %s\n", app.Version, app.Arch, app.Commit)
}

func versionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Version information",
		Run: func(cmd *cobra.Command, args []string) {
			printVersion()
		},
	}
}
