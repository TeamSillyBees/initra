# config

配置加载、运行环境选择、默认值、环境变量覆盖、启动校验和安全日志输出使用 `github.com/teamsillybees/initra/pkg/config`。

## 标准模式

业务项目在自己的 `internal/boot/config.go` 中定义配置聚合根，并组合 initra package 提供的配置结构：

```go
type Config struct {
	App        AppConfig              `mapstructure:"app"`
	Redis      redisx.Config          `mapstructure:"redis"`
	Log        logging.Config         `mapstructure:"log"`
	Storage    platformstorage.Config `mapstructure:"storage"`
	HTTPClient httpclient.Config      `mapstructure:"http_client"`
}
```

通过 `platformconfig.LoadInto` 加载：

```go
func LoadConfig(env string, configDir string) (*Config, error) {
	return platformconfig.LoadInto[Config](platformconfig.LoaderOptions{
		Env:       env,
		ConfigDir: configDir,
		Defaults:  configDefaults(),
	})
}
```

## 校验

业务配置实现 `Validate() error`。package 自带配置类型的校验要委托给对应类型：

```go
if err := c.Redis.Validate(); err != nil {
	return err
}
if err := c.Storage.Validate(); err != nil {
	return err
}
if err := c.HTTPClient.Validate(); err != nil {
	return err
}
```

## 环境规则

- 运行环境来自显式 CLI/env 输入、`APP_ENV` 或默认值 `dev`。
- 其他环境变量默认使用 `INITRA_` 前缀。
- 配置加载顺序为 defaults、`configs/config.yaml`、`configs/config.<env>.yaml`、环境变量。

## 安全日志

只暴露脱敏后的配置：

```go
func (c *Config) SafeForLog() map[string]any {
	return platformconfig.Sanitize(c, c.Log.Mask.Fields)
}
```

## 禁止写法

- 不要在业务项目中重复实现 Viper loader。
- 不要打印原始配置结构。
- `pkg/*` 已提供配置结构时，不要重新复制一份。
- 不要把业务配置聚合根放入 `pkg/config`；它应留在业务项目 `internal/boot`。

## 检查清单

- 配置是否组合 package config struct？
- 默认值是否足够支撑本地启动？
- `Validate` 是否能对关键配置 fail fast？
- secret 是否在日志输出前完成脱敏？
