package integration_test

import (
	"context"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/require"
	userinfra "github.com/teamsillybees/initra/examples/basic/internal/app/user/infra"
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

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT sys_user.id AS "id", sys_user.username AS "username", sys_user.password_hash AS "password_hash", sys_user.nickname AS "nickname", sys_user.phone AS "phone", sys_user.email AS "email", sys_user.avatar_url AS "avatar_url", sys_user.is_super_admin AS "is_super_admin", sys_user.is_enable AS "is_enable", sys_user.sort_id AS "sort_id", sys_user.created_at AS "created_at", sys_user.updated_at AS "updated_at", sys_user.deleted_at AS "deleted_at", sys_user.created_by AS "created_by", sys_user.updated_by AS "updated_by" FROM public.sys_user WHERE ( (sys_user.id = $1::bigint) AND sys_user.deleted_at IS NULL ) LIMIT $2;`)).
		WithArgs(int64(1001), 1).
		WillReturnRows(rows)

	roleRows := sqlmock.NewRows([]string{"code"}).
		AddRow("admin").
		AddRow("viewer")
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT sys_role.code AS "code" FROM public.sys_user_role INNER JOIN public.sys_role ON (sys_role.id = sys_user_role.role_id) WHERE ( (sys_user_role.user_id = $1::bigint) AND sys_user_role.deleted_at IS NULL AND sys_role.deleted_at IS NULL AND (sys_role.is_enable = $2::boolean) ) ORDER BY sys_role.sort_id ASC, sys_role.id ASC;`)).
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
