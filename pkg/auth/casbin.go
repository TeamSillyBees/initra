package auth

import (
	"github.com/casbin/casbin/v2"
	"github.com/casbin/casbin/v2/model"
	fileadapter "github.com/casbin/casbin/v2/persist/file-adapter"
)

// AccessMode 描述 HTTP 路由的访问控制模式。
type AccessMode string

const (
	// AccessModePublic 表示公开路由，不需要登录即可访问。
	AccessModePublic AccessMode = "public"
	// AccessModeAuthenticated 表示认证路由，只需要登录态即可访问。
	AccessModeAuthenticated AccessMode = "authenticated"
	// AccessModePermission 表示权限路由，需要登录态并通过 RBAC 授权。
	AccessModePermission AccessMode = "permission"
)

// RouteSecurity 描述某个 HTTP 路由的认证与授权要求。
type RouteSecurity struct {
	AccessMode AccessMode
	Resource   string
	Action     string
}

// RouteSecurityLookup 定义路由安全元数据查询能力，方便中间件与具体注册实现解耦。
type RouteSecurityLookup interface {
	Lookup(method string, path string) (RouteSecurity, bool)
}

// NewEnforcer 使用文件模型和策略创建 Casbin Enforcer。
func NewEnforcer(modelPath string, policyPath string) (*casbin.Enforcer, error) {
	m, err := model.NewModelFromFile(modelPath)
	if err != nil {
		return nil, err
	}
	adapter := fileadapter.NewAdapter(policyPath)
	return casbin.NewEnforcer(m, adapter)
}
