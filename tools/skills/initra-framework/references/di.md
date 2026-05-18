# 依赖注入与 Providers

initra 业务项目使用 `github.com/samber/do` 做显式 boot 和 module 装配。依赖解析只应发生在组合边界。

## Boot Providers

`internal/boot/providers.go` 应使用一步式调用注册框架能力：

```go
logging.Register(injector, cfg.Log)
httpclient.Register(injector, cfg.HTTPClient)
redisx.Register(injector, cfg.Redis)
storageprovider.Register(injector, cfg.Storage)
```

只有项目自身组件，例如生成的 Ent client 或模块对象，才应在 boot 中使用本地 `do.Provide` 构造函数。

## Module Providers

每个模块拥有自己的 `providers.go` 并暴露 `Provide(injector *do.Injector)`。模块内 repository/service/handler 使用命名依赖：

```go
const userServiceName = "user.service"

func Provide(injector *do.Injector) {
	do.ProvideNamed(injector, userServiceName, func(i *do.Injector) (*Service, error) {
		repo := do.MustInvokeNamed[*Repository](i, userRepositoryName)
		cache := do.MustInvokeNamed[*UserCache](i, userCacheName)
		return NewService(repo, cache), nil
	})
	do.Provide(injector, func(i *do.Injector) (*Module, error) {
		service := do.MustInvokeNamed[*Service](i, userServiceName)
		return NewModule(NewHandler(service)), nil
	})
}
```

在 provider 中解析框架依赖，再通过构造函数传入业务组件。Service 只接收构造函数参数，不感知 injector。

## Register 契约

框架 package 应暴露：

```go
func Register(injector *do.Injector, cfg Config)
```

或 options 变体：

```go
func Register(injector *do.Injector, opts RegisterOptions)
```

不要先通过 `do.ProvideValue` 单独注册配置，再调用无参数注册函数。

## 禁止写法

- 不要在 service 或 handler 方法中调用 `do.Invoke` 或 `do.MustInvoke`。
- 不要把全局 injector 传入业务逻辑。
- 不要为了复用逻辑而让一个模块 import 另一个模块的具体实现。
- 不要用 package-level global 隐藏必需依赖。

## 检查清单

- 所有框架 package 是否只在 boot 层注册一次？
- 模块依赖是否在模块自己的 `providers.go` 中注册？
- Service 构造函数是否显式列出依赖？
- 跨模块调用是否通过调用方定义的小接口表达？
- 配置是否直接传入 `Register`？
