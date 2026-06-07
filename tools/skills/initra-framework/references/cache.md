# 业务缓存

## 注册

```go
platformcache.Register(injector, platformcache.Config{
	AppName:       cfg.App.Name,
	LocalTTL:      cfg.Cache.LocalTTL,
	RemoteTTL:     cfg.Cache.RemoteTTL,
	RemoteEnabled: cfg.Redis.Enabled,
})
```

Redis 未启用时自动退化为本地缓存。

## 模块适配器

把缓存 key、序列化对象和错误转换封装在模块 `cache.go`：

```go
type UserCache struct {
	manager *platformcache.Manager
	cache   jetcache.Cache
}

func (c *UserCache) key(id idgen.ID) string {
	return c.manager.BuildKey("user", "profile", id)
}
```

Service 调用缓存适配器，不直接拼 key 或操作 jetcache 细节。

## 选择

- 详情、资料、字典等业务缓存：优先 `pkg/cache`。
- 只需要 Redis 的低层缓存、空值缓存或 singleflight：看 `pkg/redisx.NewCache[T]`。

## 禁止

- 不要在多个 service 方法里复制缓存 key 格式。
- 不要把缓存命中失败当成业务错误；区分 miss 和底层故障。
