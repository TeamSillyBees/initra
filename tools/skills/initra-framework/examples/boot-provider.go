//go:build ignore

package boot

import (
	"database/sql"

	"example.com/your-app/internal/data"
	"example.com/your-app/internal/data/ent"
	"github.com/samber/do"
	"github.com/teamsillybees/initra/pkg/auth"
	platformcache "github.com/teamsillybees/initra/pkg/cache"
	platformdatabase "github.com/teamsillybees/initra/pkg/database"
	"github.com/teamsillybees/initra/pkg/httpclient"
	"github.com/teamsillybees/initra/pkg/idgen"
	"github.com/teamsillybees/initra/pkg/logx"
	"github.com/teamsillybees/initra/pkg/observability"
	"github.com/teamsillybees/initra/pkg/redisx"
	"github.com/teamsillybees/initra/pkg/server"
	storageprovider "github.com/teamsillybees/initra/pkg/storage/provider"
)

func registerProviders(injector *do.Injector, cfg *Config, buildInfo observability.BuildInfoVO) {
	logx.Register(injector, cfg.Log)
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
