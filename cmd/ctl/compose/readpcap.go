package compose

import (
	// Common

	// External

	"errors"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	// Custom
	flags "stamus-ctl/internal/handlers"
	handlers "stamus-ctl/internal/handlers/compose"
	"stamus-ctl/internal/logging"
)

// Commands
func readPcapCmd() *cobra.Command {
	// Create cmd
	cmd := &cobra.Command{
		Use:   "readpcap <pcap file>",
		Short: "Sends a pcap file to be read by suricata",
		RunE: func(cmd *cobra.Command, args []string) error {
			return readPcap(cmd, args)
		},
		Args: cobra.ExactArgs(1),
	}
	// Add flags
	flags.Config.AddAsFlag(cmd, false)
	return cmd
}

func readPcap(cmd *cobra.Command, args []string) error {
	// Validate pcap
	if len(args) < 1 {
		return errors.New("pcap file path is required")
	}
	pcapFile := args[0]
	if err := checkFile(pcapFile, ".pcap"); err != nil {
		return err
	}
	// Get flags
	config, err := flags.Config.GetValue()
	if err != nil {
		return err
	}
	// Call handler
	params := handlers.ReadPcapParams{
		PcapPath: pcapFile,
		Config:   config.(string),
	}
	err = handlers.PcapHandler(params)
	if err != nil {
		logging.Sugar.Error(err)
	}
	return nil
}

// checkFile checks if a file exists and has the specified extension.
func checkFile(filePath, ext string) error {
	// Check if file exists
	info, err := os.Stat(filePath)
	if os.IsNotExist(err) {
		return errors.New("file does not exist")
	}
	if err != nil {
		return err
	}

	// Check if it's a regular file
	if !info.Mode().IsRegular() {
		return errors.New("not a regular file")
	}

	// Check file extension
	if filepath.Ext(filePath) != ext {
		return errors.New("file does not have the correct extension")
	}

	return nil
}
