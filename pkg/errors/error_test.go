package apperrors

import (
	"context"
	"errors"
	"testing"

	"github.com/samber/oops"
	"github.com/stretchr/testify/require"
	"github.com/teamsillybees/initra/pkg/requestctx"
)

type typedCause struct {
	message string
}

func (e *typedCause) Error() string {
	return e.message
}

// TestAppErrorErrorDoesNotExposeCause 验证 Error 文本只返回对外 message，不泄露底层错误。
func TestAppErrorErrorDoesNotExposeCause(t *testing.T) {
	cause := &typedCause{message: "driver: duplicate key"}
	err := Wrap(cause, CodeDBError, "create user failed")

	require.Equal(t, "create user failed", err.Error())
	require.NotContains(t, err.Error(), "duplicate key")
}

// TestAppErrorUnwrapKeepsStandardErrorChain 验证底层 cause 仍可通过标准错误链路识别。
func TestAppErrorUnwrapKeepsStandardErrorChain(t *testing.T) {
	cause := &typedCause{message: "driver: duplicate key"}
	err := Wrap(cause, CodeDBError, "create user failed")

	require.NotNil(t, errors.Unwrap(err))
	require.True(t, errors.Is(err, cause))

	var target *typedCause
	require.True(t, errors.As(err, &target))
	require.Same(t, cause, target)
}

// TestWrapContextWritesCauseMetadata 验证 WrapContext 只把内部元数据写入 oops cause。
func TestWrapContextWritesCauseMetadata(t *testing.T) {
	ctx := requestctx.WithTraceID(context.Background(), "trace-1")
	err := WrapContext(ctx, errors.New("redis timeout"), CodeCacheError, "read cache failed",
		WithCauseDomain(DomainCache),
		WithCauseHint(HintRedisTimeout),
		WithCauseAttr("operation", "get_user"),
		WithCauseAttr("token", "secret-token"),
	)

	oopsErr, ok := oops.AsOops(err)
	require.True(t, ok)
	require.Equal(t, DomainCache, oopsErr.Domain())
	require.Equal(t, HintRedisTimeout, oopsErr.Hint())
	require.Equal(t, "trace-1", oopsErr.Trace())
	require.Equal(t, "get_user", oopsErr.Context()["operation"])
	require.Equal(t, "secret-token", oopsErr.Context()["token"])
	require.Empty(t, err.Details)
}
