package integration_test

import (
	"context"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/require"
	authinfra "github.com/teamsillybees/initra/internal/app/auth/infra"
)

// TestAuthRepositoryFindByUsername 验证 auth 仓储能按用户名读取身份并补齐角色编码。
func TestAuthRepositoryFindByUsername(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := authinfra.NewRepository(db)

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

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT su.id AS "id", su.username AS "username", su.nickname AS "nickname", su.password_hash AS "password_hash", su.is_super_admin AS "is_super_admin", su.is_enable AS "is_enable" FROM public.sys_user AS su WHERE ( (su.username = $1::text) AND su.deleted_at IS NULL ) LIMIT $2;`)).
		WithArgs("alice", 1).
		WillReturnRows(rows)

	roleRows := sqlmock.NewRows([]string{"code"}).AddRow("admin")
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT sr.code AS "code" FROM public.sys_user_role AS sur INNER JOIN public.sys_role AS sr ON (sr.id = sur.role_id) WHERE ( (sur.user_id = $1::bigint) AND sur.deleted_at IS NULL AND sr.deleted_at IS NULL AND (sr.is_enable = $2::boolean) ) ORDER BY sr.sort_id ASC, sr.id ASC;`)).
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
