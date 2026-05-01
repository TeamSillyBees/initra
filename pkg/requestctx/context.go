package requestctx

import "context"

// requestContextKey 是当前包内部使用的上下文 key 类型，避免与外部字符串 key 冲突。
type requestContextKey string

// 请求上下文 key 常量集中维护 trace_id 与 request_id 的存储位置。
const (
	traceIDKey   requestContextKey = "trace_id"
	requestIDKey requestContextKey = "request_id"
)

// WithTraceID 将 trace_id 写入上下文，方便日志、错误响应和下游 service 共用。
func WithTraceID(ctx context.Context, traceID string) context.Context {
	return context.WithValue(ctx, traceIDKey, traceID)
}

// TraceIDFromContext 从上下文中提取 trace_id。
func TraceIDFromContext(ctx context.Context) string {
	value, _ := ctx.Value(traceIDKey).(string)
	return value
}

// WithRequestID 将 request_id 写入上下文。
func WithRequestID(ctx context.Context, requestID string) context.Context {
	return context.WithValue(ctx, requestIDKey, requestID)
}

// RequestIDFromContext 从上下文中提取 request_id。
func RequestIDFromContext(ctx context.Context) string {
	value, _ := ctx.Value(requestIDKey).(string)
	return value
}
