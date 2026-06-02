//go:build ignore

package sms

import (
	"context"
	"errors"

	"github.com/teamsillybees/initra/pkg/httpclient"

	apperrors "github.com/teamsillybees/initra/pkg/errors"
)

// Sender 封装短信服务远程调用。
type Sender struct {
	client httpclient.JSONPoster
}

// NewSender 创建短信发送器。
func NewSender(client httpclient.JSONPoster) *Sender {
	return &Sender{client: client}
}

// Send 调用远程短信服务发送验证码。
func (s *Sender) Send(ctx context.Context, mobile string, code string, traceID string) error {
	body := map[string]string{
		"mobile": mobile,
		"code":   code,
	}
	var payload struct {
		RequestID string `json:"requestId"`
	}
	err := s.client.PostJSON(ctx, "/messages", body, &payload,
		httpclient.WithHeader("X-Trace-ID", traceID),
	)
	if err != nil {
		return mapHTTPClientError(err, "发送短信失败")
	}
	return nil
}

func mapHTTPClientError(err error, message string) error {
	if err == nil {
		return nil
	}
	if clientErr, ok := errors.AsType[*httpclient.Error](err); ok {
		return apperrors.Wrap(err, apperrors.CodeInternalError, message,
			apperrors.WithDetail("service", clientErr.Service),
			apperrors.WithDetail("kind", string(clientErr.Kind)),
			apperrors.WithDetail("status_code", clientErr.StatusCode),
		)
	}
	return apperrors.Wrap(err, apperrors.CodeInternalError, message)
}
