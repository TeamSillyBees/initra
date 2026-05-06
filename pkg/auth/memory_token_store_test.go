package auth

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// TestMemoryTokenStoreConsumesRefreshTokenOnce 验证内存 token store 支持 refresh token 一次性消费。
func TestMemoryTokenStoreConsumesRefreshTokenOnce(t *testing.T) {
	store := NewMemoryTokenStore()
	record := RefreshTokenRecord{
		UserID:          1001,
		AccessTokenID:   "access-1",
		AccessExpiresAt: time.Now().Add(time.Hour),
	}

	require.NoError(t, store.StoreRefreshToken(context.Background(), "refresh-1", record, time.Hour))

	got, valid, err := store.ConsumeRefreshToken(context.Background(), "refresh-1")
	require.NoError(t, err)
	require.True(t, valid)
	require.Equal(t, record, got)

	_, valid, err = store.ValidateRefreshToken(context.Background(), "refresh-1")
	require.NoError(t, err)
	require.False(t, valid)
}

// TestMemoryTokenStoreBlacklistsAccessToken 验证内存 token store 支持 access token 黑名单。
func TestMemoryTokenStoreBlacklistsAccessToken(t *testing.T) {
	store := NewMemoryTokenStore()

	require.NoError(t, store.BlacklistAccessToken(context.Background(), "access-1", time.Hour))

	blacklisted, err := store.IsAccessTokenBlacklisted(context.Background(), "access-1")
	require.NoError(t, err)
	require.True(t, blacklisted)
}
