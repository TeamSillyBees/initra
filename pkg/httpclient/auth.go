package httpclient

import (
	"context"

	"github.com/go-resty/resty/v2"
)

// AuthHandler 将认证配置应用到单次 resty 请求。
type AuthHandler interface {
	Apply(ctx context.Context, r *resty.Request, cfg AuthConfig) error
}

type defaultAuthHandler struct{}

// Apply 根据认证类型把凭证写入当前请求，不记录或持久化凭证内容。
func (h defaultAuthHandler) Apply(_ context.Context, r *resty.Request, cfg AuthConfig) error {
	switch cfg.Type {
	case "", AuthTypeNone:
		return nil
	case AuthTypeBearer:
		r.SetAuthScheme("Bearer").SetAuthToken(cfg.Token)
	case AuthTypeBasic:
		r.SetBasicAuth(cfg.Username, cfg.Password)
	case AuthTypeAPIKey:
		r.SetHeader(cfg.Header, cfg.Value)
	case AuthTypeCustomHeader:
		if cfg.Header != "" {
			r.SetHeader(cfg.Header, cfg.Value)
		}
		for key, value := range cfg.Headers {
			r.SetHeader(key, value)
		}
	default:
		return &Error{
			Kind:    ErrorKindInternal,
			Code:    "unsupported_auth_type",
			Message: "不支持的 HTTP Client 认证类型",
		}
	}
	return nil
}
