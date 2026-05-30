# errors

统一错误码、HTTP 映射、响应结构、错误细节和错误包装由 `github.com/teamsillybees/initra/pkg/errors` 提供。业务模块不要直接依赖 `pkg/errors` 或 `github.com/samber/oops`，应通过项目内的 `internal/modules/bizerrors` 门面创建和包装错误。

`pkg/errors` 是 oops 的框架门面，不再定义自定义错误结构。业务错误码只在错误源头设置；后续层用 `Wrap` / `WrapContext` 保留根因、追加当前语义和必要排障上下文。日志、响应转换和 panic recover 只在 HTTP / Worker / CLI 等边界层统一处理。

## 标准用法

框架层或 `bizerrors` 门面可以使用 `apperrors.New` 创建校验错误和业务错误：

```go
return apperrors.New(apperrors.CodeBadRequest, "文件名不能为空")
```

框架层或 `bizerrors` 门面可以使用 `apperrors.WrapContext` 包装底层错误并补充内部 metadata：

```go
return apperrors.WrapContext(ctx, err, apperrors.CodeInternalError, "调用短信服务失败",
	apperrors.WithCauseDomain(apperrors.DomainHTTPClient),
	apperrors.WithCauseAttr("provider", "sms"),
)
```

`WithDetail` / `WithDetails` 只放允许返回给客户端的信息；SQL、token、password、secret、手机号、身份证、OSS object key、第三方完整响应体等排障字段应使用 `WithCauseAttr` 或 `WithCauseAttrs`，最终只进入边界日志。

业务专属错误码放在独立的 `internal/modules/bizerrors` package：

```go
const CodeLoginFailed apperrors.Code = "LOGIN_FAILED"

func LoginFailed() error {
	return apperrors.New(CodeLoginFailed, "登录失败", apperrors.WithStatus(http.StatusUnauthorized))
}

func WrapDBContext(ctx context.Context, err error, message string, opts ...Option) error {
	return apperrors.WrapContext(ctx, err, apperrors.CodeDBError, message,
		append([]Option{
			apperrors.WithCauseDomain(apperrors.DomainDB),
			apperrors.WithCauseHint(apperrors.HintDBConnection),
		}, opts...)...,
	)
}
```

## HTTP 响应映射

Server 已将 Huma 错误接入 `apperrors.ToHTTP`。Handler 返回 error 即可，由框架生成 JSON 响应。

不要在 handler 中手写错误响应结构。统一错误响应为：

```go
type ErrorVO struct {
	Code    string         `json:"code"`
	Message string         `json:"message"`
	Details map[string]any `json:"details,omitempty"`
	TraceID string         `json:"traceId,omitempty"`
}
```

`apperrors.ToHTTP` 从 oops 链读取最深层 code 并映射 HTTP 状态；`Public` 作为响应 message，5xx 默认返回 `internal error`。排障上下文和 stacktrace 只进入边界日志，不返回给前端。

## 错误边界

- Handler：只做必要的传输层校验，调用 service，返回 service error。
- Service：将业务错误和依赖错误转换为 `bizerrors`，优先使用 `WrapDBContext`、`WrapInternalContext`、`WrapCacheContext` 等 context 版本。
- Repository：保留数据库错误上下文；service 将 repository 错误映射为业务语义。
- 外部适配器：将框架/provider 错误映射为安全的应用错误。

## 禁止写法

- 不要新增 sentinel error 表达业务语义。
- 不要在业务模块中直接 import `github.com/samber/oops` 或 `github.com/teamsillybees/initra/pkg/errors`。
- 不要在 details 中暴露原始 SQL、token、凭证或完整远程响应体。
- 不要把底层 cause、stacktrace 或 oops attrs 返回给前端。
- 不要用 panic 表达业务错误。
- 不要吞掉错误；返回带上下文的错误。
- 不要在非边界层记录同一个错误。

## 检查清单

- 每个面向用户的错误是否都是 oops 包装后的 `error`？
- 业务错误码是否只在错误源头设置？
- 非默认业务错误码是否明确 HTTP 状态？
- 敏感细节是否已排除？
