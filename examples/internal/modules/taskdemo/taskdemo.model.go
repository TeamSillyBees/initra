package taskdemo

// PublishEmailDTO 是发布示例邮件任务的 service 入参。
type PublishEmailDTO struct {
	UserID  string
	Email   string
	TraceID string
}

// PublishedTaskResult 是任务发布后的 service 结果。
type PublishedTaskResult struct {
	TaskID string
	Type   string
	Queue  string
	State  string
	BizKey string
}

type sendEmailPayload struct {
	UserID string `json:"userId"`
	Email  string `json:"email"`
}
