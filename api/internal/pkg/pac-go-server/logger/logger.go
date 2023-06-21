package log

import (
	"log"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var Logger *zap.Logger

func init() {
	cfg := zap.NewProductionConfig()
	cfg.EncoderConfig.EncodeTime = zapcore.TimeEncoderOfLayout(time.RFC3339)
	logger, err := cfg.Build()
	if err != nil {
		log.Println("Error while initializing logger.", err)
	}
	Logger = logger
}

func GetLogger() *zap.Logger {
	return Logger
}
