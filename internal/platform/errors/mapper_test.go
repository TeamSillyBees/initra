package apperrors

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
)

// TestToHTTPMapsBusinessError 验证平台错误能映射为统一 HTTP 错误响应。
func TestToHTTPMapsBusinessError(t *testing.T) {
	err := New(
		CodeUserNotFound,
		"user not found",
		WithDetail("user_id", int64(1001)),
	)

	status, body := ToHTTP(err, "trace-001")

	require.Equal(t, http.StatusNotFound, status)
	require.Equal(t, "USER_NOT_FOUND", body.Code)
	require.Equal(t, "user not found", body.Message)
	require.Equal(t, "trace-001", body.TraceID)
	require.Equal(t, int64(1001), body.Details["user_id"])
}
