package config

import (
	"os"
	"stamus-ctl/internal/stamus"

	"github.com/jedib0t/go-pretty/v6/table"
)

const (
	Red   = "\033[31m"
	Green = "\033[32m"
	Reset = "\033[0m"
)

func ListHandler() error {
	instances, err := stamus.GetInstances()
	if err != nil {
		return err
	}
	// Prepare data
	rows := []table.Row{}
	for folder, infos := range instances {
		if infos.IsUp {
			rows = append(rows, table.Row{folder, infos.Project, infos.Version, Green + "up" + Reset})
		} else {
			rows = append(rows, table.Row{folder, infos.Project, infos.Version, "down"})
		}
	}
	// Print
	t := table.NewWriter()
	t.SetStyle(table.StyleRounded)
	header := table.Row{"Location", "Project", "Version", "Status"}
	t.SetOutputMirror(os.Stdout)
	t.AppendHeader(header)
	t.AppendRows(rows)
	t.AppendFooter(header)
	t.Render()
	return nil
}
