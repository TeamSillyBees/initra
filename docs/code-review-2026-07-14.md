# initra 代码仓库走查问题与修复状态

> 走查日期：2026-07-14
> 走查基线：`main` / `v0.1.18`
> 说明：P0 为需要立即处置的问题，P1 为阻塞企业生产使用或核心功能的问题，P2 为应近期修复的可靠性、安全性和可维护性问题。模板中的数据库凭据已确认是无效测试数据，不列为问题。

## P0

- 当前没有需要记录的 P0 问题。

## P1

> 修复日期：2026-07-15。当前没有未修复的 P1 问题。
> 项目决策：migrate diff 默认复用所选环境的业务数据库配置，以便生成基于业务库的最终 diff；该行为不再视为问题。

- **[已修复] 本地 multipart 路径穿越。** bucket、upload ID 或临时目录中的路径片段可使分片写入、合并或递归删除越过存储根目录。**修复方式：** 对路径片段和最终 containment 做双重校验，并为 UploadPart、Complete、Abort 增加根目录外文件不受影响的恶意参数回归测试。
- **[已修复] 发布版 `initra new` 依赖与半成品问题。** 直接在目标目录生成且未先完整准备依赖，失败时会留下不可用的半成品项目。**修复方式：** 在同级临时目录依次执行依赖下载、Ent 生成、全项目测试和 Git 初始化，强制 `GOWORK=off`，全部成功后再原子落盘，并由 tag CI 验证无 replace 路径。
- **[已修复] `module add` 编译错误。** 生成的 Handler 使用了过期的响应函数签名，且产物没有可靠的格式化与编译验证。**修复方式：** 改用 `response.OK(ctx, item)`，写入前执行 `go/format`，并在回归测试中运行 gofmt、checker、go test 和 go vet。
- **[已修复] 模块生成器与标准模板冲突。** 生成器仍创建 repo/model 文件，与当前 flat-package、Service 直接使用 Ent Client 的标准结构不一致。**修复方式：** 将生成结果收敛为 dto、handler、service、routes、providers 和测试文件，不再生成 repo/model。
- **[已修复] Worker/Scheduler 配置无运行时效果。** 配置和 provider 虽然存在，但 Application 没有真正解析、启动或关闭 Worker/Scheduler。**修复方式：** 按 enabled 惰性解析任务组件，将启动、失败回滚和并发关闭接入 Application 生命周期，并校验应用关闭预算能够覆盖 Worker 优雅退出时间。
- **[已修复] `/ready` 永远成功。** 就绪接口没有检查数据库、Redis 或任务后端，依赖不可用时仍会接受流量。**修复方式：** 新增带独立超时的并发 ReadinessRegistry，任一必要依赖失败或超时即返回 503，同时保持 `/health` 只表示进程存活。
- **[已修复] Snowflake 固定默认节点。** 所有实例默认使用节点 1，多副本部署时可能产生重复 ID。**修复方式：** 移除 `pkg/idgen` 的固定默认生成器，将模板默认值设为未配置状态，并在启动时强制校验每个实例显式设置唯一的 0–1023 节点号。
- **[已修复] token store 隐式退化为内存。** Redis 未启用或解析失败时会静默使用进程内状态，导致多副本认证状态不一致。**修复方式：** 默认使用 Redis 并对解析失败执行 fail-fast，内存 store 仅允许 dev/local/test 环境显式 opt-in，其他环境一律拒绝。
- **[已修复] 数据库 TLS 与连接串构造。** 手工拼接 DSN 无法安全处理特殊字符，并且固定关闭 TLS。**修复方式：** 使用结构化 PostgreSQL URL 编码凭据和数据库名，引入 `ssl_mode` 校验，仅允许 dev/local/test 使用弱 TLS，其他环境必须使用安全 TLS。
- **[已修复] 红色测试基线与 CI 缺失。** AuditHook 和模板断言已经落后于当前实现，仓库也没有持续集成阻止回归。**修复方式：** 更新失效断言并新增 GitHub Actions，覆盖 root/examples test+vet、agent-ready checker、CLI build、生成项目 test+vet+build 和 tag 发布路径。

## P2

- **当前 Go 和依赖版本存在已知可达漏洞，本轮暂缓 Go 版本升级。** Go 1.26.3 命中标准库 [GO-2026-5856](https://pkg.go.dev/vuln/GO-2026-5856)、[GO-2026-5039](https://pkg.go.dev/vuln/GO-2026-5039)、[GO-2026-5037](https://pkg.go.dev/vuln/GO-2026-5037)，`quic-go v0.59.0` 命中 [GO-2026-5676](https://pkg.go.dev/vuln/GO-2026-5676)。**建议：** 当前只记录为已接受的延期风险，不修改 Go toolchain；保留 govulncheck 结果，后续单独安排版本升级和回归验证。

- **`crud add` 名称与实际能力不符。** `cmd/initra/main.go:1298-1311` 只生成表名常量和空 struct，没有 handler、service、Ent 操作、路由、DI 或测试，不能称为 CRUD 样例。**建议：** 要么生成可编译、已接线的最小 CRUD 纵向切片，要么明确将命令降级命名为 snippet/scaffold。

- **`config add` 生成的配置不会被应用加载。** 命令创建独立 YAML 和 Config struct，但 `pkg/config/config.go:53-61` 只读取 `config.yaml` 与 `config.<env>.yaml`，生成内容也未接入聚合 Config、默认值、Validate 或 DI。**建议：** 让命令修改聚合配置并复用对应 `pkg` Config 类型，生成后增加加载和装配测试。

- **`--app-name` 只修改 README 标题。** 运行时 `app.name`、JWT issuer、容器名、数据库名和缓存 namespace 仍固定为 `initra`，不同生成项目容易互相冲突。**建议：** 区分 display name 与安全 slug，并统一用于配置、issuer、容器和基础设施命名。

- **Windows 含空格的 `--replace` 路径会生成非法 `go.mod`。** `cmd/initra/main.go:931-940` 直接拼接 replace 行，没有按 go.mod 语法处理路径。**建议：** 使用 `golang.org/x/mod/modfile` 修改 go.mod，并验证目标路径存在且属于预期模块。

- **模块和配置生成仍不是事务性的。** `initra new` 已改为临时目录验证后原子落盘，但 module/config 命令写入多个文件时仍可能在中途失败并留下部分产物。**建议：** 复用项目生成的 staged-write 思路，先预检全部目标并在临时目录完成写入，成功后再统一提交。

- **`doctor` 无法作为人或 agent 的环境 verifier。** 即使 golangci-lint、Atlas 配置缺失或 Ent 子命令不可用，`cmd/initra/main.go:903-915` 仍返回成功。**建议：** 区分 OK、BROKEN、MISSING、OPTIONAL，缺少必需项时返回非零，并提供稳定的 JSON 输出。

- **`initra skill` 无法重复执行或升级。** 第二次执行会因为已有 `SKILL.md` 而失败，长期项目无法安全接收新版规则。**建议：** 增加版本清单和 `--check`、`--force`、安全更新流程，使用临时目录原子替换。

- **内置 checker 对根仓库 `internal` import 存在漏报。** `check_initra_usage.go:213-217,268-275` 会放过业务 `internal/boot` 对根仓库 `internal/*` 的 import，路径中只要包含 `/pkg/` 也可能被误判。**建议：** 使用 Go AST 或 `go list` 分析真实 import，不使用逐行字符串和路径 substring 判定。

- **`allow_multiple_devices` 是无效配置。** 配置虽然存在，但全仓没有认证逻辑读取它，设置为 `false` 不会限制多设备会话。**建议：** 实现基于用户和设备的 session 管理，或者删除该配置以避免制造错误安全预期。

- **refresh token 在后续操作成功前就被消费。** `examples/internal/modules/auth/auth.service.go:93-114` 先原子消费旧 token，再查询数据库并签发新 token；数据库或 token store 瞬时失败会使用户无法继续刷新。**建议：** 改为 validate、查询用户、原子 replace/rotate，确保新状态持久化成功后才让旧 token 失效。

- **JWT 配置和身份约束不足。** `pkg/auth/jwt.go` 只要求 secret 非空，没有最低 HS256 密钥长度，并允许签发或解析 `UserID=0` 的 token。**建议：** 要求至少 32 字节随机 secret，在签发、解析和中间件三层强制用户 ID 为正数。

- **RedisTokenStore 的 nil client 路径会 fail-open。** `pkg/auth/token_store.go:62-65,134-148` 在 client 为空时会让保存 refresh 和加入黑名单假装成功，黑名单检查则返回 false。**建议：** 构造阶段直接返回错误，所有操作统一返回明确的 token store failure。

- **默认管理员、JWT secret 和本地网络配置容易被误用于共享或生产环境。** 模板 seed 创建固定的 `admin/admin123` 超级管理员，dev/local/test 使用已知 `change-me-*` secret；API、PostgreSQL 和 Redis 默认监听或映射全部接口。**建议：** 生产环境拒绝所有已知默认值；管理员凭据通过一次性初始化生成，本地端口默认只绑定 loopback。

- **配置加载不拒绝未知字段。** `pkg/config/config.go:32-74` 使用非严格 Viper Unmarshal，配置拼写错误会被静默忽略并回退默认值。**建议：** 增加 unknown-key 检查，至少在生产环境遇到未知配置时启动失败。

- **配置和结构化日志脱敏覆盖不完整。** 通用敏感字段没有覆盖含密码的 DSN；`pkg/logx/redact.go`、`pkg/errors/redact.go` 对 struct、指针和 typed map 会原样返回，`logx.Any` 可能泄露密码或 token。**建议：** 补充 DSN 等字段，并实现带深度和循环保护的结构化递归脱敏测试。

- **HTTP Client 默认把上游错误正文写入日志。** `pkg/httpclient/client.go:265-275,354-397` 最多记录 512 字节响应正文，而当前文本脱敏无法可靠识别 JSON token、手机号或其他 PII。**建议：** 默认只记录服务、状态码和稳定错误码；正文预览改为显式 opt-in，并按 Content-Type 结构化脱敏。

- **Retry 配置不能关闭任意 5xx 重试。** `pkg/httpclient/retry.go:33-50` 在配置状态码未命中时仍无条件重试所有 500-599，调用方无法精确控制策略。**建议：** 严格服从状态码集合，或增加显式 `RetryAll5xx` 开关。

- **数据库启动 Ping 没有 deadline。** `pkg/database/database.go:60-75` 在注册时使用 `context.Background()`，网络异常可能让启动长期受驱动或系统超时控制。**建议：** 增加 PingTimeout 配置并使用 `context.WithTimeout`。

- **数据库外键被明确关闭但缺少架构决策说明。** Atlas 迁移和测试刻意禁止 FK，user-role、role-menu 等关系可以产生孤儿数据；对于企业 auth/admin 模板，这是重要的数据一致性边界。**建议：** 更安全的默认是启用 FK；若坚持关闭，应增加 ADR，说明删除、并发、一致性和数据修复策略。

- **乐观锁能力只有字段，没有锁行为。** `pkg/entx/fieldx/optimistic_lock.go` 只定义 version 字段，没有自动 where version、递增和冲突检测，使用者容易误以为能力已经完整。**建议：** 实现完整乐观锁 hook/interceptor，或明确改名为“乐观锁字段助手”。

- **AuditHook 对缺失字段的判断依赖错误字符串且与真实 Ent 错误不匹配。** `pkg/entx/field.go:29-39` 要求错误文本包含 `fieldx`，真实 generated mutation 错误并不包含该内容。**建议：** 使用类型能力接口代替字符串匹配，并用真实 Ent mutation 增加回归测试。

- **普通本地上传存在 overwrite 检查竞态。** `pkg/storage/local/local.go:70-88` 先检查文件不存在再调用 `os.Create`，并发请求可在检查与创建之间覆盖文件，即使 `Overwrite=false`。**建议：** 使用 `O_CREATE|O_EXCL` 原子创建文件。

- **本地 presign 语义具有误导性。** `pkg/storage/local/local.go:570-589` 返回永久 URL 和提示性的 `ExpiresAt`，没有签名或到期校验，业务可能把它误当作真正的临时授权 URL。**建议：** 明确声明 local provider 不支持 presign，或实现带签名和到期校验的下载端点。

- **请求上下文无条件信任代理头。** `pkg/requestctx/request.go:96-145` 直接使用 `Forwarded/X-Forwarded-*` 构造 BaseURL，直接客户端可控制 host/proto；未来用于密码重置或回调链接时可能形成 Host Header Injection。**建议：** 只在受信代理链中采纳这些 header，并严格校验 host 和协议。

- **Redis SCAN 前缀允许 glob 元字符。** Key Builder 不禁止 `*`、`?`、`[`、反斜线，而 `UnlinkByPrefix` 直接使用 `Prefix + "*"`，错误配置可能匹配并删除超出字面前缀的 key。**建议：** 拒绝或正确转义 Redis glob 元字符。

- **可观测性配置中存在没有实际接线的开关。** metrics、tracing、pprof 以及部分 task observability 选项被声明，但没有完整注册、导出器或生命周期实现，运维可能误以为能力已经启用。**建议：** 删除装饰性配置，或补齐真实实现和集成测试，确保每个配置项都能被验证。

- **HTTP 上传大小存在两套真相。** storage 配置允许 10 MiB，但 file 路由没有设置 Huma `MaxBodyBytes`，HTTP 层可能在远小于存储限制时提前拒绝请求。**建议：** 从统一 storage limit 派生路由 body limit，并增加边界测试。

- **Docker 模板不适合作为企业部署基线。** 缺少 `.dockerignore`，`COPY . .` 会扩大构建上下文；最终镜像包含全部配置、Atlas DSN 和 seed，未明确非 root 用户，镜像未固定 digest，Compose 还直接暴露数据库和 Redis。**建议：** 最终镜像只复制二进制和必要静态文件，配置与 Secret 在运行时注入，并使用 nonroot、可复现镜像及 loopback 开发端口。

- **Logger `Sync` 同时关闭共享 sink。** `pkg/logx/logger.go:98-136` 的 With/Named logger 共享 closers，任意子 logger 调用 Sync 都可能关闭父 logger 的文件 sink，也不符合通常的 Sync 语义。**建议：** 拆分 Sync 与 Close/Shutdown，使用共享且幂等的 sink 生命周期对象。

- **模板同步工具在 Windows 上把换行差异当成内容漂移。** dry-run 在干净工作区报告 4 个更新，但归一化 CRLF/LF 后内容完全一致；存在差异时命令仍返回 0，也不能作为 CI 门禁。**建议：** 比较前统一换行，为模板固定 LF，并增加有差异时返回非零的 `--check` 模式。

- **存储默认路径在文档、代码和配置中不一致。** README 和 boot 默认值使用 `./var/uploads`，模板 YAML 实际覆盖为 `./tmp/uploads`。**建议：** 选择一个唯一默认值，并通过模板一致性测试锁定。

- **CLI 主文件职责过多。** `cmd/initra/main.go` 约 1300 行，混合命令定义、执行、模板字符串和诊断逻辑；当前生成模板落后于架构已经说明该结构开始造成维护漂移。**建议：** 保持同 package，按 new、module、crud、config、migrate、skill、doctor 等命令域拆文件，不额外引入复杂框架。

- **核心 user 模块没有测试且 service 过大。** `examples/internal/modules/user/user.service.go` 约 641 行，承担验证、事务、Ent 查询、角色、缓存和 VO 映射，但该 package 没有 `_test.go`，实际覆盖率为 0.0%。**建议：** 保持 flat package 和直接使用 Ent 的方向，按职责拆成多个同包文件，并补 service 编排、事务和并发测试。

- **云存储和部分基础能力缺少测试。** 四个云 provider、`pagination`、部分 `fieldx/indexx` 没有测试；现有测试也未覆盖真实 PostgreSQL migration、Redis/Asynq 生命周期、Docker 启动和所有 `/api/` 路由的 RouteRegistry 完整性。**建议：** 增加容器集成测试、生成项目 smoke test、云 provider 契约测试和路由安全架构测试。

- **依赖图较重。** 根模块同时携带多套云 SDK、Asynq、Redis、Web 等依赖，默认生成项目也会包含全部对象存储 provider，冷编译和依赖校验成本较高。**建议：** 在确认维护收益后，将非默认云 provider 拆为可选 submodule；不要仅为目录形式进行过度拆分。

- **工程治理和发布资产仍不完整。** 仓库现已有 root/examples、生成项目和 tag 发布路径 CI，但仍缺少 LICENSE 或内部授权声明、SECURITY、CODEOWNERS、CONTRIBUTING、CHANGELOG、golangci 配置、GoReleaser、SBOM、签名和 provenance。**建议：** 补齐维护责任、安全报告渠道和变更记录；发布二进制时再加入 checksums、SBOM、签名和自动 release。

- **注释规范未被一致执行。** 仓库要求类型、函数、常量和测试代码使用符合 Go 规范的中文注释，但部分导出声明、测试 fake 和复杂 fixture 没有说明。**建议：** 用 lint 约束导出声明；测试代码只补充有信息量的背景、fixture 和关键断言说明，避免机械化注释。
