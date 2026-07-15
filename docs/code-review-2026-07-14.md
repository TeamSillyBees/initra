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

> 修复日期：2026-07-15。33 项已修复，3 项受外部环境或所有者决策约束而部分修复，2 项按项目决策延期。

- **[已接受延期] 当前 Go 和依赖版本存在已知可达漏洞。** 当前 toolchain 与 `quic-go v0.59.0` 仍命中 [GO-2026-5856](https://pkg.go.dev/vuln/GO-2026-5856)、[GO-2026-5039](https://pkg.go.dev/vuln/GO-2026-5039)、[GO-2026-5037](https://pkg.go.dev/vuln/GO-2026-5037) 和 [GO-2026-5676](https://pkg.go.dev/vuln/GO-2026-5676)。**延期说明：** 本轮按项目决策不升级 Go；后续单独执行 toolchain、依赖升级及 `govulncheck` 回归。
- **[已修复] `crud add` 名称与能力不符。** 原命令只生成表名片段，却暗示完整 CRUD。**修复方式：** 移除 `crud` 命令，改为明确不承诺持久化和路由接线的 `snippet add`，并补命令回归测试。
- **[已修复] `config add` 产物不会被应用加载。** 原命令未接入聚合 Config 和主配置文件。**修复方式：** 通过 AST 更新 `boot.Config`，同步写入 `configs/config.yaml` 和配置类型，并验证生成代码可编译。
- **[已修复] `--app-name` 只影响 README。** 原模板仍硬编码运行时名称和基础设施标识，多个生成项目会冲突。**修复方式：** 拆分展示名与安全 `app.slug`，将 slug 用于 JWT issuer、缓存/token namespace、Compose 和数据库名。
- **[已修复] Windows 空格路径会生成非法 replace。** 原实现直接拼接 `go.mod` 文本。**修复方式：** 使用 `x/mod/modfile` 写入 replace，并校验目标目录、`go.mod` 及模块路径。
- **[已修复] module/config 生成不是事务性的。** 多文件写入失败会留下部分产物。**修复方式：** module 在临时目录完成后原子 rename，config 先预检和暂存全部变更，提交失败时逆序回滚。
- **[已修复] `doctor` 不能作为环境 verifier。** 原命令在必需工具或配置缺失时仍成功退出。**修复方式：** 引入 OK、BROKEN、MISSING、OPTIONAL 状态、必需项非零退出、命令超时和稳定 JSON schema。
- **[已修复] `initra skill` 不能幂等执行或安全升级。** 已有文件会阻止再次安装，也无法区分框架升级和本地修改。**修复方式：** 增加哈希 manifest、`--check`、`--force`、本地修改保护及带备份恢复的 staged update。
- **[已修复] 内置 checker 漏报根仓库 `internal` import。** 原规则依赖逐行字符串和目录 substring。**修复方式：** 改用 Go AST 解析真实 import，精确匹配根模块 `internal[/...]`，并覆盖 boot/pkg 路径及注释误报测试。
- **[已修复] `allow_multiple_devices` 是无效配置。** 配置存在但认证逻辑从不读取，造成错误安全预期。**修复方式：** 从 AuthConfig、模板、默认值和文档中移除该字段；多设备会话控制留待有明确业务模型时再设计。
- **[已修复] refresh token 过早消费。** 原流程在用户查询成功前就使旧 token 失效。**修复方式：** 调整为先校验 token、查询用户，再原子轮转新旧 refresh token，并增加用户查询失败不消费 token 的测试。
- **[已修复] JWT 密钥和身份约束不足。** 原实现允许短密钥及 `UserID=0`。**修复方式：** 强制至少 32 字节 secret，并在签发、解析和认证中间件中拒绝非正数用户 ID。
- **[已修复] RedisTokenStore 的 nil client 会 fail-open。** 原实现会把部分认证状态操作伪装成成功。**修复方式：** 构造阶段拒绝 nil/typed-nil client，各操作也统一执行就绪校验并返回 token store failure。
- **[已修复] 默认管理员、JWT secret 和本地网络配置易被误用。** 固定管理员密码、已知 secret 和全接口本地监听会放大误部署风险。**修复方式：** `initra new` 随机生成一次性管理员密码和各环境 JWT secret，仅写入密码哈希；生产环境拒绝示例 secret，本地 API、PostgreSQL 和 Redis 绑定 loopback。
- **[已修复] 配置加载静默忽略未知字段。** 拼写错误会退回默认值而不报错。**修复方式：** 使用 `Viper.UnmarshalExact` 严格反序列化，并补 unknown-key 回归测试。
- **[已修复] 配置和结构化日志脱敏覆盖不完整。** 原实现没有完整覆盖 DSN、URL userinfo、struct、指针、typed map、复合字段和循环引用。**修复方式：** `config`、`logx` 与 `errors` 补齐结构化递归脱敏、URL 凭据清洗及数组/字节等字段保护，并以深度和循环限制避免异常递归。
- **[已修复] HTTP Client 默认记录上游错误正文。** 默认正文预览可能泄露 token 或 PII。**修复方式：** 默认仅记录状态信息，新增显式 `error_body_preview` opt-in，并验证默认日志不包含响应正文。
- **[已修复] Retry 无法关闭任意 5xx 重试。** 原逻辑在配置未命中时仍重试全部 5xx。**修复方式：** 默认严格服从状态码集合，仅在显式 `RetryAll5xx` 时扩大范围。
- **[已修复] 数据库启动 Ping 没有 deadline。** 网络异常可能长期阻塞启动。**修复方式：** 新增并校验 `PingTimeout`，注册数据库时使用 `context.WithTimeout`，并同步示例和模板。
- **[已修复] 数据库关系缺少外键约束。** user-role、role-menu 等关系可能产生孤儿数据。**修复方式：** Atlas diff 启用 `WithForeignKeys(true)`，新增独立迁移为现有关系表补充 RESTRICT 外键，并增加生成模板断言。
- **[已修复] 乐观锁字段助手容易被误解为完整锁能力。** 当前能力仍只定义 version 字段。**修复方式：** 将入口改名为 `OptimisticLockVersion`，并在注释、文档和 skill 中明确不包含 CAS、递增或冲突检测。
- **[已修复] AuditHook 依赖错误字符串判断字段能力。** 真实 Ent mutation 错误与该字符串约定不一致。**修复方式：** 改用 typed mutation capability 接口设置审计字段，并增加支持/缺失能力回归测试。
- **[已修复] 本地上传 overwrite 检查存在竞态。** 先检查再创建会在并发下覆盖文件。**修复方式：** `Overwrite=false` 时使用 `O_CREATE|O_EXCL` 原子创建，并增加并发覆盖保护测试。
- **[已修复] local provider 的 presign 语义误导。** 原实现返回无签名、不会过期的 URL。**修复方式：** local provider 明确返回 `storage.ErrUnsupported`，避免业务误当临时授权 URL。
- **[已修复] 请求上下文无条件信任代理头。** 客户端可控制用于 BaseURL 的 host/proto。**修复方式：** 默认忽略代理头，仅可信代理 IP/CIDR 可启用，并严格校验协议与 host。
- **[已修复] Redis SCAN 前缀允许 glob 元字符。** 错误前缀可能扩大删除范围。**修复方式：** SCAN/UNLINK 参数校验拒绝 `*`、`?`、`[` 和反斜线，并增加拒绝测试。
- **[已修复] 可观测性配置包含未接线开关。** metrics、tracing、pprof 和部分 task 配置会制造能力已启用的假象。**修复方式：** 删除模板中未实现的装饰性开关，仅保留已接线的 health；Redis tracing/metrics 继续通过真实 hook 生效。
- **[已修复] HTTP 上传限制存在两套真相。** 路由限制与 storage 上限不一致会提前拒绝合法文件。**修复方式：** file 路由从 storage `max_size` 派生 `MaxBodyBytes` 并计入 multipart 开销，增加边界测试。
- **[部分修复] Docker 模板尚未形成完整可复现基线。** 已增加 `.dockerignore`、distroless nonroot 最终镜像，仅复制二进制和必要 RBAC 文件，Compose 数据服务也只绑定 loopback；基础镜像仍使用 tag，未固定 digest。**后续方向：** 在确定镜像更新策略后固定受控 digest，并由自动化定期刷新。
- **[已修复] Logger `Sync` 会关闭共享 sink。** 子 logger 调用 Sync 会影响父 logger 生命周期。**修复方式：** 将 Sync 收敛为 flush，新增共享且幂等的 Close/Shutdown 生命周期，并由 Application 统一关闭 logger。
- **[已修复] 模板同步工具把 Windows 换行视为漂移。** CRLF/LF 差异会产生假更新，旧 dry-run 也不能作为 CI 门禁。**修复方式：** 比较和写入前统一换行，并增加存在差异时非零退出的 `--check` 及回归测试。
- **[已修复] 存储默认路径不一致。** README、boot 默认值和模板 YAML 分别指向不同目录。**修复方式：** 统一为 `./var/uploads`，同步示例、模板和文档。
- **[已修复] CLI 主文件职责过多。** 原 `main.go` 混合根命令、项目生成、迁移、skill、doctor 和模板渲染。**修复方式：** 保持同 package，按 project new、migrate、scaffold、skill、doctor 拆分文件，`main.go` 只保留入口、根命令和通用帮助。
- **[已修复] 核心 user 模块无测试且 service 过大。** 验证、Ent 查询、角色、缓存和 VO 映射集中在单文件中。**修复方式：** 保持 flat package，拆为 usecase、persistence、roles、mapper 等同包文件，并补验证、缓存命中、错误传播、角色归一化和并发缓存测试。
- **[部分修复] 云存储和基础能力测试不足。** 已补云 provider 构造/契约单测、pagination、fieldx/indexx、RouteRegistry 架构测试和生成项目 smoke test；真实 PostgreSQL migration、Redis/Asynq 生命周期、Docker 启动及真实云兼容服务仍无容器级验证。**后续方向：** 优先增加可重复的 PostgreSQL、Redis/Asynq、MinIO/S3-compatible 和镜像启动集成测试；真实云测试按凭据条件执行。
- **[延期] 依赖图较重。** 根模块和默认项目仍同时携带多套云 SDK、Asynq、Redis 与 Web 依赖，冷编译成本未下降。**延期说明：** 当前不拆 submodule；只有在维护和构建收益可量化时再拆非默认 provider，避免为目录形式过度模块化。
- **[部分修复/延期] 工程治理和发布资产不完整。** 已增加 SECURITY、CONTRIBUTING、CHANGELOG 和 golangci 配置；LICENSE/内部授权、CODEOWNERS、GoReleaser、SBOM、签名和 provenance 仍缺失。**后续方向：** 先由所有者确定授权、维护责任和发布模型，再补对应资产；当前 golangci 的存量告警也需逐步清理后纳入强制门禁。
- **[已修复] 注释规范未被一致执行。** 部分导出声明缺少以标识符开头的 Go 文档注释。**修复方式：** 补齐 37 个存量声明的语义化中文注释，并增加 AST 架构测试，后续出现任何新增缺口都会直接失败。
