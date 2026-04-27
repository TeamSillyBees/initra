package infra

import (
	"context"
	"database/sql"
	"errors"
	"strings"

	. "github.com/go-jet/jet/v2/postgres"
	"github.com/teamsillybees/initra/internal/app/auth/domain"
	. "github.com/teamsillybees/initra/internal/gen/jet/table"
	platformdb "github.com/teamsillybees/initra/internal/platform/database"
	apperrors "github.com/teamsillybees/initra/internal/platform/errors"
)

// Repository 使用 Jet SQL Builder 实现 auth 模块身份仓储。
// 读取身份时只关注登录所需的最小字段，并通过角色关系表补齐角色编码集合。
type Repository struct {
	db *sql.DB
}

// NewRepository 创建 auth 仓储实例。
func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

// FindByUsername 根据用户名读取登录身份信息。
func (r *Repository) FindByUsername(ctx context.Context, username string) (*domain.Identity, error) {
	stmt := selectIdentityBase().
		FROM(SysUser).
		WHERE(
			AND(
				SysUser.Username.EQ(String(strings.TrimSpace(username))),
				SysUser.DeletedAt.IS_NULL(),
			),
		).
		LIMIT(1)
	return r.queryOne(ctx, stmt)
}

// FindByID 根据用户 ID 读取身份信息。
func (r *Repository) FindByID(ctx context.Context, id int64) (*domain.Identity, error) {
	stmt := selectIdentityBase().
		FROM(SysUser).
		WHERE(
			AND(
				SysUser.ID.EQ(Int64(id)),
				SysUser.DeletedAt.IS_NULL(),
			),
		).
		LIMIT(1)
	return r.queryOne(ctx, stmt)
}

// identityRecord 是数据库行到领域 Identity 之间的中间结构，负责承接 nullable 字段。
type identityRecord struct {
	UserID       int64
	Username     string
	Nickname     sql.NullString
	PasswordHash string
	IsSuperAdmin bool
	IsEnable     bool
}

// selectIdentityBase 定义登录身份查询所需字段，避免不同查询方法维护重复投影。
func selectIdentityBase() SelectStatement {
	return SELECT(
		SysUser.ID.AS("id"),
		SysUser.Username.AS("username"),
		SysUser.Nickname.AS("nickname"),
		SysUser.PasswordHash.AS("password_hash"),
		SysUser.IsSuperAdmin.AS("is_super_admin"),
		SysUser.IsEnable.AS("is_enable"),
	)
}

// queryOne 执行单条身份查询，并补齐用户角色编码集合。
func (r *Repository) queryOne(ctx context.Context, stmt SelectStatement) (*domain.Identity, error) {
	executor := platformdb.ExecutorFromContext(ctx, r.db)
	query, args := stmt.Sql()
	row := executor.QueryRowContext(ctx, query, args...)

	record := identityRecord{}
	if err := row.Scan(
		&record.UserID,
		&record.Username,
		&record.Nickname,
		&record.PasswordHash,
		&record.IsSuperAdmin,
		&record.IsEnable,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, apperrors.Wrap(err, apperrors.CodeDBError, "query identity failed")
	}

	roleCodes, err := r.loadRoleCodes(ctx, record.UserID)
	if err != nil {
		return nil, err
	}

	return &domain.Identity{
		UserID:       record.UserID,
		Username:     record.Username,
		Nickname:     record.Nickname.String,
		PasswordHash: record.PasswordHash,
		RoleCodes:    roleCodes,
		IsSuperAdmin: record.IsSuperAdmin,
		IsEnable:     record.IsEnable,
	}, nil
}

// loadRoleCodes 查询用户已启用角色编码，用于登录后写入 JWT claims。
func (r *Repository) loadRoleCodes(ctx context.Context, userID int64) ([]string, error) {
	executor := platformdb.ExecutorFromContext(ctx, r.db)
	stmt := SELECT(SysRole.Code.AS("code")).
		FROM(
			SysUserRole.
				INNER_JOIN(SysRole, SysRole.ID.EQ(SysUserRole.RoleID)),
		).
		WHERE(
			AND(
				SysUserRole.UserID.EQ(Int64(userID)),
				SysUserRole.DeletedAt.IS_NULL(),
				SysRole.DeletedAt.IS_NULL(),
				SysRole.IsEnable.EQ(Bool(true)),
			),
		).
		ORDER_BY(SysRole.SortID.ASC(), SysRole.ID.ASC())

	query, args := stmt.Sql()
	rows, err := executor.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, apperrors.Wrap(err, apperrors.CodeDBError, "query identity roles failed")
	}
	defer rows.Close()

	roleCodes := make([]string, 0)
	for rows.Next() {
		var roleCode string
		if err := rows.Scan(&roleCode); err != nil {
			return nil, apperrors.Wrap(err, apperrors.CodeDBError, "scan identity roles failed")
		}
		roleCodes = append(roleCodes, roleCode)
	}
	if err := rows.Err(); err != nil {
		return nil, apperrors.Wrap(err, apperrors.CodeDBError, "iterate identity roles failed")
	}
	return roleCodes, nil
}
