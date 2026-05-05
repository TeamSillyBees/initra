package integration_test

import (
	"context"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/require"
	auth "github.com/teamsillybees/initra/examples/api/internal/module/auth"
)

// TestAuthRepositoryFindByUsername 验证 auth 仓储能按用户名读取身份并补齐角色编码。
func TestAuthRepositoryFindByUsername(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := auth.NewRepository(db)

	rows := sqlmock.NewRows([]string{
		"id",
		"username",
		"nickname",
		"password_hash",
		"is_super_admin",
		"is_enable",
	}).AddRow(
		1001,
		"alice",
		"Alice",
		"hashed:secret-123",
		true,
		true,
	)

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT sys_user.id AS "id", sys_user.username AS "username", sys_user.nickname AS "nickname", sys_user.password_hash AS "password_hash", sys_user.is_super_admin AS "is_super_admin", sys_user.is_enable AS "is_enable" FROM public.sys_user WHERE ( (sys_user.username = $1::text) AND sys_user.deleted_at IS NULL ) LIMIT $2;`)).
		WithArgs("alice", 1).
		WillReturnRows(rows)

	roleRows := sqlmock.NewRows([]string{"code"}).AddRow("admin")
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT sys_role.code AS "code" FROM public.sys_user_role INNER JOIN public.sys_role ON (sys_role.id = sys_user_role.role_id) WHERE ( (sys_user_role.user_id = $1::bigint) AND sys_user_role.deleted_at IS NULL AND sys_role.deleted_at IS NULL AND (sys_role.is_enable = $2::boolean) ) ORDER BY sys_role.sort_id ASC, sys_role.id ASC;`)).
		WithArgs(int64(1001), true).
		WillReturnRows(roleRows)

	identity, err := repo.FindByUsername(context.Background(), "alice")
	require.NoError(t, err)
	require.Equal(t, int64(1001), identity.UserID)
	require.Equal(t, "Alice", identity.Nickname)
	require.Equal(t, []string{"admin"}, identity.RoleCodes)
	require.True(t, identity.IsSuperAdmin)
	require.True(t, identity.IsEnable)
	require.NoError(t, mock.ExpectationsWereMet())
}
