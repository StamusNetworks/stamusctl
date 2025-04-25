package config

import (
	"fmt"

	flags "stamus-ctl/internal/handlers"
	tmpl "stamus-ctl/internal/handlers/template"

	"github.com/spf13/cobra"
)

// Command
func keysCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "keys",
		Short: "Get list of template keys",
		RunE: func(cmd *cobra.Command, args []string) error {
			return keysHandler()
		},
	}
	// Flags
	flags.Template.AddAsFlag(cmd, false)
	flags.Markdown.AddAsFlag(cmd, false)
	return cmd
}

func keysHandler() error {
	// Validate flags
	value, err := flags.Template.GetValue()
	if err != nil {
		return err
	}
	if value == "" {
		return fmt.Errorf("template folder is required, use --template or -t to set and `config list` to list available templates")
	}
	isMd, err := flags.Markdown.GetValue()
	if err != nil {
		return err
	}
	// Call handler
	return tmpl.KeysHandler(value.(string), isMd.(bool))
}
