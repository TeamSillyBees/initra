# 任务队列能力

`pkg/task` 是 initra 的任务队列抽象层，默认由 `pkg/task/asynqadapter` 适配 Asynq。业务代码只依赖 `task.Publisher`、`task.Worker`、`task.Registry`、`task.Scheduler`、`task.Task` 和 `task.TaskMeta`，不要直接 import `github.com/hibiken/asynq`。

## Boot 装配

在 `internal/boot/providers.go` 中注册：

```go
asynqadapter.Register(injector, cfg.Task)
```

配置结构组合 `task.Config`：

```go
type Config struct {
	Task task.Config `mapstructure:"task"`
}
```

启动校验中调用：

```go
if err := c.Task.Validate(); err != nil {
	return err
}
```

## 发布任务

业务 service 定义最小接口并通过构造函数注入：

```go
type taskPublisher interface {
	Publish(ctx context.Context, item task.Task, opts ...task.PublishOption) (*task.PublishResult, error)
}
```

发布外部副作用任务时必须声明 `biz_key`：

```go
userIDText := userID.String()
result, err := publisher.Publish(ctx, task.Task{
	Type: "demo:send_email",
	Payload: sendEmailPayload{
		UserID: userID,
		Email:  email,
	},
	Meta: task.TaskMeta{
		Module:         "demo",
		Owner:          "platform",
		BizKey:         "demo:" + userIDText + ":send_email",
		BizKeyRequired: true,
		SideEffect:     true,
		Idempotent:     true,
	},
}, task.WithQueue(task.QueueDefault), task.WithMaxRetry(3))
```

## Worker 处理器

Worker 侧通过 `task.Registry` 注册 handler：

```go
registry.Register("demo:send_email",
	task.HandlerFunc(func(ctx context.Context, item task.Task) error {
		payload, err := task.DecodePayload[sendEmailPayload](item)
		if err != nil {
			return fmt.Errorf("%w: %v", task.ErrSkipRetry, err)
		}
		_ = payload
		return nil
	}),
	task.WithRegisterModule("demo"),
	task.WithRegisterOwner("platform"),
	task.WithRegisterBizKeyRequired(true),
	task.WithRegisterSideEffect(true),
)
```

Handler 必须尊重 `ctx.Done()`，并在执行业务副作用前检查业务幂等状态。`pkg/task` 按 at-least-once 模型设计，不承诺 exactly-once。

## 周期任务

第一版只推荐静态代码注册：

```go
scheduler.Register("*/5 * * * *", task.Task{
	Type:    "demo:cleanup_expired_sessions",
	Payload: map[string]string{"source": "worker"},
	Meta: task.TaskMeta{
		Module:     "demo",
		Owner:      "platform",
		Idempotent: true,
	},
}, task.WithQueue(task.QueueLow), task.WithMaxRetry(1))
```

## 禁止事项

- 业务代码直接 import Asynq。
- 记录完整 payload、密码、token、密钥、验证码、session value。
- 把 `TaskID` 或 `Unique` 当作长期业务幂等机制。
- 对会产生外部副作用的任务省略 `biz_key`。
