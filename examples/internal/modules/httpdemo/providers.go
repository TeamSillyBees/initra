package httpdemo

import (
	"github.com/samber/do"
	"github.com/teamsillybees/initra/pkg/httpclient"
)

const (
	httpBingoServiceName = "httpbingo"
	httpDemoServiceName  = "httpdemo.service"
	httpDemoHandlerName  = "httpdemo.handler"
)

// Provide 使用 do 注册 httpdemo 示例模块依赖。
func Provide(injector *do.Injector) {
	httpclient.ProvideConsumer(injector, httpDemoServiceName, httpBingoServiceName, NewService)
	do.ProvideNamed(injector, httpDemoHandlerName, func(i *do.Injector) (*Handler, error) {
		service := do.MustInvokeNamed[*Service](i, httpDemoServiceName)
		return NewHandler(service), nil
	})
	do.Provide(injector, func(i *do.Injector) (*Module, error) {
		handler := do.MustInvokeNamed[*Handler](i, httpDemoHandlerName)
		return NewModule(handler), nil
	})
}
