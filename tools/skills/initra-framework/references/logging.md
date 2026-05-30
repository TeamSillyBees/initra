# logx

zap logger 构造、console/jsonl 双输出、oops 错误字段提取和敏感字段脱敏使用 `github.com/teamsillybees/initra/pkg/logx`。

## 输出策略

- `console` 是面向终端阅读的多行文本格式，默认写入 `stderr`，用于本地开发和容器标准错误流。
- `jsonl` 是面向机器检索的 JSON Lines 文件日志，默认写入 `./var/logs/app.jsonl`。
- 不要把 `jsonl.path` 配置为 `stdout` 或 `stderr`；这两个值会被视为无效配置。

## 标准装配

在 boot 层优先注册 logger：

```go
logx.Register(injector, cfg.Log)
```

业务和框架边界日志统一从 injector 中解析 `*logx.Logger`，不要在业务代码中直接依赖 zap。

## 配置与脱敏

业务配置中组合 `logx.Config`：

```go
Log logx.Config `mapstructure:"log"`
```

推荐配置：

```yaml
log:
  console:
    enabled: true
    stack: short
    output: stderr
  jsonl:
    enabled: true
    stack: full
    path: ./var/logs/app.jsonl
```

默认脱敏字段包含 password、token、secret、authorization、cookie、body、DSN 和 access key 风格字段。项目特有敏感字段放入 `log.redact.fields`。

只通过脱敏方法打印配置：

```go
func (c *Config) SafeForLog() map[string]any {
	return platformconfig.Sanitize(c, c.Log.Redact.Fields)
}
```

## 业务用法

在 boot 或 provider 层解析 logger，并只注入真正需要日志的组件。Service 应保持聚焦，不要因为 logger 可用就到处添加日志。

优先使用结构化字段：

```go
logger.Info(ctx, "短信发送成功",
	logx.String("provider", provider),
	logx.String("trace_id", traceID),
)
```

## 禁止写法

- 不要在业务代码中调用 `zap.NewProduction`、`zap.NewDevelopment` 或构建第二套全局 logger。
- 不要把 `log.jsonl.path` 配置为 `stdout` 或 `stderr`，JSONL 只作为文件日志使用。
- 不要记录密码、token、验证码、session value、Authorization header、access key 或带密码 DSN。
- 不要记录原始 request/response body，除非内容已明确脱敏且体积很小。
- 不要把日志脱敏当成唯一防线；更好的方式是不把 secret 放入日志字段。

## 检查清单

- logx 是否早于依赖 logger 的 package 注册？
- 配置和应用元数据是否通过 `SafeForLog` 输出？
- 敏感字段名称是否覆盖完整？
- 日志字段是否结构化且低基数？
