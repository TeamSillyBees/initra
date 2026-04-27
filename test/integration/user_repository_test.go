package integration_test

import (
	"context"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/require"
	userinfra "github.com/teamsillybees/initra/internal/app/user/infra"
)

// staticRoleRelationIDGenerator 为用户角色关系提供稳定 ID，方便 SQL 断言。
type staticRoleRelationIDGenerator struct{}

// NextID 返回固定 ID。
func (staticRoleRelationIDGenerator) NextID() int64 {
	return 2001
}

// TestUserRepositoryFindByID 验证 user 仓储能按 ID 查询用户并补齐角色编码。
func TestUserRepositoryFindByID(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := userinfra.NewRepository(db, staticRoleRelationIDGenerator{})
	now := time.Date(2026, 4, 21, 10, 0, 0, 0, time.UTC)

	rows := sqlmock.NewRows([]string{
		"id",
		"username",
		"password_hash",
		"nickname",
		"phone",
		"email",
		"avatar_url",
		"is_super_admin",
		"is_enable",
		"sort_id",
		"created_at",
		"updated_at",
		"deleted_at",
		"created_by",
		"updated_by",
	}).AddRow(
		1001,
		"alice",
		"hashed:secret-123",
		"Alice",
		"13800000000",
		"alice@example.com",
		"https://example.com/avatar.png",
		true,
		true,
		1,
		now,
		now,
		nil,
		9001,
		9001,
	)

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT su.id AS "id", su.username AS "username", su.password_hash AS "password_hash", su.nickname AS "nickname", su.phone AS "phone", su.email AS "email", su.avatar_url AS "avatar_url", su.is_super_admin AS "is_super_admin", su.is_enable AS "is_enable", su.sort_id AS "sort_id", su.created_at AS "created_at", su.updated_at AS "updated_at", su.deleted_at AS "deleted_at", su.created_by AS "created_by", su.updated_by AS "updated_by" FROM public.sys_user AS su WHERE ( (su.id = $1::bigint) AND su.deleted_at IS NULL ) LIMIT $2;`)).
		WithArgs(int64(1001), 1).
		WillReturnRows(rows)

	roleRows := sqlmock.NewRows([]string{"code"}).
		AddRow("admin").
		AddRow("viewer")
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT sr.code AS "code" FROM public.sys_user_role AS sur INNER JOIN public.sys_role AS sr ON (sr.id = sur.role_id) WHERE ( (sur.user_id = $1::bigint) AND sur.deleted_at IS NULL AND sr.deleted_at IS NULL AND (sr.is_enable = $2::boolean) ) ORDER BY sr.sort_id ASC, sr.id ASC;`)).
		WithArgs(int64(1001), true).
		WillReturnRows(roleRows)

	user, err := repo.FindByID(context.Background(), 1001)
	require.NoError(t, err)
	require.Equal(t, int64(1001), user.ID)
	require.Equal(t, "Alice", user.Nickname)
	require.Equal(t, "13800000000", user.Phone)
	require.Equal(t, []string{"admin", "viewer"}, user.RoleCodes)
	require.True(t, user.IsSuperAdmin)
	require.True(t, user.IsEnable)
	require.NoError(t, mock.ExpectationsWereMet())
}
