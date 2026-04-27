#!/usr/bin/env sh
set -eu

NAME="${1:-}"

if [ -z "$NAME" ]; then
  echo "usage: $0 <module_name>" >&2
  exit 1
fi

MODULE_DIR="./internal/app/${NAME}"

mkdir -p "${MODULE_DIR}/api" "${MODULE_DIR}/domain" "${MODULE_DIR}/infra"

cat > "${MODULE_DIR}/module.go" <<EOF
package ${NAME}

// TODO: 按照脚手架约定补充模块注册逻辑。
EOF

cat > "${MODULE_DIR}/wire.go" <<EOF
package ${NAME}

// TODO: 按照 do 注入规则补充模块依赖注册。
EOF

cat > "${MODULE_DIR}/api/handler.go" <<EOF
package api

// TODO: 补充 ${NAME} 模块 API Handler。
EOF

cat > "${MODULE_DIR}/api/request.go" <<EOF
package api

// TODO: 补充 ${NAME} 模块请求 DTO。
EOF

cat > "${MODULE_DIR}/api/response.go" <<EOF
package api

// TODO: 补充 ${NAME} 模块响应 DTO。
EOF

cat > "${MODULE_DIR}/domain/entity.go" <<EOF
package domain

// TODO: 补充 ${NAME} 模块领域实体。
EOF

cat > "${MODULE_DIR}/domain/service.go" <<EOF
package domain

// TODO: 补充 ${NAME} 模块应用服务。
EOF

cat > "${MODULE_DIR}/domain/errors.go" <<EOF
package domain

// TODO: 补充 ${NAME} 模块业务错误。
EOF

cat > "${MODULE_DIR}/infra/repository.go" <<EOF
package infra

// TODO: 补充 ${NAME} 模块仓储实现。
EOF

cat > "${MODULE_DIR}/infra/cache.go" <<EOF
package infra

// TODO: 补充 ${NAME} 模块缓存实现。
EOF

cat > "${MODULE_DIR}/infra/policy.go" <<EOF
package infra

// TODO: 补充 ${NAME} 模块鉴权资源常量。
EOF

echo "created module skeleton: ${MODULE_DIR}"
