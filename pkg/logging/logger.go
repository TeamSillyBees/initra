package logging

import (
	"strings"
	"unicode"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

const maskedValue = "***"

// Config 描述日志输出格式、级别和脱敏策略。
type Config struct {
	Level  string     `mapstructure:"level"`
	Format string     `mapstructure:"format"`
	Output string     `mapstructure:"output"`
	Mask   MaskConfig `mapstructure:"mask"`
}

// MaskConfig 描述结构化日志字段脱敏配置。
type MaskConfig struct {
	Enabled bool     `mapstructure:"enabled"`
	Fields  []string `mapstructure:"fields"`
}

// NewLogger 根据配置创建 zap 日志器。
func NewLogger(cfg Config) (*zap.Logger, error) {
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
	output := strings.TrimSpace(cfg.Output)
	if output == "" {
		output = "stdout"
	}

	loggerConfig := zap.Config{
		Level:            zap.NewAtomicLevelAt(level),
		Development:      false,
		Encoding:         encoding,
		EncoderConfig:    encoderConfig,
		OutputPaths:      []string{output},
		ErrorOutputPaths: []string{"stderr"},
	}

	options := []zap.Option{}
	if cfg.Mask.Enabled {
		fields := newSensitiveFieldSet(cfg.Mask.Fields)
		options = append(options, zap.WrapCore(func(core zapcore.Core) zapcore.Core {
			return &maskCore{
				Core:   core,
				fields: fields,
			}
		}))
	}

	return loggerConfig.Build(options...)
}

type maskCore struct {
	zapcore.Core
	fields map[string]struct{}
}

func (c *maskCore) With(fields []zapcore.Field) zapcore.Core {
	return &maskCore{
		Core:   c.Core.With(maskFields(fields, c.fields)),
		fields: c.fields,
	}
}

func (c *maskCore) Check(entry zapcore.Entry, checked *zapcore.CheckedEntry) *zapcore.CheckedEntry {
	if c.Enabled(entry.Level) {
		return checked.AddCore(entry, c)
	}
	return checked
}

func (c *maskCore) Write(entry zapcore.Entry, fields []zapcore.Field) error {
	return c.Core.Write(entry, maskFields(fields, c.fields))
}

func maskFields(fields []zapcore.Field, sensitiveFields map[string]struct{}) []zapcore.Field {
	masked := make([]zapcore.Field, len(fields))
	copy(masked, fields)
	for i := range masked {
		if isSensitiveField(masked[i].Key, sensitiveFields) {
			masked[i] = zap.String(masked[i].Key, maskedValue)
		}
	}
	return masked
}

func newSensitiveFieldSet(fields []string) map[string]struct{} {
	defaults := []string{"password", "token", "secret", "authorization", "access_key"}
	result := make(map[string]struct{}, len(defaults)+len(fields))
	for _, field := range append(defaults, fields...) {
		normalized := normalizeSensitiveField(field)
		if normalized != "" {
			result[normalized] = struct{}{}
		}
	}
	return result
}

func isSensitiveField(key string, sensitiveFields map[string]struct{}) bool {
	normalized := normalizeSensitiveField(key)
	if normalized == "" {
		return false
	}
	if _, ok := sensitiveFields[normalized]; ok {
		return true
	}
	for field := range sensitiveFields {
		switch field {
		case "password", "secret", "accesskey", "authorization":
			if strings.Contains(normalized, field) {
				return true
			}
		case "token":
			if normalized == "token" || strings.HasSuffix(normalized, "token") {
				return true
			}
		}
	}
	return false
}

func normalizeSensitiveField(field string) string {
	var builder strings.Builder
	for _, r := range strings.ToLower(strings.TrimSpace(field)) {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			builder.WriteRune(r)
		}
	}
	return builder.String()
}
