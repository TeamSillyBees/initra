# cache

业务层本地/远端缓存使用 `github.com/teamsillybees/initra/pkg/cache`。更底层的 Redis 泛型缓存需求使用 `pkg/redisx`。

## 标准装配

在 Redis 注册之后注册缓存管理器：

```go
platformcache.Register(injector, platformcache.Config{
	AppName:       cfg.App.Name,
	LocalTTL:      cfg.Cache.LocalTTL,
	RemoteTTL:     cfg.Cache.RemoteTTL,
	RemoteEnabled: cfg.Redis.Enabled,
})
```

启用远端缓存时，manager 使用 DI 中的 Redis client；否则自动退化为本地缓存。

## 模块适配器模式

缓存代码放在模块内 `cache.go`。Service 依赖小接口：

```go
type userCache interface {
	Get(ctx context.Context, id int64) (*User, bool, error)
	Set(ctx context.Context, user *User) error
	Delete(ctx context.Context, id int64) error
}
```

适配器使用 `platformcache.Manager` 实现：

```go
type UserCache struct {
	manager *platformcache.Manager
	cache   jetcache.Cache
}

func NewUserCache(manager *platformcache.Manager) *UserCache {
	return &UserCache{
		manager: manager,
		cache:   manager.New("user-profile"),
	}
}
```

通过 manager 统一构造缓存 key：

```go
key := c.manager.BuildKey("user", "profile", id)
```

## 禁止写法

- 不要把缓存 key 字符串格式散落在 service 中。
- 不要让 handler 直接读写缓存。
- 不要把缓存未命中当成内部错误；合适时返回 `(nil, false, nil)`。
- 不要缓存敏感值，除非 TTL、命名空间和日志行为都已明确。

## 检查清单

- cache 注册是否绑定 app name、本地 TTL、远端 TTL 和 Redis 开关？
- 缓存访问是否隔离在模块 adapter 中？
- Key 是否使用稳定的 module/domain/id 结构？
- 缓存错误是否用 `apperrors.CodeCacheError` 包装？
