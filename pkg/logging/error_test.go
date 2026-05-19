package logging

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
	apperrors "github.com/teamsillybees/initra/pkg/errors"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"
)

// TestErrorFieldsExtractsAppAndOopsFields 验证平台错误日志会包含业务码、根因和 oops 栈。
func TestErrorFieldsExtractsAppAndOopsFields(t *testing.T) {
	cause := errors.New("driver: duplicate key")
	err := apperrors.Wrap(cause, apperrors.CodeDBError, "create user failed")
	core, logs := observer.New(zapcore.ErrorLevel)
	logger := zap.New(core)

	logger.Error("http request failed", ErrorFields(err)...)

	entries := logs.FilterMessage("http request failed").All()
	require.Len(t, entries, 1)
	fields := entries[0].ContextMap()
	require.Equal(t, "DB_ERROR", fields["error_code"])
	require.Equal(t, "create user failed", fields["error_message"])
	require.Equal(t, int64(500), fields["error_status"])
	require.Equal(t, "driver: duplicate key", fields["error_cause"])
	require.Contains(t, fields["error"].(string), "create user failed")
	require.NotEmpty(t, fields["error_stacktrace"])
}
