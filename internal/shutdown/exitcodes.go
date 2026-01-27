// Package shutdown provides graceful shutdown management for stamus-ctl.
package shutdown

// Exit codes for different shutdown scenarios.
const (
	// ExitSuccess indicates normal completion or clean shutdown.
	ExitSuccess = 0

	// ExitError indicates a general error.
	ExitError = 1

	// ExitMisuse indicates invalid arguments or configuration.
	ExitMisuse = 2

	// ExitSIGINT indicates the process was interrupted by SIGINT (Ctrl+C).
	// Standard convention: 128 + signal number (2 for SIGINT).
	ExitSIGINT = 130

	// ExitSIGTERM indicates the process was terminated by SIGTERM.
	// Standard convention: 128 + signal number (15 for SIGTERM).
	ExitSIGTERM = 143

	// ExitTimeout indicates shutdown timeout was exceeded and process was force-killed.
	ExitTimeout = 124
)
