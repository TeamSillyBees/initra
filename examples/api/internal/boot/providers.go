package boot

import (
	"context"

	"github.com/casbin/casbin/v2"
	"github.com/redis/go-redis/v9"
	"github.com/samber/do"
	"github.com/teamsillybees/initra/examples/api/internal/data"
	"github.com/teamsillybees/initra/examples/api/internal/ent"
	authmodule "github.com/teamsillybees/initra/examples/api/internal/module/auth"
	usermodule "github.com/teamsillybees/initra/examples/api/internal/module/user"
	"github.com/teamsillybees/initra/pkg/auth"
	platformcache "github.com/teamsillybees/initra/pkg/cache"
	"github.com/teamsillybees/initra/pkg/idgen"
	"github.com/teamsillybees/initra/pkg/logging"
	"github.com/teamsillybees/initra/pkg/observability"
	"github.com/teamsillybees/initra/pkg/server"
	"go.uber.org/zap"
)

func registerProviders(ctx context.Context, injector *do.Injector, cfg *Config, buildInfo observability.BuildInfoVO) {
	do.Provide(injector, func(i *do.Injector) (*zap.Logger, error) {
		return logging.NewLogger(cfg.Log)
	})
	do.Provide(injector, func(i *do.Injector) (*redis.Client, error) {
		if !cfg.Redis.Enabled {
			return nil, nil
		}
		client := redis.NewClient(&redis.Options{
			Addr:     cfg.Redis.Addr,
			Password: cfg.Redis.Password,
			DB:       cfg.Redis.DB,
			PoolSize: cfg.Redis.PoolSize,
		})
		if err := client.Ping(ctx).Err(); err != nil {
			_ = client.Close()
			return nil, err
		}
		return client, nil
	})
	do.Provide(injector, func(i *do.Injector) (*platformcache.Manager, error) {
		var client *redis.Client
		if cfg.Redis.Enabled {
			client = do.MustInvoke[*redis.Client](i)
		}
		return platformcache.NewManager(platformcache.Config{
			AppName:   cfg.App.Name,
			LocalTTL:  cfg.Cache.LocalTTL,
			RemoteTTL: cfg.Cache.RemoteTTL,
		}, client), nil
	})
	do.Provide(injector, func(i *do.Injector) (*idgen.Generator, error) {
		return idgen.NewGenerator(cfg.IDGen.Node)
	})
	do.Provide(injector, func(i *do.Injector) (*ent.Client, error) {
		generator := do.MustInvoke[*idgen.Generator](i)
		return data.NewEntClient(cfg.Database, generator)
	})
	do.Provide(injector, func(i *do.Injector) (*auth.BcryptPasswordManager, error) {
		return auth.NewBcryptPasswordManager(0), nil
	})
	do.Provide(injector, func(i *do.Injector) (*auth.JWTManager, error) {
		var store auth.TokenStore = auth.NewMemoryTokenStore()
		if cfg.Redis.Enabled {
			client := do.MustInvoke[*redis.Client](i)
			store = auth.NewRedisTokenStore(cfg.App.Name, client)
		}
		return auth.NewJWTManager(auth.JWTConfig{
			Issuer:          cfg.Auth.JWT.Issuer,
			Secret:          cfg.Auth.JWT.Secret,
			AccessTokenTTL:  cfg.Auth.AccessTokenTTL,
			RefreshTokenTTL: cfg.Auth.RefreshTokenTTL,
			Store:           store,
		})
	})
	do.Provide(injector, func(i *do.Injector) (*casbin.Enforcer, error) {
		return auth.NewEnforcer(cfg.Casbin.ModelPath, cfg.Casbin.PolicyPath)
	})
	do.Provide(injector, func(i *do.Injector) (*server.App, error) {
		logger := do.MustInvoke[*zap.Logger](i)
		jwtManager := do.MustInvoke[*auth.JWTManager](i)
		enforcer := do.MustInvoke[*casbin.Enforcer](i)
		return server.NewApp(server.Options{
			Title:   cfg.App.Name,
			Version: buildInfo.Version,
			Env:     cfg.App.Env,
		}, logger, jwtManager, enforcer)
	})
}

func registerModules(injector *do.Injector) {
	// 新增业务模块时，先在 internal/module/<module>/providers.go 暴露 Provide，
	// 再在这里把模块依赖注册进同一个 do 容器。
	usermodule.Provide(injector)
	authmodule.Provide(injector)
}

func registerRoutes(injector *do.Injector, webApp *server.App, buildInfo observability.BuildInfoVO) {
	cfg := do.MustInvoke[*Config](injector)
	if cfg.Observability.Health.Enabled {
		observability.NewModule(buildInfo).Register(webApp.API, webApp.Registry)
	}

	// 新增业务模块完成依赖注册后，需要在这里解析 Module 并调用 Register。
	// Register 内部应同时注册 Huma operation 与 RouteSecurity，避免 /api 路由因缺少安全元信息被拒绝。
	do.MustInvoke[*usermodule.Module](injector).Register(webApp.API, webApp.Registry)
	do.MustInvoke[*authmodule.Module](injector).Register(webApp.API, webApp.Registry)
}
