# 业务模块

## 目录

标准业务模块使用单一 flat package：

```text
internal/modules/<module>/
  <module>.handler.go
  <module>.dto.go
  <module>.service.go
  <module>.routes.go
  providers.go
  cache.go            可选
  *_test.go
```

默认不创建 controller/service/repository 子目录，也不创建 `<module>.repo.go` 或 `<module>.model.go`。领域实体放在 `cache.go`、`*.service.go` 或职责明确的同包文件中。

## 传输边界

`*.dto.go` 保存传输边界类型：

- Huma/Gin 请求和响应包装使用非导出的 `request`/`response` 后缀；
- 查询参数使用 `Query`，请求体使用 `Body`，multipart 表单使用 `Form`；
- 对外 JSON 类型使用 `VO`；
- 私有外部 HTTP 载荷可使用 `Payload`；
- 分页输出使用 `pagination.PageVO[T]` 或现有 cursor 类型。

Handler 负责请求解包、context 身份或 trace ID 提取、参数转换、service 调用和响应包装。成功响应使用 `response.OK(ctx, data)`；错误直接向上传递，由统一 mapper 处理。

## Service

Service 参数按用例选择现有 Body/Query、`idgen.ID`、必要的基础参数、流或职责明确的专用输入；返回 VO、分页 VO、模块内部结果或 error。不要仅为机械分层复制一套 service DTO。

需要数据库的 Service 直接依赖 `*ent.Client`。缓存、密码、HTTP Client、任务 Publisher 等可替换能力由模块定义最小私有接口。跨模块调用也由调用方定义小接口，不 import 对方具体实现。

## 装配与路由

模块 `providers.go` 注册 cache、service、handler、module；允许 `Handler -> *Service` 和 `Module -> *Handler` 的模块内具体依赖。service/handler 不访问 injector。

`*.routes.go` 同时注册 Huma operation 和 `RouteSecurity`：公开接口使用 `AccessModePublic`，登录态接口使用 `AccessModeAuthenticated`，后台或运营接口使用 `AccessModePermission` 并同步 Casbin policy。

## 验证

新增或修改模块时，至少提供与风险匹配的模块单元测试、`test/integration` 或 `test/e2e` 覆盖，并运行 checker、test 和 vet。

`initra module add <name>` 会生成 dto、handler、service、routes、providers 和测试骨架，不生成 repo/model。生成后仍需在 `internal/boot/providers.go` 的模块注册与路由注册阶段接入新模块，并按实际用例补齐依赖、权限、错误和测试。
