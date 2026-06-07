# HTTP Client

## 注册

在配置中声明远程服务，boot 层调用：

```go
httpclient.Register(injector, cfg.HTTPClient)
```

Register 会注册 `*httpclient.Factory` 和每个配置服务的命名 `*httpclient.Client`。

## 模块接入

优先使用 `ProvideConsumer` 消除 provider 胶水：

```go
httpclient.ProvideConsumer(injector, smsServiceName, "sms", NewService)
```

Service 构造函数依赖最小接口：

```go
func NewService(client httpclient.JSONPoster) *Service {
	return &Service{client: client}
}
```

也可以显式解析：

```go
client := do.MustInvokeNamed[*httpclient.Client](i, httpclient.ClientName("sms"))
```

## 调用

```go
err := s.client.PostJSON(ctx, "/messages", body, &result,
	httpclient.WithHeader("X-Trace-ID", traceID),
	httpclient.WithQuery("channel", "login"),
)
```

## 禁止

- 不要在 service 中创建 `http.Client{}` 或用 `http.DefaultClient`。
- 不要把 base URL、timeout、retry 硬编码进业务代码。
- 不要记录 Authorization header 或签名密钥。
