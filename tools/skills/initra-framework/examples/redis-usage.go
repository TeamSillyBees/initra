//go:build ignore

package verification

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
	apperrors "github.com/teamsillybees/initra/pkg/errors"
	"github.com/teamsillybees/initra/pkg/redisx"
)

type verificationStore interface {
	Set(ctx context.Context, key string, value any, expiration time.Duration) *redis.StatusCmd
	Get(ctx context.Context, key string) *redis.StringCmd
	Del(ctx context.Context, keys ...string) *redis.IntCmd
}

// Service 提供验证码缓存相关业务能力。
type Service struct {
	keys  *redisx.KeyBuilder
	store verificationStore
	ttl   time.Duration
}

// NewService 创建验证码服务。
func NewService(keys *redisx.KeyBuilder, store verificationStore, ttl time.Duration) *Service {
	return &Service{keys: keys, store: store, ttl: ttl}
}

// SaveCode 保存带 TTL 的验证码。
func (s *Service) SaveCode(ctx context.Context, mobile string, code string) error {
	key, err := s.keys.Build("verification-code", mobile)
	if err != nil {
		return apperrors.Wrap(err, apperrors.CodeBadRequest, "验证码 key 无效")
	}
	if err := s.store.Set(ctx, key, code, s.ttl).Err(); err != nil {
		return apperrors.Wrap(err, apperrors.CodeCacheError, "保存验证码失败")
	}
	return nil
}
