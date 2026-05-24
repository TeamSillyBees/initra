package bizerrors

import (
	"errors"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
	apperrors "github.com/teamsillybees/initra/pkg/errors"
)

// TestLoginFailed 使用独立业务码表达统一的登录失败语义。
func TestLoginFailed(t *testing.T) {
	err := LoginFailed()

	require.Equal(t, CodeLoginFailed, err.Code)
	require.Equal(t, http.StatusUnauthorized, err.Status)
	require.Equal(t, "login failed", err.Message)
}

// TestUserNotFound 确认用户不存在错误携带用户 ID 详情。
func TestUserNotFound(t *testing.T) {
	err := UserNotFound(1001)

	require.Equal(t, CodeUserNotFound, err.Code)
	require.Equal(t, http.StatusNotFound, err.Status)
	require.Equal(t, int64(1001), err.Details["userId"])
}

// TestWrapInternal 确认业务错误封装保留底层错误链和详情。
func TestWrapInternal(t *testing.T) {
	cause := errors.New("network failed")

	err := WrapInternal(cause, "call remote service failed", WithDetail("service", "httpbingo"))

	require.Equal(t, apperrors.CodeInternalError, err.Code)
	require.True(t, errors.Is(err, cause))
	require.Equal(t, "httpbingo", err.Details["service"])
}
