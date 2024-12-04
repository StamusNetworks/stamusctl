package utils

import (
	// Common

	"os"
	"path/filepath"
	"strings"

	// Internal
	"stamus-ctl/internal/models"
)

// Check if the folder exists
func FolderExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if os.IsNotExist(err) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}

var forbiddenChars = []string{"\\", "/", ":", "*", "?", "\"", "<", ">", "|", "$", "..", " ", "\t", "\n", "\r"}

func ValidatePath(path models.Variable) bool {
	if *path.String == "" {
		return false
	}
	for _, char := range forbiddenChars {
		if strings.Contains(*path.String, char) {
			return false
		}
	}
	return true
}

func ListFilesInFolder(folderPath string) (map[string]string, error) {
	filesMap := make(map[string]string)
	// Walk the folder
	err := filepath.Walk(folderPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() {
			relativePath, err := filepath.Rel(folderPath, path)
			if err != nil {
				return err
			}
			filesMap[relativePath] = info.Name()
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	return filesMap, nil
}
