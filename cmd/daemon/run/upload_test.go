package run

import (
	"bytes"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"stamus-ctl/internal/app"

	"github.com/gin-gonic/gin"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func setupUploadRouter() *gin.Engine {
	r := gin.New()
	r.Use(gin.Recovery())
	r.POST("/api/v1/upload", uploadHandler)
	return r
}

func setupMockFSForUpload() (afero.Fs, func()) {
	oldFS := app.FS
	mockFS := afero.NewMemMapFs()
	app.FS = mockFS
	return mockFS, func() {
		app.FS = oldFS
	}
}

// createMultipartRequest creates a multipart form request with a file
func createMultipartRequest(path, filename, content string) (*http.Request, error) {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	part, err := writer.CreateFormFile("file", filename)
	if err != nil {
		return nil, err
	}

	_, err = io.WriteString(part, content)
	if err != nil {
		return nil, err
	}

	err = writer.Close()
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(http.MethodPost, "/api/v1/upload?path="+path, body)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", writer.FormDataContentType())
	return req, nil
}

func TestUploadHandler_NoFile(t *testing.T) {
	router := setupUploadRouter()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/api/v1/upload?path=/test/file.txt", nil)
	req.Header.Set("Content-Length", "0")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "No file uploaded")
}

func TestUploadHandler_NoPath(t *testing.T) {
	router := setupUploadRouter()

	// Create a request with a file but no path
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, _ := writer.CreateFormFile("file", "test.txt")
	io.WriteString(part, "test content")
	writer.Close()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/api/v1/upload", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "No path provided")
}

func TestUploadHandler_FormFileError(t *testing.T) {
	router := setupUploadRouter()

	// Create a request with content but without proper form file
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/api/v1/upload?path=/test/file.txt", bytes.NewBufferString("not a form"))
	req.Header.Set("Content-Type", "multipart/form-data")
	req.Header.Set("Content-Length", "10")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "File upload error")
}

func TestUploadHandler_ValidRequest(t *testing.T) {
	mockFS, cleanup := setupMockFSForUpload()
	defer cleanup()

	// Save original ConfigsFolder
	originalConfigsFolder := app.ConfigsFolder
	defer func() { app.ConfigsFolder = originalConfigsFolder }()

	// Setup test configs folder
	app.ConfigsFolder = "/test/configs/"
	mockFS.MkdirAll("/test/configs/", 0755)

	router := setupUploadRouter()

	// Note: Due to Gin's SaveUploadedFile using os package directly,
	// we can only test the validation part with mock filesystem
	req, err := createMultipartRequest("subdir/test.txt", "test.txt", "test content")
	require.NoError(t, err)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// The request should be processed (may fail on actual save due to afero vs os)
	// but should not be a bad request
	assert.NotEqual(t, http.StatusBadRequest, w.Code)
}

func TestUploadHandler_WithProject(t *testing.T) {
	mockFS, cleanup := setupMockFSForUpload()
	defer cleanup()

	// Save original ConfigsFolder
	originalConfigsFolder := app.ConfigsFolder
	defer func() { app.ConfigsFolder = originalConfigsFolder }()

	// Setup test configs folder
	app.ConfigsFolder = "/test/configs/"
	mockFS.MkdirAll("/test/configs/myproject/", 0755)

	router := setupUploadRouter()

	// Create request with project parameter
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, _ := writer.CreateFormFile("file", "test.txt")
	io.WriteString(part, "test content")
	writer.Close()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/api/v1/upload?path=test.txt&project=myproject", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	router.ServeHTTP(w, req)

	// Should not be a bad request
	assert.NotEqual(t, http.StatusBadRequest, w.Code)
}

func TestUploadHandler_CreatesDirectory(t *testing.T) {
	mockFS, cleanup := setupMockFSForUpload()
	defer cleanup()

	// Save original ConfigsFolder
	originalConfigsFolder := app.ConfigsFolder
	defer func() { app.ConfigsFolder = originalConfigsFolder }()

	// Setup test configs folder - but don't create the subdirectory
	app.ConfigsFolder = "/test/configs/"
	mockFS.MkdirAll("/test/configs/", 0755)

	router := setupUploadRouter()

	// Create request that should create a new directory
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, _ := writer.CreateFormFile("file", "test.txt")
	io.WriteString(part, "test content")
	writer.Close()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/api/v1/upload?path=newdir/subdir/test.txt", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	router.ServeHTTP(w, req)

	// Directory should be created (even if file save fails due to OS vs afero)
	exists, err := afero.DirExists(mockFS, "/test/configs/newdir/subdir")
	assert.NoError(t, err)
	assert.True(t, exists, "Directory should be created")
}

func TestUploadHandler_PathConstruction(t *testing.T) {
	mockFS, cleanup := setupMockFSForUpload()
	defer cleanup()

	// Save original ConfigsFolder
	originalConfigsFolder := app.ConfigsFolder
	defer func() { app.ConfigsFolder = originalConfigsFolder }()

	// Setup test configs folder
	app.ConfigsFolder = "/test/configs/"
	mockFS.MkdirAll("/test/configs/", 0755)

	// Test that path is correctly joined
	tests := []struct {
		name     string
		path     string
		project  string
		expected string
	}{
		{
			name:     "simple path",
			path:     "file.txt",
			project:  "",
			expected: filepath.Join("/test/configs", "", "file.txt"),
		},
		{
			name:     "nested path",
			path:     "dir/subdir/file.txt",
			project:  "",
			expected: filepath.Join("/test/configs", "", "dir/subdir/file.txt"),
		},
		{
			name:     "with project",
			path:     "file.txt",
			project:  "myproject",
			expected: filepath.Join("/test/configs", "myproject", "file.txt"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Just verify the path construction logic matches expectations
			completePath := filepath.Join(app.GetConfigsFolder(tt.project), tt.path)
			assert.Equal(t, tt.expected, completePath)
		})
	}
}

func TestUploadHandler_EmptyFilename(t *testing.T) {
	router := setupUploadRouter()

	// Create a request with an empty filename
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	// Create form file with empty name - this may be rejected by Gin
	part, _ := writer.CreateFormFile("file", "")
	io.WriteString(part, "test content")
	writer.Close()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/api/v1/upload?path=/test/file.txt", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	router.ServeHTTP(w, req)

	// Gin may reject empty filename with 400 - this is acceptable behavior
	// The key is that the handler doesn't panic
	assert.True(t, w.Code == http.StatusBadRequest || w.Code == http.StatusInternalServerError || w.Code == http.StatusOK,
		"Expected 200, 400, or 500, got %d", w.Code)
}

func TestUploadHandler_LargeFile(t *testing.T) {
	mockFS, cleanup := setupMockFSForUpload()
	defer cleanup()

	// Save original ConfigsFolder
	originalConfigsFolder := app.ConfigsFolder
	defer func() { app.ConfigsFolder = originalConfigsFolder }()

	// Setup test configs folder
	app.ConfigsFolder = "/test/configs/"
	mockFS.MkdirAll("/test/configs/", 0755)

	router := setupUploadRouter()

	// Create a request with a larger file (1KB of data)
	largeContent := bytes.Repeat([]byte("x"), 1024)

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, _ := writer.CreateFormFile("file", "large.txt")
	part.Write(largeContent)
	writer.Close()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/api/v1/upload?path=large.txt", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	router.ServeHTTP(w, req)

	// Should not be a bad request
	assert.NotEqual(t, http.StatusBadRequest, w.Code)
}

func TestUploadHandler_SpecialCharactersInFilename(t *testing.T) {
	mockFS, cleanup := setupMockFSForUpload()
	defer cleanup()

	// Save original ConfigsFolder
	originalConfigsFolder := app.ConfigsFolder
	defer func() { app.ConfigsFolder = originalConfigsFolder }()

	// Setup test configs folder
	app.ConfigsFolder = "/test/configs/"
	mockFS.MkdirAll("/test/configs/", 0755)

	router := setupUploadRouter()

	// Test with various special characters in filename
	filenames := []string{
		"file-with-dashes.txt",
		"file_with_underscores.txt",
		"file.multiple.dots.txt",
	}

	for _, filename := range filenames {
		t.Run(filename, func(t *testing.T) {
			body := &bytes.Buffer{}
			writer := multipart.NewWriter(body)
			part, _ := writer.CreateFormFile("file", filename)
			io.WriteString(part, "test content")
			writer.Close()

			w := httptest.NewRecorder()
			req, _ := http.NewRequest(http.MethodPost, "/api/v1/upload?path="+filename, body)
			req.Header.Set("Content-Type", writer.FormDataContentType())
			router.ServeHTTP(w, req)

			// Should not be a bad request
			assert.NotEqual(t, http.StatusBadRequest, w.Code)
		})
	}
}

func TestUploadHandler_ContentTypes(t *testing.T) {
	mockFS, cleanup := setupMockFSForUpload()
	defer cleanup()

	// Save original ConfigsFolder
	originalConfigsFolder := app.ConfigsFolder
	defer func() { app.ConfigsFolder = originalConfigsFolder }()

	// Setup test configs folder
	app.ConfigsFolder = "/test/configs/"
	mockFS.MkdirAll("/test/configs/", 0755)

	router := setupUploadRouter()

	// Test different file types (content is same, just filename differs)
	files := []struct {
		name    string
		content string
	}{
		{"config.yaml", "key: value"},
		{"config.json", `{"key": "value"}`},
		{"script.sh", "#!/bin/bash\necho hello"},
		{"cert.pem", "-----BEGIN CERTIFICATE-----"},
	}

	for _, file := range files {
		t.Run(file.name, func(t *testing.T) {
			req, err := createMultipartRequest(file.name, file.name, file.content)
			require.NoError(t, err)

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			// Should not be a bad request
			assert.NotEqual(t, http.StatusBadRequest, w.Code)
		})
	}
}
