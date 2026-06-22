package middleware

import (
	"crypto/tls"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func setupSecurityRouter() *gin.Engine {
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(SecurityHeadersMiddleware())
	r.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, "OK")
	})
	return r
}

func setupCORSRouter() *gin.Engine {
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(CORSMiddleware())
	r.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, "OK")
	})
	r.POST("/test", func(c *gin.Context) {
		c.String(http.StatusOK, "OK")
	})
	return r
}

// Security Headers Tests

func TestSecurityHeaders_CSP(t *testing.T) {
	router := setupSecurityRouter()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/test", nil)
	router.ServeHTTP(w, req)

	csp := w.Header().Get("Content-Security-Policy")
	assert.NotEmpty(t, csp)
	assert.Contains(t, csp, "default-src 'self'")
	assert.Contains(t, csp, "script-src 'self' 'unsafe-inline'")
	assert.Contains(t, csp, "style-src 'self' 'unsafe-inline'")
	assert.Contains(t, csp, "img-src 'self' data: https:")
	assert.Contains(t, csp, "frame-ancestors 'none'")
}

func TestSecurityHeaders_XFrameOptions(t *testing.T) {
	router := setupSecurityRouter()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/test", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, "DENY", w.Header().Get("X-Frame-Options"))
}

func TestSecurityHeaders_XContentTypeOptions(t *testing.T) {
	router := setupSecurityRouter()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/test", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, "nosniff", w.Header().Get("X-Content-Type-Options"))
}

func TestSecurityHeaders_XXSSProtection(t *testing.T) {
	router := setupSecurityRouter()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/test", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, "1; mode=block", w.Header().Get("X-XSS-Protection"))
}

func TestSecurityHeaders_ReferrerPolicy(t *testing.T) {
	router := setupSecurityRouter()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/test", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, "strict-origin-when-cross-origin", w.Header().Get("Referrer-Policy"))
}

func TestSecurityHeaders_PermissionsPolicy(t *testing.T) {
	router := setupSecurityRouter()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/test", nil)
	router.ServeHTTP(w, req)

	policy := w.Header().Get("Permissions-Policy")
	assert.NotEmpty(t, policy)
	assert.Contains(t, policy, "geolocation=()")
	assert.Contains(t, policy, "microphone=()")
	assert.Contains(t, policy, "camera=()")
	assert.Contains(t, policy, "payment=()")
}

func TestSecurityHeaders_HSTS_WithoutTLS(t *testing.T) {
	router := setupSecurityRouter()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/test", nil)
	// No TLS set on request
	router.ServeHTTP(w, req)

	// HSTS should NOT be set when not using TLS
	assert.Empty(t, w.Header().Get("Strict-Transport-Security"))
}

func TestSecurityHeaders_HSTS_WithTLS(t *testing.T) {
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(SecurityHeadersMiddleware())
	r.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, "OK")
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/test", nil)
	// Simulate TLS connection
	req.TLS = &tls.ConnectionState{}
	r.ServeHTTP(w, req)

	hsts := w.Header().Get("Strict-Transport-Security")
	assert.NotEmpty(t, hsts)
	assert.Contains(t, hsts, "max-age=31536000")
	assert.Contains(t, hsts, "includeSubDomains")
}

func TestSecurityHeaders_AllHeadersPresent(t *testing.T) {
	router := setupSecurityRouter()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/test", nil)
	router.ServeHTTP(w, req)

	requiredHeaders := []string{
		"Content-Security-Policy",
		"X-Frame-Options",
		"X-Content-Type-Options",
		"X-XSS-Protection",
		"Referrer-Policy",
		"Permissions-Policy",
	}

	for _, header := range requiredHeaders {
		assert.NotEmpty(t, w.Header().Get(header), "Header %s should be present", header)
	}
}

// CORS Tests

func TestCORS_NoConfig_NoHeaders(t *testing.T) {
	// Reset viper to ensure no CORS config
	viper.Reset()

	router := setupCORSRouter()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Origin", "http://example.com")
	router.ServeHTTP(w, req)

	// When no origins are configured, CORS headers should not be set
	assert.Empty(t, w.Header().Get("Access-Control-Allow-Origin"))
}

func TestCORS_WildcardOrigin(t *testing.T) {
	// Configure wildcard origin
	viper.Reset()
	viper.Set("cors.allowed_origins", []string{"*"})
	defer viper.Reset()

	router := setupCORSRouter()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Origin", "http://any-origin.com")
	router.ServeHTTP(w, req)

	assert.Equal(t, "*", w.Header().Get("Access-Control-Allow-Origin"))
}

// TestCORS_WildcardOrigin_NoCredentials proves that when the origin is the
// wildcard "*", the middleware must NOT send Access-Control-Allow-Credentials:
// true. The Fetch/CORS spec forbids "*" combined with credentials, and browsers
// block every credentialed cross-origin request when both are present.
func TestCORS_WildcardOrigin_NoCredentials(t *testing.T) {
	viper.Reset()
	viper.Set("cors.allowed_origins", []string{"*"})
	defer viper.Reset()

	router := setupCORSRouter()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Origin", "http://any-origin.com")
	router.ServeHTTP(w, req)

	assert.Equal(t, "*", w.Header().Get("Access-Control-Allow-Origin"))
	assert.Empty(t, w.Header().Get("Access-Control-Allow-Credentials"),
		"must not send Allow-Credentials with wildcard origin (CORS spec violation)")
}

// TestCORS_SpecificOrigin_Credentials proves the credentials header is still
// sent for an explicitly allowed (non-wildcard) origin.
func TestCORS_SpecificOrigin_Credentials(t *testing.T) {
	viper.Reset()
	viper.Set("cors.allowed_origins", []string{"http://allowed-origin.com"})
	defer viper.Reset()

	router := setupCORSRouter()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Origin", "http://allowed-origin.com")
	router.ServeHTTP(w, req)

	assert.Equal(t, "http://allowed-origin.com", w.Header().Get("Access-Control-Allow-Origin"))
	assert.Equal(t, "true", w.Header().Get("Access-Control-Allow-Credentials"))
}

func TestCORS_SpecificOrigin_Allowed(t *testing.T) {
	// Configure specific origins
	viper.Reset()
	viper.Set("cors.allowed_origins", []string{"http://allowed-origin.com", "http://another-allowed.com"})
	defer viper.Reset()

	router := setupCORSRouter()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Origin", "http://allowed-origin.com")
	router.ServeHTTP(w, req)

	assert.Equal(t, "http://allowed-origin.com", w.Header().Get("Access-Control-Allow-Origin"))
}

func TestCORS_SpecificOrigin_Blocked(t *testing.T) {
	// Configure specific origins
	viper.Reset()
	viper.Set("cors.allowed_origins", []string{"http://allowed-origin.com"})
	defer viper.Reset()

	router := setupCORSRouter()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Origin", "http://blocked-origin.com")
	router.ServeHTTP(w, req)

	// Origin not in allowed list - no CORS headers
	assert.Empty(t, w.Header().Get("Access-Control-Allow-Origin"))
}

func TestCORS_PreflightOptions(t *testing.T) {
	// Configure allowed origins
	viper.Reset()
	viper.Set("cors.allowed_origins", []string{"http://allowed-origin.com"})
	defer viper.Reset()

	router := setupCORSRouter()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodOptions, "/test", nil)
	req.Header.Set("Origin", "http://allowed-origin.com")
	req.Header.Set("Access-Control-Request-Method", "POST")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNoContent, w.Code)
}

func TestCORS_MaxAge(t *testing.T) {
	// Configure allowed origins
	viper.Reset()
	viper.Set("cors.allowed_origins", []string{"http://allowed-origin.com"})
	defer viper.Reset()

	router := setupCORSRouter()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Origin", "http://allowed-origin.com")
	router.ServeHTTP(w, req)

	assert.Equal(t, "43200", w.Header().Get("Access-Control-Max-Age"))
}

func TestCORS_AllowedMethods(t *testing.T) {
	// Configure allowed origins
	viper.Reset()
	viper.Set("cors.allowed_origins", []string{"http://allowed-origin.com"})
	defer viper.Reset()

	router := setupCORSRouter()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Origin", "http://allowed-origin.com")
	router.ServeHTTP(w, req)

	methods := w.Header().Get("Access-Control-Allow-Methods")
	assert.Contains(t, methods, "GET")
	assert.Contains(t, methods, "POST")
	assert.Contains(t, methods, "PUT")
	assert.Contains(t, methods, "DELETE")
	assert.Contains(t, methods, "OPTIONS")
	assert.Contains(t, methods, "PATCH")
}

func TestCORS_AllowedHeaders(t *testing.T) {
	// Configure allowed origins
	viper.Reset()
	viper.Set("cors.allowed_origins", []string{"http://allowed-origin.com"})
	defer viper.Reset()

	router := setupCORSRouter()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Origin", "http://allowed-origin.com")
	router.ServeHTTP(w, req)

	headers := w.Header().Get("Access-Control-Allow-Headers")
	assert.Contains(t, headers, "Authorization")
	assert.Contains(t, headers, "Content-Type")
	assert.Contains(t, headers, "Origin")
}

func TestCORS_ExposeHeaders(t *testing.T) {
	// Configure allowed origins
	viper.Reset()
	viper.Set("cors.allowed_origins", []string{"http://allowed-origin.com"})
	defer viper.Reset()

	router := setupCORSRouter()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Origin", "http://allowed-origin.com")
	router.ServeHTTP(w, req)

	exposed := w.Header().Get("Access-Control-Expose-Headers")
	assert.Contains(t, exposed, "Content-Length")
	assert.Contains(t, exposed, "Content-Type")
}

func TestCORS_AllowCredentials(t *testing.T) {
	// Configure allowed origins
	viper.Reset()
	viper.Set("cors.allowed_origins", []string{"http://allowed-origin.com"})
	defer viper.Reset()

	router := setupCORSRouter()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Origin", "http://allowed-origin.com")
	router.ServeHTTP(w, req)

	assert.Equal(t, "true", w.Header().Get("Access-Control-Allow-Credentials"))
}

func TestCORS_NoOriginHeader(t *testing.T) {
	// Configure allowed origins
	viper.Reset()
	viper.Set("cors.allowed_origins", []string{"http://allowed-origin.com"})
	defer viper.Reset()

	router := setupCORSRouter()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/test", nil)
	// No Origin header set
	router.ServeHTTP(w, req)

	// No CORS headers should be set when no Origin header
	assert.Empty(t, w.Header().Get("Access-Control-Allow-Origin"))
}

func TestCORS_MultipleAllowedOrigins(t *testing.T) {
	// Configure multiple allowed origins
	viper.Reset()
	viper.Set("cors.allowed_origins", []string{
		"http://first.com",
		"http://second.com",
		"http://third.com",
	})
	defer viper.Reset()

	router := setupCORSRouter()

	tests := []struct {
		origin   string
		expected string
	}{
		{"http://first.com", "http://first.com"},
		{"http://second.com", "http://second.com"},
		{"http://third.com", "http://third.com"},
		{"http://unknown.com", ""},
	}

	for _, tt := range tests {
		t.Run(tt.origin, func(t *testing.T) {
			w := httptest.NewRecorder()
			req, _ := http.NewRequest(http.MethodGet, "/test", nil)
			req.Header.Set("Origin", tt.origin)
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expected, w.Header().Get("Access-Control-Allow-Origin"))
		})
	}
}

func TestCORS_RequestContinues(t *testing.T) {
	// Configure allowed origins
	viper.Reset()
	viper.Set("cors.allowed_origins", []string{"http://allowed-origin.com"})
	defer viper.Reset()

	router := setupCORSRouter()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Origin", "http://allowed-origin.com")
	router.ServeHTTP(w, req)

	// Request should continue to handler
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "OK", w.Body.String())
}

func TestCORS_OptionsAborts(t *testing.T) {
	// Configure allowed origins
	viper.Reset()
	viper.Set("cors.allowed_origins", []string{"http://allowed-origin.com"})
	defer viper.Reset()

	router := setupCORSRouter()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodOptions, "/test", nil)
	req.Header.Set("Origin", "http://allowed-origin.com")
	router.ServeHTTP(w, req)

	// OPTIONS should abort with 204
	assert.Equal(t, http.StatusNoContent, w.Code)
	assert.Empty(t, w.Body.String())
}
