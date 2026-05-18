# logging

zap logger 构造、输出格式、日志级别和敏感字段脱敏使用 `github.com/teamsillybees/initra/pkg/logging`。

## 标准装配

在 boot 层优先注册 logger：

```go
logging.Register(injector, cfg.Log)
```

其他 framework 注册函数可能会从 injector 中解析 `*zap.Logger`。

## 配置与脱敏

业务配置中组合 `logging.Config`：

```go
Log logging.Config `mapstructure:"log"`
```

默认脱敏字段包含 password、token、secret、authorization 和 access key 风格字段。项目特有敏感字段放入 `log.mask.fields`。

只通过脱敏方法打印配置：

```go
func (c *Config) SafeForLog() map[string]any {
	return platformconfig.Sanitize(c, c.Log.Mask.Fields)
}
```

## 业务用法

在 boot 或 provider 层解析 logger，并只注入真正需要日志的组件。Service 应保持聚焦，不要因为 logger 可用就到处添加日志。

优先使用结构化字段：

```go
logger.Info("短信发送成功",
	zap.String("provider", provider),
	zap.String("trace_id", traceID),
)
```

## 禁止写法

- 不要在业务代码中调用 `zap.NewProduction`、`zap.NewDevelopment` 或构建第二套全局 logger。
- 不要记录密码、token、验证码、session value、Authorization header、access key 或带密码 DSN。
- 不要记录原始 request/response body，除非内容已明确脱敏且体积很小。
- 不要把日志脱敏当成唯一防线；更好的方式是不把 secret 放入日志字段。

## 检查清单

- logging 是否早于依赖 `*zap.Logger` 的 package 注册？
- 配置和应用元数据是否通过 `SafeForLog` 输出？
- 敏感字段名称是否覆盖完整？
- 日志字段是否结构化且低基数？
