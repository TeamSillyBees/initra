package logging

import (
	"bytes"
	"net/url"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

// TestNewLoggerMasksSensitiveFields 验证结构化日志字段会按配置脱敏。
func TestNewLoggerMasksSensitiveFields(t *testing.T) {
	sink := &bufferSink{}
	scheme := "masktest" + strconv.FormatInt(time.Now().UnixNano(), 10)
	require.NoError(t, zap.RegisterSink(scheme, func(*url.URL) (zap.Sink, error) {
		return sink, nil
	}))

	logger, err := NewLogger(Config{
		Level:  "info",
		Format: "json",
		Output: scheme + "://",
		Mask: MaskConfig{
			Enabled: true,
			Fields:  []string{"authorization"},
		},
	})
	require.NoError(t, err)

	logger.Info("config loaded",
		zap.String("password", "plain-password"),
		zap.String("token", "plain-token"),
		zap.String("authorization", "Bearer plain-token"),
		zap.String("username", "alice"),
	)
	require.NoError(t, logger.Sync())

	body := sink.String()

	require.NotContains(t, body, "plain-password")
	require.NotContains(t, body, "plain-token")
	require.Contains(t, body, `"password":"***"`)
	require.Contains(t, body, `"token":"***"`)
	require.Contains(t, body, `"authorization":"***"`)
	require.Contains(t, body, `"username":"alice"`)
}

type bufferSink struct {
	bytes.Buffer
}

func (s *bufferSink) Sync() error {
	return nil
}

func (s *bufferSink) Close() error {
	return nil
}
