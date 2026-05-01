package domain

import (
	"context"
	"errors"
	"strings"

	"github.com/teamsillybees/initra/examples/basic/internal/app/bizerrors"
	platformauth "github.com/teamsillybees/initra/pkg/auth"
	apperrors "github.com/teamsillybees/initra/pkg/errors"
)

// Repository 定义 auth 模块读取身份信息所需的最小能力。
type Repository interface {
	FindByUsername(ctx context.Context, username string) (*Identity, error)
	FindByID(ctx context.Context, id int64) (*Identity, error)
}

// PasswordVerifier 抽象密码校验行为，便于在测试中替换为轻量实现。
type PasswordVerifier interface {
	Hash(password string) (string, error)
	Compare(hash string, password string) error
}

// TokenManager 定义 auth 模块依赖的令牌能力。
type TokenManager interface {
	IssueTokenPair(ctx context.Context, principal platformauth.Principal) (platformauth.TokenPair, error)
	ConsumeRefreshToken(ctx context.Context, token string) (*platformauth.RefreshTokenRecord, error)
}

// Service 是 auth 模块的应用服务，负责登录、刷新与当前用户信息查询。
type Service struct {
	repo      Repository
	passwords PasswordVerifier
	tokens    TokenManager
}

// NewService 构造 auth 模块应用服务。
func NewService(repo Repository, passwords PasswordVerifier, tokens TokenManager) *Service {
	return &Service{
		repo:      repo,
		passwords: passwords,
		tokens:    tokens,
	}
}

// Login 校验账号密码并签发 access JWT 与 opaque refresh token。
func (s *Service) Login(ctx context.Context, input LoginInput) (*LoginResult, error) {
	if strings.TrimSpace(input.Username) == "" || strings.TrimSpace(input.Password) == "" {
		return nil, apperrors.New(apperrors.CodeBadRequest, "username and password are required")
	}

	identity, err := s.repo.FindByUsername(ctx, strings.TrimSpace(input.Username))
	if err != nil {
		return nil, err
	}
	if identity == nil || !identity.IsEnable {
		return nil, bizerrors.LoginFailed()
	}
	if err := s.passwords.Compare(identity.PasswordHash, input.Password); err != nil {
		return nil, bizerrors.LoginFailed()
	}

	tokenPair, err := s.tokens.IssueTokenPair(ctx, platformauth.Principal{
		UserID: identity.UserID,
		Roles:  append([]string(nil), identity.RoleCodes...),
	})
	if err != nil {
		return nil, apperrors.Wrap(err, apperrors.CodeInternalError, "issue token failed")
	}

	return &LoginResult{
		AccessToken:  tokenPair.AccessToken,
		RefreshToken: tokenPair.RefreshToken,
		User: AuthenticatedUser{
			UserID:       identity.UserID,
			Username:     identity.Username,
			Nickname:     identity.Nickname,
			RoleCodes:    append([]string(nil), identity.RoleCodes...),
			IsSuperAdmin: identity.IsSuperAdmin,
			IsEnable:     identity.IsEnable,
		},
	}, nil
}

// Refresh 校验 opaque refresh token 并轮转新的 token pair。
func (s *Service) Refresh(ctx context.Context, refreshToken string) (*RefreshResult, error) {
	if strings.TrimSpace(refreshToken) == "" {
		return nil, apperrors.New(apperrors.CodeBadRequest, "refresh token is required")
	}

	record, err := s.tokens.ConsumeRefreshToken(ctx, refreshToken)
	if err != nil {
		if errors.Is(err, platformauth.ErrTokenStoreFailure) {
			return nil, apperrors.Wrap(err, apperrors.CodeInternalError, "consume refresh token failed")
		}
		return nil, apperrors.New(apperrors.CodeUnauthorized, "refresh token is invalid")
	}

	identity, err := s.repo.FindByID(ctx, record.UserID)
	if err != nil {
		return nil, err
	}
	if identity == nil || !identity.IsEnable {
		return nil, apperrors.New(apperrors.CodeUnauthorized, "refresh token is invalid")
	}

	tokenPair, err := s.tokens.IssueTokenPair(ctx, platformauth.Principal{
		UserID: identity.UserID,
		Roles:  append([]string(nil), identity.RoleCodes...),
	})
	if err != nil {
		return nil, apperrors.Wrap(err, apperrors.CodeInternalError, "issue token failed")
	}

	return &RefreshResult{
		AccessToken:  tokenPair.AccessToken,
		RefreshToken: tokenPair.RefreshToken,
	}, nil
}

// Me 查询当前登录用户信息。
func (s *Service) Me(ctx context.Context, userID int64) (*AuthenticatedUser, error) {
	identity, err := s.repo.FindByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if identity == nil {
		return nil, bizerrors.UserNotFound(userID)
	}

	return &AuthenticatedUser{
		UserID:       identity.UserID,
		Username:     identity.Username,
		Nickname:     identity.Nickname,
		RoleCodes:    append([]string(nil), identity.RoleCodes...),
		IsSuperAdmin: identity.IsSuperAdmin,
		IsEnable:     identity.IsEnable,
	}, nil
}
