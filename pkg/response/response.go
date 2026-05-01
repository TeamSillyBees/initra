package response

// SuccessResponse 是脚手架统一成功响应结构。
type SuccessResponse[T any] struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Data    T      `json:"data"`
	TraceID string `json:"trace_id,omitempty"`
}

// OK 创建标准成功响应。
func OK[T any](traceID string, data T) SuccessResponse[T] {
	return SuccessResponse[T]{
		Code:    "OK",
		Message: "success",
		Data:    data,
		TraceID: traceID,
	}
}
