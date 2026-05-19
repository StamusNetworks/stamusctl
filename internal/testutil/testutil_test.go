package testutil

// testutil_test.go — tests for the testutil helper functions.

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ---------------------------------------------------------------------------
// SetupTestRouter
// ---------------------------------------------------------------------------

func TestSetupTestRouter_WithoutAuth(t *testing.T) {
	r := SetupTestRouter(false)
	require.NotNil(t, r)
}

func TestSetupTestRouter_WithAuth(t *testing.T) {
	r := SetupTestRouter(true)
	require.NotNil(t, r)
}

// ---------------------------------------------------------------------------
// SetupTestRouterWithSecurityOnly
// ---------------------------------------------------------------------------

func TestSetupTestRouterWithSecurityOnly(t *testing.T) {
	r := SetupTestRouterWithSecurityOnly()
	require.NotNil(t, r)
}

// ---------------------------------------------------------------------------
// SetupTestRouterWithCORS
// ---------------------------------------------------------------------------

func TestSetupTestRouterWithCORS(t *testing.T) {
	r := SetupTestRouterWithCORS()
	require.NotNil(t, r)
}

// ---------------------------------------------------------------------------
// SetupMockFS
// ---------------------------------------------------------------------------

func TestSetupMockFS_ReturnsFS(t *testing.T) {
	fs, cleanup := SetupMockFS()
	defer cleanup()
	require.NotNil(t, fs)
}

func TestSetupMockFS_Cleanup_RestoresFS(t *testing.T) {
	_, cleanup := SetupMockFS()
	assert.NotPanics(t, cleanup)
}

// ---------------------------------------------------------------------------
// CreateBasicAuthHeader
// ---------------------------------------------------------------------------

func TestCreateBasicAuthHeader_Format(t *testing.T) {
	header := CreateBasicAuthHeader("user", "token123")
	assert.True(t, strings.HasPrefix(header, "Basic "), "header should start with 'Basic '")
}

func TestCreateBasicAuthHeader_NonEmpty(t *testing.T) {
	header := CreateBasicAuthHeader("admin", "secret")
	assert.NotEmpty(t, header)
}

// ---------------------------------------------------------------------------
// AssertJSONResponse
// ---------------------------------------------------------------------------

func TestAssertJSONResponse_Match(t *testing.T) {
	w := httptest.NewRecorder()
	expected := map[string]interface{}{"key": "value"}
	body, _ := json.Marshal(expected)
	w.WriteHeader(http.StatusOK)
	w.Write(body) //nolint:errcheck // test helper
	w.Header().Set("Content-Type", "application/json")

	AssertJSONResponse(t, w, http.StatusOK, expected)
}

func TestAssertJSONResponse_NilBody(t *testing.T) {
	w := httptest.NewRecorder()
	w.WriteHeader(http.StatusOK)
	// nil expectedBody — only status code is checked
	AssertJSONResponse(t, w, http.StatusOK, nil)
}

// ---------------------------------------------------------------------------
// AssertJSONContains
// ---------------------------------------------------------------------------

func TestAssertJSONContains_Match(t *testing.T) {
	w := httptest.NewRecorder()
	body, _ := json.Marshal(map[string]interface{}{"message": "ok", "extra": 42})
	w.WriteHeader(http.StatusOK)
	w.Write(body) //nolint:errcheck // test helper

	AssertJSONContains(t, w, map[string]interface{}{"message": "ok"})
}

// ---------------------------------------------------------------------------
// AssertStatusCode
// ---------------------------------------------------------------------------

func TestAssertStatusCode(t *testing.T) {
	w := httptest.NewRecorder()
	w.WriteHeader(http.StatusOK)
	AssertStatusCode(t, w, http.StatusOK)
}

// ---------------------------------------------------------------------------
// AssertHeader / AssertHeaderExists / AssertHeaderNotExists
// ---------------------------------------------------------------------------

func TestAssertHeader(t *testing.T) {
	w := httptest.NewRecorder()
	w.Header().Set("X-Custom", "myvalue")
	w.WriteHeader(http.StatusOK)
	AssertHeader(t, w, "X-Custom", "myvalue")
}

func TestAssertHeaderExists(t *testing.T) {
	w := httptest.NewRecorder()
	w.Header().Set("X-Present", "yes")
	w.WriteHeader(http.StatusOK)
	AssertHeaderExists(t, w, "X-Present")
}

func TestAssertHeaderNotExists(t *testing.T) {
	w := httptest.NewRecorder()
	w.WriteHeader(http.StatusOK)
	AssertHeaderNotExists(t, w, "X-Missing")
}

// ---------------------------------------------------------------------------
// MakeRequest / MakeGETRequest / MakePOSTRequest / MakeJSONRequest
// ---------------------------------------------------------------------------

func TestMakeRequest_GET(t *testing.T) {
	router := gin.New()
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	w := MakeRequest(router, http.MethodGet, "/test", nil, nil)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestMakeGETRequest(t *testing.T) {
	router := gin.New()
	router.GET("/get", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"method": "get"})
	})

	w := MakeGETRequest(router, "/get", nil)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestMakePOSTRequest(t *testing.T) {
	router := gin.New()
	router.POST("/post", func(c *gin.Context) {
		c.JSON(http.StatusCreated, gin.H{"created": true})
	})

	body := map[string]string{"name": "test"}
	w := MakePOSTRequest(router, "/post", body, nil)
	assert.Equal(t, http.StatusCreated, w.Code)
}

func TestMakeJSONRequest(t *testing.T) {
	router := gin.New()
	router.POST("/json", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"received": true})
	})

	body := map[string]string{"key": "val"}
	w := MakeJSONRequest(router, http.MethodPost, "/json", body, nil)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestMakeJSONRequest_NilBody(t *testing.T) {
	router := gin.New()
	router.GET("/nilbody", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	w := MakeJSONRequest(router, http.MethodGet, "/nilbody", nil, nil)
	assert.Equal(t, http.StatusOK, w.Code)
}

// ---------------------------------------------------------------------------
// CreateTestConfigDir / CreateTestFile / ReadTestFile
// ---------------------------------------------------------------------------

func TestCreateTestConfigDir(t *testing.T) {
	fs, cleanup := SetupMockFS()
	defer cleanup()

	err := CreateTestConfigDir(fs, "/myconfig")
	require.NoError(t, err)

	info, err := fs.Stat("/myconfig")
	require.NoError(t, err)
	assert.True(t, info.IsDir())
}

func TestCreateTestFile(t *testing.T) {
	fs, cleanup := SetupMockFS()
	defer cleanup()

	err := CreateTestFile(fs, "/testfile.txt", "hello")
	require.NoError(t, err)

	info, err := fs.Stat("/testfile.txt")
	require.NoError(t, err)
	assert.False(t, info.IsDir())
}

func TestReadTestFile_Success(t *testing.T) {
	fs, cleanup := SetupMockFS()
	defer cleanup()

	require.NoError(t, CreateTestFile(fs, "/readfile.txt", "content123"))

	content, err := ReadTestFile(fs, "/readfile.txt")
	require.NoError(t, err)
	assert.Equal(t, "content123", content)
}

func TestReadTestFile_MissingFile_ReturnsError(t *testing.T) {
	fs, cleanup := SetupMockFS()
	defer cleanup()

	_, err := ReadTestFile(fs, "/nonexistent.txt")
	assert.Error(t, err)
}
