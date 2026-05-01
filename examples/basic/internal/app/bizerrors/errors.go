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
