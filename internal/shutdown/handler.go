package shutdown

import (
	"context"
)

// Priority defines the order in which handlers are executed during shutdown.
// Lower numbers execute first.
type Priority int

const (
	// PriorityFirst stops accepting new work (HTTP servers, etc.).
	PriorityFirst Priority = 100

	// PriorityInFlight waits for in-flight operations to complete.
	PriorityInFlight Priority = 200

	// PriorityConnections closes client connections (Docker, Redis, etc.).
	PriorityConnections Priority = 300

	// PriorityTelemetry flushes telemetry and tracing data.
	PriorityTelemetry Priority = 400

	// PriorityLast handles final cleanup (logging, etc.).
	PriorityLast Priority = 500
)

// Handler represents a cleanup handler that runs during shutdown.
type Handler struct {
	// Name identifies this handler for logging purposes.
	Name string

	// Priority determines execution order (lower = earlier).
	Priority Priority

	// Fn is the cleanup function to execute.
	// It receives a context that will be cancelled if the shutdown timeout is exceeded.
	// The function should return an error if cleanup fails.
	Fn func(ctx context.Context) error
}

// handlerList maintains a sorted list of handlers by priority.
type handlerList []Handler

func (h handlerList) Len() int           { return len(h) }
func (h handlerList) Less(i, j int) bool { return h[i].Priority < h[j].Priority }
func (h handlerList) Swap(i, j int)      { h[i], h[j] = h[j], h[i] }
