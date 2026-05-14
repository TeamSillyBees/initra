package logging

import (
	"github.com/samber/do"
	"go.uber.org/zap"
)

// Register 将 zap Logger 注册到 DI 容器。
func Register(injector *do.Injector, cfg Config) {
	do.Provide(injector, func(i *do.Injector) (*zap.Logger, error) {
		return NewLogger(cfg)
	})
}
