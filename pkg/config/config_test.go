package config

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

type testConfig struct {
	App struct {
		Name string `mapstructure:"name"`
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
