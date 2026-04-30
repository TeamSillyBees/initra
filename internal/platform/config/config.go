package config

import (
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/go-viper/mapstructure/v2"
	"github.com/spf13/viper"
)

// LoaderOptions 描述配置加载时需要的最小输入，便于在生产环境与测试环境复用同一套逻辑。
type LoaderOptions struct {
	Env       string
	ConfigDir string
	EnvPrefix string
}

// Config 是整个应用的配置聚合根，所有配置都必须通过该结构体完成映射和校验。
type Config struct {
	App      AppConfig      `mapstructure:"app"`
	HTTP     HTTPConfig     `mapstructure:"http"`
	Logger   LoggerConfig   `mapstructure:"logger"`
	Database DatabaseConfig `mapstructure:"database"`
	Redis    RedisConfig    `mapstructure:"redis"`
	JWT      JWTConfig      `mapstructure:"jwt"`
	Casbin   CasbinConfig   `mapstructure:"casbin"`
	Cache    CacheConfig    `mapstructure:"cache"`
	IDGen    IDGenConfig    `mapstructure:"idgen"`
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

// LoggerConfig 描述日志输出格式与级别。
type LoggerConfig struct {
	Level  string `mapstructure:"level"`
	Format string `mapstructure:"format"`
}

// DatabaseConfig 描述 PostgreSQL 连接配置。
type DatabaseConfig struct {
	Driver       string `mapstructure:"driver"`
	DSN          string `mapstructure:"dsn"`
	MaxOpenConns int    `mapstructure:"max_open_conns"`
	MaxIdleConns int    `mapstructure:"max_idle_conns"`
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
// bwmarrin/snowflake 默认 10 位节点号，因此单集群内必须保证 0-1023 之间唯一，避免多实例生成重复 ID。
type IDGenConfig struct {
	Node int64 `mapstructure:"node"`
}

// Load 从配置文件、环境变量与默认值中加载应用配置。
func Load(opts LoaderOptions) (*Config, error) {
	normalized := normalizeOptions(opts)

	v := viper.New()
	v.SetConfigName(fmt.Sprintf("config.%s", normalized.Env))
	v.SetConfigType("yaml")
	v.AddConfigPath(normalized.ConfigDir)
	v.SetEnvPrefix(normalized.EnvPrefix)
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	setDefaults(v)

	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("读取配置文件失败: %w", err)
	}

	var cfg Config
	if err := v.Unmarshal(&cfg, viper.DecodeHook(mapstructure.StringToTimeDurationHookFunc())); err != nil {
		return nil, fmt.Errorf("反序列化配置失败: %w", err)
	}

	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	return &cfg, nil
}

// Validate 对启动所需关键配置进行兜底校验，尽量在启动阶段暴露问题。
func (c Config) Validate() error {
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

// normalizeOptions 为配置加载入口补齐默认环境、目录和环境变量前缀。
func normalizeOptions(opts LoaderOptions) LoaderOptions {
	if opts.Env == "" {
		opts.Env = "dev"
	}
	if opts.ConfigDir == "" {
		opts.ConfigDir = filepath.Join(".", "configs")
	}
	if opts.EnvPrefix == "" {
		opts.EnvPrefix = "INITRA"
	}
	return opts
}

// setDefaults 统一声明配置默认值，避免默认值散落在启动流程中。
func setDefaults(v *viper.Viper) {
	v.SetDefault("app.port", 8080)
	v.SetDefault("http.read_timeout", "5s")
	v.SetDefault("http.write_timeout", "10s")
	v.SetDefault("logger.level", "info")
	v.SetDefault("logger.format", "json")
	v.SetDefault("database.driver", "postgres")
	v.SetDefault("database.dsn", "")
	v.SetDefault("database.max_open_conns", 20)
	v.SetDefault("database.max_idle_conns", 10)
	v.SetDefault("redis.addr", "127.0.0.1:6379")
	v.SetDefault("redis.password", "")
	v.SetDefault("redis.db", 0)
	v.SetDefault("jwt.issuer", "")
	v.SetDefault("jwt.secret", "")
	v.SetDefault("jwt.access_token_ttl", "15m")
	v.SetDefault("jwt.refresh_token_ttl", "168h")
	v.SetDefault("casbin.model_path", "")
	v.SetDefault("casbin.policy_path", "")
	v.SetDefault("cache.local_ttl", "1m")
	v.SetDefault("cache.remote_ttl", "10m")
	v.SetDefault("idgen.node", 1)
}
