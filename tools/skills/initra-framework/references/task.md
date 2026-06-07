# 任务队列与调度

## 注册

Boot 层使用 Asynq adapter：

```go
asynqadapter.Register(injector, cfg.Task)
```

业务代码只依赖 `pkg/task`，不得 import `github.com/hibiken/asynq`。

## 发布

```go
result, err := publisher.Publish(ctx, task.Task{
	Type: "demo:send_email",
	Payload: sendEmailPayload{UserID: userID, Email: email},
	Meta: task.TaskMeta{
		Module:         "demo",
		Owner:          "platform",
		BizKey:         bizKey,
		BizKeyRequired: true,
		SideEffect:     true,
		Idempotent:     true,
		TraceID:        traceID,
	},
}, task.WithQueue(task.QueueDefault), task.WithMaxRetry(3), task.WithBizKey(bizKey))
```

任务类型必须是 `{module}:{action}`。

## 幂等

initra 任务按 at-least-once 设计，不承诺 exactly-once。`biz_key` 是业务幂等键，不是 Asynq `TaskID`，也不是 Asynq `Unique`。外部副作用任务必须由业务侧保证幂等。

## Worker 与 Scheduler

Worker 侧注册 `task.Registry` handler；周期任务使用 `task.Scheduler`。handler 内先解析 payload，再执行业务 service。

## 禁止

- 不要把 `WithTaskID` 或 `WithUnique` 当作长期业务幂等。
- 不要发布外部副作用任务时缺少 `BizKey`。
- 不要把敏感 payload 写入日志或 trace。
