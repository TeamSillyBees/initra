package logx

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
	apperrors "github.com/teamsillybees/initra/pkg/errors"
	"github.com/teamsillybees/initra/pkg/requestctx"
	"go.uber.org/zap"
)

// TestLoggerErrorSplitsConsoleAndJSONLFields 验证 Error 会分别构造 console 和 jsonl 字段。
func TestLoggerErrorSplitsConsoleAndJSONLFields(t *testing.T) {
	dir := t.TempDir()
	consolePath := filepath.Join(dir, "console.log")
	jsonlPath := filepath.Join(dir, "app.jsonl")
	logger, err := NewLogger(Config{
		Level: "debug",
		Console: ConsoleConfig{
			Enabled: true,
			Level:   "debug",
			Stack:   StackShort,
			Output:  consolePath,
		},
		JSONL: JSONLConfig{
			Enabled: true,
			Level:   "debug",
			Stack:   StackFull,
			Path:    jsonlPath,
		},
		Redact: RedactConfig{Enabled: true},
	}.Normalize())
	require.NoError(t, err)

	err = apperrors.Wrap(errors.New("driver: duplicate key"), apperrors.CodeDBError, "create user failed",
		apperrors.WithCauseTrace("trace-1"),
		apperrors.WithCauseAttr("password", "secret-password"),
	)
	ctx := requestctx.WithTraceID(context.Background(), "trace-from-context")
	logger.Error(ctx, "request failed", err, zap.String("path", "/api/v1/users"), zap.String("password", "plain-password"))
	require.NoError(t, logger.Sync())

	consoleBody := readFile(t, consolePath)
	jsonlBody := readFile(t, jsonlPath)
	require.Contains(t, consoleBody, "request failed")
	require.Contains(t, consoleBody, "DB_ERROR")
	require.NotContains(t, consoleBody, "error_context")
	require.NotContains(t, consoleBody, "secret-password")
	require.NotContains(t, consoleBody, "plain-password")
	require.Contains(t, jsonlBody, `"error_code":"DB_ERROR"`)
	require.Contains(t, jsonlBody, `"trace_id":"trace-1"`)
	require.Contains(t, jsonlBody, `"error_context"`)
	require.NotContains(t, jsonlBody, "secret-password")
	require.NotContains(t, jsonlBody, "plain-password")
}

// TestLoggerInfoRedactsFields 验证普通日志也会经过统一脱敏层。
func TestLoggerInfoRedactsFields(t *testing.T) {
	dir := t.TempDir()
	jsonlPath := filepath.Join(dir, "app.jsonl")
	logger, err := NewLogger(Config{
		Console: ConsoleConfig{Enabled: false},
		JSONL:   JSONLConfig{Enabled: true, Level: "debug", Path: jsonlPath},
		Redact:  RedactConfig{Enabled: true},
	}.Normalize())
	require.NoError(t, err)

	logger.Info(context.Background(), "config loaded", zap.String("authorization", "Bearer secret-token"))
	require.NoError(t, logger.Sync())

	body := readFile(t, jsonlPath)
	require.Contains(t, body, `"authorization":"`+RedactedValue+`"`)
	require.NotContains(t, body, "secret-token")
}

// readFile 读取测试日志文件内容。
func readFile(t *testing.T, path string) string {
	t.Helper()
	content, err := os.ReadFile(path)
	require.NoError(t, err)
	return string(content)
}
