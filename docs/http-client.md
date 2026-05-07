# HTTP Client 模块开发说明

## 1. 模块定位

`pkg/httpclient` 用于封装企业内部 Go 服务中的远程 HTTP 调用能力，基于 `go-resty` 实现。

目标不是简单包装 `http.Client`，而是提供：

* 多远程服务配置。
* 统一客户端创建与连接池复用。
* 统一认证处理。
* 统一请求与响应解析。
* 统一错误处理。
* 统一日志、超时、重试、下载等能力。

业务代码不直接使用 `resty.Client`，而是通过 `httpclient.Factory` 获取指定服务的客户端。

---

## 2. 目录结构

```text
pkg/httpclient/
  config.go        # 配置结构
  providers.go     # samber/do 自动装配入口
  factory.go       # Client 工厂，按服务名创建和缓存客户端
  client.go        # 对外请求方法
  request.go       # RequestOption 参数封装
  auth.go          # 认证处理
  response.go      # 响应解析
  errors.go        # 统一错误模型
  download.go      # 文件下载能力
```

---

## 3. 配置设计

```yaml
http_client:
  enabled: true
  timeout: 30s
  connect_timeout: 10s
  idle_conn_timeout: 90s
  max_idle_conns: 100
  max_idle_conns_per_host: 20

  services:
    user_center:
      base_url: https://api.example.com
      timeout: 30s
      headers:
        X-App-Id: initra
      auth:
        type: bearer
        token: ${USER_CENTER_TOKEN}
      retry:
        enabled: true
        count: 3
        wait_time: 500ms
        max_wait_time: 5s
      response:
        type: standard_api
```

---

## 4. 核心对象

### Factory

负责根据服务名创建和缓存客户端。

```go
type Factory interface {
    Get(serviceName string) (*Client, error)
    Clear(serviceName string)
    ClearAll()
}
```

要求：

* 每个远程服务只创建一个 `resty.Client`。
* 不允许每次请求都新建客户端。
* 未配置的服务名必须返回明确错误。
* 支持清理缓存，便于测试或配置刷新。

---

### Client

对业务暴露统一请求方法。

```go
type Client struct {
    name   string
    config ServiceConfig
    resty  *resty.Client
}
```

基础方法：

```go
func (c *Client) Get(ctx context.Context, path string, opts ...RequestOption) (*Response, error)
func (c *Client) Post(ctx context.Context, path string, body any, opts ...RequestOption) (*Response, error)
func (c *Client) Put(ctx context.Context, path string, body any, opts ...RequestOption) (*Response, error)
func (c *Client) Patch(ctx context.Context, path string, body any, opts ...RequestOption) (*Response, error)
func (c *Client) Delete(ctx context.Context, path string, opts ...RequestOption) (*Response, error)
```

泛型便捷方法：

```go
func GetJSON[T any](ctx context.Context, c *Client, path string, opts ...RequestOption) (*T, error)
func PostJSON[T any](ctx context.Context, c *Client, path string, body any, opts ...RequestOption) (*T, error)
```

---

## 5. 请求参数封装

使用 Option 模式，不要设计大量重载方法。

```go
type RequestOption func(*RequestOptions)

func WithHeader(key, value string) RequestOption
func WithHeaders(headers map[string]string) RequestOption
func WithQuery(key, value string) RequestOption
func WithQueryParams(params map[string]string) RequestOption
func WithPathParams(params map[string]string) RequestOption
func WithTimeout(timeout time.Duration) RequestOption
func WithResult(v any) RequestOption
func WithContentType(contentType string) RequestOption
func WithIdempotent(idempotent bool) RequestOption
```

使用示例：

```go
resp, err := client.Get(ctx, "/users",
    httpclient.WithQuery("page", "1"),
    httpclient.WithHeader("X-Trace-ID", traceID),
)
```

---

## 6. 认证能力

内置支持：

| 类型            | 说明                              |
| ------------- | ------------------------------- |
| none          | 不认证                             |
| bearer        | `Authorization: Bearer <token>` |
| basic         | Basic Auth                      |
| api_key       | 指定 Header 中放置 API Key           |
| custom_header | 静态认证 Header                     |

配置示例：

```yaml
auth:
  type: api_key
  header: X-API-Key
  value: ${API_KEY}
```

接口设计：

```go
type AuthHandler interface {
    Apply(ctx context.Context, r *resty.Request, cfg AuthConfig) error
}
```

---

## 7. 响应解析

支持三种响应模式：

| 类型           | 说明                          |
| ------------ | --------------------------- |
| raw          | 原始响应                        |
| json         | 直接 JSON 反序列化                |
| standard_api | 解析 `code/message/data` 标准结构 |

标准响应示例：

```json
{
  "code": 0,
  "message": "ok",
  "data": {}
}
```

默认成功规则：

* HTTP 状态码为 `2xx`。
* `code` 为 `0` 或 `200`。
* `success` 为 `true`。
* `status` 为 `success` 或 `ok`。

---

## 8. 错误模型

统一错误类型：

```go
type ErrorKind string

const (
    ErrorKindRequest  ErrorKind = "request"
    ErrorKindResponse ErrorKind = "response"
    ErrorKindInternal ErrorKind = "internal"
)

type Error struct {
    Kind       ErrorKind
    Service    string
    Method     string
    URL        string
    StatusCode int
    Code       string
    Message    string
    Cause      error
}
```

错误分类：

| 类型       | 场景                    |
| -------- | --------------------- |
| request  | 网络错误、连接失败、超时          |
| response | HTTP 非 2xx、业务 code 失败 |
| internal | 配置错误、响应解析失败、本地处理错误    |

可集成 `samber/oops`，但底层仍应保留明确的错误结构。

---

## 9. 重试策略

配置示例：

```yaml
retry:
  enabled: true
  count: 3
  wait_time: 500ms
  max_wait_time: 5s
  retry_status_codes: [408, 429, 500, 502, 503, 504]
```

约束：

* 默认只对网络错误、超时、`408`、`429`、`5xx` 重试。
* `GET`、`HEAD`、`PUT`、`DELETE` 可按配置重试。
* `POST` 默认不重试，除非显式设置 `WithIdempotent(true)`。
* 不对 `400`、`401`、`403`、`404` 重试。

---

## 10. 日志与追踪

每次请求记录：

* service name
* method
* url path
* status code
* duration
* error kind
* trace id

必须脱敏字段：

* `Authorization`
* `Cookie`
* `X-API-Key`
* `token`
* `password`
* `secret`

默认不打印完整请求体和响应体。Debug 日志只允许在开发环境开启。

---

## 11. 文件下载

提供基础下载能力：

```go
func (c *Client) DownloadBytes(ctx context.Context, path string, opts ...RequestOption) ([]byte, error)
func (c *Client) DownloadToFile(ctx context.Context, path string, target string, opts ...RequestOption) (*FileMetadata, error)
func (c *Client) Stream(ctx context.Context, path string, opts ...RequestOption) (io.ReadCloser, error)
func (c *Client) Metadata(ctx context.Context, path string, opts ...RequestOption) (*FileMetadata, error)
```

文件元数据：

```go
type FileMetadata struct {
    FileName              string
    FileSize              int64
    ContentType           string
    LastModified          time.Time
    ETag                  string
    SupportsRangeRequests bool
}
```

要求：

* 小文件可下载到内存。
* 大文件必须流式下载。
* `Stream` 返回的 `io.ReadCloser` 必须由调用方关闭。
* 下载到内存必须限制最大响应体大小。

---

## 12. 自动装配

提供 `providers.go`：

```go
func ProvideHTTPClient(i *do.Injector, cfg Config, logger *zap.Logger) {
    do.Provide(i, func(i *do.Injector) (*httpclient.Factory, error) {
        return httpclient.NewFactory(cfg.HTTPClient, logger)
    })
}
```

业务模块使用：

```go
type UserCenterClient struct {
    client *httpclient.Client
}

func NewUserCenterClient(factory *httpclient.Factory) (*UserCenterClient, error) {
    c, err := factory.Get("user_center")
    if err != nil {
        return nil, err
    }

    return &UserCenterClient{
        client: c,
    }, nil
}
```

不实现 Java 注解式 `@RemoteService`，Go 中使用构造函数显式声明依赖。

---

## 13. 开发优先级

### V1 必须完成

* 配置结构。
* Factory 创建和缓存客户端。
* GET / POST / PUT / PATCH / DELETE。
* Header / Query / Path Params。
* JSON 请求体。
* JSON 响应解析。
* Bearer / Basic / API Key 认证。
* Timeout。
* Retry。
* 统一错误模型。
* zap 日志脱敏。
* samber/do 自动装配。

### V2 后续增强

* `standard_api` 响应解析。
* 文件下载。
* HEAD 元数据读取。
* trace id 自动透传。
* Prometheus 指标。
* OAuth2 Client Credentials。

### V3 可选能力

* SSE。
* 熔断。
* 限流。
* OpenTelemetry。
* HMAC 请求签名。
* 配置热更新后自动重建客户端。

---

## 14. 实现约束

* 不要每次请求创建新的 `resty.Client`。
* 不要在业务代码中直接使用 `resty.Client`。
* 不要默认重试非幂等 POST。
* 不要输出敏感 Header。
* 不要无限制读取响应体到内存。
* 不要照搬 Java AOP、注解、反射扫描模式。
* 配置优先，代码可扩展，业务调用保持简单。
