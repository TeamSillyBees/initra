package bizerrors

import (
	"context"
	"net/http"

	apperrors "github.com/teamsillybees/initra/pkg/errors"
	"github.com/teamsillybees/initra/pkg/idgen"
)

// 示例业务错误码独立于框架公共错误码，避免业务语义泄漏到 pkg API。
const (
	CodeUserNotFound apperrors.Code = "USER_NOT_FOUND"
	CodeLoginFailed  apperrors.Code = "LOGIN_FAILED"
)

// Option 表示业务错误构造选项。
type Option = apperrors.Option

// WithDetail 为业务错误补充单个详情字段。
// Details 面向客户端，只能放允许返回给前端的信息。
func WithDetail(key string, value any) Option {
	return apperrors.WithDetail(key, value)
}

// WithCauseAttr 为底层 cause 补充内部排障字段，只进入日志，不进入 HTTP 响应。
func WithCauseAttr(key string, value any) Option {
	return apperrors.WithCauseAttr(key, value)
}

// WithCauseAttrs 为底层 cause 补充一组内部排障字段，只进入日志，不进入 HTTP 响应。
func WithCauseAttrs(attrs map[string]any) Option {
	return apperrors.WithCauseAttrs(attrs)
}

// WithCauseDomain 为底层 cause 补充错误域，只进入日志，不进入 HTTP 响应。
func WithCauseDomain(domain string) Option {
	return apperrors.WithCauseDomain(domain)
}

// WithCauseHint 为底层 cause 补充排障提示，只进入日志，不进入 HTTP 响应。
func WithCauseHint(hint string) Option {
	return apperrors.WithCauseHint(hint)
}

// WithCauseTrace 为底层 cause 补充 trace id，只进入日志，不进入 HTTP 响应。
func WithCauseTrace(traceID string) Option {
	return apperrors.WithCauseTrace(traceID)
}

// BadRequest 创建统一的请求参数错误。
func BadRequest(message string, opts ...Option) error {
	return apperrors.New(apperrors.CodeBadRequest, message, opts...)
}

// Unauthorized 创建统一的未授权错误。
func Unauthorized(message string, opts ...Option) error {
	return apperrors.New(apperrors.CodeUnauthorized, message, opts...)
}

// Internal 创建统一的服务端内部错误。
func Internal(message string, opts ...Option) error {
	return apperrors.New(apperrors.CodeInternalError, message, opts...)
}

// WrapBadRequest 将底层错误封装为请求参数错误。
func WrapBadRequest(err error, message string, opts ...Option) error {
	return apperrors.Wrap(err, apperrors.CodeBadRequest, message, opts...)
}

// WrapBadRequestContext 将底层错误封装为请求参数错误，并自动写入 trace 元数据。
func WrapBadRequestContext(ctx context.Context, err error, message string, opts ...Option) error {
	return apperrors.WrapContext(ctx, err, apperrors.CodeBadRequest, message, opts...)
}

// WrapNotFound 将底层错误封装为资源不存在错误。
func WrapNotFound(err error, message string, opts ...Option) error {
	return apperrors.Wrap(err, apperrors.CodeNotFound, message, opts...)
}

// WrapNotFoundContext 将底层错误封装为资源不存在错误，并自动写入 trace 元数据。
func WrapNotFoundContext(ctx context.Context, err error, message string, opts ...Option) error {
	return apperrors.WrapContext(ctx, err, apperrors.CodeNotFound, message, opts...)
}

// WrapInternal 将底层错误封装为服务端内部错误。
func WrapInternal(err error, message string, opts ...Option) error {
	return apperrors.Wrap(err, apperrors.CodeInternalError, message, opts...)
}

// WrapInternalContext 将底层错误封装为服务端内部错误，并自动写入 trace 元数据。
func WrapInternalContext(ctx context.Context, err error, message string, opts ...Option) error {
	return apperrors.WrapContext(ctx, err, apperrors.CodeInternalError, message, opts...)
}

// WrapDB 将底层错误封装为数据库错误。
func WrapDB(err error, message string, opts ...Option) error {
	return apperrors.Wrap(err, apperrors.CodeDBError, message, withDefaults(dbDefaults(), opts)...)
}

// WrapDBContext 将底层错误封装为数据库错误，并自动写入 trace 元数据。
func WrapDBContext(ctx context.Context, err error, message string, opts ...Option) error {
	return apperrors.WrapContext(ctx, err, apperrors.CodeDBError, message, withDefaults(dbDefaults(), opts)...)
}

// WrapCache 将底层错误封装为缓存错误。
func WrapCache(err error, message string, opts ...Option) error {
	return apperrors.Wrap(err, apperrors.CodeCacheError, message, withDefaults(cacheDefaults(), opts)...)
}

// WrapCacheContext 将底层错误封装为缓存错误，并自动写入 trace 元数据。
func WrapCacheContext(ctx context.Context, err error, message string, opts ...Option) error {
	return apperrors.WrapContext(ctx, err, apperrors.CodeCacheError, message, withDefaults(cacheDefaults(), opts)...)
}

// WrapStorage 将底层错误封装为对象存储错误。
func WrapStorage(err error, message string, opts ...Option) error {
	return apperrors.Wrap(err, apperrors.CodeInternalError, message, withDefaults(storageDefaults(), opts)...)
}

// WrapStorageContext 将底层错误封装为对象存储错误，并自动写入 trace 元数据。
func WrapStorageContext(ctx context.Context, err error, message string, opts ...Option) error {
	return apperrors.WrapContext(ctx, err, apperrors.CodeInternalError, message, withDefaults(storageDefaults(), opts)...)
}

// WrapHTTPClient 将底层错误封装为下游 HTTP 调用错误。
func WrapHTTPClient(err error, message string, opts ...Option) error {
	return apperrors.Wrap(err, apperrors.CodeInternalError, message, withDefaults(httpClientDefaults(), opts)...)
}

// WrapHTTPClientContext 将底层错误封装为下游 HTTP 调用错误，并自动写入 trace 元数据。
func WrapHTTPClientContext(ctx context.Context, err error, message string, opts ...Option) error {
	return apperrors.WrapContext(ctx, err, apperrors.CodeInternalError, message, withDefaults(httpClientDefaults(), opts)...)
}

// WrapTask 将底层错误封装为任务队列错误。
func WrapTask(err error, message string, opts ...Option) error {
	return apperrors.Wrap(err, apperrors.CodeInternalError, message, withDefaults(taskDefaults(), opts)...)
}

// WrapTaskContext 将底层错误封装为任务队列错误，并自动写入 trace 元数据。
func WrapTaskContext(ctx context.Context, err error, message string, opts ...Option) error {
	return apperrors.WrapContext(ctx, err, apperrors.CodeInternalError, message, withDefaults(taskDefaults(), opts)...)
}

// LoginFailed 创建统一的登录失败错误。
func LoginFailed() error {
	return apperrors.New(CodeLoginFailed, "login failed", apperrors.WithStatus(http.StatusUnauthorized))
}

// UserNotFound 创建统一的用户不存在错误。
func UserNotFound(userID idgen.ID) error {
	return apperrors.New(
		CodeUserNotFound,
		"user not found",
		apperrors.WithStatus(http.StatusNotFound),
		apperrors.WithDetail("userId", userID),
	)
}

func withDefaults(defaults []Option, opts []Option) []Option {
	merged := make([]Option, 0, len(defaults)+len(opts))
	merged = append(merged, defaults...)
	merged = append(merged, opts...)
	return merged
}

func dbDefaults() []Option {
	return []Option{
		apperrors.WithCauseDomain(apperrors.DomainDB),
		apperrors.WithCauseHint(apperrors.HintDBConnection),
	}
}

func cacheDefaults() []Option {
	return []Option{
		apperrors.WithCauseDomain(apperrors.DomainCache),
		apperrors.WithCauseHint(apperrors.HintRedisTimeout),
	}
}

func storageDefaults() []Option {
	return []Option{
		apperrors.WithCauseDomain(apperrors.DomainStorage),
		apperrors.WithCauseHint(apperrors.HintStorageUpload),
	}
}

func httpClientDefaults() []Option {
	return []Option{
		apperrors.WithCauseDomain(apperrors.DomainHTTPClient),
		apperrors.WithCauseHint(apperrors.HintHTTPClientCall),
	}
}

func taskDefaults() []Option {
	return []Option{
		apperrors.WithCauseDomain(apperrors.DomainTask),
	}
}
