package boot

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/samber/do"
	"github.com/teamsillybees/initra/pkg/observability"
	"github.com/teamsillybees/initra/pkg/server"
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
	Web       *server.App
	Server    *http.Server
}

// Bootstrap 完成配置加载、依赖注入、模块注册与 HTTP Server 组装。
func Bootstrap(ctx context.Context, options Options) (*Application, error) {
	cfg, err := LoadConfig(options.Env, options.ConfigDir)
	if err != nil {
		return nil, err
	}

	injector := do.New()
	do.ProvideValue(injector, cfg)

	registerProviders(ctx, injector, cfg, options.BuildInfo)
	registerModules(injector)

	logger := do.MustInvoke[*zap.Logger](injector)
	webApp := do.MustInvoke[*server.App](injector)

	registerRoutes(injector, webApp, options.BuildInfo)

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
	}, nil
}
