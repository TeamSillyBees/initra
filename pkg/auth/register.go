package auth

import (
	"github.com/casbin/casbin/v2"
	"github.com/samber/do"
	"github.com/teamsillybees/initra/pkg/redisx"
)

// RegisterOptions 描述认证相关组件注册所需的输入。
type RegisterOptions struct {
	AppName          string
	Env              string
	PasswordCost     int
	JWT              JWTConfig
	RedisEnabled     bool
	CasbinModelPath  string
	CasbinPolicyPath string
}

// Register 将密码管理器、JWT 管理器和 Casbin Enforcer 注册到 DI 容器。
func Register(injector *do.Injector, opts RegisterOptions) {
	do.Provide(injector, func(i *do.Injector) (*BcryptPasswordManager, error) {
		return NewBcryptPasswordManager(opts.PasswordCost), nil
	})
	do.Provide(injector, func(i *do.Injector) (*JWTManager, error) {
		cfg := opts.JWT
		cfg.Store = tokenStoreFromInjector(i, opts)
		return NewJWTManager(cfg)
	})
	do.Provide(injector, func(i *do.Injector) (*casbin.Enforcer, error) {
		return NewEnforcer(opts.CasbinModelPath, opts.CasbinPolicyPath)
	})
}

func tokenStoreFromInjector(injector *do.Injector, opts RegisterOptions) TokenStore {
	if !opts.RedisEnabled {
		return NewMemoryTokenStore()
	}
	client := do.MustInvoke[redisx.UniversalClient](injector)
	return NewRedisTokenStoreWithEnv(opts.AppName, opts.Env, client)
}
