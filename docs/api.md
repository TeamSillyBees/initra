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

## 默认管理员

- 用户名：`admin`
- 密码：`admin123`
