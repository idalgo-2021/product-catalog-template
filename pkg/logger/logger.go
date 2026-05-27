package logger

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func New(level string) (*zap.Logger, error) {
	switch level {
	case "debug":
		cfg := zap.NewDevelopmentConfig()
		cfg.EncoderConfig.TimeKey = "ts"
		cfg.EncoderConfig.EncodeTime = zapcore.TimeEncoderOfLayout("15:04:05.000")
		return cfg.Build()
	default:
		// info, warn, error
		cfg := zap.NewProductionConfig()
		cfg.Level = zap.NewAtomicLevelAt(parseLevel(level))
		return cfg.Build()
	}

}

func parseLevel(level string) zapcore.Level {
	switch level {
	case "warn":
		return zapcore.WarnLevel
	case "error":
		return zapcore.ErrorLevel
	default:
		return zapcore.InfoLevel
	}
}
