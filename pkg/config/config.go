package config

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/go-viper/mapstructure/v2"
	"github.com/spf13/viper"
)

// LoaderOptions 描述配置加载入口所需的最小输入。
type LoaderOptions struct {
	Env        string
	ConfigDir  string
	ConfigName string
	EnvPrefix  string
	Defaults   map[string]any
}

// Validator 允许业务配置在反序列化后执行自己的校验逻辑。
type Validator interface {
	Validate() error
}

// LoadInto 从 YAML 配置文件、默认值和环境变量中加载业务自定义配置结构。
func LoadInto[T any](opts LoaderOptions) (*T, error) {
	normalized := normalizeOptions(opts)

	v := viper.New()
	v.SetConfigName(normalized.ConfigName)
	v.SetConfigType("yaml")
	v.AddConfigPath(normalized.ConfigDir)
	v.SetEnvPrefix(normalized.EnvPrefix)
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	for key, value := range normalized.Defaults {
		v.SetDefault(key, value)
	}

	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("读取配置文件失败: %w", err)
	}

	var cfg T
	if err := v.Unmarshal(&cfg, viper.DecodeHook(mapstructure.StringToTimeDurationHookFunc())); err != nil {
		return nil, fmt.Errorf("反序列化配置失败: %w", err)
	}

	if validator, ok := any(&cfg).(Validator); ok {
		if err := validator.Validate(); err != nil {
			return nil, err
		}
	}

	return &cfg, nil
}

// normalizeOptions 为配置加载入口补齐默认环境、目录和环境变量前缀。
func normalizeOptions(opts LoaderOptions) LoaderOptions {
	if opts.Env == "" {
		opts.Env = "dev"
	}
	if opts.ConfigDir == "" {
		opts.ConfigDir = filepath.Join(".", "configs")
	}
	if opts.ConfigName == "" {
		opts.ConfigName = fmt.Sprintf("config.%s", opts.Env)
	}
	if opts.EnvPrefix == "" {
		opts.EnvPrefix = "INITRA"
	}
	if opts.Defaults == nil {
		opts.Defaults = map[string]any{}
	}
	return opts
}
