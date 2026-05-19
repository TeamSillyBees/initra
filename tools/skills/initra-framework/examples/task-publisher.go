package examples

import (
	"context"

	"github.com/teamsillybees/initra/pkg/task"
)

type demoTaskPublisher interface {
	Publish(ctx context.Context, item task.Task, opts ...task.PublishOption) (*task.PublishResult, error)
}

type demoSendEmailPayload struct {
	UserID string `json:"user_id"`
	Email  string `json:"email"`
}

func publishDemoEmail(ctx context.Context, publisher demoTaskPublisher, userID string, email string) (*task.PublishResult, error) {
	bizKey := "demo:" + userID + ":send_email"
	return publisher.Publish(ctx, task.Task{
		Type: "demo:send_email",
		Payload: demoSendEmailPayload{
			UserID: userID,
			Email:  email,
		},
		Meta: task.TaskMeta{
			Module:         "demo",
			Owner:          "platform",
			BizKey:         bizKey,
			BizKeyRequired: true,
			SideEffect:     true,
			Idempotent:     true,
		},
	}, task.WithQueue(task.QueueDefault), task.WithMaxRetry(3))
}
