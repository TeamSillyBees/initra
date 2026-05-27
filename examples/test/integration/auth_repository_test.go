package integration_test

import (
	"context"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/require"
	"github.com/teamsillybees/initra/examples/internal/modules/auth"
	"github.com/teamsillybees/initra/pkg/idgen"
)

// TestAuthRepositoryFindByUsername 验证 auth 仓储能通过 Ent 按用户名读取身份并补齐角色编码。
func TestAuthRepositoryFindByUsername(t *testing.T) {
	db, mock, client := newMockEntClient(t)
	defer db.Close()

	repo := auth.NewRepository(client)
	mock.ExpectQuery(`SELECT .*FROM "sys_user".*`).
		WillReturnRows(sysUserRows().AddRow(
			int64(1001), nil, testNow, testNow, int64(9001), int64(9001),
			"alice", "hashed:secret-123", "Alice", "13800000000", "alice@example.com",
			nil, true, true, 1,
		))
	mock.ExpectQuery(`SELECT .*FROM "sys_role".*`).
		WillReturnRows(sqlmock.NewRows([]string{"code"}).AddRow("admin"))

	identity, err := repo.FindByUsername(context.Background(), " alice ")

	require.NoError(t, err)
	require.Equal(t, idgen.New(1001), identity.UserID)
	require.Equal(t, "alice", identity.Username)
	require.Equal(t, "Alice", identity.Nickname)
	require.Equal(t, []string{"admin"}, identity.RoleCodes)
	require.True(t, identity.IsSuperAdmin)
	require.True(t, identity.IsEnable)
	require.NoError(t, mock.ExpectationsWereMet())
}
