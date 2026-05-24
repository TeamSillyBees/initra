package pagination

import apperrors "github.com/teamsillybees/initra/pkg/errors"

const (
	// DefaultPage 是未传 page 时使用的默认页码。
	DefaultPage = 1
	// DefaultPageSize 是未传 pageSize 或 limit 时使用的默认每页数量。
	DefaultPageSize = 20
	// DefaultMaxPageSize 是默认允许的最大每页数量。
	DefaultMaxPageSize = 100
)

// Options 描述分页参数归一化时使用的默认值和上限。
type Options struct {
	DefaultPage     int
	DefaultPageSize int
	MaxPageSize     int
}

// Option 调整分页参数归一化选项。
type Option func(*Options)

// WithDefaultPage 设置默认页码。
func WithDefaultPage(page int) Option {
	return func(opts *Options) {
		opts.DefaultPage = page
	}
}

// WithDefaultPageSize 设置默认每页数量。
func WithDefaultPageSize(pageSize int) Option {
	return func(opts *Options) {
		opts.DefaultPageSize = pageSize
	}
}

// WithMaxPageSize 设置最大每页数量。
func WithMaxPageSize(maxPageSize int) Option {
	return func(opts *Options) {
		opts.MaxPageSize = maxPageSize
	}
}

// PageQuery 描述 page/pageSize HTTP 查询参数，可嵌入 Huma request 结构体。
type PageQuery struct {
	Page     int `query:"page" example:"1" doc:"页码"`
	PageSize int `query:"pageSize" example:"20" doc:"每页数量"`
}

// Normalize 校验并补齐 page/pageSize 分页参数。
func (q PageQuery) Normalize(opts ...Option) (PageDTO, error) {
	normalized := normalizeOptions(opts...)

	page := q.Page
	if page == 0 {
		page = normalized.DefaultPage
	}
	if page < 0 {
		return PageDTO{}, invalidParam("page", page, "page must be greater than 0")
	}

	pageSize := q.PageSize
	if pageSize == 0 {
		pageSize = normalized.DefaultPageSize
	}
	if pageSize < 0 {
		return PageDTO{}, invalidParam("pageSize", pageSize, "pageSize must be greater than 0")
	}
	if pageSize > normalized.MaxPageSize {
		return PageDTO{}, invalidParam("pageSize", pageSize, "pageSize exceeds max page size")
	}

	return PageDTO{
		Page:     page,
		PageSize: pageSize,
	}, nil
}

// Validate 校验 page/pageSize 分页参数。
func (q PageQuery) Validate(opts ...Option) error {
	_, err := q.Normalize(opts...)
	return err
}

// PageDTO 是业务层和仓储层可复用的 page/pageSize 分页参数。
type PageDTO struct {
	Page     int
	PageSize int
}

// Normalize 校验并补齐 page/pageSize 分页参数。
func (p PageDTO) Normalize(opts ...Option) (PageDTO, error) {
	return PageQuery{
		Page:     p.Page,
		PageSize: p.PageSize,
	}.Normalize(opts...)
}

// Validate 校验 page/pageSize 分页参数。
func (p PageDTO) Validate(opts ...Option) error {
	_, err := p.Normalize(opts...)
	return err
}

// Offset 返回可直接用于 SQL OFFSET 的偏移量。
func (p PageDTO) Offset() int64 {
	return int64(p.Page-1) * int64(p.PageSize)
}

// Limit 返回可直接用于 SQL LIMIT 的数量。
func (p PageDTO) Limit() int64 {
	return int64(p.PageSize)
}

// PageResult 描述 service/repo 层通用分页查询结果。
type PageResult[T any] struct {
	Page  PageDTO
	Total int64
	Items []T
}

// OffsetQuery 描述 offset/limit HTTP 查询参数，可嵌入 Huma request 结构体。
type OffsetQuery struct {
	Offset int `query:"offset" example:"0" doc:"偏移量"`
	Limit  int `query:"limit" example:"20" doc:"返回数量"`
}

// Normalize 校验并补齐 offset/limit 分页参数。
func (q OffsetQuery) Normalize(opts ...Option) (OffsetDTO, error) {
	normalized := normalizeOptions(opts...)

	if q.Offset < 0 {
		return OffsetDTO{}, invalidParam("offset", q.Offset, "offset must be greater than or equal to 0")
	}

	limit := q.Limit
	if limit == 0 {
		limit = normalized.DefaultPageSize
	}
	if limit < 0 {
		return OffsetDTO{}, invalidParam("limit", limit, "limit must be greater than 0")
	}
	if limit > normalized.MaxPageSize {
		return OffsetDTO{}, invalidParam("limit", limit, "limit exceeds max page size")
	}

	return OffsetDTO{
		Offset: q.Offset,
		Limit:  limit,
	}, nil
}

// Validate 校验 offset/limit 分页参数。
func (q OffsetQuery) Validate(opts ...Option) error {
	_, err := q.Normalize(opts...)
	return err
}

// OffsetDTO 是业务层和仓储层可复用的 offset/limit 分页参数。
type OffsetDTO struct {
	Offset int
	Limit  int
}

// Page 返回 offset/limit 对应的当前页码。
func (p OffsetDTO) Page() int {
	if p.Limit <= 0 {
		return DefaultPage
	}
	return p.Offset/p.Limit + 1
}

// PageDTO 返回 offset/limit 等价的 page/pageSize 参数。
func (p OffsetDTO) PageDTO() PageDTO {
	return PageDTO{
		Page:     p.Page(),
		PageSize: p.Limit,
	}
}

// CursorQuery 描述 cursor pagination HTTP 查询参数，可嵌入 Huma request 结构体。
type CursorQuery struct {
	Cursor string `query:"cursor" example:"eyJpZCI6MTAwMX0" doc:"游标"`
	Limit  int    `query:"limit" example:"20" doc:"返回数量"`
}

// Normalize 校验并补齐 cursor pagination 参数。
func (q CursorQuery) Normalize(opts ...Option) (CursorDTO, error) {
	normalized := normalizeOptions(opts...)

	limit := q.Limit
	if limit == 0 {
		limit = normalized.DefaultPageSize
	}
	if limit < 0 {
		return CursorDTO{}, invalidParam("limit", limit, "limit must be greater than 0")
	}
	if limit > normalized.MaxPageSize {
		return CursorDTO{}, invalidParam("limit", limit, "limit exceeds max page size")
	}

	return CursorDTO{
		Cursor: q.Cursor,
		Limit:  limit,
	}, nil
}

// Validate 校验 cursor pagination 参数。
func (q CursorQuery) Validate(opts ...Option) error {
	_, err := q.Normalize(opts...)
	return err
}

// CursorDTO 是业务层和仓储层可复用的 cursor pagination 参数。
type CursorDTO struct {
	Cursor string
	Limit  int
}

// PageMetaVO 描述带总数分页响应中的通用 JSON 字段，可嵌入业务 VO。
type PageMetaVO struct {
	Total      int64 `json:"total"`
	Page       int   `json:"page"`
	PageSize   int   `json:"pageSize"`
	TotalPages int   `json:"totalPages"`
}

// PageVO 描述带总数的通用分页 JSON DTO。
type PageVO[T any] struct {
	Items []T `json:"items"`
	PageMetaVO
}

// NewPageMetaVO 创建 page/pageSize 分页输出元信息。
func NewPageMetaVO(total int64, params PageDTO) PageMetaVO {
	return newPageMetaVO(total, params.Page, params.PageSize)
}

// NewPageVO 创建 page/pageSize 分页输出。
func NewPageVO[T any](items []T, total int64, params PageDTO) PageVO[T] {
	return PageVO[T]{
		Items:      normalizeItems(items),
		PageMetaVO: NewPageMetaVO(total, params),
	}
}

// NewOffsetVO 创建 offset/limit 分页输出，并按 offset/limit 推导当前页。
func NewOffsetVO[T any](items []T, total int64, params OffsetDTO) PageVO[T] {
	return NewPageVO(items, total, params.PageDTO())
}

// CursorMetaVO 描述 cursor pagination 响应中的通用 JSON 字段。
type CursorMetaVO struct {
	NextCursor string `json:"nextCursor,omitempty"`
	HasMore    bool   `json:"hasMore"`
	Limit      int    `json:"limit"`
}

// CursorVO 描述 cursor pagination 通用 JSON DTO。
type CursorVO[T any] struct {
	Items []T `json:"items"`
	CursorMetaVO
}

// NewCursorVO 创建 cursor pagination 输出。
func NewCursorVO[T any](items []T, nextCursor string, hasMore bool, params CursorDTO) CursorVO[T] {
	return CursorVO[T]{
		Items: normalizeItems(items),
		CursorMetaVO: CursorMetaVO{
			NextCursor: nextCursor,
			HasMore:    hasMore,
			Limit:      params.Limit,
		},
	}
}

func normalizeOptions(opts ...Option) Options {
	normalized := Options{
		DefaultPage:     DefaultPage,
		DefaultPageSize: DefaultPageSize,
		MaxPageSize:     DefaultMaxPageSize,
	}
	for _, opt := range opts {
		opt(&normalized)
	}
	if normalized.DefaultPage <= 0 {
		normalized.DefaultPage = DefaultPage
	}
	if normalized.DefaultPageSize <= 0 {
		normalized.DefaultPageSize = DefaultPageSize
	}
	if normalized.MaxPageSize <= 0 {
		normalized.MaxPageSize = DefaultMaxPageSize
	}
	if normalized.DefaultPageSize > normalized.MaxPageSize {
		normalized.DefaultPageSize = normalized.MaxPageSize
	}
	return normalized
}

func newPageMetaVO(total int64, page int, pageSize int) PageMetaVO {
	if total < 0 {
		total = 0
	}
	return PageMetaVO{
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages(total, pageSize),
	}
}

func totalPages(total int64, pageSize int) int {
	if total <= 0 || pageSize <= 0 {
		return 0
	}
	return int((total + int64(pageSize) - 1) / int64(pageSize))
}

func normalizeItems[T any](items []T) []T {
	if items == nil {
		return []T{}
	}
	return items
}

func invalidParam(field string, value any, message string) error {
	return apperrors.New(
		apperrors.CodeBadRequest,
		message,
		apperrors.WithDetails(map[string]any{
			"field": field,
			"value": value,
		}),
	)
}
