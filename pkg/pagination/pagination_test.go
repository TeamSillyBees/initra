package pagination_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	apperrors "github.com/teamsillybees/initra/pkg/errors"
	"github.com/teamsillybees/initra/pkg/pagination"
)

func TestPageQueryNormalizeFillsDefaultsAndBuildsVO(t *testing.T) {
	params, err := pagination.PageQuery{}.Normalize()
	require.NoError(t, err)

	require.Equal(t, 1, params.Page)
	require.Equal(t, 20, params.PageSize)
	require.Equal(t, int64(0), params.Offset())
	require.Equal(t, int64(20), params.Limit())

	meta := pagination.NewPageMetaVO(42, params)
	require.Equal(t, int64(42), meta.Total)
	require.Equal(t, 1, meta.Page)
	require.Equal(t, 20, meta.PageSize)
	require.Equal(t, 3, meta.TotalPages)

	result := pagination.NewPageVO([]string{"alice"}, 42, params)
	require.Equal(t, []string{"alice"}, result.Items)
	require.Equal(t, int64(42), result.Total)
	require.Equal(t, 1, result.Page)
	require.Equal(t, 20, result.PageSize)
	require.Equal(t, 3, result.TotalPages)
}

func TestPageQueryNormalizeUsesOptions(t *testing.T) {
	params, err := pagination.PageQuery{}.Normalize(
		pagination.WithDefaultPage(2),
		pagination.WithDefaultPageSize(10),
		pagination.WithMaxPageSize(50),
	)
	require.NoError(t, err)

	require.Equal(t, 2, params.Page)
	require.Equal(t, 10, params.PageSize)
	require.Equal(t, int64(10), params.Offset())
	require.Equal(t, int64(10), params.Limit())
}

func TestPageQueryNormalizeRejectsInvalidValues(t *testing.T) {
	tests := []struct {
		name  string
		query pagination.PageQuery
	}{
		{name: "negative page", query: pagination.PageQuery{Page: -1}},
		{name: "negative page size", query: pagination.PageQuery{PageSize: -1}},
		{name: "page size exceeds max", query: pagination.PageQuery{PageSize: 101}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := tt.query.Normalize()
			require.Error(t, err)
			require.Equal(t, apperrors.CodeBadRequest, apperrors.From(err).Code)
		})
	}
}

func TestOffsetQueryNormalizeBuildsEquivalentPageVO(t *testing.T) {
	params, err := pagination.OffsetQuery{Offset: 40, Limit: 20}.Normalize()
	require.NoError(t, err)

	require.Equal(t, 40, params.Offset)
	require.Equal(t, 20, params.Limit)

	result := pagination.NewOffsetVO([]int{1, 2}, 95, params)
	require.Equal(t, []int{1, 2}, result.Items)
	require.Equal(t, int64(95), result.Total)
	require.Equal(t, 3, result.Page)
	require.Equal(t, 20, result.PageSize)
	require.Equal(t, 5, result.TotalPages)
}

func TestOffsetQueryNormalizeRejectsInvalidValues(t *testing.T) {
	tests := []struct {
		name  string
		query pagination.OffsetQuery
	}{
		{name: "negative offset", query: pagination.OffsetQuery{Offset: -1}},
		{name: "negative limit", query: pagination.OffsetQuery{Limit: -1}},
		{name: "limit exceeds max", query: pagination.OffsetQuery{Limit: 101}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := tt.query.Normalize()
			require.Error(t, err)
			require.Equal(t, apperrors.CodeBadRequest, apperrors.From(err).Code)
		})
	}
}

func TestCursorQueryNormalizeAndVO(t *testing.T) {
	params, err := pagination.CursorQuery{Cursor: "next-id"}.Normalize()
	require.NoError(t, err)

	require.Equal(t, "next-id", params.Cursor)
	require.Equal(t, 20, params.Limit)

	result := pagination.NewCursorVO([]string{"alice"}, "after-alice", true, params)
	require.Equal(t, []string{"alice"}, result.Items)
	require.Equal(t, "after-alice", result.NextCursor)
	require.True(t, result.HasMore)
	require.Equal(t, 20, result.Limit)
}

func TestCursorQueryNormalizeRejectsInvalidLimit(t *testing.T) {
	tests := []struct {
		name  string
		query pagination.CursorQuery
	}{
		{name: "negative limit", query: pagination.CursorQuery{Limit: -1}},
		{name: "limit exceeds max", query: pagination.CursorQuery{Limit: 101}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := tt.query.Normalize()
			require.Error(t, err)
			require.Equal(t, apperrors.CodeBadRequest, apperrors.From(err).Code)
		})
	}
}
