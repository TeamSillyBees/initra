package httpclient

import (
	"net/http"

	"github.com/go-resty/resty/v2"
)

// Response 描述远程 HTTP 调用的统一响应。
type Response struct {
	StatusCode int
	Header     http.Header
	Body       []byte
	Result     any
}

// String 返回响应体字符串。
func (r *Response) String() string {
	if r == nil {
		return ""
	}
	return string(r.Body)
}

func newResponse(resp *resty.Response, result any) *Response {
	if resp == nil {
		return nil
	}
	body := resp.Body()
	copiedBody := make([]byte, len(body))
	copy(copiedBody, body)
	return &Response{
		StatusCode: resp.StatusCode(),
		Header:     resp.Header(),
		Body:       copiedBody,
		Result:     result,
	}
}
