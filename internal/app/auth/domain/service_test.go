package domain

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	platformauth "github.com/teamsillybees/initra/internal/platform/auth"
)

// fakeIdentityRepository 使用内存 map 模拟身份仓储，便于专注测试 auth service 编排。
type fakeIdentityRepository struct {
	byID       map[int64]*Identity
	byUsername map[string]*Identity
}

// FindByUsername 根据用户名返回身份副本。
func (f *fakeIdentityRepository) FindByUsername(_ context.Context, username string) (*Identity, error) {
	if identity, ok := f.byUsername[username]; ok {
		cloned := *identity
		return &cloned, nil
	}
	return nil, nil
}

// FindByID 根据用户 ID 返回身份副本。
func (f *fakeIdentityRepository) FindByID(_ context.Context, id int64) (*Identity, error) {
	if identity, ok := f.byID[id]; ok {
		cloned := *identity
		return &cloned, nil
	}
	return nil, nil
}

// fakePasswordVerifier 用固定规则模拟密码校验，避免测试依赖 bcrypt 成本。
type fakePasswordVerifier struct{}

// Hash 返回测试可识别的哈希格式。
func (fakePasswordVerifier) Hash(password string) (string, error) {
	return "hashed:" + password, nil
}

// Compare 校验哈希是否符合 fake Hash 规则。
func (fakePasswordVerifier) Compare(hash string, password string) error {
	if hash != "hashed:"+password {
		return ErrLoginFailed
	}
	return nil
}

// fakeTokenStore 模拟 JWT 状态存储，记录 refresh token 的生命周期调用。
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

// StoreRefreshToken 保存 refresh token 的测试状态。
func (f *fakeTokenStore) StoreRefreshToken(_ context.Context, tokenID string, userID int64, ttl time.Duration) error {
	f.storedRefreshTokenID = tokenID
	f.storedRefreshUserID = userID
	f.storedRefreshTTL = ttl
	return nil
}

// ValidateRefreshToken 验证 refresh token 是否仍在测试状态中。
func (f *fakeTokenStore) ValidateRefreshToken(_ context.Context, tokenID string, userID int64) (bool, error) {
	if !f.refreshValid {
		return false, nil
	}
	return f.storedRefreshTokenID == tokenID && f.storedRefreshUserID == userID, nil
}

// ConsumeRefreshToken 模拟 refresh token 的一次性消费。
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

// RevokeRefreshToken 记录 refresh token 被撤销。
func (f *fakeTokenStore) RevokeRefreshToken(_ context.Context, tokenID string) error {
	f.revokedRefreshIDs = append(f.revokedRefreshIDs, tokenID)
	if f.storedRefreshTokenID == tokenID {
		f.storedRefreshTokenID = ""
	}
	return nil
}

// BlacklistAccessToken 记录 access token 黑名单写入。
func (f *fakeTokenStore) BlacklistAccessToken(_ context.Context, tokenID string, ttl time.Duration) error {
	f.blacklistedTokenID = tokenID
	f.blacklistedTTL = ttl
	return nil
}

// IsAccessTokenBlacklisted 在 auth service 测试中默认不命中黑名单。
func (f *fakeTokenStore) IsAccessTokenBlacklisted(_ context.Context, tokenID string) (bool, error) {
	return f.accessBlacklisted && f.blacklistedTokenID == tokenID, nil
}

// TestServiceLoginReturnsTokenPairForValidCredentials 验证合法账号密码能换取 access/refresh token。
func TestServiceLoginReturnsTokenPairForValidCredentials(t *testing.T) {
	store := &fakeTokenStore{refreshValid: true}
	tokenManager, err := platformauth.NewJWTManager(platformauth.JWTConfig{
		Issuer:          "initra",
		Secret:          "auth-test-secret",
		AccessTokenTTL:  time.Hour,
		RefreshTokenTTL: 24 * time.Hour,
		Store:           store,
	})
	require.NoError(t, err)

	repo := &fakeIdentityRepository{
		byID: map[int64]*Identity{
			1001: {
				UserID:       1001,
				Username:     "alice",
				Nickname:     "Alice",
				PasswordHash: "hashed:secret-123",
				RoleCodes:    []string{"admin"},
				IsEnable:     true,
			},
		},
		byUsername: map[string]*Identity{
			"alice": {
				UserID:       1001,
				Username:     "alice",
				Nickname:     "Alice",
				PasswordHash: "hashed:secret-123",
				RoleCodes:    []string{"admin"},
				IsEnable:     true,
			},
		},
	}

	service := NewService(repo, fakePasswordVerifier{}, tokenManager)

	result, err := service.Login(context.Background(), LoginInput{
		Username: "alice",
		Password: "secret-123",
	})
	require.NoError(t, err)
	require.Equal(t, int64(1001), result.User.UserID)
	require.Equal(t, "Alice", result.User.Nickname)
	require.Equal(t, []string{"admin"}, result.User.RoleCodes)
	require.NotEmpty(t, result.AccessToken)
	require.NotEmpty(t, result.RefreshToken)
	require.NotEmpty(t, store.storedRefreshTokenID)
	require.Empty(t, store.blacklistedTokenID)
}

// TestServiceRefreshIssuesNewAccessToken 验证 refresh token 轮转后能重新签发 token pair。
func TestServiceRefreshIssuesNewAccessToken(t *testing.T) {
	store := &fakeTokenStore{refreshValid: true}
	tokenManager, err := platformauth.NewJWTManager(platformauth.JWTConfig{
		Issuer:          "initra",
		Secret:          "auth-test-secret",
		AccessTokenTTL:  time.Hour,
		RefreshTokenTTL: 24 * time.Hour,
		Store:           store,
	})
	require.NoError(t, err)

	repo := &fakeIdentityRepository{
		byID: map[int64]*Identity{
			1001: {
				UserID:       1001,
				Username:     "alice",
				Nickname:     "Alice",
				PasswordHash: "hashed:secret-123",
				RoleCodes:    []string{"admin"},
				IsEnable:     true,
			},
		},
	}

	service := NewService(repo, fakePasswordVerifier{}, tokenManager)
	pair, err := tokenManager.IssueTokenPair(context.Background(), platformauth.Principal{
		UserID: 1001,
		Roles:  []string{"admin"},
	})
	require.NoError(t, err)

	result, err := service.Refresh(context.Background(), pair.RefreshToken)
	require.NoError(t, err)
	require.NotEmpty(t, result.AccessToken)
	require.NotEmpty(t, result.RefreshToken)

	claims, err := tokenManager.ParseAccessToken(context.Background(), result.AccessToken)
	require.NoError(t, err)
	require.Equal(t, int64(1001), claims.UserID)

	_, err = tokenManager.ParseRefreshToken(context.Background(), pair.RefreshToken)
	require.Error(t, err)

	_, err = tokenManager.ParseRefreshToken(context.Background(), result.RefreshToken)
	require.NoError(t, err)
}
