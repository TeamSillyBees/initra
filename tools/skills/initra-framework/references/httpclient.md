# httpclient

当业务需要远程 HTTP API、第三方服务、内部服务调用、webhook、重试、超时、请求头、认证配置、响应大小限制或结构化请求日志时，使用 `github.com/teamsillybees/initra/pkg/httpclient`。

`httpclient` 支持通过配置设置全局代理或服务级代理，例如 `http://127.0.0.1:7890`。

每个 service 可通过 `properties` 配置 appid 等自定义属性，并在代码中使用 `client.GetProperty("app_id")` 读取。

## 标准装配

在配置中添加 `http_client.services.<name>`，并在 boot 层注册一次：

```go
logx.Register(injector, cfg.Log)
httpclient.Register(injector, cfg.HTTPClient)
```

`Register` 会创建 HTTP Client factory，并为每个已配置服务注册命名 `*httpclient.Client`。

## Provider 模式

在模块 `providers.go` 中优先使用 `ProvideConsumer` 将远程服务 client 注入业务组件：

```go
httpclient.ProvideConsumer(injector, smsSenderServiceName, "sms", NewSMSSender)
```

Service 依赖 `httpclient` 提供的窄接口，避免每个模块重复声明远程调用接口：

```go
type SMSSender struct {
	client httpclient.JSONPoster
}
```

## 请求模式

使用 request option，不要手动拼接 URL：

```go
var payload smsResponse
err := c.client.PostJSON(ctx, "/messages", body, &payload,
	httpclient.WithHeader("X-Trace-ID", traceID),
)
```

在 service 边界将 `*httpclient.Error` 映射为 `apperrors.Wrap` / `apperrors.WrapContext` 返回的 oops 错误，只放入 service、kind、status code 等安全细节。不要暴露原始 Authorization header、token 或可能包含敏感信息的响应体。

## 禁止写法

- 不要为服务调用创建 `http.Client{}` 或使用 `http.DefaultClient`。
- 不要在 service 中硬编码 base URL。
- 不要重复实现 retry、timeout 或 logging middleware。
- 不要让 handler 直接调用远程服务。

## 检查清单

- 远程服务是否已在配置中声明？
- 是否调用了 `httpclient.Register(injector, cfg.HTTPClient)`？
- 是否在 `providers.go` 中优先使用 `httpclient.ProvideConsumer`？
- Service 是否依赖 `httpclient.JSONGetter`、`JSONPoster`、`ReadCaller` 或 `Caller` 等窄接口？
- 出站错误是否映射为统一 `apperrors`？
