package httpdemo

import (
	"context"

	"github.com/teamsillybees/initra/pkg/requestctx"
	"github.com/teamsillybees/initra/pkg/response"
)

// Handler 封装 httpdemo 示例模块的 HTTP 适配逻辑。
type Handler struct {
	service *Service
}

// NewHandler 创建 httpdemo 示例模块 HTTP Handler。
func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) get(ctx context.Context, input *getHTTPBingoRequest) (*getHTTPBingoResponse, error) {
	result, err := h.service.GetHTTPBingo(ctx, GetHTTPBingoDTO{
		Message: input.Message,
		TraceID: requestctx.TraceIDFromContext(ctx),
	})
	if err != nil {
		return nil, err
	}
	return &getHTTPBingoResponse{
		Body: response.OK(requestctx.TraceIDFromContext(ctx), toHTTPBingoGetVO(result)),
	}, nil
}

func (h *Handler) formPage(ctx context.Context, input *getHTTPBingoFormPageRequest) (*getHTTPBingoFormPageResponse, error) {
	result, err := h.service.GetHTTPBingoFormPage(ctx, GetHTTPBingoFormPageDTO{
		TraceID: requestctx.TraceIDFromContext(ctx),
	})
	if err != nil {
		return nil, err
	}
	return &getHTTPBingoFormPageResponse{
		Body: response.OK(requestctx.TraceIDFromContext(ctx), toHTTPBingoFormPageVO(result)),
	}, nil
}

func toHTTPBingoGetVO(result *HTTPBingoGetResult) HTTPBingoGetVO {
	if result == nil {
		return HTTPBingoGetVO{}
	}
	return HTTPBingoGetVO{
		Args:    result.Args,
		Headers: result.Headers,
		Method:  result.Method,
		Origin:  result.Origin,
		URL:     result.URL,
	}
}

func toHTTPBingoFormPageVO(result *HTTPBingoFormPage) HTTPBingoFormPageVO {
	if result == nil {
		return HTTPBingoFormPageVO{}
	}
	return HTTPBingoFormPageVO{
		ContentType: result.ContentType,
		Size:        result.Size,
		Body:        result.Body,
	}
}
