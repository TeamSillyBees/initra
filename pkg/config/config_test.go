package config

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

type testConfig struct {
	App struct {
		Name string `mapstructure:"name"`
		Env  string `mapstructure:"env"`
		Port int    `mapstructure:"port"`
	} `mapstructure:"app"`
	HTTP struct {
		ReadTimeout time.Duration `mapstructure:"read_timeout"`
	} `mapstructure:"http"`
	Database struct {
		Host string `mapstructure:"host"`
	} `mapstructure:"database"`
}

func (c *testConfig) Validate() error {
	if c.App.Name == "" {
		return errMissingAppName
	}
	return nil
}

var errMissingAppName = os.ErrInvalid

// TestLoadIntoAppliesDefaultsAndEnvironmentOverrides 验证泛型加载器不绑定具体应用配置结构。
func TestLoadIntoAppliesDefaultsAndEnvironmentOverrides(t *testing.T) {
	t.Setenv("APP_ENV", "prod")
	t.Setenv("INITRA_APP_PORT", "9091")

	configDir := t.TempDir()
	writeConfigYAML(t, configDir, "config.yaml", `
app:
  name: initra
`)
	writeConfigYAML(t, configDir, "config.local.yaml", `
app:
  name: initra-local
`)

	cfg, err := LoadInto[testConfig](LoaderOptions{
		Env:       "local",
		ConfigDir: configDir,
		Defaults: map[string]any{
			"app.port":          8080,
			"http.read_timeout": "5s",
		},
	})

	require.NoError(t, err)
	require.Equal(t, "initra-local", cfg.App.Name)
	require.Equal(t, "local", cfg.App.Env)
	require.Equal(t, 9091, cfg.App.Port)
	require.Equal(t, 5*time.Second, cfg.HTTP.ReadTimeout)
}

// TestLoadIntoSupportsEnvironmentOnlyConfig 验证没有配置文件时仍可完全依赖环境变量加载。
func TestLoadIntoSupportsEnvironmentOnlyConfig(t *testing.T) {
	t.Setenv("APP_ENV", "prod")
	t.Setenv("INITRA_APP_NAME", "initra-env")
	t.Setenv("INITRA_APP_PORT", "9092")
	t.Setenv("INITRA_HTTP_READ_TIMEOUT", "3s")

	cfg, err := LoadInto[testConfig](LoaderOptions{
		ConfigDir: t.TempDir(),
	})

	require.NoError(t, err)
	require.Equal(t, "initra-env", cfg.App.Name)
	require.Equal(t, "prod", cfg.App.Env)
	require.Equal(t, 9092, cfg.App.Port)
	require.Equal(t, 3*time.Second, cfg.HTTP.ReadTimeout)
}

// TestLoadIntoAllowsBaseConfigWithoutEnvironmentConfig 验证只提供基础配置文件时不会强制要求环境配置文件。
func TestLoadIntoAllowsBaseConfigWithoutEnvironmentConfig(t *testing.T) {
	configDir := t.TempDir()
	writeConfigYAML(t, configDir, "config.yaml", `
app:
  name: initra
  port: 8080
`)

	cfg, err := LoadInto[testConfig](LoaderOptions{
		Env:       "local",
		ConfigDir: configDir,
	})

	require.NoError(t, err)
	require.Equal(t, "local", cfg.App.Env)
	require.Equal(t, 8080, cfg.App.Port)
}

// TestLoadIntoAllowsEnvironmentConfigWithoutBaseConfig 验证只提供环境配置文件时不会强制要求基础配置文件。
func TestLoadIntoAllowsEnvironmentConfigWithoutBaseConfig(t *testing.T) {
	configDir := t.TempDir()
	writeConfigYAML(t, configDir, "config.local.yaml", `
app:
  name: initra-local
  port: 18080
`)

	cfg, err := LoadInto[testConfig](LoaderOptions{
		Env:       "local",
		ConfigDir: configDir,
	})

	require.NoError(t, err)
	require.Equal(t, "local", cfg.App.Env)
	require.Equal(t, "initra-local", cfg.App.Name)
	require.Equal(t, 18080, cfg.App.Port)
}

// TestLoadIntoRejectsInvalidExistingConfig 验证存在但格式错误的配置文件仍会 fail fast。
func TestLoadIntoRejectsInvalidExistingConfig(t *testing.T) {
	configDir := t.TempDir()
	writeConfigYAML(t, configDir, "config.yaml", "app:\n  name: initra\n  port: [broken\n")

	_, err := LoadInto[testConfig](LoaderOptions{
		ConfigDir: configDir,
	})

	require.Error(t, err)
	require.Contains(t, err.Error(), "读取基础配置文件 config.yaml 失败")
}

// TestLoadIntoDefaultsToDevEnv 验证未指定运行环境时默认读取 dev 配置。
func TestLoadIntoDefaultsToDevEnv(t *testing.T) {
	t.Setenv("APP_ENV", "")

	configDir := t.TempDir()
	writeConfigYAML(t, configDir, "config.yaml", `
app:
  name: initra
  port: 8080
`)
	writeConfigYAML(t, configDir, "config.dev.yaml", `
app:
  port: 8081
`)

	cfg, err := LoadInto[testConfig](LoaderOptions{
		ConfigDir: configDir,
	})

	require.NoError(t, err)
	require.Equal(t, "dev", cfg.App.Env)
	require.Equal(t, 8081, cfg.App.Port)
}

// TestLoadIntoUsesAPPENVToSelectEnvironment 验证运行环境由无前缀 APP_ENV 选择，且不接受 YAML 中的 app.env。
func TestLoadIntoUsesAPPENVToSelectEnvironment(t *testing.T) {
	t.Setenv("APP_ENV", "prod")
	t.Setenv("INITRA_APP_PORT", "10080")

	configDir := t.TempDir()
	writeConfigYAML(t, configDir, "config.yaml", `
app:
  name: initra
  env: file-local
  port: 8080
database:
  host: "127.0.0.1"
`)
	writeConfigYAML(t, configDir, "config.prod.yaml", `
app:
  env: file-prod
  port: 9090
database:
  host: ""
`)

	cfg, err := LoadInto[testConfig](LoaderOptions{
		ConfigDir: configDir,
	})

	require.NoError(t, err)
	require.Equal(t, "prod", cfg.App.Env)
	require.Equal(t, 10080, cfg.App.Port)
	require.Empty(t, cfg.Database.Host)
}

// TestLoadIntoRunsCustomValidator 验证业务项目可以在自己的配置结构上定义启动校验。
func TestLoadIntoRunsCustomValidator(t *testing.T) {
	configDir := t.TempDir()
	writeConfigYAML(t, configDir, "config.yaml", "app:\n  port: 8080\n")
	writeConfigYAML(t, configDir, "config.local.yaml", "{}\n")

	_, err := LoadInto[testConfig](LoaderOptions{
		Env:       "local",
		ConfigDir: configDir,
	})

	require.ErrorIs(t, err, errMissingAppName)
}

// TestSanitizeMasksSensitiveConfig 验证配置打印前会递归脱敏常见敏感字段。
func TestSanitizeMasksSensitiveConfig(t *testing.T) {
	cfg := struct {
		Database struct {
			Password string `mapstructure:"password"`
		} `mapstructure:"database"`
		Redis struct {
			Password string `mapstructure:"password"`
		} `mapstructure:"redis"`
		Auth struct {
			JWT struct {
				Secret         string        `mapstructure:"secret"`
				AccessTokenTTL time.Duration `mapstructure:"access_token_ttl"`
			} `mapstructure:"jwt"`
		} `mapstructure:"auth"`
		Headers map[string]string `mapstructure:"headers"`
	}{}
	cfg.Database.Password = "db-password"
	cfg.Redis.Password = "redis-password"
	cfg.Auth.JWT.Secret = "jwt-secret"
	cfg.Auth.JWT.AccessTokenTTL = 15 * time.Minute
	cfg.Headers = map[string]string{"authorization": "Bearer raw-token"}

	sanitized := Sanitize(cfg, []string{"authorization"})
	printable := fmt.Sprint(sanitized)

	require.NotContains(t, printable, "db-password")
	require.NotContains(t, printable, "redis-password")
	require.NotContains(t, printable, "jwt-secret")
	require.NotContains(t, printable, "raw-token")
	database := sanitized["database"].(map[string]any)
	require.Equal(t, "***", database["password"])

	auth := sanitized["auth"].(map[string]any)
	jwt := auth["jwt"].(map[string]any)
	require.Equal(t, "***", jwt["secret"])
	require.Equal(t, 15*time.Minute, jwt["access_token_ttl"])
}

func writeConfigYAML(t *testing.T, dir string, name string, content string) {
	t.Helper()
	require.NoError(t, os.WriteFile(filepath.Join(dir, name), []byte(content), 0o600))
}
