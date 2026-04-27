package user

import (
	"database/sql"
	"time"

	"github.com/samber/do"
	"github.com/teamsillybees/initra/internal/app/user/api"
	"github.com/teamsillybees/initra/internal/app/user/domain"
	"github.com/teamsillybees/initra/internal/app/user/infra"
	platformcache "github.com/teamsillybees/initra/internal/platform/cache"
	"github.com/teamsillybees/initra/internal/platform/idgen"
	"github.com/teamsillybees/initra/internal/shared/utils"
)

// do 服务名常量用于区分 user 模块内部的仓储、缓存、服务和 Handler。
const (
	userRepositoryServiceName = "user.repository"
	userCacheServiceName      = "user.cache"
	userServiceServiceName    = "user.service"
	userHandlerServiceName    = "user.handler"
)

// Provide 使用 do 注册 user 模块依赖。
func Provide(injector *do.Injector) {
	do.ProvideNamed(injector, userRepositoryServiceName, func(i *do.Injector) (*infra.Repository, error) {
		db := do.MustInvoke[*sql.DB](i)
		generator := do.MustInvoke[*idgen.Generator](i)
		return infra.NewRepository(db, generator), nil
	})
	do.ProvideNamed(injector, userCacheServiceName, func(i *do.Injector) (*infra.UserCache, error) {
		manager := do.MustInvoke[*platformcache.Manager](i)
		return infra.NewUserCache(manager), nil
	})
	do.ProvideNamed(injector, userServiceServiceName, func(i *do.Injector) (*domain.Service, error) {
		repo := do.MustInvokeNamed[*infra.Repository](i, userRepositoryServiceName)
		cache := do.MustInvokeNamed[*infra.UserCache](i, userCacheServiceName)
		generator := do.MustInvoke[*idgen.Generator](i)
		passwords := do.MustInvoke[*utils.BcryptPasswordManager](i)
		return domain.NewService(repo, cache, generator, passwords, time.Now), nil
	})
	do.ProvideNamed(injector, userHandlerServiceName, func(i *do.Injector) (*api.Handler, error) {
		service := do.MustInvokeNamed[*domain.Service](i, userServiceServiceName)
		return api.NewHandler(service), nil
	})
	do.Provide(injector, func(i *do.Injector) (*Module, error) {
		handler := do.MustInvokeNamed[*api.Handler](i, userHandlerServiceName)
		return NewModule(handler), nil
	})
}
