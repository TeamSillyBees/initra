package logx

import (
	"fmt"
	"strings"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// newConsoleLogger 创建面向人眼阅读的 console zap logger。
func newConsoleLogger(cfg Config) (*zap.Logger, func(), error) {
	level, err := parseLevel(cfg.Console.Level)
	if err != nil {
		return nil, nil, err
	}
	encoderConfig := zap.NewDevelopmentEncoderConfig()
	encoderConfig.TimeKey = "ts"
	encoderConfig.LevelKey = "level"
	encoderConfig.NameKey = "logger"
	encoderConfig.CallerKey = "caller"
	encoderConfig.MessageKey = "msg"
	encoderConfig.StacktraceKey = ""
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	if cfg.Console.Color {
		encoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	} else {
		encoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder
	}

	sink, closeSink, err := zap.Open(cfg.Console.Output)
	if err != nil {
		return nil, nil, err
	}
	core := zapcore.NewCore(zapcore.NewConsoleEncoder(encoderConfig), sink, level)
	return zap.New(core, loggerOptions(cfg.Redact)...), closeSink, nil
}

// newJSONLLogger 创建面向机器检索的 JSON Lines zap logger。
func newJSONLLogger(cfg Config) (*zap.Logger, func(), error) {
	level, err := parseLevel(cfg.JSONL.Level)
	if err != nil {
		return nil, nil, err
	}
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.TimeKey = "ts"
	encoderConfig.LevelKey = "level"
	encoderConfig.NameKey = "logger"
	encoderConfig.CallerKey = "caller"
	encoderConfig.MessageKey = "msg"
	encoderConfig.StacktraceKey = ""
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder

	sink, closeSink, err := zap.Open(cfg.JSONL.Path)
	if err != nil {
		return nil, nil, err
	}
	core := zapcore.NewCore(zapcore.NewJSONEncoder(encoderConfig), sink, level)
	return zap.New(core, loggerOptions(cfg.Redact)...), closeSink, nil
}

// loggerOptions 返回所有底层 zap logger 共用的选项。
func loggerOptions(redact RedactConfig) []zap.Option {
	options := []zap.Option{zap.AddCaller(), zap.AddCallerSkip(1)}
	if redact.Enabled {
		options = append(options, zap.WrapCore(func(core zapcore.Core) zapcore.Core {
			return &redactCore{Core: core, cfg: redact}
		}))
	}
	return options
}

// syncLogger 刷新单个 zap logger，并忽略 stdout/stderr 常见的 sync 错误。
func syncLogger(logger *zap.Logger) error {
	if logger == nil {
		return nil
	}
	if err := logger.Sync(); err != nil && !isIgnorableSyncError(err) {
		return err
	}
	return nil
}

// isIgnorableSyncError 判断错误是否属于标准输出流不支持 Sync 的场景。
func isIgnorableSyncError(err error) bool {
	if err == nil {
		return true
	}
	text := err.Error()
	return text == "sync /dev/stderr: invalid argument" ||
		text == "sync /dev/stdout: invalid argument" ||
		text == "sync stdout: invalid argument" ||
		text == "sync stderr: invalid argument" ||
		text == "invalid argument"
}

// combineSyncErrors 合并两个 Sync 错误，保留第二个错误的 unwrap 链。
func combineSyncErrors(first error, second error) error {
	switch {
	case first == nil:
		return second
	case second == nil:
		return first
	default:
		return fmt.Errorf("%v; %w", first, second)
	}
}

// ParseLevel 将配置中的日志级别解析为 zapcore.Level。
func parseLevel(level string) (zapcore.Level, error) {
	parsed := zapcore.InfoLevel
	if strings.TrimSpace(level) == "" {
		return parsed, nil
	}
	if err := parsed.UnmarshalText([]byte(strings.ToLower(strings.TrimSpace(level)))); err != nil {
		return parsed, err
	}
	return parsed, nil
}
