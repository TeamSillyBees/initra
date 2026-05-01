package boot

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/casbin/casbin/v2"
	"github.com/redis/go-redis/v9"
	"github.com/samber/do"
	authmodule "github.com/teamsillybees/initra/examples/basic/internal/app/auth"
	usermodule "github.com/teamsillybees/initra/examples/basic/internal/app/user"
	"github.com/teamsillybees/initra/pkg/auth"
	platformcache "github.com/teamsillybees/initra/pkg/cache"
	"github.com/teamsillybees/initra/pkg/database"
	"github.com/teamsillybees/initra/pkg/idgen"
	"github.com/teamsillybees/initra/pkg/logging"
	"github.com/teamsillybees/initra/pkg/observability"
	"github.com/teamsillybees/initra/pkg/password"
	"github.com/teamsillybees/initra/pkg/web"
	"go.uber.org/zap"
)

// Options 描述应用启动时的外部输入。
type Options struct {
	Env       string
	ConfigDir string
	BuildInfo observability.BuildInfo
}

// Application 是启动完成后的应用聚合根。
type Application struct {
	Container *do.Injector
	Config    *Config
	Logger    *zap.Logger
	Web       *web.App
	Server    *http.Server
	DB        *sql.DB
	Redis     *redis.Client
}

// Bootstrap 完成配置加载、依赖注入、模块注册与 HTTP Server 组装。
func Bootstrap(ctx context.Context, options Options) (*Application, error) {
	cfg, err := LoadConfig(options.Env, options.ConfigDir)
	if err != nil {
		return nil, err
	}

	injector := do.New()
	do.ProvideValue(injector, cfg)

	do.Provide(injector, func(i *do.Injector) (*zap.Logger, error) {
		return logging.NewLogger(cfg.Logger)
	})
	do.Provide(injector, func(i *do.Injector) (*sql.DB, error) {
		return database.Open(ctx, cfg.Database)
	})
	do.Provide(injector, func(i *do.Injector) (*redis.Client, error) {
		client := redis.NewClient(&redis.Options{
			Addr:     cfg.Redis.Addr,
			Password: cfg.Redis.Password,
			DB:       cfg.Redis.DB,
		})
		if err := client.Ping(ctx).Err(); err != nil {
			_ = client.Close()
			return nil, err
		}
		return client, nil
	})
	do.Provide(injector, func(i *do.Injector) (*platformcache.Manager, error) {
		client := do.MustInvoke[*redis.Client](i)
		return platformcache.NewManager(platformcache.Config{
			AppName:   cfg.App.Name,
			LocalTTL:  cfg.Cache.LocalTTL,
			RemoteTTL: cfg.Cache.RemoteTTL,
		}, client), nil
	})
	do.Provide(injector, func(i *do.Injector) (*idgen.Generator, error) {
		return idgen.NewGenerator(cfg.IDGen.Node)
	})
	do.Provide(injector, func(i *do.Injector) (*password.BcryptPasswordManager, error) {
		return password.NewBcryptPasswordManager(0), nil
	})
	do.Provide(injector, func(i *do.Injector) (*auth.JWTManager, error) {
		client := do.MustInvoke[*redis.Client](i)
		return auth.NewJWTManager(auth.JWTConfig{
			Issuer:          cfg.JWT.Issuer,
			Secret:          cfg.JWT.Secret,
			AccessTokenTTL:  cfg.JWT.AccessTokenTTL,
			RefreshTokenTTL: cfg.JWT.RefreshTokenTTL,
			Store:           auth.NewRedisTokenStore(cfg.App.Name, client),
		})
	})
	do.Provide(injector, func(i *do.Injector) (*casbin.Enforcer, error) {
		return auth.NewEnforcer(cfg.Casbin.ModelPath, cfg.Casbin.PolicyPath)
	})
	do.Provide(injector, func(i *do.Injector) (*web.App, error) {
		logger := do.MustInvoke[*zap.Logger](i)
		jwtManager := do.MustInvoke[*auth.JWTManager](i)
		enforcer := do.MustInvoke[*casbin.Enforcer](i)
		return web.NewApp(web.Options{
			Title:   cfg.App.Name,
			Version: options.BuildInfo.Version,
			Env:     cfg.App.Env,
		}, logger, jwtManager, enforcer)
	})

	// 新增业务模块时，先在 internal/app/<module>/wire.go 暴露 Provide，
	// 再在这里把模块依赖注册进同一个 do 容器。
	usermodule.Provide(injector)
	authmodule.Provide(injector)

	logger := do.MustInvoke[*zap.Logger](injector)
	db := do.MustInvoke[*sql.DB](injector)
	redisClient := do.MustInvoke[*redis.Client](injector)
	webApp := do.MustInvoke[*web.App](injector)

	observability.NewModule(options.BuildInfo).Register(webApp.API, webApp.Registry)
	// 新增业务模块完成依赖注册后，需要在这里解析 Module 并调用 Register。
	// Register 内部应同时注册 Huma operation 与 RouteSecurity，避免 /api 路由因缺少安全元信息被拒绝。
	do.MustInvoke[*usermodule.Module](injector).Register(webApp.API, webApp.Registry)
	do.MustInvoke[*authmodule.Module](injector).Register(webApp.API, webApp.Registry)

	server := &http.Server{
		Addr:              fmt.Sprintf(":%d", cfg.App.Port),
		Handler:           webApp.Engine,
		ReadTimeout:       cfg.HTTP.ReadTimeout,
		WriteTimeout:      cfg.HTTP.WriteTimeout,
		ReadHeaderTimeout: 3 * time.Second,
	}

	return &Application{
		Container: injector,
		Config:    cfg,
		Logger:    logger,
		Web:       webApp,
		Server:    server,
		DB:        db,
		Redis:     redisClient,
	}, nil
}

// Run 启动 HTTP 服务并在收到上游取消信号后优雅关闭。
func (a *Application) Run(ctx context.Context) error {
	errCh := make(chan error, 1)

	go func() {
		if err := a.Server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
		}
	}()

	select {
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		return a.Shutdown(shutdownCtx)
	case err := <-errCh:
		return err
	}
}

// Shutdown 优雅关闭 HTTP Server 与底层资源。
func (a *Application) Shutdown(ctx context.Context) error {
	var firstErr error

	if err := a.Server.Shutdown(ctx); err != nil && firstErr == nil {
		firstErr = err
	}
	if err := a.DB.Close(); err != nil && firstErr == nil {
		firstErr = err
	}
	if err := a.Redis.Close(); err != nil && firstErr == nil {
		firstErr = err
	}
	if err := a.Logger.Sync(); err != nil && firstErr == nil {
		firstErr = err
	}

	return firstErr
}
