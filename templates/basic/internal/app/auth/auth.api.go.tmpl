package auth

import (
	"context"

	platformauth "github.com/teamsillybees/initra/pkg/auth"
	apperrors "github.com/teamsillybees/initra/pkg/errors"
	"github.com/teamsillybees/initra/pkg/requestctx"
	"github.com/teamsillybees/initra/pkg/response"
)

// LoginRequest 描述登录请求体。
type LoginRequest struct {
	Username string `json:"username" example:"alice"`
	Password string `json:"password" example:"secret-123"`
}

// LoginInput 描述登录接口输入。
type LoginInput struct {
	Body LoginRequest
}

// RefreshRequest 描述刷新 token 请求体。
type RefreshRequest struct {
	RefreshToken string `json:"refresh_token" example:"token"`
}

// RefreshInput 描述刷新接口输入。
type RefreshInput struct {
	Body RefreshRequest
}

// MeInput 仅作为当前用户接口的占位输入。
type MeInput struct{}

// LoginResponse 描述登录成功响应体。
type LoginResponse struct {
	AccessToken  string       `json:"access_token"`
	RefreshToken string       `json:"refresh_token"`
	User         UserIdentity `json:"user"`
}

// RefreshResponse 描述刷新 token 响应体。
type RefreshResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

// UserIdentity 描述当前登录用户的公开身份信息。
type UserIdentity struct {
	UserID       int64    `json:"user_id"`
	Username     string   `json:"username"`
	Nickname     string   `json:"nickname"`
	RoleCodes    []string `json:"role_codes"`
	IsSuperAdmin bool     `json:"is_super_admin"`
	IsEnable     bool     `json:"is_enable"`
}

// loginOutput 是 Huma 登录接口响应包装。
type loginOutput struct {
	Body response.SuccessResponse[LoginResponse]
}

// refreshOutput 是 Huma 刷新接口响应包装。
type refreshOutput struct {
	Body response.SuccessResponse[RefreshResponse]
}

// meOutput 是 Huma 当前用户接口响应包装。
type meOutput struct {
	Body response.SuccessResponse[UserIdentity]
}

// Handler 封装 auth 模块的 HTTP 适配逻辑。
type Handler struct {
	service *Service
}

// NewHandler 创建 auth 模块 Handler。
func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

// Login 执行账号密码登录。
func (h *Handler) Login(ctx context.Context, input *LoginInput) (*loginOutput, error) {
	result, err := h.service.Login(ctx, LoginParams{
		Username: input.Body.Username,
		Password: input.Body.Password,
	})
	if err != nil {
		return nil, err
	}

	return &loginOutput{
		Body: response.OK(requestctx.TraceIDFromContext(ctx), LoginResponse{
			AccessToken:  result.AccessToken,
			RefreshToken: result.RefreshToken,
			User:         toUserIdentity(result.User),
		}),
	}, nil
}

// Refresh 使用 opaque refresh token 轮转新的 token pair。
func (h *Handler) Refresh(ctx context.Context, input *RefreshInput) (*refreshOutput, error) {
	result, err := h.service.Refresh(ctx, input.Body.RefreshToken)
	if err != nil {
		return nil, err
	}

	return &refreshOutput{
		Body: response.OK(requestctx.TraceIDFromContext(ctx), RefreshResponse{
			AccessToken:  result.AccessToken,
			RefreshToken: result.RefreshToken,
		}),
	}, nil
}

// Me 返回当前登录用户信息。
func (h *Handler) Me(ctx context.Context, _ *MeInput) (*meOutput, error) {
	principal, ok := platformauth.PrincipalFromContext(ctx)
	if !ok {
		return nil, apperrors.New(apperrors.CodeUnauthorized, "user principal is missing")
	}

	user, err := h.service.Me(ctx, principal.UserID)
	if err != nil {
		return nil, err
	}

	return &meOutput{
		Body: response.OK(requestctx.TraceIDFromContext(ctx), toUserIdentity(user)),
	}, nil
}

func toUserIdentity(identity *Identity) UserIdentity {
	return UserIdentity{
		UserID:       identity.UserID,
		Username:     identity.Username,
		Nickname:     identity.Nickname,
		RoleCodes:    append([]string(nil), identity.RoleCodes...),
		IsSuperAdmin: identity.IsSuperAdmin,
		IsEnable:     identity.IsEnable,
	}
}
