# redisx

当业务需要 Redis client、Key 规范、TTL 临时数据、短时分布式锁、Redis 承载的 token/session、Redis 泛型缓存、Lua 脚本或安全扫描时，使用 `github.com/teamsillybees/initra/pkg/redisx`。

## 标准装配

在 logger 注册之后，在 boot 层统一注册 Redis：

```go
logging.Register(injector, cfg.Log)
redisx.Register(injector, cfg.Redis)
```

`redisx.Register` 会根据 `redis.mode` 注册 `redis.UniversalClient`，支持 standalone 和 sentinel。initra 不支持 Redis cluster，不要新增 cluster client 用法。

## 业务用法

Service 依赖窄接口，不初始化 client：

```go
type verificationStore interface {
	Set(ctx context.Context, key string, value any, expiration time.Duration) *redis.StatusCmd
	Get(ctx context.Context, key string) *redis.StringCmd
	Del(ctx context.Context, keys ...string) *redis.IntCmd
}
```

在 `providers.go` 或 boot 层解析真实 Redis client：

```go
client := do.MustInvoke[redis.UniversalClient](i)
return NewService(client), nil
```

## Key 规则

使用 `redisx.NewKeyBuilder(redisx.KeyConfig{App: cfg.App.Name, Env: cfg.App.Env})` 维护共享 Key 前缀。初始化期注册前缀，业务代码通过 builder 构造 Key。

- Key 格式为 `{app}:{env}:{module}:{biz}:{id}`。
- 临时值必须设置 TTL。
- 不要记录密码、token、验证码、session value 或原始 payload。
- 生产环境不要使用 `KEYS`。优先使用已注册前缀和 SCAN/UNLINK helper。

## 缓存与锁

当需要 JSON/msgpack payload、TTL jitter、空值缓存和 singleflight 防击穿时，使用 `redisx.NewCache[T]`。当需要短时间互斥或效率锁时，使用 `redisx.NewLocker(client)`；它不是强一致事务方案。

强一致场景应使用数据库事务、唯一约束、乐观锁、幂等表或 fencing token。

## 检查清单

- Redis 是否只在 boot 层注册一次？
- Service 是否只接收窄接口？
- Key 是否按 app/env/module/domain 命名空间隔离？
- 每个临时值是否设置 TTL？
- 日志是否避开敏感值？
- 扫描逻辑是否使用 SCAN 风格而不是 `KEYS`？
