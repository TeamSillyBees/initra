# errors

统一错误码、HTTP 映射、响应结构、错误细节和错误包装使用 `github.com/teamsillybees/initra/pkg/errors`，在业务代码中通常别名为 `apperrors`。

## 标准用法

使用 `apperrors.New` 创建校验错误和业务错误：

```go
return apperrors.New(apperrors.CodeBadRequest, "文件名不能为空")
```

使用 `apperrors.Wrap` 包装底层错误并补充上下文：

```go
return apperrors.Wrap(err, apperrors.CodeInternalError, "调用短信服务失败",
	apperrors.WithDetail("provider", "sms"),
)
```

业务专属错误码放在独立的 `internal/module/bizerrors` package：

```go
const CodeLoginFailed apperrors.Code = "LOGIN_FAILED"

func LoginFailed() *apperrors.AppError {
	return apperrors.New(CodeLoginFailed, "登录失败", apperrors.WithStatus(http.StatusUnauthorized))
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
	TraceID string         `json:"trace_id,omitempty"`
}
```

## 错误边界

- Handler：只做必要的传输层校验，调用 service，返回 service error。
- Service：将业务错误和依赖错误转换为 `apperrors`。
- Repository：保留数据库错误上下文；service 将 repository 错误映射为业务语义。
- 外部适配器：将框架/provider 错误映射为安全的应用错误。

## 禁止写法

- 不要新增 sentinel error 表达业务语义。
- 不要在 details 中暴露原始 SQL、token、凭证或完整远程响应体。
- 不要用 panic 表达业务错误。
- 不要吞掉错误；返回带上下文的错误。

## 检查清单

- 每个面向用户的错误是否都是 `apperrors.AppError` 或已包装成它？
- 业务错误码是否集中定义？
- 非默认业务错误码是否明确 HTTP 状态？
- 敏感细节是否已排除？
