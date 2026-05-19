package task

import (
	"context"
	"time"
)

type disabledPublisher struct{}

// NewDisabledPublisher 返回未启用任务队列时的 Publisher。
func NewDisabledPublisher() Publisher {
	return disabledPublisher{}
}

func (disabledPublisher) Publish(context.Context, Task, ...PublishOption) (*PublishResult, error) {
	return nil, ErrDisabled
}

func (disabledPublisher) PublishAt(context.Context, Task, time.Time, ...PublishOption) (*PublishResult, error) {
	return nil, ErrDisabled
}

func (disabledPublisher) PublishIn(context.Context, Task, time.Duration, ...PublishOption) (*PublishResult, error) {
	return nil, ErrDisabled
}

func (disabledPublisher) Close() error {
	return nil
}

type disabledWorker struct{}

// NewDisabledWorker 返回未启用消费端时的 Worker。
func NewDisabledWorker() Worker {
	return disabledWorker{}
}

func (disabledWorker) Start(context.Context) error {
	return ErrDisabled
}

func (disabledWorker) Stop(context.Context) error {
	return nil
}

func (disabledWorker) Shutdown(context.Context) error {
	return nil
}

type disabledScheduler struct{}

// NewDisabledScheduler 返回未启用调度器时的 Scheduler。
func NewDisabledScheduler() Scheduler {
	return disabledScheduler{}
}

func (disabledScheduler) Register(string, Task, ...ScheduleOption) error {
	return ErrDisabled
}

func (disabledScheduler) Start(context.Context) error {
	return ErrDisabled
}

func (disabledScheduler) Shutdown(context.Context) error {
	return nil
}
