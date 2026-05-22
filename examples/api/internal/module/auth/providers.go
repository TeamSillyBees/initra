package auth

import (
	"github.com/samber/do"
	"github.com/teamsillybees/initra/examples/api/internal/data/ent"
	platformauth "github.com/teamsillybees/initra/pkg/auth"
)

const (
	authRepositoryServiceName = "auth.repository"
	authServiceServiceName    = "auth.service"
	authHandlerServiceName    = "auth.handler"
)

// Provide 使用 do 注册 auth 模块依赖。
func Provide(injector *do.Injector) {
	do.ProvideNamed(injector, authRepositoryServiceName, func(i *do.Injector) (*Repository, error) {
		client := do.MustInvoke[*ent.Client](i)
		return NewRepository(client), nil
	})
	do.ProvideNamed(injector, authServiceServiceName, func(i *do.Injector) (*Service, error) {
		repo := do.MustInvokeNamed[*Repository](i, authRepositoryServiceName)
		passwords := do.MustInvoke[*platformauth.BcryptPasswordManager](i)
		jwtManager := do.MustInvoke[*platformauth.JWTManager](i)
		return NewService(repo, passwords, jwtManager), nil
	})
	do.ProvideNamed(injector, authHandlerServiceName, func(i *do.Injector) (*Handler, error) {
		service := do.MustInvokeNamed[*Service](i, authServiceServiceName)
		return NewHandler(service), nil
	})
	do.Provide(injector, func(i *do.Injector) (*Module, error) {
		handler := do.MustInvokeNamed[*Handler](i, authHandlerServiceName)
		return NewModule(handler), nil
	})
}
