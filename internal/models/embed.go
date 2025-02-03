package models

import (
	"embed"
	"log"
	"stamus-ctl/internal/app"
	"strings"

	"github.com/spf13/afero"
)

func ExtractEmbedTo(embedName string, embed embed.FS, outputFolder string) error {
	files := getAllFilesEmbed(embedName, embed)

	for _, file := range files {
		data, err := embed.ReadFile(file)
		if err != nil {
			return err
		}
		err = app.FS.MkdirAll(outputFolder+"/"+extractPath(file), 0755)
		if err != nil {
			return err
		}
		err = afero.WriteFile(app.FS, outputFolder+"/"+extractPath(file)+"/"+extractFileName(file), data, 0644)
		if err != nil {
			return err
		}
	}
	return nil
}

func getAllFilesEmbed(inputFolder string, embed embed.FS) []string {
	var files []string
	entries, err := embed.ReadDir(inputFolder)
	if err != nil {
		log.Println("Error reading dir", inputFolder)
		log.Fatal(err)
	}
	for _, entry := range entries {
		if entry.IsDir() {
			files = append(files, getAllFilesEmbed(inputFolder+"/"+entry.Name(), embed)...)
		} else {
			files = append(files, inputFolder+"/"+entry.Name())
		}
	}
	return files
}

func extractPath(path string) string {
	// returns everything before the last /
	array := strings.Split(path, "/")
	return strings.Join(array[1:len(array)-1], "/")
}

func extractFileName(path string) string {
	// returns everything before the last /
	array := strings.Split(path, "/")
	return strings.Join(array[len(array)-1:], "/")
}
