package user

import (
	"database/sql"
	"time"

	"github.com/samber/do"
	"github.com/teamsillybees/initra/pkg/auth"
	platformcache "github.com/teamsillybees/initra/pkg/cache"
	"github.com/teamsillybees/initra/pkg/idgen"
)

const (
	userRepositoryServiceName = "user.repository"
	userCacheServiceName      = "user.cache"
	userServiceServiceName    = "user.service"
	userHandlerServiceName    = "user.handler"
)

// Provide 使用 do 注册 user 模块依赖。
func Provide(injector *do.Injector) {
	do.ProvideNamed(injector, userRepositoryServiceName, func(i *do.Injector) (*Repository, error) {
		db := do.MustInvoke[*sql.DB](i)
		generator := do.MustInvoke[*idgen.Generator](i)
		return NewRepository(db, generator), nil
	})
	do.ProvideNamed(injector, userCacheServiceName, func(i *do.Injector) (*UserCache, error) {
		manager := do.MustInvoke[*platformcache.Manager](i)
		return NewUserCache(manager), nil
	})
	do.ProvideNamed(injector, userServiceServiceName, func(i *do.Injector) (*Service, error) {
		repo := do.MustInvokeNamed[*Repository](i, userRepositoryServiceName)
		cache := do.MustInvokeNamed[*UserCache](i, userCacheServiceName)
		generator := do.MustInvoke[*idgen.Generator](i)
		passwords := do.MustInvoke[*auth.BcryptPasswordManager](i)
		return NewService(repo, cache, generator, passwords, time.Now), nil
	})
	do.ProvideNamed(injector, userHandlerServiceName, func(i *do.Injector) (*Handler, error) {
		service := do.MustInvokeNamed[*Service](i, userServiceServiceName)
		return NewHandler(service), nil
	})
	do.Provide(injector, func(i *do.Injector) (*Module, error) {
		handler := do.MustInvokeNamed[*Handler](i, userHandlerServiceName)
		return NewModule(handler), nil
	})
}
