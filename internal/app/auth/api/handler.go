package api

import (
	"context"

	"github.com/teamsillybees/initra/internal/app/auth/domain"
	platformauth "github.com/teamsillybees/initra/internal/platform/auth"
	apperrors "github.com/teamsillybees/initra/internal/platform/errors"
	sharedtypes "github.com/teamsillybees/initra/internal/shared/types"
)

// Handler 封装 auth 模块的 HTTP 适配逻辑。
type Handler struct {
	service *domain.Service
}

// NewHandler 创建 auth 模块 Handler。
func NewHandler(service *domain.Service) *Handler {
	return &Handler{service: service}
}

// Login 执行账号密码登录。
func (h *Handler) Login(ctx context.Context, input *LoginInput) (*loginOutput, error) {
	result, err := h.service.Login(ctx, domain.LoginInput{
		Username: input.Body.Username,
		Password: input.Body.Password,
	})
	if err != nil {
		return nil, err
	}

	return &loginOutput{
		Body: sharedtypes.OK(sharedtypes.TraceIDFromContext(ctx), LoginResponse{
			AccessToken:  result.AccessToken,
			RefreshToken: result.RefreshToken,
			User: UserClaims{
				UserID:       result.User.UserID,
				Username:     result.User.Username,
				Nickname:     result.User.Nickname,
				RoleCodes:    append([]string(nil), result.User.RoleCodes...),
				IsSuperAdmin: result.User.IsSuperAdmin,
				IsEnable:     result.User.IsEnable,
			},
		}),
	}, nil
}

// Refresh 刷新一对 JWT。
func (h *Handler) Refresh(ctx context.Context, input *RefreshInput) (*refreshOutput, error) {
	result, err := h.service.Refresh(ctx, input.Body.RefreshToken)
	if err != nil {
		return nil, err
	}

	return &refreshOutput{
		Body: sharedtypes.OK(sharedtypes.TraceIDFromContext(ctx), RefreshResponse{
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
		Body: sharedtypes.OK(sharedtypes.TraceIDFromContext(ctx), UserClaims{
			UserID:       user.UserID,
			Username:     user.Username,
			Nickname:     user.Nickname,
			RoleCodes:    append([]string(nil), user.RoleCodes...),
			IsSuperAdmin: user.IsSuperAdmin,
			IsEnable:     user.IsEnable,
		}),
	}, nil
}
