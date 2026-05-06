package requestctx

import (
	"net/http"
	"strings"
)

const (
	defaultFormMemory = 32 << 20

	headerAcceptLanguage  = "Accept-Language"
	headerForwarded       = "Forwarded"
	headerForwardedHost   = "X-Forwarded-Host"
	headerForwardedProto  = "X-Forwarded-Proto"
	headerOriginalForward = "X-Forwarded-Scheme"
	headerOrigin          = "Origin"
)

// QueryParam 从请求 URL 中提取 query 参数。
func QueryParam(r *http.Request, name string) string {
	if r == nil || r.URL == nil {
		return ""
	}
	return r.URL.Query().Get(name)
}

// PathParam 从请求路径参数中提取值。
func PathParam(r *http.Request, name string) string {
	if r == nil {
		return ""
	}
	return r.PathValue(name)
}

// HeaderValue 从请求头中提取值。
func HeaderValue(r *http.Request, name string) string {
	if r == nil {
		return ""
	}
	return r.Header.Get(name)
}

// CookieValue 从请求 Cookie 中提取值。
func CookieValue(r *http.Request, name string) string {
	if r == nil {
		return ""
	}
	cookie, err := r.Cookie(name)
	if err != nil {
		return ""
	}
	return cookie.Value
}

// FormParam 从请求表单中提取值。
func FormParam(r *http.Request, name string) string {
	if r == nil {
		return ""
	}
	if r.Form == nil && r.PostForm == nil && r.MultipartForm == nil {
		_ = r.ParseMultipartForm(defaultFormMemory)
	}
	if r.PostForm != nil {
		if value := r.PostForm.Get(name); value != "" {
			return value
		}
	}
	if r.MultipartForm != nil && r.MultipartForm.Value != nil {
		if values := r.MultipartForm.Value[name]; len(values) > 0 {
			return values[0]
		}
	}
	return ""
}

// AcceptLanguage 从 Accept-Language 请求头中提取语言偏好。
func AcceptLanguage(r *http.Request) string {
	return HeaderValue(r, headerAcceptLanguage)
}

// Language 优先从自定义语言头提取语言，未命中时回退到 Accept-Language。
func Language(r *http.Request, headerNames ...string) string {
	for _, name := range headerNames {
		if value := strings.TrimSpace(HeaderValue(r, name)); value != "" {
			return value
		}
	}
	return AcceptLanguage(r)
}

// Origin 从 Origin 请求头中提取请求来源。
func Origin(r *http.Request) string {
	return HeaderValue(r, headerOrigin)
}

// Host 提取请求 Host，优先使用反向代理传入的 Host 信息。
func Host(r *http.Request) string {
	if r == nil {
		return ""
	}
	if value := firstForwardedParam(r.Header.Values(headerForwarded), "host"); value != "" {
		return value
	}
	if value := firstCSV(HeaderValue(r, headerForwardedHost)); value != "" {
		return value
	}
	if r.Host != "" {
		return r.Host
	}
	if r.URL != nil {
		return r.URL.Host
	}
	return ""
}

// Scheme 提取请求协议，优先使用反向代理传入的协议信息。
func Scheme(r *http.Request) string {
	if r == nil {
		return ""
	}
	if r.URL != nil && r.URL.Scheme != "" {
		return strings.ToLower(r.URL.Scheme)
	}
	if value := firstForwardedParam(r.Header.Values(headerForwarded), "proto"); value != "" {
		return strings.ToLower(value)
	}
	if value := firstCSV(HeaderValue(r, headerForwardedProto)); value != "" {
		return strings.ToLower(value)
	}
	if value := firstCSV(HeaderValue(r, headerOriginalForward)); value != "" {
		return strings.ToLower(value)
	}
	if r.TLS != nil {
		return "https"
	}
	return "http"
}

// BaseURL 根据请求协议和 Host 构造基础 URL。
func BaseURL(r *http.Request) string {
	host := Host(r)
	if host == "" {
		return ""
	}
	return Scheme(r) + "://" + host
}
