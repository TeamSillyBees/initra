package httpdemo

import (
	"github.com/samber/do"
	"github.com/teamsillybees/initra/pkg/httpclient"
)

const (
	httpBingoServiceName   = "httpbingo"
	httpDemoServiceName    = "httpdemo.service"
	httpDemoHandlerName    = "httpdemo.handler"
	httpDemoHTTPClientName = "httpdemo.httpclient"
)

// Provide 使用 do 注册 httpdemo 示例模块依赖。
func Provide(injector *do.Injector) {
	do.ProvideNamed(injector, httpDemoHTTPClientName, func(i *do.Injector) (*httpclient.Client, error) {
		factory := do.MustInvoke[httpclient.Factory](i)
		return factory.Get(httpBingoServiceName)
	})
	do.ProvideNamed(injector, httpDemoServiceName, func(i *do.Injector) (*Service, error) {
		client := do.MustInvokeNamed[*httpclient.Client](i, httpDemoHTTPClientName)
		return NewService(client), nil
	})
	do.ProvideNamed(injector, httpDemoHandlerName, func(i *do.Injector) (*Handler, error) {
		service := do.MustInvokeNamed[*Service](i, httpDemoServiceName)
		return NewHandler(service), nil
	})
	do.Provide(injector, func(i *do.Injector) (*Module, error) {
		handler := do.MustInvokeNamed[*Handler](i, httpDemoHandlerName)
		return NewModule(handler), nil
	})
}
