package integration_test

import (
	"context"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/require"
	"github.com/teamsillybees/initra/examples/internal/modules/auth"
	"github.com/teamsillybees/initra/pkg/idgen"
)

// TestAuthServiceFindByID 验证 auth 服务能通过 Ent 按用户 ID 读取身份并补齐角色编码。
func TestAuthServiceFindByID(t *testing.T) {
	db, mock, client := newMockEntClient(t)
	defer db.Close()

	// Me 只使用 FindByID，不访问 passwords 或 tokens，故可传 nil。
	svc := auth.NewService(client, nil, nil)
	mock.ExpectQuery(`SELECT .*FROM "sys_user".*`).
		WillReturnRows(sysUserRows().AddRow(
			int64(1001), nil, testNow, testNow, int64(9001), int64(9001),
			"alice", "hashed:secret-123", "Alice", "13800000000", "alice@example.com",
			nil, true, true, 1,
		))
	mock.ExpectQuery(`SELECT .*FROM "sys_role".*`).
		WillReturnRows(sqlmock.NewRows([]string{"code"}).AddRow("admin"))

	vo, err := svc.Me(context.Background(), idgen.New(1001))

	require.NoError(t, err)
	require.Equal(t, idgen.New(1001), vo.UserID)
	require.Equal(t, "alice", vo.Username)
	require.Equal(t, "Alice", vo.Nickname)
	require.Equal(t, []string{"admin"}, vo.RoleCodes)
	require.True(t, vo.IsSuperAdmin)
	require.True(t, vo.IsEnable)
	require.NoError(t, mock.ExpectationsWereMet())
}
