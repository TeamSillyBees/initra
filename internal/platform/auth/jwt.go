package auth

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// Token 类型常量用于区分访问令牌和刷新令牌，避免 refresh token 被误用于访问接口。
const (
	// TokenTypeAccess 表示短期访问令牌。
	TokenTypeAccess = "access"
	// TokenTypeRefresh 表示长期刷新令牌。
	TokenTypeRefresh = "refresh"
)

// JWTConfig 描述 JWT 服务的最小配置输入。
type JWTConfig struct {
	Issuer          string
	Secret          string
	AccessTokenTTL  time.Duration
	RefreshTokenTTL time.Duration
	Store           TokenStore
	Now             func() time.Time
}

// Principal 表示当前登录用户在请求链路中的最小身份载体。
type Principal struct {
	UserID   int64
	Roles    []string
	TenantID string
}

// Claims 是脚手架统一的 JWT Claims。
type Claims struct {
	UserID    int64    `json:"user_id"`
	Roles     []string `json:"roles"`
	TenantID  string   `json:"tenant_id,omitempty"`
	TokenType string   `json:"token_type"`
	jwt.RegisteredClaims
}

// TokenPair 表示一次签发返回的访问令牌与刷新令牌。
type TokenPair struct {
	AccessToken      string    `json:"access_token"`
	RefreshToken     string    `json:"refresh_token"`
	AccessExpiresAt  time.Time `json:"access_expires_at"`
	RefreshExpiresAt time.Time `json:"refresh_expires_at"`
}

// JWTManager 负责签发与解析 JWT。
type JWTManager struct {
	issuer          string
	secret          []byte
	accessTokenTTL  time.Duration
	refreshTokenTTL time.Duration
	store           TokenStore
	now             func() time.Time
}

// JWT 错误哨兵用于让上层区分 token 本身无效、已被吊销和状态存储异常三类场景。
var (
	// ErrTokenInvalid 表示 token 本身签名、类型、时效或声明不合法。
	ErrTokenInvalid = errors.New("token invalid")
	// ErrTokenRevoked 表示 token 已被服务端主动吊销或 refresh token 已失效。
	ErrTokenRevoked = errors.New("token revoked")
	// ErrTokenStoreFailure 表示 Redis 等状态存储校验失败，无法安全完成 token 生命周期判断。
	ErrTokenStoreFailure = errors.New("token store failure")
)

// NewJWTManager 构造 JWT 管理器。
func NewJWTManager(cfg JWTConfig) (*JWTManager, error) {
	switch {
	case cfg.Issuer == "":
		return nil, errors.New("issuer 不能为空")
	case cfg.Secret == "":
		return nil, errors.New("secret 不能为空")
	case cfg.AccessTokenTTL <= 0:
		return nil, errors.New("access token ttl 必须大于 0")
	case cfg.RefreshTokenTTL <= 0:
		return nil, errors.New("refresh token ttl 必须大于 0")
	case cfg.RefreshTokenTTL <= cfg.AccessTokenTTL:
		return nil, errors.New("refresh token ttl 必须大于 access token ttl")
	default:
		if cfg.Now == nil {
			cfg.Now = time.Now
		}
		return &JWTManager{
			issuer:          cfg.Issuer,
			secret:          []byte(cfg.Secret),
			accessTokenTTL:  cfg.AccessTokenTTL,
			refreshTokenTTL: cfg.RefreshTokenTTL,
			store:           cfg.Store,
			now:             cfg.Now,
		}, nil
	}
}

// IssueTokenPair 为指定身份生成一对可直接返回给客户端的 JWT。
func (m *JWTManager) IssueTokenPair(ctx context.Context, principal Principal) (TokenPair, error) {
	now := m.now()

	accessExpiresAt := now.Add(m.accessTokenTTL)
	refreshExpiresAt := now.Add(m.refreshTokenTTL)

	accessToken, err := m.issue(principal, TokenTypeAccess, accessExpiresAt, now)
	if err != nil {
		return TokenPair{}, err
	}

	refreshToken, err := m.issue(principal, TokenTypeRefresh, refreshExpiresAt, now)
	if err != nil {
		return TokenPair{}, err
	}

	refreshClaims, err := m.parse(refreshToken, TokenTypeRefresh)
	if err != nil {
		return TokenPair{}, err
	}
	if m.store != nil {
		if err := m.store.StoreRefreshToken(ctx, refreshClaims.ID, refreshClaims.UserID, refreshExpiresAt.Sub(now)); err != nil {
			return TokenPair{}, fmt.Errorf("%w: store refresh token: %v", ErrTokenStoreFailure, err)
		}
	}

	return TokenPair{
		AccessToken:      accessToken,
		RefreshToken:     refreshToken,
		AccessExpiresAt:  accessExpiresAt,
		RefreshExpiresAt: refreshExpiresAt,
	}, nil
}

// ParseAccessToken 解析并验证访问令牌。
func (m *JWTManager) ParseAccessToken(ctx context.Context, token string) (*Claims, error) {
	claims, err := m.parse(token, TokenTypeAccess)
	if err != nil {
		return nil, err
	}
	if m.store == nil {
		return claims, nil
	}

	blacklisted, err := m.store.IsAccessTokenBlacklisted(ctx, claims.ID)
	if err != nil {
		return nil, fmt.Errorf("%w: check access token blacklist: %v", ErrTokenStoreFailure, err)
	}
	if blacklisted {
		return nil, ErrTokenRevoked
	}
	return claims, nil
}

// ParseRefreshToken 解析并验证刷新令牌。
func (m *JWTManager) ParseRefreshToken(ctx context.Context, token string) (*Claims, error) {
	claims, err := m.parse(token, TokenTypeRefresh)
	if err != nil {
		return nil, err
	}
	if m.store == nil {
		return claims, nil
	}

	valid, err := m.store.ValidateRefreshToken(ctx, claims.ID, claims.UserID)
	if err != nil {
		return nil, fmt.Errorf("%w: validate refresh token: %v", ErrTokenStoreFailure, err)
	}
	if !valid {
		return nil, ErrTokenRevoked
	}
	return claims, nil
}

// ConsumeRefreshToken 原子校验并消费 refresh token，用于 refresh token 轮转场景。
func (m *JWTManager) ConsumeRefreshToken(ctx context.Context, token string) (*Claims, error) {
	claims, err := m.parse(token, TokenTypeRefresh)
	if err != nil {
		return nil, err
	}
	if m.store == nil {
		return claims, nil
	}

	valid, err := m.store.ConsumeRefreshToken(ctx, claims.ID, claims.UserID)
	if err != nil {
		return nil, fmt.Errorf("%w: consume refresh token: %v", ErrTokenStoreFailure, err)
	}
	if !valid {
		return nil, ErrTokenRevoked
	}
	return claims, nil
}

// BlacklistAccessToken 把 access token 加入 Redis 黑名单，TTL 等于该 token 剩余寿命。
func (m *JWTManager) BlacklistAccessToken(ctx context.Context, token string) error {
	if m.store == nil {
		return fmt.Errorf("%w: access token blacklist store is not configured", ErrTokenStoreFailure)
	}

	claims, err := m.parse(token, TokenTypeAccess)
	if err != nil {
		return err
	}
	if claims.ExpiresAt == nil {
		return fmt.Errorf("%w: access token expires_at is missing", ErrTokenInvalid)
	}

	ttl := claims.ExpiresAt.Time.Sub(m.now())
	if ttl <= 0 {
		return nil
	}
	if err := m.store.BlacklistAccessToken(ctx, claims.ID, ttl); err != nil {
		return fmt.Errorf("%w: blacklist access token: %v", ErrTokenStoreFailure, err)
	}
	return nil
}

// issue 根据统一 Claims 结构签发指定类型 token。
// 调用方必须传入同一个 now 值，确保 iat、nbf 和 exp 基于一致时钟计算。
func (m *JWTManager) issue(principal Principal, tokenType string, expiresAt, now time.Time) (string, error) {
	claims := Claims{
		UserID:    principal.UserID,
		Roles:     append([]string(nil), principal.Roles...),
		TenantID:  principal.TenantID,
		TokenType: tokenType,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    m.issuer,
			Subject:   strconv.FormatInt(principal.UserID, 10),
			ID:        uuid.NewString(),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(expiresAt),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(m.secret)
}

// parse 校验签名、签发方、算法、过期时间和 token 类型，并返回业务 Claims。
// 这里强制使用注入时钟，避免测试时钟和真实时钟混用造成生命周期判断漂移。
func (m *JWTManager) parse(token string, expectedType string) (*Claims, error) {
	parsed, err := jwt.ParseWithClaims(
		token,
		&Claims{},
		func(token *jwt.Token) (any, error) {
			return m.secret, nil
		},
		jwt.WithIssuer(m.issuer),
		jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}),
		jwt.WithExpirationRequired(),
		jwt.WithIssuedAt(),
		jwt.WithTimeFunc(m.now),
	)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrTokenInvalid, err)
	}

	claims, ok := parsed.Claims.(*Claims)
	if !ok || !parsed.Valid {
		return nil, fmt.Errorf("%w: token 无效", ErrTokenInvalid)
	}
	if claims.TokenType != expectedType {
		return nil, fmt.Errorf("%w: token type mismatch: expect=%s actual=%s", ErrTokenInvalid, expectedType, claims.TokenType)
	}
	if claims.ID == "" {
		return nil, fmt.Errorf("%w: token id is missing", ErrTokenInvalid)
	}
	if claims.Subject != strconv.FormatInt(claims.UserID, 10) {
		return nil, fmt.Errorf("%w: token subject mismatch", ErrTokenInvalid)
	}
	return claims, nil
}

// principalContextKey 是请求上下文中保存 Principal 的私有 key，避免与外部包 key 冲突。
type principalContextKey struct{}

// WithPrincipal 将登录用户信息写入上下文，便于后续中间件和 service 统一读取。
func WithPrincipal(ctx context.Context, principal Principal) context.Context {
	return context.WithValue(ctx, principalContextKey{}, principal)
}

// PrincipalFromContext 从上下文中提取当前登录用户。
func PrincipalFromContext(ctx context.Context) (Principal, bool) {
	principal, ok := ctx.Value(principalContextKey{}).(Principal)
	return principal, ok
}
