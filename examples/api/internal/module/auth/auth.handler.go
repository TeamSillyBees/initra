package auth

import (
	"context"

	platformauth "github.com/teamsillybees/initra/pkg/auth"
	apperrors "github.com/teamsillybees/initra/pkg/errors"
	"github.com/teamsillybees/initra/pkg/requestctx"
	"github.com/teamsillybees/initra/pkg/response"
)

// Handler 封装 auth 模块的 HTTP 适配逻辑。
type Handler struct {
	service *Service
}

// NewHandler 创建 auth 模块 Handler。
func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) login(ctx context.Context, input *loginRequest) (*loginResponse, error) {
	identity, tokenPair, err := h.service.Login(ctx, LoginDTO{
		Username: input.Body.Username,
		Password: input.Body.Password,
	})
	if err != nil {
		return nil, err
	}

	return &loginResponse{
		Body: response.OK(requestctx.TraceIDFromContext(ctx), LoginVO{
			AccessToken:  tokenPair.AccessToken,
			RefreshToken: tokenPair.RefreshToken,
			User:         toUserIdentityVO(identity),
		}),
	}, nil
}

func (h *Handler) refresh(ctx context.Context, input *refreshRequest) (*refreshResponse, error) {
	tokenPair, err := h.service.Refresh(ctx, input.Body.RefreshToken)
	if err != nil {
		return nil, err
	}

	return &refreshResponse{
		Body: response.OK(requestctx.TraceIDFromContext(ctx), RefreshVO{
			AccessToken:  tokenPair.AccessToken,
			RefreshToken: tokenPair.RefreshToken,
		}),
	}, nil
}

func (h *Handler) me(ctx context.Context, _ *meRequest) (*meResponse, error) {
	principal, ok := platformauth.PrincipalFromContext(ctx)
	if !ok {
		return nil, apperrors.New(apperrors.CodeUnauthorized, "user principal is missing")
	}

	user, err := h.service.Me(ctx, principal.UserID)
	if err != nil {
		return nil, err
	}

	return &meResponse{
		Body: response.OK(requestctx.TraceIDFromContext(ctx), toUserIdentityVO(user)),
	}, nil
}

func toUserIdentityVO(identity *Identity) UserIdentityVO {
	return UserIdentityVO{
		UserID:       identity.UserID,
		Username:     identity.Username,
		Nickname:     identity.Nickname,
		RoleCodes:    append([]string(nil), identity.RoleCodes...),
		IsSuperAdmin: identity.IsSuperAdmin,
		IsEnable:     identity.IsEnable,
	}
}
