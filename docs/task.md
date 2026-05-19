# initra Asynq 任务队列封装开发指导

> 目标：基于 Asynq 为 `initra` 封装立即任务、指定时间任务、周期性定时任务、任务处理器注册、队列配置、可观测性元数据、重试与超时策略，并统一使用 `Publisher / Worker` 命名。

---

## 1. 定位与边界

- `pkg/task` 是 `initra` 的任务队列抽象层。
- 底层默认实现为 Asynq。
- Asynq `Client` 在框架中命名为 `Publisher`。
- Asynq `Server` 在框架中命名为 `Worker`。
- 业务代码不直接依赖 `github.com/hibiken/asynq`。
- Asynq 类型只允许出现在 `pkg/task/asynqadapter`。
- 框架不把 Asynq 抽象成通用消息队列。
- 框架不承诺 exactly-once。
- 框架按 at-least-once 模型设计。
- 外部副作用任务必须由业务侧保证幂等。
- `biz_key` 是业务幂等键，不是 Asynq `TaskID`，也不是 Asynq `Unique`。

---

## 2. 核心概念

- `Publisher`：发布任务；对应 Asynq `Client`；支持立即、延迟、指定时间任务。
- `Worker`：消费任务；对应 Asynq `Server`；接入应用生命周期。
- `Registry`：注册任务处理器；维护 `task_type -> handler` 映射。
- `Scheduler`：注册周期性任务；底层优先适配 Asynq `PeriodicTaskManager`。
- `Task`：框架层任务定义；包含 `type`、`payload`、`meta`、`options`。
- `TaskMeta`：任务治理、审计、观测、幂等约束元数据。
- `biz_key`：不可重复业务动作的幂等键；由业务定义、框架校验、日志记录。

---

## 3. 推荐目录结构

- `pkg/task/config.go`：任务队列配置。
- `pkg/task/task.go`：`Task`、`TaskMeta`、`TaskOption`。
- `pkg/task/publisher.go`：`Publisher` 接口。
- `pkg/task/worker.go`：`Worker` 接口。
- `pkg/task/registry.go`：`Registry`、`Handler`。
- `pkg/task/scheduler.go`：周期任务接口。
- `pkg/task/retry.go`：重试策略抽象。
- `pkg/task/middleware.go`：处理器中间件。
- `pkg/task/observability.go`：日志、指标、链路追踪元数据。
- `pkg/task/errors.go`：任务错误语义。
- `pkg/task/asynqadapter/publisher.go`：适配 Asynq Client。
- `pkg/task/asynqadapter/worker.go`：适配 Asynq Server。
- `pkg/task/asynqadapter/registry.go`：适配 Asynq ServeMux。
- `pkg/task/asynqadapter/scheduler.go`：适配 Asynq PeriodicTaskManager。
- `pkg/task/asynqadapter/options.go`：映射框架选项到 Asynq options。
- `pkg/task/asynqadapter/retry.go`：映射重试策略到 Asynq `RetryDelayFunc`。

---

## 4. 配置设计

```yaml
task:
  enabled: true
  backend: asynq
  redis:
    mode: single # single | sentinel | cluster
    addr: 127.0.0.1:6379
    username:
    password:
    db: 2
    tls: false
    master_name:
    sentinel_addrs: []
    cluster_addrs: []
  publisher:
    default_queue: default
    default_max_retry: 3
    default_timeout: 5m
    default_retention: 24h
    enforce_biz_key: true
  worker:
    enabled: true
    concurrency: 10
    shutdown_timeout: 30s
    health_check_interval: 15s
    delayed_task_check_interval: 5s
    task_check_interval: 1s
    strict_priority: false
    queues:
      critical: 6
      default: 3
      low: 1
  scheduler:
    enabled: false
    sync_interval: 3m
    timezone: Asia/Shanghai
  observability:
    logging: true
    metrics: true
    tracing: true
    include_payload_in_log: false
    include_payload_in_trace: false
```

- `task.enabled=false` 时不初始化任务队列。
- `publisher.default_queue` 用于未显式指定队列的任务。
- `publisher.default_max_retry` 用于未显式指定重试次数的任务。
- `publisher.default_timeout` 用于未显式指定超时时间的任务。
- 默认重试延迟策略使用 Asynq 官方策略。
- 仅当配置自定义重试策略时才设置 Asynq `RetryDelayFunc`。
- `worker.queues` 映射到 Asynq 队列权重。
- `worker.strict_priority=true` 时启用严格优先级。
- payload 默认不得进入日志或 trace。

---

## 5. Redis 连接策略

- 推荐复用 `initra` 已有 Redis 配置与连接池。
- 复用外部 Redis 客户端时，由 `initra` 管理连接生命周期。
- Asynq adapter 自建 Redis 客户端时，由 adapter 负责关闭连接。
- 任务队列 Redis DB 建议与业务缓存 Redis DB 隔离。
- 生产环境必须设置 Redis 密码。
- 跨公网连接 Redis 必须启用 TLS 或专线网络。

---

## 6. Task 与 TaskMeta

```go
type Task struct {
    Type    string
    Payload any
    Meta    TaskMeta
    Options []TaskOption
}

type TaskMeta struct {
    Name           string
    Description    string
    Module         string
    Owner          string
    Scenario       string
    BizKey         string
    BizKeyRequired bool
    SideEffect     bool
    CostLevel      string // low | medium | high
    Idempotent     bool
    TraceID        string
    TenantID       string
    CorrelationID  string
    Tags           map[string]string
}
```

- `Task.Type` 必须非空。
- `Task.Type` 命名格式：`{module}:{action}`。
- 示例：`user:send_welcome_email`、`order:sync_payment_result`。
- `Payload` 必须可 JSON 序列化。
- `Payload` 只放任务执行所需的最小信息。
- `Payload` 不放完整领域对象、密码、token、密钥、身份证号等敏感字段。
- Handler 内部应重新读取数据库获取最新状态。
- `TaskMeta` 用于日志、指标、trace、审计、任务治理和 `biz_key` 校验。
- `Owner` 必须能定位责任团队或责任人。
- `CostLevel` 用于区分低成本任务和高成本任务。
- `Idempotent` 用于声明业务是否已经实现幂等。

---

## 7. biz_key 规则

- 凡是会产生外部副作用的任务，必须声明 `biz_key`。
- 凡是成本较高的任务，必须声明 `biz_key`。
- 凡是不可重复执行的任务，必须声明 `biz_key`。
- 必须声明 `biz_key` 的典型任务：短信、邮件、第三方支付、优惠券、权益发放、账单创建、外部系统状态变更、Webhook、文件转码、大批量导入导出。
- 可不强制 `biz_key` 的典型任务：清理临时文件、刷新缓存、低价值异步日志、可重复执行且无外部副作用的统计任务。
- `biz_key` 建议格式：`{module}:{business_id}:{action}`。
- 示例：`order:123456:sync_payment_result`。
- 示例：`user:10001:send_welcome_email`。
- `biz_key` 必须进入结构化日志字段。
- `biz_key` 可以进入 Asynq Header。
- `biz_key` 不应直接作为 Prometheus label。
- `biz_key` 不应只依赖 Asynq `Unique` 实现幂等。
- `biz_key` 不应只依赖 Asynq `TaskID` 实现长期幂等。
- 不可重复外部副作用任务建议使用业务表或幂等表唯一约束：`task_type + biz_key + effect_type`。

---

## 8. Publisher 接口

```go
type Publisher interface {
    Publish(ctx context.Context, task Task, opts ...PublishOption) (*PublishResult, error)
    PublishAt(ctx context.Context, task Task, at time.Time, opts ...PublishOption) (*PublishResult, error)
    PublishIn(ctx context.Context, task Task, delay time.Duration, opts ...PublishOption) (*PublishResult, error)
    Close() error
}
```

- `Publish` 发布立即任务。
- `PublishAt` 发布指定时间任务。
- `PublishIn` 发布延迟任务。
- `PublishOption` 建议支持：`WithQueue`、`WithMaxRetry`、`WithTimeout`、`WithDeadline`、`WithRetention`、`WithUnique`、`WithTaskID`、`WithHeader`、`WithBizKey`。
- `PublishResult` 建议包含：`TaskID`、`Type`、`Queue`、`State`、`ProcessAt`、`BizKey`。
- `Publish` 前必须校验任务类型、payload、队列名和 `biz_key`。
- `Publish` 不应吞掉重复任务错误。
- `Publish` 默认不记录完整 payload。
- trace 上下文应通过 header 传递给 Worker。

---

## 9. Handler 与 Registry

```go
type Handler interface {
    HandleTask(ctx context.Context, task Task) error
}

type Registry interface {
    Register(taskType string, handler Handler, opts ...RegisterOption) error
}
```

- 一个 `task_type` 只能注册一个 handler。
- 重复注册必须返回错误。
- 空 `task_type` 必须拒绝。
- 空 handler 必须拒绝。
- Registry 负责组装 middleware 链。
- Registry 负责生成 Asynq `ServeMux`。
- 注册时可附加：`module`、`owner`、`description`、`biz_key_required`、`side_effect`、`cost_level`、`default_queue`、`default_max_retry`、`default_timeout`。
- Handler 必须尊重 `ctx.Done()`。
- Handler 必须处理重复执行。
- Handler 必须处理任务超时和业务状态变化。
- Handler 不应在内部启动失控 goroutine。

---

## 10. Worker 设计

```go
type Worker interface {
    Start(ctx context.Context) error
    Stop(ctx context.Context) error
    Shutdown(ctx context.Context) error
}
```

- Worker 必须接入 `initra.Application` 生命周期。
- Worker 启动时必须使用已完成注册的 Registry。
- Worker 关闭时必须触发优雅关闭。
- 推荐运行模式：`api` 只启用 Publisher，`worker` 启用 Worker，`all` 同进程启用 API 和 Worker。
- 生产环境推荐 API 与 Worker 分进程部署。
- Worker 并发数、监听队列、shutdown timeout 由配置控制。
- Worker 应暴露 Redis ping 健康状态。
- Worker 应在启动时输出已注册任务类型清单。

---

## 11. 队列配置

- 默认队列：`default`。
- 推荐内置队列：`critical`、`default`、`low`。
- 队列权重示例：`critical: 6`、`default: 3`、`low: 1`。
- `strict_priority=false` 时使用加权优先级。
- `strict_priority=true` 时使用严格优先级。
- 高优先级任务才允许进入 `critical`。
- 低价值批处理任务建议进入 `low`。
- 队列名必须由框架校验。
- 业务代码不允许随意构造队列名。
- 未声明队列名应拒绝发布。

---

## 12. 重试与超时

- 默认重试延迟逻辑使用 Asynq 官方默认策略。
- 框架允许设置最大重试次数、任务超时时间、任务截止时间、自定义重试策略。
- 单个任务可覆盖默认重试次数和默认超时时间。
- 自定义重试策略建议配置在 Worker 级别。
- 不建议每个任务随意定义复杂重试策略。
- 推荐内置重试策略：`official`、`fixed`、`linear`、`exponential`。
- `official` 表示不设置 Asynq `RetryDelayFunc`。
- 第三方限流任务应使用较低并发和较长重试间隔。
- 用户可感知通知任务应限制最大重试次数。
- 不可重复外部副作用任务在重试前必须检查业务幂等状态。

---

## 13. 周期性定时任务

```go
type Scheduler interface {
    Register(cron string, task Task, opts ...ScheduleOption) error
    Start(ctx context.Context) error
    Shutdown(ctx context.Context) error
}
```

- 周期任务不应通过业务代码手动循环实现。
- 周期任务统一通过 `Scheduler` 注册。
- 底层优先使用 Asynq `PeriodicTaskManager`。
- 第一版建议只支持静态代码注册。
- 第二版可支持配置文件或数据库动态注册。
- 周期任务必须声明：cron 表达式、task type、payload、queue、timeout、max retry、owner。
- 周期任务配置错误应导致 Worker 启动失败。
- 周期任务必须在启动时输出注册清单。
- 周期任务如果产生外部副作用，必须声明 `biz_key` 生成规则。
- 周期任务 `biz_key` 可包含执行时间窗口。
- 示例：`report:daily:2026-05-19`、`sync:inventory:2026-05-19T10`。

---

## 14. 可观测性

- 必须内置结构化日志。
- 必须预留 metrics。
- 必须预留 tracing。
- 日志字段建议包含：`task_id`、`task_type`、`task_name`、`queue`、`biz_key`、`module`、`owner`、`retry_count`、`max_retry`、`timeout`、`duration_ms`、`trace_id`、`tenant_id`、`correlation_id`、`error`。
- Metrics 指标建议包含：`task_publish_total`、`task_publish_failed_total`、`task_processed_total`、`task_failed_total`、`task_retry_total`、`task_duration_seconds`、`task_queue_latency_seconds`、`task_active`、`task_pending`、`task_scheduled`、`task_archived`。
- Metrics label 建议包含：`task_type`、`queue`、`module`、`result`。
- Metrics label 不应包含：`biz_key`、`task_id`、`trace_id`、`user_id`、`order_id`。
- Trace attribute 建议包含：`messaging.system=asynq`、`messaging.destination.name`、`messaging.operation`、`task.type`、`task.queue`、`task.biz_key`。
- Publisher 发布任务时应将 trace 上下文写入 header。
- Worker 处理任务时应从 header 恢复 trace 上下文。
- payload 默认不得进入日志或 trace。
- 开启 payload 调试时必须注意脱敏。

---

## 15. Middleware

- Worker Handler 必须支持 middleware 链。
- 推荐内置 middleware：`RecoverMiddleware`、`TracingMiddleware`、`LoggingMiddleware`、`MetricsMiddleware`、`BizKeyValidationMiddleware`、`IdempotencyMiddleware`。
- `RecoverMiddleware` 捕获 panic 并转换为错误。
- `TracingMiddleware` 创建任务处理 span。
- `LoggingMiddleware` 记录开始、成功、失败、耗时。
- `MetricsMiddleware` 记录处理次数、失败次数、耗时。
- `BizKeyValidationMiddleware` 校验强制 `biz_key`。
- `IdempotencyMiddleware` 第一版可只预留接口。
- 推荐顺序：Recover → Tracing → Logging → Metrics → BizKeyValidation → Idempotency → BusinessHandler。

---

## 16. 错误语义

- 推荐框架错误：`ErrSkipRetry`、`ErrRevoke`、`ErrInvalidTask`、`ErrMissingBizKey`、`ErrDuplicateTask`、`ErrPublishFailed`。
- `ErrSkipRetry` 映射到 Asynq `SkipRetry`。
- `ErrRevoke` 映射到 Asynq `RevokeTask`。
- 临时网络错误应允许重试。
- 第三方服务限流应允许重试。
- JSON 解析错误通常应跳过重试。
- 业务对象不存在通常应跳过重试。
- 业务状态已完成通常应返回成功或撤销。
- Handler 不应把所有错误都包装成 `ErrSkipRetry`。

---

## 17. 生命周期

- `Bootstrap` 阶段：加载配置、初始化 Redis、初始化 Publisher、初始化 Registry、注册任务处理器、初始化 Worker、初始化 Scheduler。
- `Application.Run(ctx)` 阶段：按运行模式启动 Worker 和 Scheduler，监听 `ctx.Done()`，触发优雅关闭。
- 关闭顺序：停止接收新 HTTP 请求 → 停止发布新任务 → 停止 Scheduler → 停止 Worker 拉取新任务 → 等待活跃任务完成 → 关闭 Publisher → 关闭 Redis。
- Worker 停止必须尊重 `shutdown_timeout`。
- Handler 必须正确响应 context cancellation。

---

## 18. 模板与示例

- `templates/api`：默认启用 Publisher；默认不启用 Worker；提供立即任务发布示例。
- `templates/worker`：默认启用 Worker；提供任务处理器示例；提供周期任务示例。
- `examples/api`：演示从 HTTP 请求发布异步任务。
- `examples/worker`：演示消费异步任务。
- 示例任务：`demo:send_email`、`demo:generate_report`、`demo:cleanup_expired_sessions`。

---

## 19. 第一阶段交付清单

- `task.Config`、`Task`、`TaskMeta`。
- `Publisher` 接口和 Asynq Publisher adapter。
- `Publish`、`PublishAt`、`PublishIn`。
- `Registry`、`Handler`、`HandlerFunc`。
- Asynq Registry adapter。
- `Worker` 生命周期封装。
- 队列配置映射。
- 默认重试次数与超时时间配置。
- 官方重试策略默认保留。
- 自定义 `RetryDelayFunc` 适配。
- `biz_key` 强制校验。
- zap 日志 middleware。
- metrics 和 tracing middleware 预留。
- API 模板发布任务示例。
- Worker 模板处理任务示例。

---

## 20. 第二阶段交付清单

- 周期任务配置文件注册。
- 周期任务数据库动态注册。
- 幂等表与 `IdempotencyMiddleware`。
- Outbox 可靠投递。
- asynqmon 可选集成。
- Prometheus 指标完善。
- OpenTelemetry 链路完善。
- 任务管理接口。
- `initra task` CLI 子命令。

---

## 21. 不建议第一版实现

- 不建议第一版做复杂任务编排。
- 不建议第一版做 DAG 工作流。
- 不建议第一版做跨服务事件总线。
- 不建议第一版做完整任务管理后台。
- 不建议第一版支持任意动态脚本任务。
- 不建议第一版暴露全部 Asynq option 给业务。
- 不建议第一版提供 payload 查询展示能力。
- 不建议第一版承诺 exactly-once。

---

## 22. 开发注意事项

- `initra` 的抽象必须薄。
- 不要重新实现 Asynq 已提供的调度、重试、队列能力。
- 框架重点应放在命名规范、生命周期集成、配置治理、幂等约束、可观测性、模板示例、错误语义。
- Asynq 当前仍处于 `v0.x` 主版本，必须固定依赖版本。
- `go.mod` 不建议使用无约束 `@latest`。
- adapter 层应隔离 Asynq API 变化。
- 业务代码不应 import `github.com/hibiken/asynq`。
- 文档中必须明确：Asynq 是 at-least-once；`biz_key` 是业务幂等键；`Unique` 不是长期幂等机制；`TaskID` 不是完整业务幂等机制；外部副作用必须业务侧幂等。

---

## 23. 参考依据

- Asynq README：Client 入队，Server 拉取队列并启动 worker goroutine，支持 scheduling、retries、weighted priority queues、strict priority queues、timeout、periodic tasks、Prometheus、Web UI。
- Asynq `client.go`：`MaxRetry`、`Queue`、`Timeout`、`Deadline`、`Unique`、`ProcessAt`、`ProcessIn`、`TaskID`、`Retention`、`Header`。
- Asynq `server.go`：`Config.Queues`、`StrictPriority`、`RetryDelayFunc`、`ErrorHandler`、`ShutdownTimeout`、`HealthCheckFunc`、`Handler`、`SkipRetry`、`RevokeTask`。
- Asynq `periodic_task_manager.go`：`PeriodicTaskManager`、`PeriodicTaskConfigProvider`、`PeriodicTaskConfig`、`SyncInterval`。
