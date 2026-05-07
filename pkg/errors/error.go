package apperrors

import (
	"errors"

	"github.com/samber/oops"
)

// AppError 是平台统一错误模型，负责承接错误码、消息、细节与底层原因。
type AppError struct {
	Code    Code
	Message string
	Status  int
	Details map[string]any
	Cause   error
}

// Error 让 AppError 满足 error 接口。
func (e *AppError) Error() string {
	if e == nil {
		return ""
	}
	if e.Cause != nil {
		return e.Cause.Error()
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

// Option 用于以较小的成本扩展错误细节。
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
	appErr.Cause = oops.Wrapf(err, "%s", message)
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
