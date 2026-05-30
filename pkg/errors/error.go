package apperrors

import (
	"context"
	"errors"
	"strings"

	"github.com/samber/oops"
	"github.com/teamsillybees/initra/pkg/requestctx"
)

// AppError 是平台统一业务错误模型，负责承接业务错误码、HTTP 状态码、对外消息和对外细节。
// 底层 cause 通过 oops 增强后保留在 Cause 中，仅供日志、排障和 errors.Is/errors.As 使用。
type AppError struct {
	Code    Code
	Message string
	Status  int
	Details map[string]any
	Cause   error

	causeDomain string
	causeHint   string
	causeTrace  string
	causeAttrs  map[string]any
}

// Error 让 AppError 满足 error 接口。
func (e *AppError) Error() string {
	if e == nil {
		return ""
	}
	return e.Message
}

// Unwrap 暴露底层错误，便于 errors.Is / errors.As 继续工作。
func (e *AppError) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.Cause
}

// Option 用于以较小的成本扩展错误细节或底层 cause 调试元数据。
type Option func(*AppError)

// WithDetail 为错误响应补充单个细节字段。
func WithDetail(key string, value any) Option {
	return func(err *AppError) {
		if err.Details == nil {
			err.Details = map[string]any{}
		}
		err.Details[key] = value
	}
}

// WithDetails 为错误响应补充一组细节字段。
func WithDetails(details map[string]any) Option {
	return func(err *AppError) {
		if err.Details == nil {
			err.Details = map[string]any{}
		}
		for key, value := range details {
			err.Details[key] = value
		}
	}
}

// WithCauseDomain 为底层 cause 补充内部错误域，仅进入日志，不进入 HTTP 响应。
func WithCauseDomain(domain string) Option {
	return func(err *AppError) {
		err.causeDomain = strings.TrimSpace(domain)
	}
}

// WithCauseHint 为底层 cause 补充排障提示，仅进入日志，不进入 HTTP 响应。
func WithCauseHint(hint string) Option {
	return func(err *AppError) {
		err.causeHint = strings.TrimSpace(hint)
	}
}

// WithCauseTrace 为底层 cause 补充 trace id，仅进入日志，不进入 HTTP 响应。
func WithCauseTrace(traceID string) Option {
	return func(err *AppError) {
		err.causeTrace = strings.TrimSpace(traceID)
	}
}

// WithCauseAttr 为底层 cause 补充单个内部调试字段，仅进入日志，不进入 HTTP 响应。
func WithCauseAttr(key string, value any) Option {
	return func(err *AppError) {
		key = strings.TrimSpace(key)
		if key == "" {
			return
		}
		if err.causeAttrs == nil {
			err.causeAttrs = map[string]any{}
		}
		err.causeAttrs[key] = value
	}
}

// WithCauseAttrs 为底层 cause 补充一组内部调试字段，仅进入日志，不进入 HTTP 响应。
func WithCauseAttrs(attrs map[string]any) Option {
	return func(err *AppError) {
		if len(attrs) == 0 {
			return
		}
		if err.causeAttrs == nil {
			err.causeAttrs = map[string]any{}
		}
		for key, value := range attrs {
			key = strings.TrimSpace(key)
			if key == "" {
				continue
			}
			err.causeAttrs[key] = value
		}
	}
}

// WithStatus 允许调用方覆盖默认 HTTP 状态码。
func WithStatus(status int) Option {
	return func(err *AppError) {
		err.Status = status
	}
}

// New 创建一个新的平台错误。
func New(code Code, message string, opts ...Option) *AppError {
	appErr := &AppError{
		Code:    code,
		Message: message,
		Status:  statusOf(code),
		Details: map[string]any{},
	}
	for _, opt := range opts {
		opt(appErr)
	}
	return appErr
}

// Wrap 使用 oops 包装底层错误后，再转换成平台统一错误。
func Wrap(err error, code Code, message string, opts ...Option) *AppError {
	if err == nil {
		return New(code, message, opts...)
	}
	appErr := New(code, message, opts...)
	appErr.Cause = wrapCause(err, message, appErr)
	return appErr
}

// WrapContext 使用 oops 包装底层错误，并自动从 context 中提取 trace id 写入 cause metadata。
func WrapContext(ctx context.Context, err error, code Code, message string, opts ...Option) *AppError {
	if err == nil {
		return New(code, message, opts...)
	}
	appErr := New(code, message, opts...)
	if appErr.causeTrace == "" {
		if traceID, ok := requestctx.TraceIDFromContext(ctx); ok {
			appErr.causeTrace = strings.TrimSpace(traceID)
		}
	}
	appErr.Cause = wrapCause(err, message, appErr)
	return appErr
}

// From 尝试从任意错误链中提取 AppError。
func From(err error) *AppError {
	if err == nil {
		return nil
	}
	if appErr, ok := errors.AsType[*AppError](err); ok {
		return appErr
	}
	return nil
}

// statusOf 返回错误码对应的 HTTP 状态码，未知错误码默认视为服务端错误。
func statusOf(code Code) int {
	if status, ok := defaultStatuses[code]; ok {
		return status
	}
	return httpStatusInternalError
}

// httpStatusInternalError 避免在错误模型底层额外引入 net/http 依赖。
const httpStatusInternalError = 500

func wrapCause(err error, message string, appErr *AppError) error {
	builder := oops.With()
	if appErr.causeDomain != "" {
		builder = builder.In(appErr.causeDomain)
	}
	if appErr.causeHint != "" {
		builder = builder.Hint(appErr.causeHint)
	}
	if appErr.causeTrace != "" {
		builder = builder.Trace(appErr.causeTrace)
	}
	if len(appErr.causeAttrs) > 0 {
		builder = builder.With(causeAttrPairs(appErr.causeAttrs)...)
	}
	return builder.Wrapf(err, "%s", message)
}

func causeAttrPairs(attrs map[string]any) []any {
	pairs := make([]any, 0, len(attrs)*2)
	for key, value := range attrs {
		pairs = append(pairs, key, value)
	}
	return pairs
}
