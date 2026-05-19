package main

import (
	"fmt"
	"os"
	"path/filepath"

	"stamus-ctl/cmd/ctl"

	"github.com/spf13/cobra/doc"
)

func main() {
	// Get output directory from args or use default
	outputDir := "./docs/man"
	if len(os.Args) > 1 {
		outputDir = os.Args[1]
	}

	// Create output directory if it doesn't exist
	if err := os.MkdirAll(outputDir, 0o755); err != nil {
		fmt.Fprintf(os.Stderr, "Error creating output directory: %v\n", err)
		os.Exit(1)
	}

	// Get the root command
	cmd := ctl.RootCmdForDoc()

	// Generate man pages
	header := &doc.GenManHeader{
		Title:   "STAMUSCTL",
		Section: "1",
		Source:  "Stamus Networks",
		Manual:  "Stamus Networks Control Tool",
	}

	if err := doc.GenManTree(cmd, header, outputDir); err != nil {
		fmt.Fprintf(os.Stderr, "Error generating man pages: %v\n", err)
		os.Exit(1)
	}

	// Count generated files
	files, _ := filepath.Glob(filepath.Join(outputDir, "*.1"))
	fmt.Printf("Generated %d man page(s) in %s\n", len(files), outputDir)
}
