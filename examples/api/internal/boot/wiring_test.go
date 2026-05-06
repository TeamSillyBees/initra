package boot

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/samber/do"
	"github.com/stretchr/testify/require"
	authmodule "github.com/teamsillybees/initra/examples/api/internal/module/auth"
	usermodule "github.com/teamsillybees/initra/examples/api/internal/module/user"
)

// TestModuleProvideDoesNotCollide 验证 auth 和 user 模块使用具名依赖后不会在 DI 容器中冲突。
func TestModuleProvideDoesNotCollide(t *testing.T) {
	injector := do.New()

	require.NotPanics(t, func() {
		usermodule.Provide(injector)
		authmodule.Provide(injector)
	})
}

// TestLoadConfigUsesRecommendedConfigShape 验证示例项目使用新的分组配置规范。
func TestLoadConfigUsesRecommendedConfigShape(t *testing.T) {
	configDir := t.TempDir()
	content := []byte(`
app:
  name: example-api
  env: local
  version: 0.1.0
  instance_id: local-1

server:
  addr: ":18080"
  read_timeout: 10s
  write_timeout: 30s
  idle_timeout: 60s
  shutdown_timeout: 20s

database:
  driver: postgres
  dsn: "postgres://postgres:postgres@127.0.0.1:5432/example?sslmode=disable"
  max_open_conns: 20
  max_idle_conns: 10
  conn_max_lifetime: 1h

redis:
  enabled: false
  addr: "127.0.0.1:6379"
  password: ""
  db: 0
  pool_size: 10

auth:
  enabled: true
  access_token_ttl: 15m
  refresh_token_ttl: 720h
  allow_multiple_devices: true
  jwt:
    issuer: example-api
    secret: "change-me"

log:
  level: info
  format: json
  output: stdout
  mask:
    enabled: true
    fields: ["password", "token", "secret", "authorization"]

observability:
  metrics:
    enabled: true
  tracing:
    enabled: false
  pprof:
    enabled: false
  health:
    enabled: true

casbin:
  model_path: ./configs/rbac_model.conf
  policy_path: ./configs/rbac_policy.csv

cache:
  local_ttl: 1m
  remote_ttl: 10m

idgen:
  node: 1
`)
	require.NoError(t, os.WriteFile(filepath.Join(configDir, "config.local.yaml"), content, 0o600))

	cfg, err := LoadConfig("local", configDir)

	require.NoError(t, err)
	require.Equal(t, "example-api", cfg.App.Name)
	require.Equal(t, "0.1.0", cfg.App.Version)
	require.Equal(t, "local-1", cfg.App.InstanceID)
	require.Equal(t, ":18080", cfg.Server.Addr)
	require.Equal(t, 20*time.Second, cfg.Server.ShutdownTimeout)
	require.Equal(t, time.Hour, cfg.Database.ConnMaxLifetime)
	require.False(t, cfg.Redis.Enabled)
	require.Equal(t, 10, cfg.Redis.PoolSize)
	require.True(t, cfg.Auth.Enabled)
	require.Equal(t, 720*time.Hour, cfg.Auth.RefreshTokenTTL)
	require.True(t, cfg.Auth.AllowMultipleDevices)
	require.Equal(t, "example-api", cfg.Auth.JWT.Issuer)
	require.Equal(t, "change-me", cfg.Auth.JWT.Secret)
	require.Equal(t, []string{"password", "token", "secret", "authorization"}, cfg.Log.Mask.Fields)
	require.True(t, cfg.Observability.Metrics.Enabled)

	safe := cfg.SafeForLog()
	auth := safe["auth"].(map[string]any)
	jwt := auth["jwt"].(map[string]any)
	require.Equal(t, "***", jwt["secret"])
}
