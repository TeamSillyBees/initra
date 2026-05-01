package apperrors

// ErrorResponse 是脚手架统一错误响应体。
type ErrorResponse struct {
	Code    string         `json:"code"`
	Message string         `json:"message"`
	Details map[string]any `json:"details,omitempty"`
	TraceID string         `json:"trace_id,omitempty"`
}

// ToHTTP 将任意 error 归一化为 HTTP 状态码和统一响应体。
func ToHTTP(err error, traceID string) (int, ErrorResponse) {
	appErr := From(err)
	if appErr == nil {
		return statusOf(CodeInternalError), ErrorResponse{
			Code:    string(CodeInternalError),
			Message: "internal error",
			TraceID: traceID,
		}
	}

	return appErr.Status, ErrorResponse{
		Code:    string(appErr.Code),
		Message: appErr.Message,
		Details: appErr.Details,
		TraceID: traceID,
	}
}
