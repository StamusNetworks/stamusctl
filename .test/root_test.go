package tests

import (
	// Core

	"testing"

	// External
	"github.com/stretchr/testify/assert"

	// Internal
	"stamus-ctl/internal/app"
	"stamus-ctl/pkg"
)

func TestPing(t *testing.T) {
	res, _ := newRequest("GET", "/api/v1/ping", nil)
	assert.Equal(t, 200, res.Code)
	assert.Equal(t, "{\"message\":\"pong\"}", res.Body.String())
}

func TestComposeInit(t *testing.T) {
	InitUnitTest(t)
	initRequest := pkg.InitRequest{
		IsDefault: true,
		Values: map[string]string{
			"nginx.exec": "nginx",
		},
	}

	res, _ := newRequest("POST", "/api/v1/compose/init", initRequest)

	if t != nil {
		assert.Equal(t, 200, res.Code)
		assert.Equal(t, "{\"message\":\"ok\"}", res.Body.String())
		current := "config"
		compareDirs(t, app.GetConfigsFolder(current), "./outputs/compose-init")
	}
}

func TestComposeInitSet(t *testing.T) {
	InitUnitTest(t)
	initRequest := pkg.InitRequest{
		IsDefault: true,
		Values: map[string]string{
			"nginx.exec":         "nginx",
			"websocket.response": "lel",
		},
	}

	res, _ := newRequest("POST", "/api/v1/compose/init", initRequest)
	assert.Equal(t, 200, res.Code)
	assert.Equal(t, "{\"message\":\"ok\"}", res.Body.String())

	current := "config"
	compareDirs(t, app.GetConfigsFolder(current), "./outputs/compose-init-set")
}

func TestComposeInitOptional(t *testing.T) {
	InitUnitTest(t)
	initRequest := pkg.InitRequest{
		IsDefault: true,
		Values: map[string]string{
			"nginx": "false",
		},
	}

	res, _ := newRequest("POST", "/api/v1/compose/init", initRequest)
	assert.Equal(t, 200, res.Code)
	assert.Equal(t, "{\"message\":\"ok\"}", res.Body.String())

	current := "config"
	compareDirs(t, app.GetConfigsFolder(current), "./outputs/compose-init-optional")
}

func TestComposeInitArbitrary(t *testing.T) {
	InitUnitTest(t)
	initRequest := pkg.InitRequest{
		IsDefault: true,
		Values: map[string]string{
			"nginx.exec":     "nginx",
			"websocket.port": "6969",
		},
	}

	res, _ := newRequest("POST", "/api/v1/compose/init", initRequest)
	assert.Equal(t, 200, res.Code)
	assert.Equal(t, "{\"message\":\"ok\"}", res.Body.String())

	current := "config"
	compareDirs(t, app.GetConfigsFolder(current), "./outputs/compose-init-arbitrary")
}

func TestConfigSet(t *testing.T) {
	InitUnitTest(t)
	// Setup
	TestComposeInit(nil)
	// Test
	setRequest := pkg.SetRequest{
		Values: map[string]string{
			"websocket.response": "lel",
		},
	}
	res, _ := newRequest("POST", "/api/v1/config", setRequest)
	assert.Equal(t, 200, res.Code)
	assert.Equal(t, "{\"message\":\"ok\"}", res.Body.String())
	// Compare
	current := "config"
	compareDirs(t, app.GetConfigsFolder(current), "./outputs/compose-init-set")
}

func TestConfigReload(t *testing.T) {
	InitUnitTest(t)
	// Setup
	TestComposeInit(nil)
	setRequest := pkg.SetRequest{
		Values: map[string]string{
			"websocket.port": "6969",
		},
	}
	newRequest("POST", "/api/v1/config", setRequest)
	// Test
	setRequest = pkg.SetRequest{
		Reload: true,
	}
	res, _ := newRequest("POST", "/api/v1/config", setRequest)
	assert.Equal(t, 200, res.Code)
	assert.Equal(t, "{\"message\":\"ok\"}", res.Body.String())
	// Compare
	current := "config"
	compareDirs(t, app.GetConfigsFolder(current), "./outputs/compose-init")
}

// func TestUpload(t *testing.T) {
// 	// Setup
// 	TestComposeInit(nil)

// 	// Prepare file to upload
// 	filePath := "../inputs/values.yaml"
// 	file, err := os.Open(filePath)
// 	if t != nil {
// 		assert.NoError(t, err)
// 	}
// 	defer file.Close()

// 	// Create buffer / multipart writer
// 	var buffer bytes.Buffer
// 	writer := multipart.NewWriter(&buffer)
// 	// Create form file field
// 	part, err := writer.CreateFormFile("file", filepath.Base(filePath))
// 	if t != nil {
// 		assert.NoError(t, err)
// 	}
// 	// Write content into filed
// 	_, err = io.Copy(part, file)
// 	if t != nil {
// 		assert.NoError(t, err)
// 	}
// 	// Close writer
// 	err = writer.Close()
// 	if t != nil {
// 		assert.NoError(t, err)
// 	}

// 	// Create router
// 	router := root.SetupRouter(func(string) {})
// 	// Create request with buffer body and content type
// 	req, err := http.NewRequest("POST", "/api/v1/upload?path=/", &buffer)
// 	req.Header.Set("Content-Type", writer.FormDataContentType())
// 	if t != nil {
// 		assert.NoError(t, err)
// 	}
// 	// Serve request
// 	w := httptest.NewRecorder()
// 	router.ServeHTTP(w, req)

// 	// Check response
// 	if t != nil {
// 		assert.Equal(t, 200, w.Code)
// 		assert.Equal(t, "{\"message\":\"Uploaded file\"}", w.Body.String())
// 	}
// }

// func TestValuesFromFile(t *testing.T) {
// 	// Setup
// 	TestUpload(nil)

// 	// Test
// 	setRequest := pkg.SetRequest{
// 		ValuesPath: "../inputs/values.yaml",
// 	}
// 	res, _ := newRequest("POST", "/api/v1/config", setRequest)
// 	assert.Equal(t, 200, res.Code)
// 	assert.Equal(t, "{\"message\":\"ok\"}", res.Body.String())

// 	// Compare
// 	compareDirs(t, "./tmp", "../outputs/compose-init-arbitrary")
// }

// func TestValueAsFile(t *testing.T) {
// 	// Setup
// 	TestUpload(nil)

// 	// Test
// 	setRequest := pkg.SetRequest{
// 		FromFile: map[string]string{
// 			"websocket.value": "../inputs/values.json",
// 		},
// 	}
// 	res, _ := newRequest("POST", "/api/v1/config", setRequest)
// 	assert.Equal(t, 200, res.Code)
// 	assert.Equal(t, "{\"message\":\"ok\"}", res.Body.String())

// 	// Compare
// 	compareDirs(t, "./tmp", "../outputs/compose-init-file")

// 	// Read inputs/values.json file content
// 	content, err := os.ReadFile("../inputs/values.json")
// 	if t != nil {
// 		assert.NoError(t, err)
// 	}
// 	// content to string
// 	t.Log(string(content))

// }
