package logging

import (
	"net/http"
)

// LogError logs an error message with request context and tracing information.
func LogError(request *http.Request, err string) {
	LoggerWithRequest(request).Error(err)
}
