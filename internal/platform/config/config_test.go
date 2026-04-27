package config

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// TestLoadAppliesDefaultsAndEnvironmentOverrides 验证配置默认值和环境变量覆盖同时生效。
func TestLoadAppliesDefaultsAndEnvironmentOverrides(t *testing.T) {
	t.Setenv("INITRA_APP_PORT", "9091")
	t.Setenv("INITRA_JWT_SECRET", "override-secret")

	configDir := t.TempDir()
	configPath := filepath.Join(configDir, "config.local.yaml")

	content := []byte(`
app:
  name: initra
  env: local
database:
  dsn: postgres://postgres:postgres@127.0.0.1:5432/initra?sslmode=disable
redis:
  addr: 127.0.0.1:6379
casbin:
  model_path: ./configs/rbac_model.conf
  policy_path: ./configs/rbac_policy.csv
jwt:
  issuer: initra
`)
	require.NoError(t, os.WriteFile(configPath, content, 0o600))

	cfg, err := Load(LoaderOptions{
		Env:       "local",
		ConfigDir: configDir,
	})
	require.NoError(t, err)
	require.Equal(t, "initra", cfg.App.Name)
	require.Equal(t, "local", cfg.App.Env)
	require.Equal(t, 9091, cfg.App.Port)
	require.Equal(t, "override-secret", cfg.JWT.Secret)
	require.Equal(t, "json", cfg.Logger.Format)
	require.Equal(t, 5*time.Second, cfg.HTTP.ReadTimeout)
	require.Equal(t, 10*time.Minute, cfg.Cache.RemoteTTL)
	require.Equal(t, int64(1), cfg.IDGen.Node)
}

// TestValidateRejectsRefreshTokenTTLNotGreaterThanAccessTokenTTL 验证 refresh token 生命周期必须长于 access token。
func TestValidateRejectsRefreshTokenTTLNotGreaterThanAccessTokenTTL(t *testing.T) {
	cfg := Config{
		App: AppConfig{
			Name: "initra",
			Env:  "test",
			Port: 8080,
		},
		Database: DatabaseConfig{
			DSN: "postgres://postgres:postgres@127.0.0.1:5432/initra?sslmode=disable",
		},
		JWT: JWTConfig{
			Issuer:          "initra",
			Secret:          "test-secret",
			AccessTokenTTL:  30 * time.Minute,
			RefreshTokenTTL: 15 * time.Minute,
		},
		Casbin: CasbinConfig{
			ModelPath:  "./configs/rbac_model.conf",
			PolicyPath: "./configs/rbac_policy.csv",
		},
	}

	err := cfg.Validate()
	require.Error(t, err)
	require.Contains(t, err.Error(), "jwt.refresh_token_ttl")
}

// TestValidateRejectsIDGenNodeOutsideSnowflakeRange 验证雪花节点号不能超出 10 位节点范围。
func TestValidateRejectsIDGenNodeOutsideSnowflakeRange(t *testing.T) {
	cfg := validConfigForTest()
	cfg.IDGen.Node = 1024

	err := cfg.Validate()
	require.Error(t, err)
	require.Contains(t, err.Error(), "idgen.node")
}

// validConfigForTest 构造一份通过基础校验的配置，供校验失败测试定向修改。
func validConfigForTest() Config {
	return Config{
		App: AppConfig{
			Name: "initra",
			Env:  "test",
			Port: 8080,
		},
		Database: DatabaseConfig{
			DSN: "postgres://postgres:postgres@127.0.0.1:5432/initra?sslmode=disable",
		},
		JWT: JWTConfig{
			Issuer:          "initra",
			Secret:          "test-secret",
			AccessTokenTTL:  15 * time.Minute,
			RefreshTokenTTL: 24 * time.Hour,
		},
		Casbin: CasbinConfig{
			ModelPath:  "./configs/rbac_model.conf",
			PolicyPath: "./configs/rbac_policy.csv",
		},
		IDGen: IDGenConfig{
			Node: 1,
		},
	}
}
