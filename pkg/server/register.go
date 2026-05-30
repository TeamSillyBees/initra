package server

import (
	"github.com/casbin/casbin/v2"
	"github.com/samber/do"
	platformauth "github.com/teamsillybees/initra/pkg/auth"
	"github.com/teamsillybees/initra/pkg/logx"
)

// Register 将 Web 应用注册到 DI 容器。
func Register(injector *do.Injector, opts Options) {
	do.Provide(injector, func(i *do.Injector) (*App, error) {
		logger := do.MustInvoke[*logx.Logger](i)
		jwtManager := do.MustInvoke[*platformauth.JWTManager](i)
		enforcer := do.MustInvoke[*casbin.Enforcer](i)
		return NewApp(opts, logger, jwtManager, enforcer)
	})
}
