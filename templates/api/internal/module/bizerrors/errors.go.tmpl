package bizerrors

import (
	"net/http"

	apperrors "github.com/teamsillybees/initra/pkg/errors"
)

// 示例业务错误码独立于框架公共错误码，避免业务语义泄漏到 pkg API。
const (
	CodeUserNotFound apperrors.Code = "USER_NOT_FOUND"
	CodeLoginFailed  apperrors.Code = "LOGIN_FAILED"
)

// Option 表示业务错误构造选项。
type Option = apperrors.Option

// WithDetail 为业务错误补充单个详情字段。
func WithDetail(key string, value any) Option {
	return apperrors.WithDetail(key, value)
}

// BadRequest 创建统一的请求参数错误。
func BadRequest(message string, opts ...Option) *apperrors.AppError {
	return apperrors.New(apperrors.CodeBadRequest, message, opts...)
}

// Unauthorized 创建统一的未授权错误。
func Unauthorized(message string, opts ...Option) *apperrors.AppError {
	return apperrors.New(apperrors.CodeUnauthorized, message, opts...)
}

// Internal 创建统一的服务端内部错误。
func Internal(message string, opts ...Option) *apperrors.AppError {
	return apperrors.New(apperrors.CodeInternalError, message, opts...)
}

// WrapBadRequest 将底层错误封装为请求参数错误。
func WrapBadRequest(err error, message string, opts ...Option) *apperrors.AppError {
	return apperrors.Wrap(err, apperrors.CodeBadRequest, message, opts...)
}

// WrapNotFound 将底层错误封装为资源不存在错误。
func WrapNotFound(err error, message string, opts ...Option) *apperrors.AppError {
	return apperrors.Wrap(err, apperrors.CodeNotFound, message, opts...)
}

// WrapInternal 将底层错误封装为服务端内部错误。
func WrapInternal(err error, message string, opts ...Option) *apperrors.AppError {
	return apperrors.Wrap(err, apperrors.CodeInternalError, message, opts...)
}

// WrapDB 将底层错误封装为数据库错误。
func WrapDB(err error, message string, opts ...Option) *apperrors.AppError {
	return apperrors.Wrap(err, apperrors.CodeDBError, message, opts...)
}

// WrapCache 将底层错误封装为缓存错误。
func WrapCache(err error, message string, opts ...Option) *apperrors.AppError {
	return apperrors.Wrap(err, apperrors.CodeCacheError, message, opts...)
}

// LoginFailed 创建统一的登录失败错误。
func LoginFailed() *apperrors.AppError {
	return apperrors.New(CodeLoginFailed, "login failed", apperrors.WithStatus(http.StatusUnauthorized))
}

// UserNotFound 创建统一的用户不存在错误。
func UserNotFound(userID int64) *apperrors.AppError {
	return apperrors.New(
		CodeUserNotFound,
		"user not found",
		apperrors.WithStatus(http.StatusNotFound),
		apperrors.WithDetail("user_id", userID),
	)
}
