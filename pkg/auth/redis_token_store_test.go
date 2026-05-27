package auth

import (
	"context"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/stretchr/testify/require"
	"github.com/teamsillybees/initra/pkg/idgen"
	"github.com/teamsillybees/initra/pkg/redisx"
)

func TestRedisTokenStoreUsesRedisxKeyBuilderAndScripts(t *testing.T) {
	server := miniredis.RunT(t)
	ctx := context.Background()
	client, err := redisx.NewClient(ctx, redisx.Config{
		Enabled: true,
		Mode:    redisx.ModeStandalone,
		Addr:    server.Addr(),
		Pool: redisx.PoolConfig{
			Size: 2,
		},
	}, nil)
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, client.Close()) })

	store := NewRedisTokenStoreWithEnv("initra", "dev", client)
	record := RefreshTokenRecord{
		UserID:          idgen.New(1001),
		AccessTokenID:   "access-1",
		AccessExpiresAt: time.Now().Add(time.Hour),
	}

	require.NoError(t, store.StoreRefreshToken(ctx, "refresh-1", record, time.Hour))
	_, err = server.Get("initra:dev:auth:refresh:refresh-1")
	require.NoError(t, err)

	got, ok, err := store.ValidateRefreshToken(ctx, "refresh-1")
	require.NoError(t, err)
	require.True(t, ok)
	require.Equal(t, record.UserID, got.UserID)
	require.Equal(t, record.AccessTokenID, got.AccessTokenID)

	got, ok, err = store.ConsumeRefreshToken(ctx, "refresh-1")
	require.NoError(t, err)
	require.True(t, ok)
	require.Equal(t, record.UserID, got.UserID)

	_, ok, err = store.ValidateRefreshToken(ctx, "refresh-1")
	require.NoError(t, err)
	require.False(t, ok)

	require.NoError(t, store.BlacklistAccessToken(ctx, "access-1", time.Hour))
	blacklisted, err := store.IsAccessTokenBlacklisted(ctx, "access-1")
	require.NoError(t, err)
	require.True(t, blacklisted)
	require.True(t, server.Exists("initra:dev:auth:blacklist:access-1"))
}
