package httpdemo

// GetHTTPBingoDTO 描述调用 HTTPBingo GET 示例所需的输入。
type GetHTTPBingoDTO struct {
	Message string
	TraceID string
}

// HTTPBingoGetResult 描述 HTTPBingo GET 示例返回的领域结果。
type HTTPBingoGetResult struct {
	Args    map[string][]string
	Headers map[string][]string
	Method  string
	Origin  string
	URL     string
}

// GetHTTPBingoFormPageDTO 描述获取 HTTPBingo 表单页所需的输入。
type GetHTTPBingoFormPageDTO struct {
	TraceID string
}

// HTTPBingoFormPage 描述 HTTPBingo 表单页响应。
type HTTPBingoFormPage struct {
	ContentType string
	Size        int
	Body        string
}
