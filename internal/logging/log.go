package logging

import (
	"stamus-ctl/internal/app"

	"github.com/spf13/viper"
	"github.com/uptrace/opentelemetry-go-extra/otelzap"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	envType = "dev"
	Logger  *zap.Logger
	Sugar   *zap.SugaredLogger
	levels  = [...]zapcore.Level{zap.WarnLevel, zap.InfoLevel, zap.DebugLevel}
)

func NewLogger() *zap.Logger {
	verbosity := viper.GetInt("verbose")
	encoder := zapcore.EncoderConfig{
		// Keys can be anything except the empty string.
		TimeKey:        "T",
		LevelKey:       "L",
		NameKey:        "N",
		CallerKey:      "C",
		FunctionKey:    zapcore.OmitKey,
		MessageKey:     "M",
		StacktraceKey:  "S",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.StringDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
		EncodeLevel:    zapcore.CapitalColorLevelEncoder,
	}

	if verbosity >= len(levels) {
		verbosity = len(levels) - 1
	}

	if envType == "prd" {
		encoder.StacktraceKey = zapcore.OmitKey

		if app.Name == app.CtlName {
			encoder.TimeKey = zapcore.OmitKey
		}
	}

	config := zap.Config{
		Level:             zap.NewAtomicLevelAt(levels[verbosity]),
		Development:       true,
		Encoding:          "console",
		EncoderConfig:     encoder,
		OutputPaths:       []string{"stdout"},
		ErrorOutputPaths:  []string{"stderr"},
		DisableStacktrace: true,
	}

	log, _ := config.Build()
	defer log.Sync()
	return log
}

func SetLogger() {
	Logger = NewLogger()
	Sugar = Logger.Sugar()

	config := zap.NewProductionConfig()
	logger, _ := config.Build()
	otellogger := otelzap.New(logger)

	otelzap.ReplaceGlobals(otellogger)
}

func init() {
	noop := zap.NewNop()

	Logger = noop
	Sugar = noop.Sugar()
}
