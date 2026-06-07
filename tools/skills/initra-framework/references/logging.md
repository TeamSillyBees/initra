# 日志

## Boot 注册

```go
logx.Register(injector, cfg.Log)
```

业务代码注入 `*logx.Logger`，不要初始化第二套 zap logger。

## 使用方式

```go
logger.Info(ctx, "user profile refreshed",
	logx.String("user_id", userID.String()),
	logx.Duration("cost", cost),
)
```

错误日志传入 error，让 logx 提取 oops code、domain、trace 等上下文：

```go
logger.Error(ctx, "publish task failed", err, logx.String("task_type", taskType))
```

## 脱敏

- 配置日志使用项目 `SafeForLog()`。
- HTTP Client、request log 和错误上下文不要包含密码、token、验证码、session value、Authorization、access key 或带密码 DSN。
- `log.redact.fields` 应包含 `password`、`token`、`secret`、`authorization`。

## 禁止

- 不要在业务代码调用 `zap.NewProduction` / `zap.NewDevelopment`。
- 不要把 JSONL 日志路径配置成 `stderr`。
