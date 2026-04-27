package logging

import (
	"strings"

	"github.com/teamsillybees/initra/internal/platform/config"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// NewLogger 根据配置创建 zap 日志器。
func NewLogger(cfg config.LoggerConfig) (*zap.Logger, error) {
	level := zapcore.InfoLevel
	if err := level.UnmarshalText([]byte(strings.ToLower(cfg.Level))); err != nil {
		return nil, err
	}

	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.TimeKey = "timestamp"
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder

	encoding := cfg.Format
	if encoding == "" {
		encoding = "json"
	}

	loggerConfig := zap.Config{
		Level:            zap.NewAtomicLevelAt(level),
		Development:      false,
		Encoding:         encoding,
		EncoderConfig:    encoderConfig,
		OutputPaths:      []string{"stdout"},
		ErrorOutputPaths: []string{"stderr"},
	}

	return loggerConfig.Build()
}
