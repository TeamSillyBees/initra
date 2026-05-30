# errors

统一错误码、HTTP 映射、响应结构、错误细节和错误包装由 `github.com/teamsillybees/initra/pkg/errors` 提供。业务模块不要直接依赖 `pkg/errors` 或 `github.com/samber/oops`，应通过项目内的 `internal/modules/bizerrors` 门面创建和包装错误。

`AppError` 是业务错误模型，负责业务错误码、HTTP 状态码、对外 message 和对外 details；`oops` 只作为底层 cause 增强器，负责 stacktrace、domain、hint、trace 和内部 attrs。

## 标准用法

框架层或 `bizerrors` 门面可以使用 `apperrors.New` 创建校验错误和业务错误：

```go
return apperrors.New(apperrors.CodeBadRequest, "文件名不能为空")
```

框架层或 `bizerrors` 门面可以使用 `apperrors.WrapContext` 包装底层错误并补充内部 cause metadata：

```go
return apperrors.WrapContext(ctx, err, apperrors.CodeInternalError, "调用短信服务失败",
	apperrors.WithCauseDomain(apperrors.DomainHTTPClient),
	apperrors.WithCauseAttr("provider", "sms"),
)
```

`Details` 只放允许返回给客户端的信息；SQL、token、password、secret、手机号、身份证、OSS object key、第三方完整响应体等排障字段应使用 `WithCauseAttr` 或 `WithCauseAttrs`，最终只进入日志。

业务专属错误码放在独立的 `internal/modules/bizerrors` package：

```go
const CodeLoginFailed apperrors.Code = "LOGIN_FAILED"

func LoginFailed() *apperrors.AppError {
	return apperrors.New(CodeLoginFailed, "登录失败", apperrors.WithStatus(http.StatusUnauthorized))
}

func WrapDBContext(ctx context.Context, err error, message string, opts ...Option) *apperrors.AppError {
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

## 检查清单

- 每个面向用户的错误是否都是 `apperrors.AppError` 或已包装成它？
- 业务错误码是否集中定义？
- 非默认业务错误码是否明确 HTTP 状态？
- 敏感细节是否已排除？
