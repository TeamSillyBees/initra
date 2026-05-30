package apperrors

// ErrorVO 是脚手架统一错误响应 JSON DTO。
type ErrorVO struct {
	Code    string         `json:"code"`
	Message string         `json:"message"`
	Details map[string]any `json:"details,omitempty"`
	TraceID string         `json:"traceId,omitempty"`
}

// ToHTTP 将任意 error 归一化为 HTTP 状态码和统一响应体。
func ToHTTP(err error, traceID string) (int, ErrorVO) {
	appErr := From(err)
	if appErr == nil {
		return statusOf(CodeInternalError), ErrorVO{
			Code:    string(CodeInternalError),
			Message: "internal error",
			TraceID: traceID,
		}
	}

	var details map[string]any
	if len(appErr.Details) > 0 {
		details = SanitizeMap(appErr.Details)
	}
	return appErr.Status, ErrorVO{
		Code:    string(appErr.Code),
		Message: appErr.Message,
		Details: details,
		TraceID: traceID,
	}
}
