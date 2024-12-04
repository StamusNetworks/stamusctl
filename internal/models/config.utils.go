package models

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"stamus-ctl/internal/logging"
	"strings"
	"text/template"

	"github.com/Masterminds/sprig/v3"
)

func deleteEmptyFiles(folderPath string) error {
	err := filepath.Walk(folderPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		// Check if it's a regular file and empty
		if !info.IsDir() && info.Size() == 0 {
			err := os.Remove(path)
			if err != nil {
				return err
			}
		}
		return nil
	})
	return err
}

func deleteEmptyFolders(folderPath string) error {
	err := filepath.Walk(folderPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		// Check if it's a directory and empty
		if info.IsDir() {
			removeDirIfEmpty(path)
		}
		return nil
	})
	return err
}

func removeDirIfEmpty(path string) error {
	isEmpty, err := isDirEmpty(path)
	if err != nil {
		log.Println("Error checking if directory is empty", err)
		return err
	}
	if isEmpty {
		err := os.Remove(path)
		if err != nil {
			return err
		}
	}
	return nil
}

func isDirEmpty(name string) (bool, error) {
	f, err := os.Open(name)
	if err != nil {
		return false, err
	}
	defer f.Close()

	_, err = f.Readdirnames(1) // Or f.Readdir(1)
	if err == io.EOF {
		return true, nil
	}
	return false, err // Either not empty or error, suits both cases
}

// Get all files in a folder with a specific extension
func getAllFiles(folderPath string, extension string) ([]string, error) {
	var files []string
	err := filepath.Walk(folderPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && filepath.Ext(path) == extension {
			files = append(files, path)
		}
		return nil
	})
	return files, err
}

// Nests a flat map into a nested map
func nestMap(input map[string]interface{}) map[string]interface{} {
	output := make(map[string]interface{})

	for key, value := range input {
		parts := strings.Split(key, ".")
		currentMap := output

		for i, part := range parts {
			if i == len(parts)-1 {
				// Last part, set the value
				currentMap[part] = value
			} else {
				// Intermediate part, ensure the map exists
				if _, ok := currentMap[part]; !ok {
					currentMap[part] = make(map[string]interface{})
				}
				// Move to the next level in the map
				currentMap = currentMap[part].(map[string]interface{})
			}
		}
	}

	return output
}

// Process templates from a folder to another with a data nested map
func processTemplates(inputFolder string, outputFolder string, data map[string]interface{}) error {
	tpls, err := getAllFiles(inputFolder, ".tpl")
	if err != nil {
		return err
	}

	// Walk the source directory and process templates
	err = filepath.Walk(inputFolder, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		rel, err := filepath.Rel(inputFolder, path)
		if err != nil {
			return err
		}

		destPath := filepath.Join(outputFolder, rel)

		if info.IsDir() {
			return os.MkdirAll(destPath, info.Mode())
		}

		tmpl, err := template.New(filepath.Base(path)).Funcs(sprig.FuncMap()).ParseFiles(append([]string{path}, tpls...)...)
		if err != nil {
			logging.Sugar.Info("Error parsing template", path, err)
			return err
		}

		destFile, err := os.Create(destPath)
		if err != nil {
			return err
		}
		defer destFile.Close()

		err = tmpl.Execute(destFile, data)
		if err != nil {
			return err
		}

		if filepath.Ext(path) == ".sh" {
			err = os.Chmod(destPath, 0755)
			if err != nil {
				return err
			}
		}

		return nil
	})
	if err != nil {
		return err
	}
	logging.Sugar.Info("Configuration saved to: ", outputFolder)
	return nil
}

func addValuePrefix(key string) string {
	return fmt.Sprintf("Values.%s", key)
}

func removeEmptyStrings(s []string) []string {
	var r []string
	for _, str := range s {
		if str != "" {
			r = append(r, str)
		}
	}
	return r
}
