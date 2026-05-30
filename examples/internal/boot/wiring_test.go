package boot

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	authmodule "github.com/teamsillybees/initra/examples/internal/modules/auth"
	filemodule "github.com/teamsillybees/initra/examples/internal/modules/file"
	httpdemomodule "github.com/teamsillybees/initra/examples/internal/modules/httpdemo"
	taskdemomodule "github.com/teamsillybees/initra/examples/internal/modules/taskdemo"
	usermodule "github.com/teamsillybees/initra/examples/internal/modules/user"

	"github.com/samber/do"
	"github.com/stretchr/testify/require"
	"github.com/teamsillybees/initra/pkg/observability"
)

// TestModuleProvideDoesNotCollide 验证 auth、user 和 file 模块使用具名依赖后不会在 DI 容器中冲突。
func TestModuleProvideDoesNotCollide(t *testing.T) {
	injector := do.New()

	require.NotPanics(t, func() {
		usermodule.Provide(injector)
		authmodule.Provide(injector)
		filemodule.Provide(injector)
		httpdemomodule.Provide(injector)
		taskdemomodule.Provide(injector)
	})
}

// TestRegisterProvidersDoesNotConnectDuringRegistration 验证注册 provider 本身不触发数据库连接。
func TestRegisterProvidersDoesNotConnectDuringRegistration(t *testing.T) {
	injector := do.New()
	cfg := &Config{
		App: AppConfig{
			Name: "example-api",
			Env:  "test",
		},
		Auth: AuthConfig{
			AccessTokenTTL:  time.Minute,
			RefreshTokenTTL: time.Hour,
			JWT: JWTConfig{
				Issuer: "example-api",
				Secret: "secret",
			},
		},
		IDGen: IDGenConfig{Node: 1},
	}

	require.NotPanics(t, func() {
		registerProviders(injector, cfg, observability.BuildInfoVO{Version: "test"})
	})
}

// TestLoadConfigUsesRecommendedConfigShape 验证示例项目使用新的分组配置规范。
func TestLoadConfigUsesRecommendedConfigShape(t *testing.T) {
	configDir := t.TempDir()
	content := []byte(`
app:
  name: example-api
  version: 0.1.0
  instance_id: base-1

server:
  addr: ":8080"
  read_timeout: 10s
  write_timeout: 30s
  idle_timeout: 60s
  shutdown_timeout: 20s

database:
  host: "127.0.0.1"
  port: 5432
  user: "postgres"
  password: "postgres"
  dbname: "example"
  max_open_conns: 20
  max_idle_conns: 10
  conn_max_lifetime: 1h

redis:
  enabled: false
  addr: "127.0.0.1:6379"
  password: ""
  db: 0
  pool:
    size: 10

auth:
  enabled: true
  access_token_ttl: 15m
  refresh_token_ttl: 720h
  allow_multiple_devices: true
  jwt:
    issuer: example-api
    secret: "base-secret"

log:
  level: info
  console:
    enabled: true
    level: debug
    output: stderr
  jsonl:
    enabled: true
    level: info
    path: stdout
  redact:
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

storage:
  enabled: true
  provider: local
  presign_default_ttl: 15m
  local:
    root_dir: ./var/uploads
    temp_dir: .multipart
    generate_date_path: true
    allowed_extensions: ["txt", "md", "png", "jpg", "jpeg", "gif", "pdf"]
    max_size: 10485760

http_client:
  enabled: true
  timeout: 30s
  connect_timeout: 10s
  idle_conn_timeout: 90s
  max_idle_conns: 100
  max_idle_conns_per_host: 20
  max_response_body_size: 10485760
  proxy: http://127.0.0.1:7890
  services:
    httpbingo:
      base_url: https://httpbingo.org
      timeout: 10s
      proxy: http://127.0.0.1:7891
      headers:
        X-App-Id: example-api
      retry:
        enabled: true
        count: 2
        wait_time: 200ms
        max_wait_time: 2s

idgen:
  node: 1
`)
	override := []byte(`
app:
  instance_id: local-1

server:
  addr: ":18080"

auth:
  jwt:
    secret: "change-me"
`)
	require.NoError(t, os.WriteFile(filepath.Join(configDir, "config.yaml"), content, 0o600))
	require.NoError(t, os.WriteFile(filepath.Join(configDir, "config.local.yaml"), override, 0o600))

	cfg, err := LoadConfig("local", configDir)

	require.NoError(t, err)
	require.Equal(t, "example-api", cfg.App.Name)
	require.Equal(t, "local", cfg.App.Env)
	require.Equal(t, "0.1.0", cfg.App.Version)
	require.Equal(t, "local-1", cfg.App.InstanceID)
	require.Equal(t, ":18080", cfg.Server.Addr)
	require.Equal(t, 20*time.Second, cfg.Server.ShutdownTimeout)
	require.Equal(t, time.Hour, cfg.Database.ConnMaxLifetime)
	require.False(t, cfg.Redis.Enabled)
	require.Equal(t, 10, cfg.Redis.Pool.Size)
	require.True(t, cfg.Auth.Enabled)
	require.Equal(t, 720*time.Hour, cfg.Auth.RefreshTokenTTL)
	require.True(t, cfg.Auth.AllowMultipleDevices)
	require.Equal(t, "example-api", cfg.Auth.JWT.Issuer)
	require.Equal(t, "change-me", cfg.Auth.JWT.Secret)
	require.Equal(t, []string{"password", "token", "secret", "authorization"}, cfg.Log.Redact.Fields)
	require.True(t, cfg.Observability.Metrics.Enabled)
	require.True(t, cfg.Storage.Enabled)
	require.Equal(t, "./var/uploads", cfg.Storage.Local.RootDir)
	require.Equal(t, int64(10*1024*1024), cfg.Storage.Local.MaxSize)
	require.True(t, cfg.HTTPClient.Enabled)
	require.Equal(t, 30*time.Second, cfg.HTTPClient.Timeout)
	require.Equal(t, "http://127.0.0.1:7890", cfg.HTTPClient.Proxy)
	require.Equal(t, "https://httpbingo.org", cfg.HTTPClient.Services["httpbingo"].BaseURL)
	require.Equal(t, "http://127.0.0.1:7891", cfg.HTTPClient.Services["httpbingo"].Proxy)
	require.Equal(t, "example-api", cfg.HTTPClient.Services["httpbingo"].Headers["x-app-id"])

	safe := cfg.SafeForLog()
	auth := safe["auth"].(map[string]any)
	jwt := auth["jwt"].(map[string]any)
	require.Equal(t, "***", jwt["secret"])
}
