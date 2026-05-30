package httpclient

import (
	"github.com/samber/do"
	"github.com/teamsillybees/initra/pkg/logx"
)

const clientNamePrefix = "httpclient.client."

// ClientName 返回指定服务名在 DI 容器中的 HTTP Client 命名依赖。
func ClientName(serviceName string) string {
	return clientNamePrefix + serviceName
}

// Register 将 HTTP Client 工厂和已配置服务的命名 Client 注册到 DI 容器。
//
// 调用方需要先向容器注册 *logx.Logger。业务模块可通过
// do.MustInvokeNamed[*httpclient.Client](injector, httpclient.ClientName("service"))
// 直接依赖指定服务的 Client，避免自行感知 Factory 创建细节。
func Register(injector *do.Injector, cfg Config) {
	do.Provide(injector, func(i *do.Injector) (Factory, error) {
		logger := do.MustInvoke[*logx.Logger](i)
		return NewFactory(cfg, logger)
	})
	for serviceName := range cfg.Services {
		registerClient(injector, serviceName)
	}
}

// ProvideHTTPClient 将 HTTP Client 工厂和已配置服务的命名 Client 注册到 DI 容器。
func ProvideHTTPClient(injector *do.Injector, cfg Config) {
	Register(injector, cfg)
}

func registerClient(injector *do.Injector, serviceName string) {
	do.ProvideNamed(injector, ClientName(serviceName), func(i *do.Injector) (*Client, error) {
		factory := do.MustInvoke[Factory](i)
		return factory.Get(serviceName)
	})
}
