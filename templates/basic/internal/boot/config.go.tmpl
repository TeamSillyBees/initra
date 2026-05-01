package boot

import (
	"fmt"
	"time"

	platformconfig "github.com/teamsillybees/initra/pkg/config"
	"github.com/teamsillybees/initra/pkg/database"
	"github.com/teamsillybees/initra/pkg/logging"
)

// Config 是示例项目的应用配置聚合根。
type Config struct {
	App      AppConfig       `mapstructure:"app"`
	HTTP     HTTPConfig      `mapstructure:"http"`
	Logger   logging.Config  `mapstructure:"logger"`
	Database database.Config `mapstructure:"database"`
	Redis    RedisConfig     `mapstructure:"redis"`
	JWT      JWTConfig       `mapstructure:"jwt"`
	Casbin   CasbinConfig    `mapstructure:"casbin"`
	Cache    CacheConfig     `mapstructure:"cache"`
	IDGen    IDGenConfig     `mapstructure:"idgen"`
}

// AppConfig 描述应用基础元信息。
type AppConfig struct {
	Name string `mapstructure:"name"`
	Env  string `mapstructure:"env"`
	Port int    `mapstructure:"port"`
}

// HTTPConfig 描述 HTTP Server 的读写超时等基础参数。
type HTTPConfig struct {
	ReadTimeout  time.Duration `mapstructure:"read_timeout"`
	WriteTimeout time.Duration `mapstructure:"write_timeout"`
}

// RedisConfig 描述 Redis 连接配置。
type RedisConfig struct {
	Addr     string `mapstructure:"addr"`
	Password string `mapstructure:"password"`
	DB       int    `mapstructure:"db"`
}

// JWTConfig 描述 Token 签发所需配置。
type JWTConfig struct {
	Issuer          string        `mapstructure:"issuer"`
	Secret          string        `mapstructure:"secret"`
	AccessTokenTTL  time.Duration `mapstructure:"access_token_ttl"`
	RefreshTokenTTL time.Duration `mapstructure:"refresh_token_ttl"`
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

// Validate 对启动所需关键配置进行兜底校验。
func (c *Config) Validate() error {
	switch {
	case c.App.Name == "":
		return fmt.Errorf("app.name 不能为空")
	case c.App.Env == "":
		return fmt.Errorf("app.env 不能为空")
	case c.App.Port <= 0:
		return fmt.Errorf("app.port 必须大于 0")
	case c.Database.DSN == "":
		return fmt.Errorf("database.dsn 不能为空")
	case c.JWT.Issuer == "":
		return fmt.Errorf("jwt.issuer 不能为空")
	case c.JWT.Secret == "":
		return fmt.Errorf("jwt.secret 不能为空")
	case c.JWT.AccessTokenTTL <= 0:
		return fmt.Errorf("jwt.access_token_ttl 必须大于 0")
	case c.JWT.RefreshTokenTTL <= 0:
		return fmt.Errorf("jwt.refresh_token_ttl 必须大于 0")
	case c.JWT.RefreshTokenTTL <= c.JWT.AccessTokenTTL:
		return fmt.Errorf("jwt.refresh_token_ttl 必须大于 jwt.access_token_ttl")
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
		"app.port":                8080,
		"http.read_timeout":       "5s",
		"http.write_timeout":      "10s",
		"logger.level":            "info",
		"logger.format":           "json",
		"database.driver":         "postgres",
		"database.dsn":            "",
		"database.max_open_conns": 20,
		"database.max_idle_conns": 10,
		"redis.addr":              "127.0.0.1:6379",
		"redis.password":          "",
		"redis.db":                0,
		"jwt.issuer":              "",
		"jwt.secret":              "",
		"jwt.access_token_ttl":    "15m",
		"jwt.refresh_token_ttl":   "168h",
		"casbin.model_path":       "",
		"casbin.policy_path":      "",
		"cache.local_ttl":         "1m",
		"cache.remote_ttl":        "10m",
		"idgen.node":              1,
	}
}
