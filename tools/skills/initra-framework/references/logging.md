# logx

zap logger 构造、console/jsonl 双输出、oops 错误字段提取和敏感字段脱敏使用 `github.com/teamsillybees/initra/pkg/logx`。

## 输出策略

- `console` 是面向终端阅读的多行文本格式，默认写入 `stderr`，用于本地开发和容器标准错误流。
- `jsonl` 是面向机器检索的 JSON Lines 日志，可写入 `stdout` 供日志采集系统采集，也可写入文件。
- `jsonl.path: stdout` 时会忽略并关闭 `console` 输出，避免同一标准输出流混入人眼格式日志。
- 文件 JSONL 支持可选 `rotation`，启用后按日期生成文件，并可配置单文件大小阈值。
- 不要把 `jsonl.path` 配置为 `stderr`；JSONL 只支持 stdout 或文件路径。

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
    rotation:
      enabled: true
      date_format: "2006-01-02"
      max_size_mb: 100
```

生产环境可直接输出 JSON Lines 到 stdout：

```yaml
log:
  console:
    enabled: false
  jsonl:
    enabled: true
    stack: full
    path: stdout
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
- 不要把 `log.jsonl.path` 配置为 `stderr`；stdout 仅用于 JSON Lines 采集模式。
- 不要记录密码、token、验证码、session value、Authorization header、access key 或带密码 DSN。
- 不要记录原始 request/response body，除非内容已明确脱敏且体积很小。
- 不要把日志脱敏当成唯一防线；更好的方式是不把 secret 放入日志字段。

## 检查清单

- logx 是否早于依赖 logger 的 package 注册？
- 配置和应用元数据是否通过 `SafeForLog` 输出？
- 敏感字段名称是否覆盖完整？
- 日志字段是否结构化且低基数？
