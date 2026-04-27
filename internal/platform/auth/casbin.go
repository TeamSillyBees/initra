package auth

import (
	"github.com/casbin/casbin/v2"
	"github.com/casbin/casbin/v2/model"
	fileadapter "github.com/casbin/casbin/v2/persist/file-adapter"
)

// RouteSecurity 描述某个 HTTP 路由的认证与授权要求。
type RouteSecurity struct {
	Public   bool
	Resource string
	Action   string
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
