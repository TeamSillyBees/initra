package boot

import (
	"fmt"
	"time"

	"github.com/teamsillybees/initra/examples/api/internal/data"
	platformconfig "github.com/teamsillybees/initra/pkg/config"
	"github.com/teamsillybees/initra/pkg/logging"
)

// Config 是示例项目的应用配置聚合根。
type Config struct {
	App           AppConfig           `mapstructure:"app"`
	Server        ServerConfig        `mapstructure:"server"`
	Database      data.DatabaseConfig `mapstructure:"database"`
	Redis         RedisConfig         `mapstructure:"redis"`
	Auth          AuthConfig          `mapstructure:"auth"`
	Log           logging.Config      `mapstructure:"log"`
	Observability ObservabilityConfig `mapstructure:"observability"`
	Casbin        CasbinConfig        `mapstructure:"casbin"`
	Cache         CacheConfig         `mapstructure:"cache"`
	IDGen         IDGenConfig         `mapstructure:"idgen"`
}

// AppConfig 描述应用基础元信息。
type AppConfig struct {
	Name       string `mapstructure:"name"`
	Env        string `mapstructure:"env"`
	Version    string `mapstructure:"version"`
	InstanceID string `mapstructure:"instance_id"`
}

// ServerConfig 描述 HTTP Server 的监听地址、超时和关闭参数。
type ServerConfig struct {
	Addr            string        `mapstructure:"addr"`
	ReadTimeout     time.Duration `mapstructure:"read_timeout"`
	WriteTimeout    time.Duration `mapstructure:"write_timeout"`
	IdleTimeout     time.Duration `mapstructure:"idle_timeout"`
	ShutdownTimeout time.Duration `mapstructure:"shutdown_timeout"`
}

// RedisConfig 描述 Redis 连接配置。
type RedisConfig struct {
	Enabled  bool   `mapstructure:"enabled"`
	Addr     string `mapstructure:"addr"`
	Password string `mapstructure:"password"`
	DB       int    `mapstructure:"db"`
	PoolSize int    `mapstructure:"pool_size"`
}

// AuthConfig 描述认证与令牌配置。
type AuthConfig struct {
	Enabled              bool          `mapstructure:"enabled"`
	AccessTokenTTL       time.Duration `mapstructure:"access_token_ttl"`
	RefreshTokenTTL      time.Duration `mapstructure:"refresh_token_ttl"`
	AllowMultipleDevices bool          `mapstructure:"allow_multiple_devices"`
	JWT                  JWTConfig     `mapstructure:"jwt"`
}

// JWTConfig 描述 JWT 签发所需配置。
type JWTConfig struct {
	Issuer string `mapstructure:"issuer"`
	Secret string `mapstructure:"secret"`
}

// ObservabilityConfig 描述可观测性能力开关。
type ObservabilityConfig struct {
	Metrics FeatureConfig `mapstructure:"metrics"`
	Tracing FeatureConfig `mapstructure:"tracing"`
	Pprof   FeatureConfig `mapstructure:"pprof"`
	Health  FeatureConfig `mapstructure:"health"`
}

// FeatureConfig 描述单项能力是否启用。
type FeatureConfig struct {
	Enabled bool `mapstructure:"enabled"`
}

// CasbinConfig 描述权限模型与策略文件路径。
type CasbinConfig struct {
	ModelPath  string `mapstructure:"model_path"`
	PolicyPath string `mapstructure:"policy_path"`
}

// CacheConfig 描述多级缓存的默认 TTL。
type CacheConfig struct {
	LocalTTL  time.Duration `mapstructure:"local_ttl"`
	RemoteTTL time.Duration `mapstructure:"remote_ttl"`
}

// IDGenConfig 描述雪花算法节点配置。
type IDGenConfig struct {
	Node int64 `mapstructure:"node"`
}

// LoadConfig 加载并校验示例项目配置。
func LoadConfig(env string, configDir string) (*Config, error) {
	return platformconfig.LoadInto[Config](platformconfig.LoaderOptions{
		Env:       env,
		ConfigDir: configDir,
		Defaults:  configDefaults(),
	})
}

// SafeForLog 返回脱敏后的配置副本，可安全用于结构化日志打印。
func (c *Config) SafeForLog() map[string]any {
	return platformconfig.Sanitize(c, c.Log.Mask.Fields)
}

// Validate 对启动所需关键配置进行兜底校验。
func (c *Config) Validate() error {
	switch {
	case c.App.Name == "":
		return fmt.Errorf("app.name 不能为空")
	case c.App.Env == "":
		return fmt.Errorf("app.env 不能为空")
	case c.Server.Addr == "":
		return fmt.Errorf("server.addr 不能为空")
	case c.Server.ReadTimeout <= 0:
		return fmt.Errorf("server.read_timeout 必须大于 0")
	case c.Server.WriteTimeout <= 0:
		return fmt.Errorf("server.write_timeout 必须大于 0")
	case c.Server.IdleTimeout <= 0:
		return fmt.Errorf("server.idle_timeout 必须大于 0")
	case c.Server.ShutdownTimeout <= 0:
		return fmt.Errorf("server.shutdown_timeout 必须大于 0")
	case c.Database.Host == "":
		return fmt.Errorf("database.host 不能为空")
	case c.Database.Port <= 0 || c.Database.Port > 65535:
		return fmt.Errorf("database.port 必须在 1 到 65535 之间")
	case c.Database.User == "":
		return fmt.Errorf("database.user 不能为空")
	case c.Database.DBName == "":
		return fmt.Errorf("database.dbname 不能为空")
	case c.Redis.Enabled && c.Redis.Addr == "":
		return fmt.Errorf("redis.addr 不能为空")
	case c.Redis.Enabled && c.Redis.PoolSize <= 0:
		return fmt.Errorf("redis.pool_size 必须大于 0")
	case !c.Auth.Enabled:
		return fmt.Errorf("auth.enabled 当前必须为 true")
	case c.Auth.JWT.Issuer == "":
		return fmt.Errorf("auth.jwt.issuer 不能为空")
	case c.Auth.JWT.Secret == "":
		return fmt.Errorf("auth.jwt.secret 不能为空")
	case c.Auth.AccessTokenTTL <= 0:
		return fmt.Errorf("auth.access_token_ttl 必须大于 0")
	case c.Auth.RefreshTokenTTL <= 0:
		return fmt.Errorf("auth.refresh_token_ttl 必须大于 0")
	case c.Auth.RefreshTokenTTL <= c.Auth.AccessTokenTTL:
		return fmt.Errorf("auth.refresh_token_ttl 必须大于 auth.access_token_ttl")
	case c.Casbin.ModelPath == "":
		return fmt.Errorf("casbin.model_path 不能为空")
	case c.Casbin.PolicyPath == "":
		return fmt.Errorf("casbin.policy_path 不能为空")
	case c.IDGen.Node < 0 || c.IDGen.Node > 1023:
		return fmt.Errorf("idgen.node 必须在 0 到 1023 之间")
	default:
		return nil
	}
}

// configDefaults 统一声明示例项目配置默认值。
func configDefaults() map[string]any {
	return map[string]any{
		"app.version":                   "dev",
		"app.instance_id":               "local-1",
		"server.addr":                   ":8080",
		"server.read_timeout":           "10s",
		"server.write_timeout":          "30s",
		"server.idle_timeout":           "60s",
		"server.shutdown_timeout":       "20s",
		"database.host":                 "",
		"database.port":                 5432,
		"database.user":                 "",
		"database.password":             "",
		"database.dbname":               "",
		"database.max_open_conns":       20,
		"database.max_idle_conns":       10,
		"database.conn_max_lifetime":    "1h",
		"redis.enabled":                 false,
		"redis.addr":                    "127.0.0.1:6379",
		"redis.password":                "",
		"redis.db":                      0,
		"redis.pool_size":               10,
		"auth.enabled":                  true,
		"auth.access_token_ttl":         "15m",
		"auth.refresh_token_ttl":        "720h",
		"auth.allow_multiple_devices":   true,
		"auth.jwt.issuer":               "",
		"auth.jwt.secret":               "",
		"log.level":                     "info",
		"log.format":                    "json",
		"log.output":                    "stdout",
		"log.mask.enabled":              true,
		"log.mask.fields":               []string{"password", "token", "secret", "authorization"},
		"observability.metrics.enabled": true,
		"observability.tracing.enabled": false,
		"observability.pprof.enabled":   false,
		"observability.health.enabled":  true,
		"casbin.model_path":             "",
		"casbin.policy_path":            "",
		"cache.local_ttl":               "1m",
		"cache.remote_ttl":              "10m",
		"idgen.node":                    1,
	}
}
