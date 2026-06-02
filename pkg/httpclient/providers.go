package httpclient

import (
	"fmt"

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
	do.Provide(injector, func(i *do.Injector) (*Factory, error) {
		logger := do.MustInvoke[*logx.Logger](i)
		return NewFactory(cfg, logger)
	})
	for serviceName := range cfg.Services {
		registerClient(injector, serviceName)
	}
}

// ProvideConsumer 注册依赖指定远程服务 Client 的业务组件。
//
// constructor 的入参可以是 *Client，也可以是 Getter、ReadCaller、Caller 等
// 由 *Client 实现的接口，用于减少业务模块中重复解析命名 Client 的胶水代码。
func ProvideConsumer[T any, D any](injector *do.Injector, providerName string, serviceName string, constructor func(D) *T) {
	do.ProvideNamed(injector, providerName, func(i *do.Injector) (*T, error) {
		client := do.MustInvokeNamed[*Client](i, ClientName(serviceName))
		dependency, ok := any(client).(D)
		if !ok {
			return nil, fmt.Errorf("%w: %s client does not satisfy consumer dependency", ErrUnsupported, serviceName)
		}
		return constructor(dependency), nil
	})
}

func registerClient(injector *do.Injector, serviceName string) {
	do.ProvideNamed(injector, ClientName(serviceName), func(i *do.Injector) (*Client, error) {
		factory := do.MustInvoke[*Factory](i)
		return factory.Get(serviceName)
	})
}
