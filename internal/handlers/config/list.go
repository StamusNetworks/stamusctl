package config

import (
	"fmt"
	"io"
	"os"

	"stamus-ctl/internal/stamus"

	"github.com/jedib0t/go-pretty/v6/table"
)

const (
	Red    = "\033[31m"
	Green  = "\033[32m"
	Yellow = "\033[33m"
	Reset  = "\033[0m"
)

// getInstances is a mockable function for testing
var getInstances = stamus.GetInstances

// outputWriter is the destination for table output (mockable for testing)
var outputWriter io.Writer = os.Stdout

// formatStatus returns a colored status string with container counts
func formatStatus(infos stamus.Infos) string {
	cs := infos.Containers
	countStr := ""
	if cs.Total > 0 {
		countStr = fmt.Sprintf(" (%d/%d)", cs.Running, cs.Total)
	}

	switch infos.Status {
	case stamus.StatusUp:
		return Green + "up" + countStr + Reset
	case stamus.StatusPartial:
		return Yellow + "partial" + countStr + Reset
	case stamus.StatusUnhealthy:
		return Red + "unhealthy" + countStr + Reset
	default:
		return "down" + countStr
	}
}

func ListHandler() error {
	instances, err := getInstances()
	if err != nil {
		return err
	}
	// Prepare data
	rows := []table.Row{}
	for folder, infos := range instances {
		rows = append(rows, table.Row{folder, infos.Project, infos.Version, formatStatus(infos)})
	}
	// Print
	t := table.NewWriter()
	t.SetStyle(table.StyleRounded)
	header := table.Row{"Location", "Project", "Version", "Status"}
	t.SetOutputMirror(outputWriter)
	t.AppendHeader(header)
	t.AppendRows(rows)
	t.AppendFooter(header)
	t.Render()
	return nil
}
