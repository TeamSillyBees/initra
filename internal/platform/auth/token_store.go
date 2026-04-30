package auth

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// TokenStore 定义 JWT 生命周期管理所需的最小状态存储能力。
// access token 默认保持无状态，仅在被主动吊销时写入黑名单。
// refresh token 是 opaque token，Redis 只保存其指纹与 access token jti 的绑定记录。
type TokenStore interface {
	StoreRefreshToken(ctx context.Context, tokenID string, record RefreshTokenRecord, ttl time.Duration) error
	ValidateRefreshToken(ctx context.Context, tokenID string) (RefreshTokenRecord, bool, error)
	ConsumeRefreshToken(ctx context.Context, tokenID string) (RefreshTokenRecord, bool, error)
	BlacklistAccessToken(ctx context.Context, tokenID string, ttl time.Duration) error
	IsAccessTokenBlacklisted(ctx context.Context, tokenID string) (bool, error)
}

// RedisTokenStore 是基于 Redis 的 token 状态存储实现。
type RedisTokenStore struct {
	keyPrefix string
	client    redis.Cmdable
}

// NewRedisTokenStore 创建 Redis token store。
func NewRedisTokenStore(appName string, client redis.Cmdable) *RedisTokenStore {
	return &RedisTokenStore{
		keyPrefix: appName,
		client:    client,
	}
}

// StoreRefreshToken 记录 refresh token 指纹与 access token jti 的绑定，并设置与 token 等长的 TTL。
func (s *RedisTokenStore) StoreRefreshToken(ctx context.Context, tokenID string, record RefreshTokenRecord, ttl time.Duration) error {
	if s == nil || s.client == nil {
		return nil
	}
	if ttl <= 0 {
		return fmt.Errorf("refresh token ttl must be positive")
	}
	payload, err := json.Marshal(record)
	if err != nil {
		return err
	}
	return s.client.Set(ctx, s.refreshTokenKey(tokenID), payload, ttl).Err()
}

// ValidateRefreshToken 校验 refresh token 是否仍在 Redis 中处于有效状态。
func (s *RedisTokenStore) ValidateRefreshToken(ctx context.Context, tokenID string) (RefreshTokenRecord, bool, error) {
	if s == nil || s.client == nil {
		return RefreshTokenRecord{}, false, nil
	}

	value, err := s.client.Get(ctx, s.refreshTokenKey(tokenID)).Result()
	if errors.Is(err, redis.Nil) {
		return RefreshTokenRecord{}, false, nil
	}
	if err != nil {
		return RefreshTokenRecord{}, false, err
	}
	record, err := decodeRefreshTokenRecord(value)
	if err != nil {
		return RefreshTokenRecord{}, false, err
	}
	return record, true, nil
}

// ConsumeRefreshToken 以原子方式读取并删除 refresh token，避免旧 token 被重复使用。
func (s *RedisTokenStore) ConsumeRefreshToken(ctx context.Context, tokenID string) (RefreshTokenRecord, bool, error) {
	if s == nil || s.client == nil {
		return RefreshTokenRecord{}, false, nil
	}

	value, err := consumeRefreshTokenScript.Run(
		ctx,
		s.client,
		[]string{s.refreshTokenKey(tokenID)},
	).Text()
	if errors.Is(err, redis.Nil) {
		return RefreshTokenRecord{}, false, nil
	}
	if err != nil {
		return RefreshTokenRecord{}, false, err
	}
	if value == "" {
		return RefreshTokenRecord{}, false, nil
	}
	record, err := decodeRefreshTokenRecord(value)
	if err != nil {
		return RefreshTokenRecord{}, false, err
	}
	return record, true, nil
}

// BlacklistAccessToken 把 access token 的 jti 写入黑名单，并附带剩余有效期作为 TTL。
func (s *RedisTokenStore) BlacklistAccessToken(ctx context.Context, tokenID string, ttl time.Duration) error {
	if s == nil || s.client == nil || ttl <= 0 {
		return nil
	}
	return s.client.Set(ctx, s.accessBlacklistKey(tokenID), "1", ttl).Err()
}

// IsAccessTokenBlacklisted 判断 access token 是否已经被服务端主动吊销。
func (s *RedisTokenStore) IsAccessTokenBlacklisted(ctx context.Context, tokenID string) (bool, error) {
	if s == nil || s.client == nil {
		return false, nil
	}

	count, err := s.client.Exists(ctx, s.accessBlacklistKey(tokenID)).Result()
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// refreshTokenKey 生成 refresh token 状态缓存 key。
func (s *RedisTokenStore) refreshTokenKey(tokenID string) string {
	return fmt.Sprintf("%s:auth:refresh:%s", s.keyPrefix, tokenID)
}

// accessBlacklistKey 生成 access token 黑名单 key。
func (s *RedisTokenStore) accessBlacklistKey(tokenID string) string {
	return fmt.Sprintf("%s:auth:blacklist:%s", s.keyPrefix, tokenID)
}

// decodeRefreshTokenRecord 解析 Redis 中保存的 refresh token 绑定记录。
func decodeRefreshTokenRecord(value string) (RefreshTokenRecord, error) {
	var record RefreshTokenRecord
	if err := json.Unmarshal([]byte(value), &record); err != nil {
		return RefreshTokenRecord{}, err
	}
	return record, nil
}

// consumeRefreshTokenScript 通过 Lua 保证“读取 refresh token 绑定记录”和“删除旧 token”原子完成。
var consumeRefreshTokenScript = redis.NewScript(`
local value = redis.call("GET", KEYS[1])
if not value then
    return ""
end
redis.call("DEL", KEYS[1])
return value
`)
