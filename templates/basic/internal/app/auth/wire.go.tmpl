package auth

import (
	"database/sql"

	"github.com/samber/do"
	platformauth "github.com/teamsillybees/initra/pkg/auth"
	"github.com/teamsillybees/initra/pkg/password"
)

const (
	authRepositoryServiceName = "auth.repository"
	authServiceServiceName    = "auth.service"
	authHandlerServiceName    = "auth.handler"
)

// Provide 使用 do 注册 auth 模块依赖。
func Provide(injector *do.Injector) {
	do.ProvideNamed(injector, authRepositoryServiceName, func(i *do.Injector) (*Repository, error) {
		db := do.MustInvoke[*sql.DB](i)
		return NewRepository(db), nil
	})
	do.ProvideNamed(injector, authServiceServiceName, func(i *do.Injector) (*Service, error) {
		repo := do.MustInvokeNamed[*Repository](i, authRepositoryServiceName)
		passwords := do.MustInvoke[*password.BcryptPasswordManager](i)
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
