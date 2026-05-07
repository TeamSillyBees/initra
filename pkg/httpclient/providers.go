package httpclient

import (
	"github.com/samber/do"
	"go.uber.org/zap"
)

// Register 将 HTTP Client 工厂注册到 DI 容器。
func Register(injector *do.Injector, cfg Config, logger *zap.Logger) {
	do.Provide(injector, func(i *do.Injector) (Factory, error) {
		return NewFactory(cfg, logger)
	})
}

// ProvideHTTPClient 将 HTTP Client 工厂注册到 DI 容器，供模板或业务启动代码直接调用。
func ProvideHTTPClient(injector *do.Injector, cfg Config, logger *zap.Logger) {
	Register(injector, cfg, logger)
}
