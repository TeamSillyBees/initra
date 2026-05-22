package user

import (
	"github.com/samber/do"
	"github.com/teamsillybees/initra/examples/api/internal/data/ent"
	"github.com/teamsillybees/initra/pkg/auth"
	platformcache "github.com/teamsillybees/initra/pkg/cache"
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
		client := do.MustInvoke[*ent.Client](i)
		return NewRepository(client), nil
	})
	do.ProvideNamed(injector, userCacheServiceName, func(i *do.Injector) (*UserCache, error) {
		manager := do.MustInvoke[*platformcache.Manager](i)
		return NewUserCache(manager), nil
	})
	do.ProvideNamed(injector, userServiceServiceName, func(i *do.Injector) (*Service, error) {
		repo := do.MustInvokeNamed[*Repository](i, userRepositoryServiceName)
		cache := do.MustInvokeNamed[*UserCache](i, userCacheServiceName)
		passwords := do.MustInvoke[*auth.BcryptPasswordManager](i)
		return NewService(repo, cache, passwords), nil
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
