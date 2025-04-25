package embeds

import (
	"embed"
	"runtime/debug"

	"stamus-ctl/internal/app"
	"stamus-ctl/internal/models"
	"stamus-ctl/internal/utils"
)

//go:embed clearndr/*
var AllConf embed.FS

// Create ClearNDR folder if it does not exist
func InitClearNDRFolder(path string) {
	clearndrConfigExist, err := utils.FolderExists(path)
	if err != nil {
		debug.PrintStack()
		panic(err)
	}
	if !clearndrConfigExist && app.Embed.IsTrue() {
		err = models.ExtractEmbedTo("clearndr", AllConf,
			app.TemplatesFolder+"clearndr/embedded/")
		if err != nil {
			debug.PrintStack()
			panic(err)
		}
	}
}
