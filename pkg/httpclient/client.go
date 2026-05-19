package httpclient

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/go-resty/resty/v2"
	"go.uber.org/zap"
)

const jsonContentType = "application/json"

// Client 封装单个远程服务的 HTTP 调用能力。
type Client struct {
	name        string
	config      ServiceConfig
	resty       *resty.Client
	logger      *zap.Logger
	authHandler AuthHandler
}

func newClient(name string, global Config, cfg ServiceConfig, logger *zap.Logger) (*Client, error) {
	if logger == nil {
		logger = zap.NewNop()
	}
	if err := validateServiceConfig(name, cfg); err != nil {
		return nil, err
	}
	return &Client{
		name:        name,
		config:      cfg,
		resty:       newRestyClient(global, cfg),
		logger:      logger,
		authHandler: defaultAuthHandler{},
	}, nil
}

// Name 返回远程服务名称。
func (c *Client) Name() string {
	if c == nil {
		return ""
	}
	return c.name
}

// Get 发送 HTTP GET 请求。
func (c *Client) Get(ctx context.Context, path string, opts ...RequestOption) (*Response, error) {
	return c.do(ctx, http.MethodGet, path, nil, opts...)
}

// Post 发送 HTTP POST 请求，body 会作为 JSON 请求体提交。
func (c *Client) Post(ctx context.Context, path string, body any, opts ...RequestOption) (*Response, error) {
	return c.do(ctx, http.MethodPost, path, body, opts...)
}

// Put 发送 HTTP PUT 请求，body 会作为 JSON 请求体提交。
func (c *Client) Put(ctx context.Context, path string, body any, opts ...RequestOption) (*Response, error) {
	return c.do(ctx, http.MethodPut, path, body, opts...)
}

// Patch 发送 HTTP PATCH 请求，body 会作为 JSON 请求体提交。
func (c *Client) Patch(ctx context.Context, path string, body any, opts ...RequestOption) (*Response, error) {
	return c.do(ctx, http.MethodPatch, path, body, opts...)
}

// Delete 发送 HTTP DELETE 请求。
func (c *Client) Delete(ctx context.Context, path string, opts ...RequestOption) (*Response, error) {
	return c.do(ctx, http.MethodDelete, path, nil, opts...)
}

// GetJSON 发送 GET 请求并将 JSON 响应反序列化为 T。
func GetJSON[T any](ctx context.Context, c *Client, path string, opts ...RequestOption) (*T, error) {
	var result T
	options := append([]RequestOption{WithResult(&result)}, opts...)
	if _, err := c.Get(ctx, path, options...); err != nil {
		return nil, err
	}
	return &result, nil
}

// PostJSON 发送 POST 请求并将 JSON 响应反序列化为 T。
func PostJSON[T any](ctx context.Context, c *Client, path string, body any, opts ...RequestOption) (*T, error) {
	var result T
	options := append([]RequestOption{WithResult(&result)}, opts...)
	if _, err := c.Post(ctx, path, body, options...); err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *Client) do(ctx context.Context, method string, path string, body any, opts ...RequestOption) (*Response, error) {
	if c == nil {
		return nil, &Error{
			Kind:    ErrorKindInternal,
			Method:  method,
			URL:     safeErrorURL(path),
			Code:    "nil_client",
			Message: "HTTP Client 不能为空",
		}
	}
	if ctx == nil {
		ctx = context.Background()
	}
	options := newRequestOptions(opts)
	requestCtx, cancel := requestContext(ctx, options)
	if cancel != nil {
		defer cancel()
	}

	req := c.resty.R().
		SetContext(requestCtx).
		SetQueryParams(options.QueryParams).
		SetPathParams(options.PathParams)

	if options.Result != nil {
		req.SetResult(options.Result)
	}
	if body != nil {
		req.SetBody(body)
		if options.ContentType == "" && !hasHeader(options.Headers, "Content-Type") {
			req.SetHeader("Content-Type", jsonContentType)
		}
	}
	if options.ContentType != "" {
		req.SetHeader("Content-Type", options.ContentType)
	}
	if err := c.authHandler.Apply(requestCtx, req, c.config.Auth); err != nil {
		httpErr := c.enrichError(err, ErrorKindInternal, method, path, 0)
		c.logRequest(method, path, options.Headers, 0, time.Duration(0), httpErr)
		return nil, httpErr
	}
	req.SetHeaders(options.Headers)

	start := time.Now()
	resp, err := req.Execute(method, path)
	duration := time.Since(start)
	statusCode := 0
	if resp != nil {
		statusCode = resp.StatusCode()
	}

	if err != nil {
		httpErr := c.errorFromExecution(err, resp, method, path)
		c.logRequest(method, path, options.Headers, statusCode, duration, httpErr)
		return nil, httpErr
	}
	if resp == nil {
		httpErr := &Error{
			Kind:    ErrorKindInternal,
			Service: c.name,
			Method:  method,
			URL:     safeErrorURL(path),
			Code:    "empty_response",
			Message: "HTTP Client 未返回响应",
		}
		c.logRequest(method, path, options.Headers, statusCode, duration, httpErr)
		return nil, httpErr
	}
	if !resp.IsSuccess() {
		httpErr := &Error{
			Kind:       ErrorKindResponse,
			Service:    c.name,
			Method:     method,
			URL:        safeErrorURL(resp.Request.URL),
			StatusCode: statusCode,
			Code:       fmt.Sprint(statusCode),
			Message:    responseMessage(resp),
		}
		c.logRequest(method, path, options.Headers, statusCode, duration, httpErr)
		return nil, httpErr
	}

	result := options.Result
	if result == nil {
		result = resp.Result()
	}
	wrapped := newResponse(resp, result)
	c.logRequest(method, path, options.Headers, statusCode, duration, nil)
	return wrapped, nil
}

func requestContext(ctx context.Context, opts RequestOptions) (context.Context, context.CancelFunc) {
	if opts.Idempotent {
		ctx = context.WithValue(ctx, retryIdempotentContextKey{}, true)
	}
	if opts.Timeout <= 0 {
		return ctx, nil
	}
	return context.WithTimeout(ctx, opts.Timeout)
}

func (c *Client) errorFromExecution(err error, resp *resty.Response, method string, path string) *Error {
	statusCode := 0
	url := path
	kind := ErrorKindRequest
	if resp != nil {
		statusCode = resp.StatusCode()
		if resp.Request != nil && resp.Request.URL != "" {
			url = resp.Request.URL
		}
		if resp.RawResponse != nil {
			kind = ErrorKindInternal
		}
	}
	if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		kind = ErrorKindRequest
	}
	return &Error{
		Kind:       kind,
		Service:    c.name,
		Method:     method,
		URL:        safeErrorURL(url),
		StatusCode: statusCode,
		Code:       executionErrorCode(kind, err),
		Message:    err.Error(),
		Cause:      err,
	}
}

func (c *Client) enrichError(err error, kind ErrorKind, method string, path string, statusCode int) *Error {
	if httpErr, ok := errors.AsType[*Error](err); ok {
		if httpErr.Service == "" {
			httpErr.Service = c.name
		}
		if httpErr.Method == "" {
			httpErr.Method = method
		}
		if httpErr.URL == "" {
			httpErr.URL = safeErrorURL(path)
		}
		if httpErr.StatusCode == 0 {
			httpErr.StatusCode = statusCode
		}
		return httpErr
	}
	return &Error{
		Kind:       kind,
		Service:    c.name,
		Method:     method,
		URL:        safeErrorURL(path),
		StatusCode: statusCode,
		Code:       "internal_error",
		Message:    err.Error(),
		Cause:      err,
	}
}

func (c *Client) logRequest(method string, path string, headers map[string]string, statusCode int, duration time.Duration, err *Error) {
	if c.logger == nil {
		return
	}
	fields := []zap.Field{
		zap.String("service", c.name),
		zap.String("method", method),
		zap.String("url_path", safeLogPath(path)),
		zap.Int("status_code", statusCode),
		zap.Duration("duration", duration),
		zap.String("trace_id", firstTraceID(headers)),
	}
	if err != nil {
		fields = append(fields,
			zap.String("error_kind", string(err.Kind)),
			zap.String("error_code", err.Code),
		)
		c.logger.Warn("http client request failed", fields...)
		return
	}
	fields = append(fields, zap.String("error_kind", ""))
	c.logger.Info("http client request completed", fields...)
}

func (c *Client) closeIdleConnections() {
	if c == nil || c.resty == nil || c.resty.GetClient() == nil {
		return
	}
	c.resty.GetClient().CloseIdleConnections()
}

func responseMessage(resp *resty.Response) string {
	if resp == nil {
		return ""
	}
	body := strings.TrimSpace(string(resp.Body()))
	if body == "" {
		return resp.Status()
	}
	const maxMessageLength = 512
	if len(body) > maxMessageLength {
		return body[:maxMessageLength]
	}
	return body
}

func executionErrorCode(kind ErrorKind, err error) string {
	if kind == ErrorKindRequest {
		if errors.Is(err, context.DeadlineExceeded) {
			return "timeout"
		}
		if errors.Is(err, context.Canceled) {
			return "canceled"
		}
		return "request_error"
	}
	return "response_parse_error"
}

func hasHeader(headers map[string]string, key string) bool {
	for header := range headers {
		if strings.EqualFold(header, key) {
			return true
		}
	}
	return false
}
