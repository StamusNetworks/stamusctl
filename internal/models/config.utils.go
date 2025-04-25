package models

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"stamus-ctl/internal/app"
	"stamus-ctl/internal/logging"

	"github.com/Masterminds/sprig/v3"
	"github.com/spf13/afero"
	"go.uber.org/zap"
)

func deleteEmptyFiles(folderPath string) error {
	err := afero.Walk(app.FS, folderPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		// Check if it's a regular file and empty
		if !info.IsDir() && info.Size() == 0 {
			err := app.FS.Remove(path)
			if err != nil && !os.IsPermission(err) {
				return err
			}
		}
		return nil
	})
	return err
}

func deleteEmptyFolders(folderPath string) error {
	empty, _ := isDirEmpty(folderPath)
	if empty {
		return nil
	}
	err := afero.Walk(app.FS, folderPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			if os.IsNotExist(err) {
				return nil
			}
			return err
		}
		// Check if it's a directory and empty
		isDir, err := afero.IsDir(app.FS, path)
		if err != nil {
			return err
		}
		if isDir {
			return removeDirIfEmpty(path)
		}
		return nil
	})
	return err
}

func removeDirIfEmpty(path string) error {
	isEmpty, err := isDirEmpty(path)
	if err != nil {
		return err
	}

	if isEmpty {
		err := app.FS.Remove(path)
		if err != nil {
			return err
		}
	}
	return nil
}

func isDirEmpty(name string) (bool, error) {
	f, err := app.FS.Open(name)
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
	err := afero.Walk(app.FS, folderPath, func(path string, info os.FileInfo, err error) error {
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

func getAllFilesContent(folderPath string, extension string) ([]string, error) {
	files, err := getAllFiles(folderPath, extension)
	if err != nil {
		return nil, err
	}
	for i, file := range files {
		content, err := afero.ReadFile(app.FS, file)
		if err != nil {
			return nil, err
		}
		files[i] = string(content)
	}
	return files, nil
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
	tpls, err := getAllFilesContent(inputFolder, ".tpl")
	logging.Sugar.Debug("walking in: ", inputFolder, " to: ", outputFolder)
	if err != nil {
		return err
	}

	// Walk the source directory and process templates
	err = afero.Walk(app.FS, inputFolder, func(path string, info os.FileInfo, err error) error {
		logger := logging.Sugar.With("path", path, "isDir", info.IsDir(), "mod", info.Mode().String())
		if err != nil {
			logger.Error(err)
			return err
		}
		return processTemplate(data, tpls, path, inputFolder, outputFolder, info, logger)
	})
	if err != nil {
		return err
	}
	logging.Sugar.Info("Configuration saved to: ", outputFolder)
	return nil
}

func processTemplate(data map[string]interface{}, tpls []string,
	path, inputFolder, outputFolder string, info os.FileInfo, logger *zap.SugaredLogger,
) error {
	// Pass through non-template files
	if filepath.Ext(info.Name()) == ".tpl" {
		return nil
	}

	// Get the relative path to the input folder
	rel, err := filepath.Rel(inputFolder, path)
	if err != nil {
		logger.Error(err)
		return err
	}
	destPath := filepath.Join(outputFolder, rel)
	logger = logger.With("destPath", destPath)

	// Create the destination directory if it doesn't exist
	logger.Debug("Walkin")
	if info.IsDir() {
		return app.FS.MkdirAll(destPath, info.Mode())
	}

	// Extract file content
	pathContent, err := afero.ReadFile(app.FS, path)
	if err != nil {
		logger.Error(err)
		return err
	}
	content := strings.Join(append([]string{string(pathContent)}, tpls...), "\n")
	if len(content) == 0 {
		return nil
	}
	if content[len(content)-1] != '\n' {
		content = content + "\n"
	}

	// Process template
	tmpl, err := template.New(filepath.Base(path)).Funcs(sprig.FuncMap()).Parse(content)
	if err != nil {
		logger.Error(err)
		return err
	}

	// Create destination file
	destFile, err := app.FS.Create(destPath)
	if err != nil {
		logger.Error(err)
		return err
	}
	defer destFile.Close()

	// Execute the template
	err = tmpl.Execute(destFile, data)
	if err != nil {
		splited := strings.Split(err.Error(), "error calling fail: ")
		fmt.Println("Failed instanciating template.", splited[1])
		return err
	}

	// Set permissions for shell scripts
	if filepath.Ext(path) == ".sh" {
		err = app.FS.Chmod(destPath, 0o755)
		if err != nil {
			logger.Error(err)
			return err
		}
	}

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
