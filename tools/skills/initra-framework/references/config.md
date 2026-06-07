# 配置

## 入口

业务项目在自己的 `internal/boot/config.go` 定义聚合配置，并组合 pkg 配置结构：

```go
type Config struct {
	Redis      redisx.Config     `mapstructure:"redis"`
	Log        logx.Config       `mapstructure:"log"`
	Storage    storage.Config    `mapstructure:"storage"`
	HTTPClient httpclient.Config `mapstructure:"http_client"`
	Task       task.Config       `mapstructure:"task"`
}
```

加载使用 `pkg/config`：

```go
func LoadConfig(env string, configDir string) (*Config, error) {
	return platformconfig.LoadInto[Config](platformconfig.LoaderOptions{
		Env:       env,
		ConfigDir: configDir,
		Defaults:  configDefaults(),
	})
}
```

## 校验与脱敏

- 在 `Validate()` 中调用各 pkg 配置的 `Validate()`，再校验业务字段。
- 用 `platformconfig.Sanitize(c, c.Log.Redact.Fields)` 或项目 `SafeForLog()` 打印配置。
- `APP_ENV` 表示运行环境；其他环境变量默认使用 `INITRA_` 前缀。

## 禁止

- 不要在业务项目重新实现 Viper loader。
- 不要在日志输出原始密码、token、secret、Authorization、access key 或带密码 DSN。
