package auth

import (
	"context"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/require"
)

// fakeTokenStore 记录 JWTManager 对 TokenStore 的调用，便于断言 Redis TTL 和黑名单行为。
type fakeTokenStore struct {
	storedRefreshTokenID string
	storedRefreshUserID  int64
	storedRefreshTTL     time.Duration
	revokedRefreshIDs    []string
	blacklistedTokenID   string
	blacklistedTTL       time.Duration
	refreshValid         bool
	accessBlacklisted    bool
}

// StoreRefreshToken 保存 refresh token 写入参数。
func (f *fakeTokenStore) StoreRefreshToken(_ context.Context, tokenID string, userID int64, ttl time.Duration) error {
	f.storedRefreshTokenID = tokenID
	f.storedRefreshUserID = userID
	f.storedRefreshTTL = ttl
	return nil
}

// ValidateRefreshToken 模拟 Redis 中 refresh token 的存在性和归属校验。
func (f *fakeTokenStore) ValidateRefreshToken(_ context.Context, tokenID string, userID int64) (bool, error) {
	if !f.refreshValid {
		return false, nil
	}
	return f.storedRefreshTokenID == tokenID && f.storedRefreshUserID == userID, nil
}

// RevokeRefreshToken 记录被撤销的 refresh token ID。
func (f *fakeTokenStore) RevokeRefreshToken(_ context.Context, tokenID string) error {
	f.revokedRefreshIDs = append(f.revokedRefreshIDs, tokenID)
	if f.storedRefreshTokenID == tokenID {
		f.storedRefreshTokenID = ""
	}
	return nil
}

// ConsumeRefreshToken 模拟 refresh token 轮转时的一次性消费语义。
func (f *fakeTokenStore) ConsumeRefreshToken(_ context.Context, tokenID string, userID int64) (bool, error) {
	if !f.refreshValid {
		return false, nil
	}
	if f.storedRefreshTokenID != tokenID || f.storedRefreshUserID != userID {
		return false, nil
	}
	f.revokedRefreshIDs = append(f.revokedRefreshIDs, tokenID)
	f.storedRefreshTokenID = ""
	return true, nil
}

// BlacklistAccessToken 记录 access token 黑名单写入参数。
func (f *fakeTokenStore) BlacklistAccessToken(_ context.Context, tokenID string, ttl time.Duration) error {
	f.blacklistedTokenID = tokenID
	f.blacklistedTTL = ttl
	return nil
}

// IsAccessTokenBlacklisted 模拟 access token 黑名单命中判断。
func (f *fakeTokenStore) IsAccessTokenBlacklisted(_ context.Context, tokenID string) (bool, error) {
	return f.accessBlacklisted && f.blacklistedTokenID == tokenID, nil
}

// TestJWTManagerIssuesAndParsesTokens 验证 token pair 能正常签发、解析并缓存 refresh token。
func TestJWTManagerIssuesAndParsesTokens(t *testing.T) {
	store := &fakeTokenStore{refreshValid: true}
	manager, err := NewJWTManager(JWTConfig{
		Issuer:          "initra",
		Secret:          "unit-test-secret",
		AccessTokenTTL:  time.Hour,
		RefreshTokenTTL: 24 * time.Hour,
		Store:           store,
	})
	require.NoError(t, err)

	pair, err := manager.IssueTokenPair(context.Background(), Principal{
		UserID: 1001,
		Roles:  []string{"admin"},
	})
	require.NoError(t, err)
	require.NotEmpty(t, pair.AccessToken)
	require.NotEmpty(t, pair.RefreshToken)

	accessClaims, err := manager.ParseAccessToken(context.Background(), pair.AccessToken)
	require.NoError(t, err)
	require.Equal(t, int64(1001), accessClaims.UserID)
	require.Equal(t, []string{"admin"}, accessClaims.Roles)
	require.Equal(t, TokenTypeAccess, accessClaims.TokenType)
	require.NotEmpty(t, accessClaims.ID)

	refreshClaims, err := manager.ParseRefreshToken(context.Background(), pair.RefreshToken)
	require.NoError(t, err)
	require.Equal(t, int64(1001), refreshClaims.UserID)
	require.Equal(t, TokenTypeRefresh, refreshClaims.TokenType)
	require.NotEmpty(t, refreshClaims.ID)
	require.NotEqual(t, accessClaims.ID, refreshClaims.ID)
	require.Equal(t, refreshClaims.ID, store.storedRefreshTokenID)
	require.Equal(t, int64(1001), store.storedRefreshUserID)
	require.Positive(t, store.storedRefreshTTL)
	require.Empty(t, store.blacklistedTokenID)
}

// TestJWTManagerRejectsExpiredAccessToken 验证过期 access token 无法通过解析。
func TestJWTManagerRejectsExpiredAccessToken(t *testing.T) {
	manager, err := NewJWTManager(JWTConfig{
		Issuer:          "initra",
		Secret:          "unit-test-secret",
		AccessTokenTTL:  time.Hour,
		RefreshTokenTTL: 24 * time.Hour,
	})
	require.NoError(t, err)

	token, err := manager.issue(Principal{UserID: 1001}, TokenTypeAccess, time.Now().Add(-time.Minute), time.Now().Add(-2*time.Minute))
	require.NoError(t, err)

	_, err = manager.ParseAccessToken(context.Background(), token)
	require.Error(t, err)
}

// TestJWTManagerRejectsAccessTokenWithoutExpiration 验证 JWT 必须显式携带 exp。
func TestJWTManagerRejectsAccessTokenWithoutExpiration(t *testing.T) {
	manager, err := NewJWTManager(JWTConfig{
		Issuer:          "initra",
		Secret:          "unit-test-secret",
		AccessTokenTTL:  time.Hour,
		RefreshTokenTTL: 24 * time.Hour,
	})
	require.NoError(t, err)

	claims := Claims{
		UserID:    1001,
		TokenType: TokenTypeAccess,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:  "initra",
			Subject: "1001",
			ID:      "token-without-exp",
		},
	}
	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte("unit-test-secret"))
	require.NoError(t, err)

	_, err = manager.ParseAccessToken(context.Background(), token)
	require.Error(t, err)
}

// TestJWTManagerRejectsBlacklistedAccessToken 验证服务端黑名单可以主动吊销 access token。
func TestJWTManagerRejectsBlacklistedAccessToken(t *testing.T) {
	store := &fakeTokenStore{refreshValid: true}
	manager, err := NewJWTManager(JWTConfig{
		Issuer:          "initra",
		Secret:          "unit-test-secret",
		AccessTokenTTL:  time.Hour,
		RefreshTokenTTL: 24 * time.Hour,
		Store:           store,
	})
	require.NoError(t, err)

	pair, err := manager.IssueTokenPair(context.Background(), Principal{UserID: 1001})
	require.NoError(t, err)

	claims, err := manager.parse(pair.AccessToken, TokenTypeAccess)
	require.NoError(t, err)
	store.blacklistedTokenID = claims.ID
	store.accessBlacklisted = true

	_, err = manager.ParseAccessToken(context.Background(), pair.AccessToken)
	require.Error(t, err)
}

// TestJWTManagerRejectsRefreshTokenMissingFromStore 验证未在状态存储中登记的 refresh token 会被拒绝。
func TestJWTManagerRejectsRefreshTokenMissingFromStore(t *testing.T) {
	store := &fakeTokenStore{refreshValid: false}
	manager, err := NewJWTManager(JWTConfig{
		Issuer:          "initra",
		Secret:          "unit-test-secret",
		AccessTokenTTL:  time.Hour,
		RefreshTokenTTL: 24 * time.Hour,
		Store:           store,
	})
	require.NoError(t, err)

	pair, err := manager.IssueTokenPair(context.Background(), Principal{UserID: 1001})
	require.NoError(t, err)

	_, err = manager.ParseRefreshToken(context.Background(), pair.RefreshToken)
	require.Error(t, err)
}

// TestJWTManagerBlacklistAccessTokenStoresRemainingTTL 验证黑名单 TTL 使用 access token 剩余有效期。
func TestJWTManagerBlacklistAccessTokenStoresRemainingTTL(t *testing.T) {
	store := &fakeTokenStore{refreshValid: true}
	manager, err := NewJWTManager(JWTConfig{
		Issuer:          "initra",
		Secret:          "unit-test-secret",
		AccessTokenTTL:  time.Hour,
		RefreshTokenTTL: 24 * time.Hour,
		Store:           store,
	})
	require.NoError(t, err)

	pair, err := manager.IssueTokenPair(context.Background(), Principal{UserID: 1001})
	require.NoError(t, err)

	err = manager.BlacklistAccessToken(context.Background(), pair.AccessToken)
	require.NoError(t, err)
	require.NotEmpty(t, store.blacklistedTokenID)
	require.Positive(t, store.blacklistedTTL)
}

// TestJWTManagerUsesInjectedClockForRefreshTTL 验证 refresh token 缓存 TTL 基于注入时钟计算。
func TestJWTManagerUsesInjectedClockForRefreshTTL(t *testing.T) {
	now := time.Now().Add(-time.Hour)
	store := &fakeTokenStore{refreshValid: true}
	manager, err := NewJWTManager(JWTConfig{
		Issuer:          "initra",
		Secret:          "unit-test-secret",
		AccessTokenTTL:  time.Hour,
		RefreshTokenTTL: 24 * time.Hour,
		Store:           store,
		Now: func() time.Time {
			return now
		},
	})
	require.NoError(t, err)

	_, err = manager.IssueTokenPair(context.Background(), Principal{UserID: 1001})
	require.NoError(t, err)

	require.Equal(t, 24*time.Hour, store.storedRefreshTTL)
}

// TestJWTManagerUsesInjectedClockForBlacklistTTL 验证黑名单 TTL 基于注入时钟计算。
func TestJWTManagerUsesInjectedClockForBlacklistTTL(t *testing.T) {
	now := time.Now().Add(-30 * time.Minute)
	store := &fakeTokenStore{refreshValid: true}
	manager, err := NewJWTManager(JWTConfig{
		Issuer:          "initra",
		Secret:          "unit-test-secret",
		AccessTokenTTL:  time.Hour,
		RefreshTokenTTL: 24 * time.Hour,
		Store:           store,
		Now: func() time.Time {
			return now
		},
	})
	require.NoError(t, err)

	pair, err := manager.IssueTokenPair(context.Background(), Principal{UserID: 1001})
	require.NoError(t, err)

	err = manager.BlacklistAccessToken(context.Background(), pair.AccessToken)
	require.NoError(t, err)
	require.Greater(t, store.blacklistedTTL, 50*time.Minute)
}
