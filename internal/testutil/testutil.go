// Package testutil provides shared test helpers for daemon tests.
package testutil

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"stamus-ctl/internal/app"
	"stamus-ctl/internal/auth"
	"stamus-ctl/internal/middleware"

	"github.com/gin-gonic/gin"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
)

func init() {
	gin.SetMode(gin.TestMode)
}

// SetupTestRouter creates a test router with optional authentication.
// If withAuth is true, the auth middleware is enabled.
func SetupTestRouter(withAuth bool) *gin.Engine {
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(middleware.SecurityHeadersMiddleware())
	r.Use(middleware.CORSMiddleware())
	if withAuth {
		r.Use(auth.AuthMiddleware())
	}
	return r
}

// SetupTestRouterWithSecurityOnly creates a test router with only security middleware.
func SetupTestRouterWithSecurityOnly() *gin.Engine {
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(middleware.SecurityHeadersMiddleware())
	return r
}

// SetupTestRouterWithCORS creates a test router with CORS middleware.
func SetupTestRouterWithCORS() *gin.Engine {
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(middleware.CORSMiddleware())
	return r
}

// SetupMockFS creates an in-memory filesystem for testing.
// Returns the mock filesystem and a cleanup function to restore the original FS.
func SetupMockFS() (afero.Fs, func()) {
	oldFS := app.FS
	mockFS := afero.NewMemMapFs()
	app.FS = mockFS
	return mockFS, func() {
		app.FS = oldFS
	}
}

// CreateBasicAuthHeader creates a Basic Auth header value for testing.
// Format: "Basic base64(username:token)"
func CreateBasicAuthHeader(username, token string) string {
	credentials := username + ":" + token
	encoded := base64.StdEncoding.EncodeToString([]byte(credentials))
	return "Basic " + encoded
}

// AssertJSONResponse checks that the response has the expected status code
// and that the body matches the expected JSON structure.
func AssertJSONResponse(t *testing.T, w *httptest.ResponseRecorder, expectedStatus int, expectedBody map[string]interface{}) {
	t.Helper()

	assert.Equal(t, expectedStatus, w.Code, "status code mismatch")

	if expectedBody != nil {
		var actualBody map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &actualBody)
		assert.NoError(t, err, "failed to unmarshal response body")

		for key, expectedValue := range expectedBody {
			actualValue, exists := actualBody[key]
			assert.True(t, exists, "expected key %q not found in response", key)
			assert.Equal(t, expectedValue, actualValue, "value mismatch for key %q", key)
		}
	}
}

// AssertJSONContains checks that the response body contains the expected key-value pairs.
func AssertJSONContains(t *testing.T, w *httptest.ResponseRecorder, expectedPairs map[string]interface{}) {
	t.Helper()

	var actualBody map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &actualBody)
	assert.NoError(t, err, "failed to unmarshal response body")

	for key, expectedValue := range expectedPairs {
		actualValue, exists := actualBody[key]
		assert.True(t, exists, "expected key %q not found in response", key)
		assert.Equal(t, expectedValue, actualValue, "value mismatch for key %q", key)
	}
}

// AssertStatusCode checks that the response has the expected status code.
func AssertStatusCode(t *testing.T, w *httptest.ResponseRecorder, expectedStatus int) {
	t.Helper()
	assert.Equal(t, expectedStatus, w.Code, "status code mismatch")
}

// AssertHeader checks that the response has the expected header value.
func AssertHeader(t *testing.T, w *httptest.ResponseRecorder, header, expectedValue string) {
	t.Helper()
	actualValue := w.Header().Get(header)
	assert.Equal(t, expectedValue, actualValue, "header %q mismatch", header)
}

// AssertHeaderExists checks that the response has the specified header.
func AssertHeaderExists(t *testing.T, w *httptest.ResponseRecorder, header string) {
	t.Helper()
	actualValue := w.Header().Get(header)
	assert.NotEmpty(t, actualValue, "header %q not found", header)
}

// AssertHeaderNotExists checks that the response does not have the specified header.
func AssertHeaderNotExists(t *testing.T, w *httptest.ResponseRecorder, header string) {
	t.Helper()
	actualValue := w.Header().Get(header)
	assert.Empty(t, actualValue, "header %q should not exist", header)
}

// MakeRequest creates and performs an HTTP request on the given router.
func MakeRequest(router *gin.Engine, method, path string, body io.Reader, headers map[string]string) *httptest.ResponseRecorder {
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(method, path, body)

	for key, value := range headers {
		req.Header.Set(key, value)
	}

	router.ServeHTTP(w, req)
	return w
}

// MakeJSONRequest creates and performs a JSON HTTP request on the given router.
func MakeJSONRequest(router *gin.Engine, method, path string, body interface{}, headers map[string]string) *httptest.ResponseRecorder {
	var bodyReader io.Reader
	if body != nil {
		jsonBytes, _ := json.Marshal(body)
		bodyReader = bytes.NewBuffer(jsonBytes)
	}

	if headers == nil {
		headers = make(map[string]string)
	}
	headers["Content-Type"] = "application/json"

	return MakeRequest(router, method, path, bodyReader, headers)
}

// MakeGETRequest creates and performs a GET request on the given router.
func MakeGETRequest(router *gin.Engine, path string, headers map[string]string) *httptest.ResponseRecorder {
	return MakeRequest(router, http.MethodGet, path, nil, headers)
}

// MakePOSTRequest creates and performs a POST request on the given router.
func MakePOSTRequest(router *gin.Engine, path string, body interface{}, headers map[string]string) *httptest.ResponseRecorder {
	return MakeJSONRequest(router, http.MethodPost, path, body, headers)
}

// CreateTestConfigDir creates a test configuration directory structure.
func CreateTestConfigDir(fs afero.Fs, configPath string) error {
	return fs.MkdirAll(configPath, 0755)
}

// CreateTestFile creates a test file with the given content.
func CreateTestFile(fs afero.Fs, path string, content string) error {
	return afero.WriteFile(fs, path, []byte(content), 0644)
}

// ReadTestFile reads a test file and returns its content.
func ReadTestFile(fs afero.Fs, path string) (string, error) {
	data, err := afero.ReadFile(fs, path)
	if err != nil {
		return "", err
	}
	return string(data), nil
}
