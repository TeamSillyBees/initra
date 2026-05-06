package requestctx

import (
	"context"

	"go.uber.org/zap"
)

// requestContextKey 是当前包内部使用的上下文 key 类型，避免与外部字符串 key 冲突。
type requestContextKey string

// 请求上下文 key 常量集中维护请求级数据的存储位置。
const (
	requestIDKey requestContextKey = "request_id"
	traceIDKey   requestContextKey = "trace_id"
	userIDKey    requestContextKey = "user_id"
	tenantIDKey  requestContextKey = "tenant_id"
	sessionIDKey requestContextKey = "session_id"
	appIDKey     requestContextKey = "app_id"
)

// Values 表示可随请求上下文传递的请求级数据。
type Values struct {
	RequestID string
	TraceID   string
	UserID    string
	TenantID  string
	SessionID string
	AppID     string
}

// WithValues 将一组请求级数据写入上下文。
func WithValues(ctx context.Context, values Values) context.Context {
	ctx = WithRequestID(ctx, values.RequestID)
	ctx = WithTraceID(ctx, values.TraceID)
	ctx = WithUserID(ctx, values.UserID)
	ctx = WithTenantID(ctx, values.TenantID)
	ctx = WithSessionID(ctx, values.SessionID)
	ctx = WithAppID(ctx, values.AppID)
	return ctx
}

// ValuesFromContext 从上下文中提取请求级数据。
func ValuesFromContext(ctx context.Context) Values {
	return Values{
		RequestID: RequestIDFromContext(ctx),
		TraceID:   TraceIDFromContext(ctx),
		UserID:    UserIDFromContext(ctx),
		TenantID:  TenantIDFromContext(ctx),
		SessionID: SessionIDFromContext(ctx),
		AppID:     AppIDFromContext(ctx),
	}
}

// WithRequestID 将 request_id 写入上下文。
func WithRequestID(ctx context.Context, requestID string) context.Context {
	return withString(ctx, requestIDKey, requestID)
}

// RequestIDFromContext 从上下文中提取 request_id。
func RequestIDFromContext(ctx context.Context) string {
	return stringFromContext(ctx, requestIDKey)
}

// WithTraceID 将 trace_id 写入上下文，方便日志、错误响应和下游 service 共用。
func WithTraceID(ctx context.Context, traceID string) context.Context {
	return withString(ctx, traceIDKey, traceID)
}

// TraceIDFromContext 从上下文中提取 trace_id。
func TraceIDFromContext(ctx context.Context) string {
	return stringFromContext(ctx, traceIDKey)
}

// WithUserID 将 user_id 写入上下文。
func WithUserID(ctx context.Context, userID string) context.Context {
	return withString(ctx, userIDKey, userID)
}

// UserIDFromContext 从上下文中提取 user_id。
func UserIDFromContext(ctx context.Context) string {
	return stringFromContext(ctx, userIDKey)
}

// WithTenantID 将 tenant_id 写入上下文。
func WithTenantID(ctx context.Context, tenantID string) context.Context {
	return withString(ctx, tenantIDKey, tenantID)
}

// TenantIDFromContext 从上下文中提取 tenant_id。
func TenantIDFromContext(ctx context.Context) string {
	return stringFromContext(ctx, tenantIDKey)
}

// WithSessionID 将 session_id 写入上下文。
func WithSessionID(ctx context.Context, sessionID string) context.Context {
	return withString(ctx, sessionIDKey, sessionID)
}

// SessionIDFromContext 从上下文中提取 session_id。
func SessionIDFromContext(ctx context.Context) string {
	return stringFromContext(ctx, sessionIDKey)
}

// WithAppID 将 app_id 写入上下文。
func WithAppID(ctx context.Context, appID string) context.Context {
	return withString(ctx, appIDKey, appID)
}

// AppIDFromContext 从上下文中提取 app_id。
func AppIDFromContext(ctx context.Context) string {
	return stringFromContext(ctx, appIDKey)
}

// LogFields 根据请求上下文构造 zap 日志字段，空值不会输出。
func LogFields(ctx context.Context) []zap.Field {
	values := ValuesFromContext(ctx)
	fields := make([]zap.Field, 0, 6)
	if values.RequestID != "" {
		fields = append(fields, zap.String("request_id", values.RequestID))
	}
	if values.TraceID != "" {
		fields = append(fields, zap.String("trace_id", values.TraceID))
	}
	if values.UserID != "" {
		fields = append(fields, zap.String("user_id", values.UserID))
	}
	if values.TenantID != "" {
		fields = append(fields, zap.String("tenant_id", values.TenantID))
	}
	if values.SessionID != "" {
		fields = append(fields, zap.String("session_id", values.SessionID))
	}
	if values.AppID != "" {
		fields = append(fields, zap.String("app_id", values.AppID))
	}
	return fields
}

func withString(ctx context.Context, key requestContextKey, value string) context.Context {
	return context.WithValue(ctx, key, value)
}

func stringFromContext(ctx context.Context, key requestContextKey) string {
	if ctx == nil {
		return ""
	}
	value, _ := ctx.Value(key).(string)
	return value
}
