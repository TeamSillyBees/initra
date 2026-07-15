# 统一错误

## 入口

统一使用 `pkg/errors`，通常别名为 `apperrors`：

```go
return apperrors.New(apperrors.CodeBadRequest, "用户名不能为空")
return apperrors.WrapContext(ctx, err, apperrors.CodeDBError, "查询用户失败",
	apperrors.WithCauseDomain(apperrors.DomainDB),
)
```

业务专属错误码放在业务错误包中，用 `apperrors.New` / `Wrap` 封装，不在模块内散落 sentinel error。

## HTTP 映射

HTTP 边界复用框架 mapper 和响应机制。`pkg/server` 的 Huma error mapper 通过 `apperrors.ToHTTP(err, traceID)` 生成：

```json
{"code":"BAD_REQUEST","message":"用户名不能为空","traceId":"..."}
```

## 公开信息

- `WithPublic` 和 `WithDetail` 可以进入 HTTP 响应。
- `WithCauseDomain`、`WithCauseHint`、`WithCauseAttr` 只用于日志排障。
- `WrapContext` 会从 context 提取 trace id。

## 禁止

- 不要在 handler 手写错误 JSON。
- 不要把数据库、Redis、对象存储、下游 HTTP 的原始敏感错误直接公开给用户。
- 不要吞掉错误；向上返回并保留上下文。
