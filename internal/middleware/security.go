package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
)

// SecurityHeadersMiddleware adds standard security headers to all API responses
func SecurityHeadersMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Content-Security-Policy: Restrict resource loading to prevent XSS attacks
		// default-src 'self': Only allow resources from same origin
		// script-src 'self' 'unsafe-inline': Allow scripts from same origin and inline scripts (for Swagger UI)
		// style-src 'self' 'unsafe-inline': Allow styles from same origin and inline styles (for Swagger UI)
		// img-src 'self' data: https:: Allow images from same origin, data URIs, and HTTPS sources
		c.Header("Content-Security-Policy", "default-src 'self'; script-src 'self' 'unsafe-inline'; style-src 'self' 'unsafe-inline'; img-src 'self' data: https:; font-src 'self' data:; object-src 'none'; base-uri 'self'; form-action 'self'; frame-ancestors 'none'")

		// Strict-Transport-Security (HSTS): Force HTTPS connections for 1 year including subdomains
		// Only add HSTS if the connection is over HTTPS to avoid browser warnings
		if c.Request.TLS != nil {
			c.Header("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		}

		// X-Frame-Options: Prevent clickjacking attacks by disabling iframe embedding
		c.Header("X-Frame-Options", "DENY")

		// X-Content-Type-Options: Prevent MIME type sniffing
		c.Header("X-Content-Type-Options", "nosniff")

		// X-XSS-Protection: Enable browser's XSS protection (legacy header but still useful)
		c.Header("X-XSS-Protection", "1; mode=block")

		// Referrer-Policy: Control referrer information
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")

		// Permissions-Policy: Disable unnecessary browser features
		c.Header("Permissions-Policy", "geolocation=(), microphone=(), camera=(), payment=()")

		// Continue processing the request
		c.Next()
	}
}

// CORSMiddleware configures CORS policies for the API
// CORS allows controlled access from web browsers running on different domains
func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get allowed origins from configuration, default to localhost for development
		allowedOrigins := viper.GetStringSlice("cors.allowed_origins")
		if len(allowedOrigins) == 0 {
			// Default to no CORS if not configured
			// This is secure by default - only same-origin requests allowed
			allowedOrigins = []string{}
		}

		// Check if the request origin is allowed
		origin := c.Request.Header.Get("Origin")
		allowed := false
		wildcard := false

		// If wildcard is configured, allow all origins (not recommended for production)
		for _, allowedOrigin := range allowedOrigins {
			if allowedOrigin == "*" {
				c.Header("Access-Control-Allow-Origin", "*")
				allowed = true
				wildcard = true
				break
			}
			if allowedOrigin == origin {
				c.Header("Access-Control-Allow-Origin", origin)
				allowed = true
				break
			}
		}

		// Only set CORS headers if origin is allowed
		if allowed {
			// Allow credentials (cookies, authorization headers, etc.).
			// The Fetch/CORS spec forbids combining credentials with the "*"
			// wildcard origin: browsers block such responses outright. Only
			// advertise credentials when echoing a specific origin.
			if !wildcard {
				c.Header("Access-Control-Allow-Credentials", "true")
			}

			// Specify allowed HTTP methods
			c.Header("Access-Control-Allow-Methods",
				"GET, POST, PUT, DELETE, OPTIONS, PATCH")

			// Specify allowed headers in requests
			c.Header("Access-Control-Allow-Headers",
				"Origin, Content-Type, Content-Length, Accept-Encoding, Authorization, X-CSRF-Token, Accept, X-Requested-With")

			// Specify which response headers can be exposed to the browser
			c.Header("Access-Control-Expose-Headers", "Content-Length, Content-Type")

			// Cache preflight requests for 12 hours
			c.Header("Access-Control-Max-Age", "43200")
		}

		// Handle preflight OPTIONS requests
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}
