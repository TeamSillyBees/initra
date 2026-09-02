# HTTP Client

## 最小接入

配置通常只声明服务地址，其余连接池、超时和响应大小使用安全默认值：

```yaml
http_client:
  enabled: true
  services:
    sms:
      base_url: https://sms.example.com
```

boot 统一注册，模块把命名客户端直接注入 `httpclient.Executor`：

```go
httpclient.Register(injector, cfg.HTTPClient)

func NewService(client httpclient.Executor) *Service {
	return &Service{client: client}
}

func Provide(injector *do.Injector) {
	httpclient.Provide(injector, "sms", NewService)
}
```

只有需要 Client 专有能力时才显式解析 `do.MustInvokeNamed[*httpclient.Client](i, httpclient.ClientName("sms"))`。

## 构建请求

JSON 请求优先使用泛型便捷函数；路径、查询和 Header 通过 Option 就地补充：

```go
result, err := httpclient.PostJSON[sendResult](ctx, s.client, "/messages", body,
	httpclient.WithPath("tenant", tenantID),
	httpclient.WithQuery("channel", "login"),
)
```

同一模型支持多值查询、query struct、表单、文件和原始正文：

```go
result, err := httpclient.PostForm[resultVO](ctx, s.client, "/token", url.Values{"scope": {"read", "write"}})
result, err = httpclient.PostMultipart[resultVO](ctx, s.client, "/files",
	httpclient.WithMultipartField("folder", "inbox"),
	httpclient.WithFile("file", name, reader),
)
resp, err := s.client.Do(ctx, http.MethodPut, "/objects/{id}",
	httpclient.WithPath("id", id),
	httpclient.WithRawBody(content, "application/octet-stream"),
)
```

可选入口包括 `GetJSON`、`PostJSON`、`PutJSON`、`PatchJSON`、`DeleteJSON`、`PostForm`、`PostMultipart`、`DoBytes`；任意 Method 或特殊协议使用 `Executor.Do`。大文件下载需要显式依赖 `httpclient.Streamer` 并关闭 `StreamResponse`。

## 错误与扩展

- 非 2xx 同时返回 `Response` 和 `*httpclient.Error`；用 `WithErrorResult` 解析结构化错误，用 `httpclient.IsStatus` 判断状态。
- trace/request ID 自动从 `context.Context` 透传，不要在每个业务调用中重复设置。
- 动态 Token 或请求签名在 boot 注册 `httpclient.WithServiceHooks`，不要把签名实现和密钥放进 Service。
- 不要在 Service 中创建 `http.Client{}`、使用 `http.DefaultClient`、硬编码 base URL/timeout/retry，或记录 Authorization、签名密钥及响应正文。
