package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/teamsillybees/initra/pkg/redisx"
)

const (
	refreshTokenPrefixName    = "auth_refresh"
	accessBlacklistPrefixName = "auth_blacklist"
	consumeRefreshScriptName  = "auth_consume_refresh_token"
	defaultRedisTokenEnv      = "default"
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
	keys    *redisx.KeyBuilder
	client  redisx.CommandClient
	store   *redisx.Store
	scripts *redisx.ScriptRegistry
}

// NewRedisTokenStore 创建 Redis token store。
func NewRedisTokenStore(appName string, client redisx.CommandClient) *RedisTokenStore {
	return NewRedisTokenStoreWithEnv(appName, defaultRedisTokenEnv, client)
}

// NewRedisTokenStoreWithEnv 创建带 app/env 命名空间的 Redis token store。
func NewRedisTokenStoreWithEnv(appName string, env string, client redisx.CommandClient) *RedisTokenStore {
	if strings.TrimSpace(env) == "" {
		env = defaultRedisTokenEnv
	}
	keys := redisx.NewKeyBuilder(redisx.KeyConfig{App: appName, Env: env})
	mustRegisterRedisTokenPrefix(keys, refreshTokenPrefixName, "auth", "refresh")
	mustRegisterRedisTokenPrefix(keys, accessBlacklistPrefixName, "auth", "blacklist")

	return &RedisTokenStore{
		keys:    keys,
		client:  client,
		store:   redisx.NewStore(client),
		scripts: newRedisTokenScriptRegistry(),
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
	key, err := s.refreshTokenKey(tokenID)
	if err != nil {
		return err
	}
	return s.store.SetString(ctx, key, payload, ttl)
}

// ValidateRefreshToken 校验 refresh token 是否仍在 Redis 中处于有效状态。
func (s *RedisTokenStore) ValidateRefreshToken(ctx context.Context, tokenID string) (RefreshTokenRecord, bool, error) {
	if s == nil || s.client == nil {
		return RefreshTokenRecord{}, false, nil
	}

	key, err := s.refreshTokenKey(tokenID)
	if err != nil {
		return RefreshTokenRecord{}, false, err
	}
	value, found, err := s.store.GetString(ctx, key)
	if !found {
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

	key, err := s.refreshTokenKey(tokenID)
	if err != nil {
		return RefreshTokenRecord{}, false, err
	}
	value, err := s.scripts.Run(
		ctx,
		s.client,
		consumeRefreshScriptName,
		[]string{key},
	).Text()
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
	key, err := s.accessBlacklistKey(tokenID)
	if err != nil {
		return err
	}
	return s.store.SetString(ctx, key, "1", ttl)
}

// IsAccessTokenBlacklisted 判断 access token 是否已经被服务端主动吊销。
func (s *RedisTokenStore) IsAccessTokenBlacklisted(ctx context.Context, tokenID string) (bool, error) {
	if s == nil || s.client == nil {
		return false, nil
	}

	key, err := s.accessBlacklistKey(tokenID)
	if err != nil {
		return false, err
	}
	return s.store.Exists(ctx, key)
}

// refreshTokenKey 生成 refresh token 状态缓存 key。
func (s *RedisTokenStore) refreshTokenKey(tokenID string) (string, error) {
	return s.keys.Build(refreshTokenPrefixName, tokenID)
}

// accessBlacklistKey 生成 access token 黑名单 key。
func (s *RedisTokenStore) accessBlacklistKey(tokenID string) (string, error) {
	return s.keys.Build(accessBlacklistPrefixName, tokenID)
}

// decodeRefreshTokenRecord 解析 Redis 中保存的 refresh token 绑定记录。
func decodeRefreshTokenRecord(value string) (RefreshTokenRecord, error) {
	var record RefreshTokenRecord
	if err := json.Unmarshal([]byte(value), &record); err != nil {
		return RefreshTokenRecord{}, err
	}
	return record, nil
}

func mustRegisterRedisTokenPrefix(keys *redisx.KeyBuilder, name string, module string, biz string) {
	if err := keys.RegisterPrefix(name, module, biz); err != nil {
		panic(err)
	}
}

func newRedisTokenScriptRegistry() *redisx.ScriptRegistry {
	registry := redisx.NewScriptRegistry()
	if err := registry.Register(redisx.ScriptDefinition{
		Name:   consumeRefreshScriptName,
		Source: consumeRefreshTokenScriptSource,
		Keys: []redisx.ScriptArgument{{
			Name:        "refresh_token_key",
			Description: "refresh token 状态记录 key",
		}},
	}); err != nil {
		panic(err)
	}
	return registry
}

// consumeRefreshTokenScriptSource 通过 Lua 保证“读取 refresh token 绑定记录”和“删除旧 token”原子完成。
const consumeRefreshTokenScriptSource = `
local value = redis.call("GET", KEYS[1])
if not value then
    return ""
end
redis.call("DEL", KEYS[1])
return value
`
