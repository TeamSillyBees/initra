package bizerrors

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
	apperrors "github.com/teamsillybees/initra/pkg/errors"
	"github.com/teamsillybees/initra/pkg/idgen"
	logfields "github.com/teamsillybees/initra/pkg/logging"
	"github.com/teamsillybees/initra/pkg/requestctx"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"
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
	err := UserNotFound(idgen.New(1001))

	require.Equal(t, CodeUserNotFound, err.Code)
	require.Equal(t, http.StatusNotFound, err.Status)
	require.Equal(t, idgen.New(1001), err.Details["userId"])
}

// TestWrapInternal 确认业务错误封装保留底层错误链和详情。
func TestWrapInternal(t *testing.T) {
	cause := errors.New("network failed")

	err := WrapInternal(cause, "call remote service failed", WithDetail("service", "httpbingo"))

	require.Equal(t, apperrors.CodeInternalError, err.Code)
	require.True(t, errors.Is(err, cause))
	require.Equal(t, "httpbingo", err.Details["service"])
}

// TestWrapDBContext 确认数据库错误会写入 trace、domain 和 hint 等内部 cause metadata。
func TestWrapDBContext(t *testing.T) {
	ctx := requestctx.WithTraceID(context.Background(), "trace-1")
	cause := errors.New("driver: duplicate key")

	err := WrapDBContext(ctx, cause, "query user failed", WithCauseAttr("sql", "select * from users"))
	core, logs := observer.New(zapcore.ErrorLevel)
	logger := zap.New(core)

	logger.Error("query failed", logfields.ErrorFields(err)...)

	entries := logs.FilterMessage("query failed").All()
	require.Len(t, entries, 1)
	fields := entries[0].ContextMap()
	require.Equal(t, apperrors.DomainDB, fields["error_domain"])
	require.Equal(t, apperrors.HintDBConnection, fields["error_hint"])
	require.Equal(t, "trace-1", fields["error_trace"])
	require.Equal(t, "[REDACTED]", fields["error_attrs"].(map[string]any)["sql"])
	require.Empty(t, err.Details)
}
