package boot

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"time"

	"github.com/samber/do"
	"github.com/teamsillybees/initra/pkg/observability"
	"github.com/teamsillybees/initra/pkg/redisx"
	"github.com/teamsillybees/initra/pkg/server"
	"go.uber.org/zap"
)

// Options 描述应用启动时的外部输入。
type Options struct {
	Env       string
	ConfigDir string
	BuildInfo observability.BuildInfoVO
}

// Application 是启动完成后的应用聚合根。
type Application struct {
	Container *do.Injector
	Config    *Config
	Logger    *zap.Logger
	Web       *server.App
	Server    *http.Server
	DB        *sql.DB
	Redis     redisx.UniversalClient
}

// Bootstrap 完成配置加载、依赖注入、模块注册与 HTTP Server 组装。
func Bootstrap(ctx context.Context, options Options) (*Application, error) {
	cfg, err := LoadConfig(options.Env, options.ConfigDir)
	if err != nil {
		return nil, err
	}

	injector := do.New()
	do.ProvideValue(injector, cfg)

	registerProviders(injector, cfg, options.BuildInfo)
	registerModules(injector)

	logger := do.MustInvoke[*zap.Logger](injector)
	db, err := do.Invoke[*sql.DB](injector)
	if err != nil {
		return nil, fmt.Errorf("resolve database client: %w", err)
	}
	var redisClient redisx.UniversalClient
	if cfg.Redis.Enabled {
		redisClient, err = do.Invoke[redisx.UniversalClient](injector)
		if err != nil {
			_ = db.Close()
			return nil, fmt.Errorf("resolve redis client: %w", err)
		}
	}
	if err := checkStartupConnectivity(ctx, db, cfg.Redis.Enabled, redisClient); err != nil {
		closeStartupResources(db, redisClient)
		return nil, err
	}
	webApp := do.MustInvoke[*server.App](injector)

	registerRoutes(injector, webApp, options.BuildInfo)

	s := &http.Server{
		Addr:              cfg.Server.Addr,
		Handler:           webApp.Engine,
		ReadTimeout:       cfg.Server.ReadTimeout,
		WriteTimeout:      cfg.Server.WriteTimeout,
		IdleTimeout:       cfg.Server.IdleTimeout,
		ReadHeaderTimeout: 3 * time.Second,
	}

	return &Application{
		Container: injector,
		Config:    cfg,
		Logger:    logger,
		Web:       webApp,
		Server:    s,
		DB:        db,
		Redis:     redisClient,
	}, nil
}

func checkStartupConnectivity(ctx context.Context, db *sql.DB, redisEnabled bool, redisClient redisx.UniversalClient) error {
	if db == nil {
		return fmt.Errorf("database client 不能为空")
	}
	if err := db.PingContext(ctx); err != nil {
		return fmt.Errorf("database startup ping failed: %w", err)
	}
	if redisEnabled {
		if err := redisx.Ping(ctx, redisClient); err != nil {
			return fmt.Errorf("redis startup ping failed: %w", err)
		}
	}
	return nil
}

func closeStartupResources(db *sql.DB, redisClient redisx.UniversalClient) {
	if redisClient != nil {
		_ = redisClient.Close()
	}
	if db != nil {
		_ = db.Close()
	}
}
