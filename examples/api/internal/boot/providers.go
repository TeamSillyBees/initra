package boot

import (
	"context"

	"github.com/casbin/casbin/v2"
	"github.com/samber/do"
	"github.com/teamsillybees/initra/pkg/auth"
	"github.com/teamsillybees/initra/pkg/logging"
	"github.com/teamsillybees/initra/pkg/observability"
	"github.com/teamsillybees/initra/pkg/server"
	"go.uber.org/zap"
)

func registerProviders(ctx context.Context, injector *do.Injector, cfg *Config, buildInfo observability.BuildInfo) {
	_ = ctx
	do.Provide(injector, func(i *do.Injector) (*zap.Logger, error) {
		return logging.NewLogger(cfg.Logger)
	})
	do.Provide(injector, func(i *do.Injector) (*auth.JWTManager, error) {
		return auth.NewJWTManager(auth.JWTConfig{
			Issuer:          cfg.JWT.Issuer,
			Secret:          cfg.JWT.Secret,
			AccessTokenTTL:  cfg.JWT.AccessTokenTTL,
			RefreshTokenTTL: cfg.JWT.RefreshTokenTTL,
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
	_ = injector
}

func registerRoutes(injector *do.Injector, webApp *server.App, buildInfo observability.BuildInfo) {
	_ = injector
	observability.NewModule(buildInfo).Register(webApp.API, webApp.Registry)
}
