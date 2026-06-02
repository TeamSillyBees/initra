package httpdemo

import (
	"context"
	"errors"
	"strings"

	"github.com/teamsillybees/initra/examples/internal/modules/bizerrors"
	"github.com/teamsillybees/initra/pkg/httpclient"
)

const defaultMessage = "hello from initra"

// Service 是 httpdemo 示例模块的应用服务。
type Service struct {
	client httpclient.ReadCaller
}

// NewService 构造 httpdemo 示例模块应用服务。
func NewService(client httpclient.ReadCaller) *Service {
	return &Service{client: client}
}

// GetHTTPBingo 调用 HTTPBingo /get 并返回 JSON 回显结果。
func (s *Service) GetHTTPBingo(ctx context.Context, message string, traceID string) (HTTPBingoGetVO, error) {
	if err := s.ensureClient(); err != nil {
		return HTTPBingoGetVO{}, err
	}
	message = strings.TrimSpace(message)
	if message == "" {
		message = defaultMessage
	}
	payload := httpBingoGetPayload{}
	err := s.client.GetJSON(ctx, "/get", &payload,
		httpclient.WithQuery("message", message),
		httpclient.WithHeader("X-Trace-ID", strings.TrimSpace(traceID)),
	)
	if err != nil {
		return HTTPBingoGetVO{}, mapHTTPClientError(ctx, err, "call httpbingo get failed")
	}
	return httpBingoGetVOFromPayload(payload), nil
}

// GetHTTPBingoFormPage 调用 HTTPBingo /forms/post 并返回 HTML 表单页内容。
func (s *Service) GetHTTPBingoFormPage(ctx context.Context, traceID string) (HTTPBingoFormPageVO, error) {
	if err := s.ensureClient(); err != nil {
		return HTTPBingoFormPageVO{}, err
	}
	body, resp, err := s.client.GetBytes(ctx, "/forms/post",
		httpclient.WithHeader("X-Trace-ID", strings.TrimSpace(traceID)),
	)
	if err != nil {
		return HTTPBingoFormPageVO{}, mapHTTPClientError(ctx, err, "call httpbingo form page failed")
	}
	return HTTPBingoFormPageVO{
		ContentType: resp.Header.Get("Content-Type"),
		Size:        int32(len(body)),
		Body:        string(body),
	}, nil
}

func (s *Service) ensureClient() error {
	if s == nil || s.client == nil {
		return bizerrors.Internal("httpbingo client is not configured")
	}
	return nil
}

func httpBingoGetVOFromPayload(payload httpBingoGetPayload) HTTPBingoGetVO {
	return HTTPBingoGetVO{
		Args:    payload.Args,
		Headers: payload.Headers,
		Method:  payload.Method,
		Origin:  payload.Origin,
		URL:     payload.URL,
	}
}

func mapHTTPClientError(ctx context.Context, err error, message string) error {
	if err == nil {
		return nil
	}
	if clientErr, ok := errors.AsType[*httpclient.Error](err); ok {
		return bizerrors.WrapHTTPClientContext(ctx, err, message,
			bizerrors.WithCauseAttr("service", clientErr.Service),
			bizerrors.WithCauseAttr("kind", string(clientErr.Kind)),
			bizerrors.WithCauseAttr("status_code", clientErr.StatusCode),
		)
	}
	return bizerrors.WrapHTTPClientContext(ctx, err, message)
}
