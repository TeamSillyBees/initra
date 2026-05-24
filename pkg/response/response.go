package response

// SuccessVO 是脚手架统一成功响应 JSON DTO。
type SuccessVO[T any] struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Data    T      `json:"data"`
	TraceID string `json:"traceId,omitempty"`
}

// OK 创建标准成功响应。
func OK[T any](traceID string, data T) SuccessVO[T] {
	return SuccessVO[T]{
		Code:    "OK",
		Message: "success",
		Data:    data,
		TraceID: traceID,
	}
}
