package apperrors

import "net/http"

// Code 定义全局统一错误码，所有平台错误和业务错误都应从这里衍生。
type Code string

// 标准错误码常量是 HTTP 层、业务层和测试断言之间的稳定契约。
const (
	CodeOK            Code = "OK"
	CodeBadRequest    Code = "BAD_REQUEST"
	CodeUnauthorized  Code = "UNAUTHORIZED"
	CodeForbidden     Code = "FORBIDDEN"
	CodeNotFound      Code = "NOT_FOUND"
	CodeInternalError Code = "INTERNAL_ERROR"
	CodeDBError       Code = "DB_ERROR"
	CodeCacheError    Code = "CACHE_ERROR"
	CodeUserNotFound  Code = "USER_NOT_FOUND"
	CodeLoginFailed   Code = "LOGIN_FAILED"
)

// defaultStatuses 描述错误码到 HTTP 状态码的默认映射。
var defaultStatuses = map[Code]int{
	CodeBadRequest:    http.StatusBadRequest,
	CodeUnauthorized:  http.StatusUnauthorized,
	CodeForbidden:     http.StatusForbidden,
	CodeNotFound:      http.StatusNotFound,
	CodeInternalError: http.StatusInternalServerError,
	CodeDBError:       http.StatusInternalServerError,
	CodeCacheError:    http.StatusInternalServerError,
	CodeUserNotFound:  http.StatusNotFound,
	CodeLoginFailed:   http.StatusUnauthorized,
}
