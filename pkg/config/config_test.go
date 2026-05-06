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
	t.Setenv("INITRA_APP_PORT", "9091")

	configDir := t.TempDir()
	configPath := filepath.Join(configDir, "config.local.yaml")
	content := []byte(`
app:
  name: initra
`)
	require.NoError(t, os.WriteFile(configPath, content, 0o600))

	cfg, err := LoadInto[testConfig](LoaderOptions{
		Env:       "local",
		ConfigDir: configDir,
		Defaults: map[string]any{
			"app.port":          8080,
			"http.read_timeout": "5s",
		},
	})

	require.NoError(t, err)
	require.Equal(t, "initra", cfg.App.Name)
	require.Equal(t, 9091, cfg.App.Port)
	require.Equal(t, 5*time.Second, cfg.HTTP.ReadTimeout)
}

// TestLoadIntoAllowsAPPENVToOverrideAppEnv 验证运行环境可以用无前缀 APP_ENV 覆盖配置文件中的 app.env。
func TestLoadIntoAllowsAPPENVToOverrideAppEnv(t *testing.T) {
	t.Setenv("APP_ENV", "prod")

	configDir := t.TempDir()
	configPath := filepath.Join(configDir, "config.local.yaml")
	content := []byte(`
app:
  name: initra
  env: local
`)
	require.NoError(t, os.WriteFile(configPath, content, 0o600))

	cfg, err := LoadInto[testConfig](LoaderOptions{
		Env:       "local",
		ConfigDir: configDir,
	})

	require.NoError(t, err)
	require.Equal(t, "prod", cfg.App.Env)
}

// TestLoadIntoRunsCustomValidator 验证业务项目可以在自己的配置结构上定义启动校验。
func TestLoadIntoRunsCustomValidator(t *testing.T) {
	configDir := t.TempDir()
	configPath := filepath.Join(configDir, "config.local.yaml")
	require.NoError(t, os.WriteFile(configPath, []byte("app:\n  port: 8080\n"), 0o600))

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
			DSN string `mapstructure:"dsn"`
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
	cfg.Database.DSN = "postgres://postgres:db-password@127.0.0.1:5432/initra?sslmode=disable"
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
	require.Contains(t, fmt.Sprint(sanitized["database"]), "postgres:***@127.0.0.1")

	auth := sanitized["auth"].(map[string]any)
	jwt := auth["jwt"].(map[string]any)
	require.Equal(t, "***", jwt["secret"])
	require.Equal(t, 15*time.Minute, jwt["access_token_ttl"])
}
