package logx

import (
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// TestRedactFieldsMasksSensitiveKeys 验证结构化敏感字段会被统一脱敏。
func TestRedactFieldsMasksSensitiveKeys(t *testing.T) {
	fields := RedactFields([]zap.Field{
		zap.String("password", "plain-password"),
		zap.String("authorization", "Bearer plain-token"),
		zap.String("username", "alice"),
	}, RedactConfig{Enabled: true})

	got := stringFields(fields)
	require.Equal(t, RedactedValue, got["password"])
	require.Equal(t, RedactedValue, got["authorization"])
	require.Equal(t, "alice", got["username"])
}

// TestRedactValueSanitizesNestedData 验证 map 和 slice 中的敏感字段会递归脱敏。
func TestRedactValueSanitizesNestedData(t *testing.T) {
	got := RedactValue("payload", map[string]any{
		"token": "secret-token",
		"profile": map[string]any{
			"email":    "alice@example.com",
			"nickname": "alice",
		},
	}).(map[string]any)

	require.Equal(t, RedactedValue, got["token"])
	require.Equal(t, RedactedValue, got["profile"].(map[string]any)["email"])
	require.Equal(t, "alice", got["profile"].(map[string]any)["nickname"])
}

// TestRedactTextMasksAssignments 验证字符串中的敏感赋值片段会被脱敏。
func TestRedactTextMasksAssignments(t *testing.T) {
	got := RedactText("dsn=postgres://user:pass@localhost token:abc username=alice")

	require.Contains(t, got, "dsn="+RedactedValue)
	require.Contains(t, got, "token:"+RedactedValue)
	require.Contains(t, got, "username=alice")
	require.NotContains(t, got, "abc")
	require.NotContains(t, got, "postgres://")
}

// stringFields 提取测试中需要断言的字符串字段。
func stringFields(fields []zap.Field) map[string]string {
	result := make(map[string]string, len(fields))
	for _, field := range fields {
		if field.Type == zapcore.StringType {
			result[field.Key] = field.String
		}
	}
	return result
}
