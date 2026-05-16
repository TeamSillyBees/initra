package boot

import (
	"database/sql"

	"github.com/samber/do"
	"github.com/teamsillybees/initra/examples/api/internal/data"
	"github.com/teamsillybees/initra/examples/api/internal/ent"
	authmodule "github.com/teamsillybees/initra/examples/api/internal/module/auth"
	filemodule "github.com/teamsillybees/initra/examples/api/internal/module/file"
	httpdemomodule "github.com/teamsillybees/initra/examples/api/internal/module/httpdemo"
	usermodule "github.com/teamsillybees/initra/examples/api/internal/module/user"
	"github.com/teamsillybees/initra/pkg/auth"
	platformcache "github.com/teamsillybees/initra/pkg/cache"
	platformdatabase "github.com/teamsillybees/initra/pkg/database"
	"github.com/teamsillybees/initra/pkg/httpclient"
	"github.com/teamsillybees/initra/pkg/idgen"
	"github.com/teamsillybees/initra/pkg/logging"
	"github.com/teamsillybees/initra/pkg/observability"
	"github.com/teamsillybees/initra/pkg/redisx"
	"github.com/teamsillybees/initra/pkg/server"
	storageprovider "github.com/teamsillybees/initra/pkg/storage/provider"
)

func registerProviders(injector *do.Injector, cfg *Config, buildInfo observability.BuildInfoVO) {
	logging.Register(injector, cfg.Log)
	httpclient.Register(injector, cfg.HTTPClient)
	redisx.Register(injector, cfg.Redis)
	platformcache.Register(injector, platformcache.Config{
		AppName:       cfg.App.Name,
		LocalTTL:      cfg.Cache.LocalTTL,
		RemoteTTL:     cfg.Cache.RemoteTTL,
		RemoteEnabled: cfg.Redis.Enabled,
	})
	idgen.Register(injector, cfg.IDGen.Node)
	storageprovider.Register(injector, cfg.Storage)
	platformdatabase.Register(injector, data.SQLDBConfig(cfg.Database))
	do.Provide(injector, func(i *do.Injector) (*ent.Client, error) {
		generator := do.MustInvoke[*idgen.Generator](i)
		db := do.MustInvoke[*sql.DB](i)
		return data.NewEntClientFromDB(db, generator), nil
	})
	auth.Register(injector, auth.RegisterOptions{
		AppName:      cfg.App.Name,
		Env:          cfg.App.Env,
		RedisEnabled: cfg.Redis.Enabled,
		JWT: auth.JWTConfig{
			Issuer:          cfg.Auth.JWT.Issuer,
			Secret:          cfg.Auth.JWT.Secret,
			AccessTokenTTL:  cfg.Auth.AccessTokenTTL,
			RefreshTokenTTL: cfg.Auth.RefreshTokenTTL,
		},
		CasbinModelPath:  cfg.Casbin.ModelPath,
		CasbinPolicyPath: cfg.Casbin.PolicyPath,
	})
	server.Register(injector, server.Options{
		Title:   cfg.App.Name,
		Version: buildInfo.Version,
		Env:     cfg.App.Env,
	})
}

func registerModules(injector *do.Injector) {
	// 新增业务模块时，先在 internal/module/<module>/providers.go 暴露 Provide，
	// 再在这里把模块依赖注册进同一个 do 容器。
	usermodule.Provide(injector)
	authmodule.Provide(injector)
	filemodule.Provide(injector)
	httpdemomodule.Provide(injector)
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
	do.MustInvoke[*filemodule.Module](injector).Register(webApp.API, webApp.Registry)
	do.MustInvoke[*httpdemomodule.Module](injector).Register(webApp.API, webApp.Registry)
}
