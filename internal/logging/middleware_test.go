package logging

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestSetRequestIdInResponse(t *testing.T) {
	// Initialize tracing
	InitTracer("", "")

	// Setup test router
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(SetRequestIDInResponse())

	// Add a test route
	router.GET("/test", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "success"})
	})

	// Make request
	recorder := httptest.NewRecorder()
	req, _ := http.NewRequestWithContext(context.Background(), http.MethodGet, "/test", nil)
	router.ServeHTTP(recorder, req)

	// Check response
	assert.Equal(t, 200, recorder.Code)

	// Check that X-Request-ID header is set
	requestID := recorder.Header().Get("X-Request-ID")
	assert.NotEmpty(t, requestID)

	// Request ID should be a valid trace ID format (32 hex characters)
	assert.Len(t, requestID, 32)
}

func TestLogRequestResponse(t *testing.T) {
	// Initialize tracing
	InitTracer("", "")

	// Setup test router
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(LogRequestResponse())

	// Add a test route
	router.GET("/test", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "success"})
	})

	// Make request
	recorder := httptest.NewRecorder()
	req, _ := http.NewRequestWithContext(context.Background(), http.MethodGet, "/test", nil)
	router.ServeHTTP(recorder, req)

	// Check response
	assert.Equal(t, 200, recorder.Code)

	// The middleware should not affect the response
	assert.Contains(t, recorder.Body.String(), "success")
}

func TestLogRequestResponse_WithDifferentMethods(t *testing.T) {
	// Initialize tracing
	InitTracer("", "")

	testCases := []struct {
		method string
		path   string
	}{
		{"GET", "/get-test"},
		{"POST", "/post-test"},
		{"PUT", "/put-test"},
		{"DELETE", "/delete-test"},
		{"PATCH", "/patch-test"},
	}

	for _, testCase := range testCases {
		t.Run(testCase.method, func(t *testing.T) {
			// Setup test router
			gin.SetMode(gin.TestMode)
			router := gin.New()
			router.Use(LogRequestResponse())

			// Add routes for different methods
			router.Handle(testCase.method, testCase.path, func(c *gin.Context) {
				c.JSON(200, gin.H{"method": testCase.method})
			})

			// Make request
			recorder := httptest.NewRecorder()
			req, _ := http.NewRequestWithContext(context.Background(), testCase.method, testCase.path, nil)
			router.ServeHTTP(recorder, req)

			// Check response
			assert.Equal(t, 200, recorder.Code)
		})
	}
}

func TestLogRequestResponse_WithDifferentStatusCodes(t *testing.T) {
	// Initialize tracing
	InitTracer("", "")

	testCases := []struct {
		name       string
		statusCode int
	}{
		{"OK", 200},
		{"Created", 201},
		{"Bad Request", 400},
		{"Unauthorized", 401},
		{"Not Found", 404},
		{"Internal Server Error", 500},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			// Setup test router
			gin.SetMode(gin.TestMode)
			router := gin.New()
			router.Use(LogRequestResponse())

			// Add a test route that returns the specific status code
			router.GET("/test", func(c *gin.Context) {
				c.JSON(testCase.statusCode, gin.H{"status": testCase.statusCode})
			})

			// Make request
			recorder := httptest.NewRecorder()
			req, _ := http.NewRequestWithContext(context.Background(), http.MethodGet, "/test", nil)
			router.ServeHTTP(recorder, req)

			// Check response
			assert.Equal(t, testCase.statusCode, recorder.Code)
		})
	}
}

func TestMiddleware_ChainedTogether(t *testing.T) {
	// Test that both middleware work together
	InitTracer("", "")

	// Setup test router with both middleware
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(SetRequestIDInResponse())
	router.Use(LogRequestResponse())

	// Add a test route
	router.GET("/test", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "success"})
	})

	// Make request
	recorder := httptest.NewRecorder()
	req, _ := http.NewRequestWithContext(context.Background(), http.MethodGet, "/test", nil)
	router.ServeHTTP(recorder, req)

	// Check response
	assert.Equal(t, 200, recorder.Code)
	assert.Contains(t, recorder.Body.String(), "success")

	// Check that X-Request-ID header is set
	requestID := recorder.Header().Get("X-Request-ID")
	assert.NotEmpty(t, requestID)
}

func TestMiddleware_WithComplexPaths(t *testing.T) {
	// Test with complex paths and query parameters
	InitTracer("", "")

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(SetRequestIDInResponse())
	router.Use(LogRequestResponse())

	// Add routes with parameters
	router.GET("/users/:id/posts/:postId", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"userId": c.Param("id"),
			"postId": c.Param("postId"),
		})
	})

	// Make request with path parameters and query string
	recorder := httptest.NewRecorder()
	req, _ := http.NewRequestWithContext(
		context.Background(),
		http.MethodGet,
		"/users/123/posts/456?sort=date&order=desc",
		nil,
	)
	router.ServeHTTP(recorder, req)

	// Check response
	assert.Equal(t, 200, recorder.Code)

	// Check that X-Request-ID header is set
	requestID := recorder.Header().Get("X-Request-ID")
	assert.NotEmpty(t, requestID)
}
