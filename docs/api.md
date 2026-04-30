# API 说明

## 文档入口

- OpenAPI JSON：`GET /openapi.json`
- OpenAPI YAML：`GET /openapi.yaml`
- 在线文档：`GET /docs`

## 观测接口

- `GET /health`
- `GET /ready`
- `GET /version`

## Auth 模块

- `POST /api/v1/auth/login`
- `POST /api/v1/auth/refresh`
- `GET /api/v1/auth/me`

## User 模块

- `GET /api/v1/users`
- `POST /api/v1/users`
- `GET /api/v1/users/{id}`
- `PUT /api/v1/users/{id}`
- `DELETE /api/v1/users/{id}`

## 新增业务接口说明

新增业务接口时，先在对应模块的 `module.go` 里维护 Huma operation，这会自动进入 OpenAPI；再同步更新本文档中的人工索引。所有 `/api/` 接口还必须登记 `RouteSecurity`，否则鉴权中间件会按 fail-closed 处理并拒绝访问。

## 默认管理员

- 用户名：`admin`
- 密码：`admin123`
