package tests

import (
	"bytes"
	"embed"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"

	"github.com/kylelemons/godebug/diff"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"

	root "stamus-ctl/cmd/daemon/run"
	"stamus-ctl/internal/app"
	"stamus-ctl/internal/models"
)

//go:embed config/*
var inputConf embed.FS

//go:embed outputs/*
var outputConfs embed.FS

// InitUnitTest initializes the test environment, exports the embedded files and extracts them to the outputs folder in the in memory filesystem
func InitUnitTest(t *testing.T) {
	app.Embed.Set("true")
	err := models.ExtractEmbedTo("config", inputConf, app.TemplatesFolder+"clearndr/embedded/")
	assert.NoError(t, err)
	err = models.ExtractEmbedTo("outputs", outputConfs, "./outputs")
	assert.NoError(t, err)
}

func newRequest(method string, url string, body interface{}) (*httptest.ResponseRecorder, error) {
	// Create router
	router := root.SetupRouter(func(string) {})
	// Create a new request
	w := httptest.NewRecorder()
	req, err := http.NewRequest(method, url, newBody(body))
	if err != nil {
		return nil, err
	}
	// Serve the request
	router.ServeHTTP(w, req)

	return w, nil
}

// Generic input body for POST requests
func newBody(body interface{}) io.Reader {
	bodyJson, _ := json.Marshal(body)
	return bytes.NewReader(bodyJson)
}

// compareDirs compares the content of two directories
func compareDirs(t *testing.T, dir1, dir2 string) {
	folder1, err := getFolderContent(dir1)
	assert.NoError(t, err, fmt.Sprintf("failed to read directory %s with error %s", dir1, err))
	folder2, err := getFolderContent(dir2)
	assert.NoError(t, err, fmt.Sprintf("failed to read directory %s with error %s", dir2, err))

	err = compareFolderContent(folder1, folder2)
	assert.NoError(t, err, fmt.Sprintf("directories have different content with error %s", err))
	err = compareFolderContent(folder2, folder1)
	assert.NoError(t, err, fmt.Sprintf("directories have different content with error %s", err))

	folder1Names := []string{}
	for name := range folder1 {
		folder1Names = append(folder1Names, name)
	}
	folder2Names := []string{}
	for name := range folder2 {
		folder2Names = append(folder2Names, name)
	}
	log.Println("Names in folder1:", len(folder1Names), folder1Names, "Names in folder2:", len(folder2Names), folder2Names)
}

func getFolderContent(folder string) (map[string]string, error) {
	fileMap := make(map[string]string)
	err := readFolder(folder, folder, fileMap)
	if err != nil {
		return nil, err
	}
	return fileMap, nil
}

func readFolder(basePath, currentPath string, fileMap map[string]string) error {
	files, err := afero.ReadDir(app.FS, currentPath)
	if err != nil {
		return err
	}
	for _, file := range files {
		relativePath := filepath.Join(currentPath, file.Name())
		if file.IsDir() {
			err := readFolder(basePath, relativePath, fileMap)
			if err != nil {
				return err
			}
		} else {
			content, err := afero.ReadFile(app.FS, relativePath)
			if err != nil {
				return err
			}
			// Generate the key as a relative path from the base folder
			key, err := filepath.Rel(basePath, relativePath)
			if err != nil {
				return err
			}
			fileMap[key] = string(content)
		}
	}
	return nil
}

func compareFolderContent(folder1 map[string]string, folder2 map[string]string) error {
	if len(folder1) != len(folder2) {
		return fmt.Errorf("directories have different number of files")
	}
	for name, content1 := range folder1 {
		content2, ok := folder2[name]
		if !ok {
			return fmt.Errorf("file %s is missing in directory", name)
		}
		if RemoveLinesWithSeed(strings.TrimSuffix(content1, "\n")) !=
			RemoveLinesWithSeed(strings.TrimSuffix(content2, "\n")) {
			log.Println("Content diff for", name, "content1", content1, "content2", content2)
			return fmt.Errorf("file content mismatch for %s \nContent diff:\n%s", name, diff.Diff(content1, content2))
		}
	}
	return nil
}

// RemoveLinesWithSeed removes all lines containing the word "seed" from the input string.
func RemoveLinesWithSeed(input string) string {
	var result []string
	lines := strings.Split(input, "\n")
	for _, line := range lines {
		if !strings.Contains(line, "seed") {
			result = append(result, line)
		}
	}
	return strings.Join(result, "\n")
}
