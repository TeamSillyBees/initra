package auth

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	platformauth "github.com/teamsillybees/initra/pkg/auth"
)

var errLoginFailed = errors.New("login failed")

type fakeIdentityRepository struct {
	byID       map[int64]*Identity
	byUsername map[string]*Identity
}

func (f *fakeIdentityRepository) FindByUsername(_ context.Context, username string) (*Identity, error) {
	if identity, ok := f.byUsername[username]; ok {
		cloned := *identity
		return &cloned, nil
	}
	return nil, nil
}

func (f *fakeIdentityRepository) FindByID(_ context.Context, id int64) (*Identity, error) {
	if identity, ok := f.byID[id]; ok {
		cloned := *identity
		return &cloned, nil
	}
	return nil, nil
}

type fakePasswordVerifier struct{}

func (fakePasswordVerifier) Hash(password string) (string, error) {
	return "hashed:" + password, nil
}

func (fakePasswordVerifier) Compare(hash string, password string) error {
	if hash != "hashed:"+password {
		return errLoginFailed
	}
	return nil
}

type fakeTokenStore struct {
	storedRefreshTokenID string
	storedRefreshRecord  platformauth.RefreshTokenRecord
	storedRefreshTTL     time.Duration
	revokedRefreshIDs    []string
	blacklistedTokenID   string
	blacklistedTTL       time.Duration
	refreshValid         bool
	accessBlacklisted    bool
}

func (f *fakeTokenStore) StoreRefreshToken(_ context.Context, tokenID string, record platformauth.RefreshTokenRecord, ttl time.Duration) error {
	f.storedRefreshTokenID = tokenID
	f.storedRefreshRecord = record
	f.storedRefreshTTL = ttl
	return nil
}

func (f *fakeTokenStore) ValidateRefreshToken(_ context.Context, tokenID string) (platformauth.RefreshTokenRecord, bool, error) {
	if !f.refreshValid {
		return platformauth.RefreshTokenRecord{}, false, nil
	}
	return f.storedRefreshRecord, f.storedRefreshTokenID == tokenID, nil
}

func (f *fakeTokenStore) ConsumeRefreshToken(_ context.Context, tokenID string) (platformauth.RefreshTokenRecord, bool, error) {
	if !f.refreshValid {
		return platformauth.RefreshTokenRecord{}, false, nil
	}
	if f.storedRefreshTokenID != tokenID {
		return platformauth.RefreshTokenRecord{}, false, nil
	}
	record := f.storedRefreshRecord
	f.revokedRefreshIDs = append(f.revokedRefreshIDs, tokenID)
	f.storedRefreshTokenID = ""
	return record, true, nil
}

func (f *fakeTokenStore) RevokeRefreshToken(_ context.Context, tokenID string) error {
	f.revokedRefreshIDs = append(f.revokedRefreshIDs, tokenID)
	if f.storedRefreshTokenID == tokenID {
		f.storedRefreshTokenID = ""
	}
	return nil
}

func (f *fakeTokenStore) BlacklistAccessToken(_ context.Context, tokenID string, ttl time.Duration) error {
	f.blacklistedTokenID = tokenID
	f.blacklistedTTL = ttl
	f.accessBlacklisted = true
	return nil
}

func (f *fakeTokenStore) IsAccessTokenBlacklisted(_ context.Context, tokenID string) (bool, error) {
	return f.accessBlacklisted && f.blacklistedTokenID == tokenID, nil
}

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

	result, err := service.Login(context.Background(), LoginParams{
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

	_, err = tokenManager.ValidateRefreshToken(context.Background(), pair.RefreshToken)
	require.Error(t, err)

	_, err = tokenManager.ValidateRefreshToken(context.Background(), result.RefreshToken)
	require.NoError(t, err)
}
