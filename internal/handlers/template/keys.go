package config

import (
	"os"
	"stamus-ctl/internal/models"

	"github.com/jedib0t/go-pretty/v6/table"
)

func KeysHandler(templatePath string, isMd bool) error {
	// Get template keys
	confFile, err := models.CreateFile(templatePath, "config.yaml")
	if err != nil {
		return err
	}
	config, err := models.ConfigFromFile(confFile)
	if err != nil {
		return err
	}
	params, _, err := config.ExtractParams()
	if err != nil {
		return err
	}
	// Prepare data
	rows := []table.Row{}
	for _, name := range params.GetOrdered() {
		param := params.Get(name)
		choices := ""
		if param.Choices != nil {
			for nb, choice := range param.Choices {
				choices = choices + choice.AsString()
				if nb != len(param.Choices)-1 {
					choices = choices + ", "
					if nb%4 == 3 {
						choices = choices + "\n"
					}
				}
			}
		}
		rows = append(rows, table.Row{name, param.Type, param.Default.AsString(), choices})
	}
	// Print
	t := table.NewWriter()
	t.SetStyle(table.StyleRounded)
	header := table.Row{"Key", "Type", "Default", "Choices"}
	t.SetOutputMirror(os.Stdout)
	t.AppendHeader(header)
	t.AppendRows(rows)
	t.AppendFooter(header)
	if isMd {
		t.RenderMarkdown()
	} else {
		t.Render()
	}
	return nil
}
