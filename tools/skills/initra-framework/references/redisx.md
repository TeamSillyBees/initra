# Redis

## 注册与注入

Boot 层：

```go
redisx.Register(injector, cfg.Redis)
```

业务侧依赖接口，例如 `redisx.UniversalClient` 或更小的本地接口；不要创建新的 client。

## Key Builder

用 `KeyBuilder` 管理命名空间：

```go
keys := redisx.NewKeyBuilder(redisx.KeyConfig{App: cfg.App.Name, Env: cfg.App.Env})
_ = keys.RegisterPrefix("verification-code", "auth", "verification")
key, err := keys.Build("verification-code", mobile)
```

## 扫描删除

生产环境不用 `KEYS`。使用允许前缀的 SCAN：

```go
result, err := redisx.UnlinkByPrefix(ctx, client, redisx.ScanOptions{
	Prefix:          prefix,
	AllowedPrefixes: keys.AllowedPrefixes(),
	MaxKeys:         10000,
	BatchSize:       500,
})
```

## 短锁

`redisx.NewLocker(client)` 只适合短时间互斥或效率优化；强一致业务仍要用数据库唯一约束、事务、乐观锁、幂等表或 fencing token。

## Redis 泛型缓存

`redisx.NewCache[T]` 适合 Redis-only 缓存，支持 TTL jitter、空值缓存和 singleflight。业务详情缓存通常优先看 `pkg/cache` 的 jetcache 管理器。

## 禁止

- 不要使用 `redis.NewClusterClient`；当前配置支持 standalone/sentinel。
- 不要记录密码、token、验证码、session value。
- 不要把锁当成事务正确性的唯一保障。
