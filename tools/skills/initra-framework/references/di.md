# 依赖注入与 Providers

## Boot 层

`internal/boot/providers.go` 只负责应用级装配。框架能力用一步式 `Register`：

```go
logx.Register(injector, cfg.Log)
httpclient.Register(injector, cfg.HTTPClient)
redisx.Register(injector, cfg.Redis)
asynqadapter.Register(injector, cfg.Task)
storageprovider.Register(injector, cfg.Storage)
```

只有项目自身组件，例如 Ent client、SQL DB 和模块对象，才在 boot 中写本地 `do.Provide`。

## Module 层

每个模块暴露 `Provide(injector *do.Injector)`。在 provider 中解析框架依赖，通过构造函数传给 service：

```go
do.ProvideNamed(injector, userServiceName, func(i *do.Injector) (*Service, error) {
	client := do.MustInvoke[*ent.Client](i)
	cache := do.MustInvokeNamed[*UserCache](i, userCacheName)
	return NewService(client, cache), nil
})
```

Service 和 handler 不感知 injector。

## 禁止

- 不要在 service/handler 方法里调用 `do.Invoke` 或 `do.MustInvoke`。
- 不要先 `do.ProvideValue(injector, cfg.HTTPClient)` 再 `httpclient.Register(injector)`。
- 不要为了复用逻辑让一个业务模块 import 另一个模块的具体实现。
