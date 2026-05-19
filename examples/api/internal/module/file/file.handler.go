package file

import (
	"context"
	"mime"

	"github.com/teamsillybees/initra/examples/api/internal/module/bizerrors"
	"github.com/teamsillybees/initra/pkg/requestctx"
	"github.com/teamsillybees/initra/pkg/response"
)

// Handler 封装 file 示例模块的 HTTP 适配逻辑。
type Handler struct {
	service *Service
}

// NewHandler 创建 file 示例模块 HTTP Handler。
func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) upload(ctx context.Context, input *uploadLocalFileRequest) (*uploadLocalFileResponse, error) {
	if input.RawBody.Form != nil {
		defer input.RawBody.Form.RemoveAll()
	}
	data := input.RawBody.Data()
	if data == nil || !data.File.IsSet {
		return nil, bizerrors.BadRequest("file is required")
	}
	defer data.File.Close()

	file, err := h.service.UploadLocal(ctx, UploadLocalFileDTO{
		FileName:    data.File.Filename,
		ContentType: data.File.ContentType,
		Size:        data.File.Size,
		Body:        data.File.File,
	})
	if err != nil {
		return nil, err
	}
	return &uploadLocalFileResponse{
		Body: response.OK(requestctx.TraceIDFromContext(ctx), toLocalFileVO(file)),
	}, nil
}

func (h *Handler) download(ctx context.Context, input *downloadLocalFileRequest) (*downloadLocalFileResponse, error) {
	result, err := h.service.DownloadLocal(ctx, input.Key)
	if err != nil {
		return nil, err
	}
	return &downloadLocalFileResponse{
		ContentType:        result.File.ContentType,
		ContentDisposition: attachmentDisposition(result.File.FileName),
		Body:               result.Body,
	}, nil
}

func (h *Handler) stat(ctx context.Context, input *statLocalFileRequest) (*statLocalFileResponse, error) {
	file, err := h.service.StatLocal(ctx, input.Key)
	if err != nil {
		return nil, err
	}
	return &statLocalFileResponse{
		Body: response.OK(requestctx.TraceIDFromContext(ctx), toLocalFileVO(file)),
	}, nil
}

func (h *Handler) delete(ctx context.Context, input *deleteLocalFileRequest) (*deleteLocalFileResponse, error) {
	if err := h.service.DeleteLocal(ctx, input.Key); err != nil {
		return nil, err
	}
	return &deleteLocalFileResponse{
		Body: response.OK(requestctx.TraceIDFromContext(ctx), map[string]any{}),
	}, nil
}

func toLocalFileVO(file *LocalFile) LocalFileVO {
	if file == nil {
		return LocalFileVO{}
	}
	return LocalFileVO{
		Key:          file.Key,
		FileName:     file.FileName,
		Size:         file.Size,
		ContentType:  file.ContentType,
		URL:          file.URL,
		LastModified: file.LastModified,
	}
}

func attachmentDisposition(fileName string) string {
	if fileName == "" {
		fileName = "download"
	}
	return mime.FormatMediaType("attachment", map[string]string{"filename": fileName})
}
