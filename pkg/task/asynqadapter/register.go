package asynqadapter

import (
	"github.com/samber/do"
	"github.com/teamsillybees/initra/pkg/task"
	"go.uber.org/zap"
)

// Register 将任务队列 Publisher、Registry、Worker 和 Scheduler 注册到 DI 容器。
func Register(injector *do.Injector, cfg task.Config) {
	do.Provide(injector, func(i *do.Injector) (task.Registry, error) {
		logger := do.MustInvoke[*zap.Logger](i)
		return task.NewRegistry(
			task.RecoverMiddleware(logger),
			task.TracingMiddleware(nil),
			task.LoggingMiddleware(logger),
			task.MetricsMiddleware(nil),
			task.BizKeyValidationMiddleware(),
			task.IdempotencyMiddleware(nil),
		), nil
	})
	do.Provide(injector, func(i *do.Injector) (task.Publisher, error) {
		logger := do.MustInvoke[*zap.Logger](i)
		return NewPublisher(cfg, logger)
	})
	do.Provide(injector, func(i *do.Injector) (task.Worker, error) {
		logger := do.MustInvoke[*zap.Logger](i)
		registry := do.MustInvoke[task.Registry](i)
		return NewWorker(cfg, registry, logger)
	})
	do.Provide(injector, func(i *do.Injector) (task.Scheduler, error) {
		logger := do.MustInvoke[*zap.Logger](i)
		return NewScheduler(cfg, logger)
	})
}
