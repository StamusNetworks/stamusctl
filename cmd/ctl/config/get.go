package config

import (
	// Core
	"fmt"

	// External
	"github.com/spf13/cobra"

	// Internal

	flags "stamus-ctl/internal/handlers"
	config "stamus-ctl/internal/handlers/config"
	handlers "stamus-ctl/internal/handlers/config"
)

// Command
func getCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get [keys...]",
		Short: "Get compose config file parameters values",
		Long: `Get compose config file parameters values
Input the keys of parameters to get
Example: get scirius`,
		Args: cobra.ArbitraryArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			getHandler(cmd, args)
			return nil
		},
	}
	// Subcommands
	cmd.AddCommand(getContentCmd())
	// Flags
	flags.Config.AddAsFlag(cmd, false)
	return cmd
}

func versionCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "version",
		Short: "Get config version",
		Args:  cobra.ArbitraryArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			versionHandler()
			return nil
		},
	}
	// Flags
	flags.Config.AddAsFlag(cmd, false)
	return cmd
}

// Handlers
func versionHandler() error {
	// Get properties
	conf, err := flags.Config.GetValue()
	if err != nil {
		return err
	}
	// Get the version
	version := handlers.GetVersion(conf.(string))
	// Print the version
	fmt.Println(version)
	return nil
}

// Subcommands
func getContentCmd() *cobra.Command {
	// Command
	cmd := &cobra.Command{
		Use:   "content",
		Short: "Get config content architecture",
		RunE: func(cmd *cobra.Command, args []string) error {
			getContent(cmd, args)
			return nil
		},
	}
	// Flags
	flags.Config.AddAsFlag(cmd, false)
	return cmd
}

// Handlers
func getHandler(cmd *cobra.Command, args []string) error {
	// Get properties
	reload, err := flags.Reload.GetValue()
	if err != nil {
		return err
	}
	conf, err := flags.Config.GetValue()
	if err != nil {
		return err
	}
	// Load the config values
	groupedValues, err := config.GetGroupedConfig(conf.(string), args, reload.(bool))
	if err != nil {
		return err
	}
	// Print the values
	printGroupedValues(groupedValues, "")
	return nil
}

func getContent(cmd *cobra.Command, args []string) error {
	// Get properties
	conf, err := flags.Config.GetValue()
	if err != nil {
		return err
	}
	// Call handler
	groupedContent, err := config.GetGroupedContent(conf.(string), args)
	if err != nil {
		return err
	}
	// Print the content
	printColoredGroupedValues(groupedContent, "")
	return nil
}

// From the grouped values, print the values in a readable format
func printGroupedValues(group map[string]interface{}, prefix string) {
	for key, value := range group {
		switch v := value.(type) {
		case string:
			fmt.Printf("%s%s: %s\n", prefix, key, v)
		case map[string]interface{}:
			fmt.Printf("%s%s:\n", prefix, key)
			printGroupedValues(v, prefix+"  ")
		}
	}
}

func printColoredGroupedValues(group map[string]interface{}, prefix string) {
	for key, value := range group {
		switch v := value.(type) {
		case string:
			fmt.Printf("\033[2m%s%s\033[0m\n", prefix, key)
		case map[string]interface{}:
			fmt.Printf("\033[2m%s\033[0m\033[1m%s/\033[0m\n", prefix, key)
			printColoredGroupedValues(v, prefix+"|  ")
		}
	}
}
