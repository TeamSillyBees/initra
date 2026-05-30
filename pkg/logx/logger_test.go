package logx

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	apperrors "github.com/teamsillybees/initra/pkg/errors"
	"github.com/teamsillybees/initra/pkg/requestctx"
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
	logger.Error(ctx, "request failed", err,
		String("method", "GET"),
		String("path", "/api/v1/users"),
		Int("status", 404),
		String("operation", "GetProfile"),
		String("password", "plain-password"),
	)
	require.NoError(t, logger.Sync())

	consoleBody := readFile(t, consolePath)
	jsonlBody := readFile(t, jsonlPath)
	require.NotContains(t, consoleBody, `{"`)
	require.Contains(t, consoleBody, "request failed")
	require.Contains(t, consoleBody, "GET /api/v1/users -> 404")
	require.Contains(t, consoleBody, "code=DB_ERROR")
	require.Contains(t, consoleBody, "trace=trace-1")
	require.Contains(t, consoleBody, "public=internal error")
	require.Contains(t, consoleBody, "ctx: operation=GetProfile")
	require.Contains(t, consoleBody, "stack:")
	require.LessOrEqual(t, len(consoleStackLines(consoleBody)), shortStackFrameLimit)
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

	logger.Info(context.Background(), "config loaded", String("authorization", "Bearer secret-token"))
	require.NoError(t, logger.Sync())

	body := readFile(t, jsonlPath)
	require.Contains(t, body, `"authorization":"`+RedactedValue+`"`)
	require.NotContains(t, body, "secret-token")
}

// TestNewLoggerRejectsJSONLStdIOPath 验证 JSONL 不允许写入标准输出流。
func TestNewLoggerRejectsJSONLStdIOPath(t *testing.T) {
	for _, path := range []string{"stdout", "stderr"} {
		t.Run(path, func(t *testing.T) {
			_, err := NewLogger(Config{
				Console: ConsoleConfig{Enabled: false},
				JSONL:   JSONLConfig{Enabled: true, Path: path},
			}.Normalize())

			require.Error(t, err)
			require.Contains(t, err.Error(), "log.jsonl.path")
		})
	}
}

// TestNewLoggerCreatesJSONLParentDirectory 验证 JSONL 文件父目录会被自动创建。
func TestNewLoggerCreatesJSONLParentDirectory(t *testing.T) {
	jsonlPath := filepath.Join(t.TempDir(), "missing", "logs", "app.jsonl")
	logger, err := NewLogger(Config{
		Console: ConsoleConfig{Enabled: false},
		JSONL:   JSONLConfig{Enabled: true, Level: "debug", Path: jsonlPath},
		Redact:  RedactConfig{Enabled: true},
	}.Normalize())
	require.NoError(t, err)

	logger.Info(context.Background(), "jsonl ready", String("operation", "CreateParent"))
	require.NoError(t, logger.Sync())

	require.FileExists(t, jsonlPath)
	require.Contains(t, readFile(t, jsonlPath), `"msg":"jsonl ready"`)
}

// readFile 读取测试日志文件内容。
func readFile(t *testing.T, path string) string {
	t.Helper()
	content, err := os.ReadFile(path)
	require.NoError(t, err)
	return string(content)
}

// consoleStackLines 提取 console 输出中的短栈行。
func consoleStackLines(body string) []string {
	lines := strings.Split(body, "\n")
	stackIndex := -1
	for index, line := range lines {
		if strings.TrimSpace(line) == "stack:" {
			stackIndex = index
			break
		}
	}
	if stackIndex < 0 {
		return nil
	}
	result := make([]string, 0)
	for _, line := range lines[stackIndex+1:] {
		if strings.TrimSpace(line) == "" {
			continue
		}
		if !strings.HasPrefix(line, "  ") {
			break
		}
		result = append(result, strings.TrimSpace(line))
	}
	return result
}
