package log

import (
	"log"
	"os"
	"strings"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var Logger *zap.Logger

func init() {
	Logger = InitLogger()
}

// InitLogger initializes the logger with configurable log level from environment
func InitLogger() *zap.Logger {
	// Get log level from environment variable, default to "info"
	logLevel := strings.ToLower(os.Getenv("LOG_LEVEL"))
	if logLevel == "" {
		logLevel = "info"
	}

	// Parse log level
	var level zapcore.Level
	switch logLevel {
	case "debug":
		level = zapcore.DebugLevel
	case "info":
		level = zapcore.InfoLevel
	case "warn", "warning":
		level = zapcore.WarnLevel
	case "error":
		level = zapcore.ErrorLevel
	case "fatal":
		level = zapcore.FatalLevel
	default:
		level = zapcore.InfoLevel
		log.Printf("Invalid LOG_LEVEL '%s', defaulting to 'info'\n", logLevel)
	}

	// Create logger configuration
	cfg := zap.NewProductionConfig()
	cfg.Level = zap.NewAtomicLevelAt(level)
	cfg.EncoderConfig.EncodeTime = zapcore.TimeEncoderOfLayout(time.RFC3339)

	logger, err := cfg.Build()
	if err != nil {
		log.Println("Error while initializing logger.", err)
		// Return a no-op logger as fallback
		return zap.NewNop()
	}

	logger.Info("Logger initialized", zap.String("level", logLevel))
	return logger
}

func GetLogger() *zap.Logger {
	return Logger
}
