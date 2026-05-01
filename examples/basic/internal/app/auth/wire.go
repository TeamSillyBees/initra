package auth

import (
	"database/sql"

	"github.com/samber/do"
	"github.com/teamsillybees/initra/examples/basic/internal/app/auth/api"
	"github.com/teamsillybees/initra/examples/basic/internal/app/auth/domain"
	"github.com/teamsillybees/initra/examples/basic/internal/app/auth/infra"
	platformauth "github.com/teamsillybees/initra/pkg/auth"
	"github.com/teamsillybees/initra/pkg/password"
)

// do 服务名常量用于区分同类型依赖，避免模块间命名冲突。
const (
	authRepositoryServiceName = "auth.repository"
	authServiceServiceName    = "auth.service"
	authHandlerServiceName    = "auth.handler"
)

// Provide 使用 do 注册 auth 模块依赖。
func Provide(injector *do.Injector) {
	do.ProvideNamed(injector, authRepositoryServiceName, func(i *do.Injector) (*infra.Repository, error) {
		db := do.MustInvoke[*sql.DB](i)
		return infra.NewRepository(db), nil
	})
	do.ProvideNamed(injector, authServiceServiceName, func(i *do.Injector) (*domain.Service, error) {
		repo := do.MustInvokeNamed[*infra.Repository](i, authRepositoryServiceName)
		passwords := do.MustInvoke[*password.BcryptPasswordManager](i)
		jwtManager := do.MustInvoke[*platformauth.JWTManager](i)
		return domain.NewService(repo, passwords, jwtManager), nil
	})
	do.ProvideNamed(injector, authHandlerServiceName, func(i *do.Injector) (*api.Handler, error) {
		service := do.MustInvokeNamed[*domain.Service](i, authServiceServiceName)
		return api.NewHandler(service), nil
	})
	do.Provide(injector, func(i *do.Injector) (*Module, error) {
		handler := do.MustInvokeNamed[*api.Handler](i, authHandlerServiceName)
		return NewModule(handler), nil
	})
}
