package httpclient

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/teamsillybees/initra/pkg/logx"
)

const jsonContentType = "application/json"

// Client 封装单个远程服务的 HTTP 调用能力。
type Client struct {
	name        string
	config      ServiceConfig
	resty       *resty.Client
	logger      *logx.Logger
	authHandler AuthHandler
}

// Getter 表示 HTTP GET 调用能力。
type Getter interface {
	Get(ctx context.Context, path string, opts ...RequestOption) (*Response, error)
}

// Poster 表示 HTTP POST 调用能力。
type Poster interface {
	Post(ctx context.Context, path string, body any, opts ...RequestOption) (*Response, error)
}

// Putter 表示 HTTP PUT 调用能力。
type Putter interface {
	Put(ctx context.Context, path string, body any, opts ...RequestOption) (*Response, error)
}

// Patcher 表示 HTTP PATCH 调用能力。
type Patcher interface {
	Patch(ctx context.Context, path string, body any, opts ...RequestOption) (*Response, error)
}

// Deleter 表示 HTTP DELETE 调用能力。
type Deleter interface {
	Delete(ctx context.Context, path string, opts ...RequestOption) (*Response, error)
}

// JSONGetter 表示 GET 并解析 JSON 响应的能力。
type JSONGetter interface {
	GetJSON(ctx context.Context, path string, result any, opts ...RequestOption) error
}

// JSONPoster 表示 POST JSON 请求并解析 JSON 响应的能力。
type JSONPoster interface {
	PostJSON(ctx context.Context, path string, body any, result any, opts ...RequestOption) error
}

// BytesGetter 表示 GET 并返回原始响应体的能力。
type BytesGetter interface {
	GetBytes(ctx context.Context, path string, opts ...RequestOption) ([]byte, *Response, error)
}

// ReadCaller 表示常见读取类远程调用能力。
type ReadCaller interface {
	JSONGetter
	BytesGetter
}

// Caller 表示常见 HTTP 远程调用能力。
type Caller interface {
	Getter
	Poster
	Putter
	Patcher
	Deleter
	JSONGetter
	JSONPoster
	BytesGetter
}

func newClient(name string, global Config, cfg ServiceConfig, logger *logx.Logger) (*Client, error) {
	if logger == nil {
		logger = logx.NewNop()
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

// GetProperty 返回当前服务配置中的自定义属性值。
func (c *Client) GetProperty(key string) (string, bool) {
	if c == nil || c.config.Properties == nil {
		return "", false
	}
	value, ok := c.config.Properties[key]
	return value, ok
}

// Properties 返回当前服务配置中的自定义属性副本。
func (c *Client) Properties() map[string]string {
	if c == nil || c.config.Properties == nil {
		return map[string]string{}
	}
	properties := make(map[string]string, len(c.config.Properties))
	for key, value := range c.config.Properties {
		properties[key] = value
	}
	return properties
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

// GetJSON 发送 GET 请求并将 JSON 响应反序列化到 result。
func (c *Client) GetJSON(ctx context.Context, path string, result any, opts ...RequestOption) error {
	options := append([]RequestOption{WithResult(result)}, opts...)
	_, err := c.Get(ctx, path, options...)
	return err
}

// PostJSON 发送 POST 请求并将 JSON 响应反序列化到 result。
func (c *Client) PostJSON(ctx context.Context, path string, body any, result any, opts ...RequestOption) error {
	options := append([]RequestOption{WithResult(result)}, opts...)
	_, err := c.Post(ctx, path, body, options...)
	return err
}

// GetBytes 发送 GET 请求并返回原始响应体副本与响应元数据。
func (c *Client) GetBytes(ctx context.Context, path string, opts ...RequestOption) ([]byte, *Response, error) {
	resp, err := c.Get(ctx, path, opts...)
	if err != nil {
		return nil, nil, err
	}
	body := make([]byte, len(resp.Body))
	copy(body, resp.Body)
	return body, resp, nil
}

// GetJSON 发送 GET 请求并将 JSON 响应反序列化为 T。
func GetJSON[T any](ctx context.Context, c *Client, path string, opts ...RequestOption) (*T, error) {
	var result T
	if err := c.GetJSON(ctx, path, &result, opts...); err != nil {
		return nil, err
	}
	return &result, nil
}

// PostJSON 发送 POST 请求并将 JSON 响应反序列化为 T。
func PostJSON[T any](ctx context.Context, c *Client, path string, body any, opts ...RequestOption) (*T, error) {
	var result T
	if err := c.PostJSON(ctx, path, body, &result, opts...); err != nil {
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
		c.logRequest(requestCtx, method, path, options.Headers, 0, time.Duration(0), httpErr)
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
		c.logRequest(requestCtx, method, path, options.Headers, statusCode, duration, httpErr)
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
		c.logRequest(requestCtx, method, path, options.Headers, statusCode, duration, httpErr)
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
		c.logRequest(requestCtx, method, path, options.Headers, statusCode, duration, httpErr)
		return nil, httpErr
	}

	result := options.Result
	if result == nil {
		result = resp.Result()
	}
	wrapped := newResponse(resp, result)
	c.logRequest(requestCtx, method, path, options.Headers, statusCode, duration, nil)
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

func (c *Client) logRequest(ctx context.Context, method string, path string, headers map[string]string, statusCode int, duration time.Duration, err *Error) {
	if c.logger == nil {
		return
	}
	fields := []logx.Field{
		logx.String("service", c.name),
		logx.String("method", method),
		logx.String("url_path", safeLogPath(path)),
		logx.Int("status_code", statusCode),
		logx.Duration("duration", duration),
		logx.String("trace_id", firstTraceID(headers)),
	}
	if err != nil {
		fields = append(fields,
			logx.String("error_kind", string(err.Kind)),
			logx.String("error_code", err.Code),
		)
		c.logger.Error(ctx, "http client request failed", err, fields...)
		return
	}
	fields = append(fields, logx.String("error_kind", ""))
	c.logger.Info(ctx, "http client request completed", fields...)
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
