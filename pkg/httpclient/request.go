package httpclient

import "time"

// RequestOption 用于按请求覆盖 Header、查询参数、超时等设置。
type RequestOption func(*RequestOptions)

// RequestOptions 描述单次 HTTP 请求的可选参数。
type RequestOptions struct {
	Headers     map[string]string
	QueryParams map[string]string
	PathParams  map[string]string
	Timeout     time.Duration
	Result      any
	ContentType string
	Idempotent  bool
}

// WithHeader 设置单个请求 Header。
func WithHeader(key, value string) RequestOption {
	return func(opts *RequestOptions) {
		opts.Headers[key] = value
	}
}

// WithHeaders 批量设置请求 Header。
func WithHeaders(headers map[string]string) RequestOption {
	return func(opts *RequestOptions) {
		for key, value := range headers {
			opts.Headers[key] = value
		}
	}
}

// WithQuery 设置单个查询参数。
func WithQuery(key, value string) RequestOption {
	return func(opts *RequestOptions) {
		opts.QueryParams[key] = value
	}
}

// WithQueryParams 批量设置查询参数。
func WithQueryParams(params map[string]string) RequestOption {
	return func(opts *RequestOptions) {
		for key, value := range params {
			opts.QueryParams[key] = value
		}
	}
}

// WithPathParams 批量设置路径参数。
func WithPathParams(params map[string]string) RequestOption {
	return func(opts *RequestOptions) {
		for key, value := range params {
			opts.PathParams[key] = value
		}
	}
}

// WithTimeout 设置单次请求超时。
func WithTimeout(timeout time.Duration) RequestOption {
	return func(opts *RequestOptions) {
		opts.Timeout = timeout
	}
}

// WithResult 设置 JSON 响应反序列化目标。
func WithResult(v any) RequestOption {
	return func(opts *RequestOptions) {
		opts.Result = v
	}
}

// WithContentType 设置请求 Content-Type。
func WithContentType(contentType string) RequestOption {
	return func(opts *RequestOptions) {
		opts.ContentType = contentType
	}
}

// WithIdempotent 标记当前请求可安全重试。
func WithIdempotent(idempotent bool) RequestOption {
	return func(opts *RequestOptions) {
		opts.Idempotent = idempotent
	}
}

func newRequestOptions(options []RequestOption) RequestOptions {
	opts := RequestOptions{
		Headers:     map[string]string{},
		QueryParams: map[string]string{},
		PathParams:  map[string]string{},
	}
	for _, option := range options {
		if option != nil {
			option(&opts)
		}
	}
	return opts
}
