package logging

import (
	"os"

	"github.com/uptrace/opentelemetry-go-extra/otelzap"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	envType = "dev"
	// Logger is the global zap logger instance used throughout the application.
	Logger *zap.Logger
	// Sugar is the global sugared logger instance for convenient logging.
	Sugar  *zap.SugaredLogger
	levels = [...]zapcore.Level{zap.WarnLevel, zap.InfoLevel, zap.DebugLevel}
)

// NewLogger creates a new zap logger with optional file logging.
func NewLogger(filelogger bool) *zap.Logger {
	verbosity := 2
	encoder := zapcore.EncoderConfig{
		// Keys can be anything except the empty string.
		TimeKey:        "timestamp",
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		LevelKey:       "level",
		NameKey:        "name",
		CallerKey:      "caller",
		FunctionKey:    zapcore.OmitKey,
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.CapitalColorLevelEncoder,
		EncodeDuration: zapcore.StringDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	if verbosity >= len(levels) {
		verbosity = len(levels) - 1
	}

	if envType == "prd" {
		verbosity = 1

		encoder.StacktraceKey = zapcore.OmitKey
	}

	consoleEncoder := zapcore.NewConsoleEncoder(encoder)

	encoder.EncodeLevel = zapcore.CapitalLevelEncoder

	jsonEncoder := zapcore.NewJSONEncoder(encoder)

	level := zap.NewAtomicLevelAt(levels[verbosity])

	var core zapcore.Core
	if filelogger {
		f, err := os.OpenFile("/var/log/stamus-ctl/stamus-ctl.log",
			os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0o644)
		if err != nil {
			panic(err)
		}

		core = zapcore.NewTee(
			zapcore.NewCore(jsonEncoder, zapcore.AddSync(f), level),
			zapcore.NewCore(consoleEncoder, zapcore.AddSync(os.Stdout), level),
		)
	} else {
		core = zapcore.NewCore(consoleEncoder, zapcore.AddSync(os.Stdout), level)
	}

	log := zap.New(core)
	defer func() {
		_ = log.Sync()
	}()

	return log
}

// SetLogger initializes the global logger instances with OpenTelemetry integration.
func SetLogger() {
	filelogger := os.Getenv("FILE_LOGGER") == "true"
	Logger = NewLogger(filelogger)
	Sugar = Logger.Sugar()

	config := zap.NewProductionConfig()
	logger, _ := config.Build()
	otellogger := otelzap.New(logger)

	zap.ReplaceGlobals(Logger)
	otelzap.ReplaceGlobals(otellogger)
}

func init() {
	Logger = NewLogger(false)
	Sugar = Logger.Sugar()
}
