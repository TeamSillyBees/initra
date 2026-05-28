package pagination

import apperrors "github.com/teamsillybees/initra/pkg/errors"

const (
	// DefaultPage 是未传 page 时使用的默认页码。
	DefaultPage int32 = 1
	// DefaultPageSize 是未传 pageSize 或 limit 时使用的默认每页数量。
	DefaultPageSize int32 = 20
	// DefaultMaxPageSize 是默认允许的最大每页数量。
	DefaultMaxPageSize int32 = 100
)

// Options 描述分页参数归一化时使用的默认值和上限。
type Options struct {
	DefaultPage     int32
	DefaultPageSize int32
	MaxPageSize     int32
}

// Option 调整分页参数归一化选项。
type Option func(*Options)

// WithDefaultPage 设置默认页码。
func WithDefaultPage(page int32) Option {
	return func(opts *Options) {
		opts.DefaultPage = page
	}
}

// WithDefaultPageSize 设置默认每页数量。
func WithDefaultPageSize(pageSize int32) Option {
	return func(opts *Options) {
		opts.DefaultPageSize = pageSize
	}
}

// WithMaxPageSize 设置最大每页数量。
func WithMaxPageSize(maxPageSize int32) Option {
	return func(opts *Options) {
		opts.MaxPageSize = maxPageSize
	}
}

// PageQuery 描述 page/pageSize HTTP 查询参数，可嵌入 Huma request 结构体。
type PageQuery struct {
	Page     int32 `query:"page" example:"1" minimum:"1" doc:"当前页码，从 1 开始，不传默认为 1"`
	PageSize int32 `query:"pageSize" example:"20" minimum:"1" maximum:"100" doc:"每页记录数，不传默认为 20，最大 100"`
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
	Page     int32
	PageSize int32
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
	Offset int32 `query:"offset" example:"0" minimum:"0" doc:"偏移量，从 0 开始，不传默认为 0"`
	Limit  int32 `query:"limit" example:"20" minimum:"1" maximum:"100" doc:"返回记录数，不传默认为 20，最大 100"`
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
	Offset int32
	Limit  int32
}

// Page 返回 offset/limit 对应的当前页码。
func (p OffsetDTO) Page() int32 {
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
	Cursor string `query:"cursor" example:"eyJpZCI6MTAwMX0" doc:"分页游标，首次查询不传或传空字符串"`
	Limit  int32  `query:"limit" example:"20" minimum:"1" maximum:"100" doc:"返回记录数，不传默认为 20，最大 100"`
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
	Limit  int32
}

// PageMetaVO 描述带总数分页响应中的通用 JSON 字段，可嵌入业务 VO。
type PageMetaVO struct {
	Total      int64 `json:"total" doc:"总记录数"`
	Page       int32 `json:"page" doc:"当前页码"`
	PageSize   int32 `json:"pageSize" doc:"每页记录数"`
	TotalPages int32 `json:"totalPages" doc:"总页数"`
}

// PageVO 描述带总数的通用分页 JSON DTO。
type PageVO[T any] struct {
	Items []T `json:"items" doc:"当前页数据列表"`
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
	NextCursor string `json:"nextCursor,omitempty" doc:"下一页游标，为空字符串表示已是最后一页"`
	HasMore    bool   `json:"hasMore" doc:"是否还有更多数据"`
	Limit      int32  `json:"limit" doc:"每页记录数"`
}

// CursorVO 描述 cursor pagination 通用 JSON DTO。
type CursorVO[T any] struct {
	Items []T `json:"items" doc:"当前页数据列表"`
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

func newPageMetaVO(total int64, page int32, pageSize int32) PageMetaVO {
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

func totalPages(total int64, pageSize int32) int32 {
	if total <= 0 || pageSize <= 0 {
		return 0
	}
	return int32((total + int64(pageSize) - 1) / int64(pageSize))
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
