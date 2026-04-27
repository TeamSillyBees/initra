package infra

import (
	"context"
	"errors"

	jetcache "github.com/mgtv-tech/jetcache-go"
	"github.com/teamsillybees/initra/internal/app/user/domain"
	platformcache "github.com/teamsillybees/initra/internal/platform/cache"
	apperrors "github.com/teamsillybees/initra/internal/platform/errors"
)

// UserCache 负责 user 模块详情缓存的读写与 Key 规范封装。
type UserCache struct {
	manager *platformcache.Manager
	cache   jetcache.Cache
}

// NewUserCache 创建用户详情缓存实现。
func NewUserCache(manager *platformcache.Manager) *UserCache {
	return &UserCache{
		manager: manager,
		cache:   manager.New("user-profile"),
	}
}

// Get 读取用户详情缓存。
func (c *UserCache) Get(ctx context.Context, id int64) (*domain.User, bool, error) {
	var user domain.User
	err := c.cache.Get(ctx, c.key(id), &user)
	if err == nil {
		return &user, true, nil
	}
	if errors.Is(err, jetcache.ErrCacheMiss) {
		return nil, false, nil
	}
	return nil, false, apperrors.Wrap(err, apperrors.CodeCacheError, "get user cache failed")
}

// Set 写入用户详情缓存。
func (c *UserCache) Set(ctx context.Context, user *domain.User) error {
	return c.cache.Set(ctx, c.key(user.ID), jetcache.Value(user))
}

// Delete 删除用户详情缓存。
func (c *UserCache) Delete(ctx context.Context, id int64) error {
	if err := c.cache.Delete(ctx, c.key(id)); err != nil {
		return apperrors.Wrap(err, apperrors.CodeCacheError, "delete user cache failed")
	}
	return nil
}

// key 使用平台统一缓存 key 规范生成用户详情缓存 key。
func (c *UserCache) key(id int64) string {
	return c.manager.BuildKey("user", "profile", id)
}
