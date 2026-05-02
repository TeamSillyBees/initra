package user

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"time"

	. "github.com/go-jet/jet/v2/postgres"
	. "github.com/teamsillybees/initra/examples/basic/internal/gen/jet/table"
	platformdb "github.com/teamsillybees/initra/pkg/db"
	apperrors "github.com/teamsillybees/initra/pkg/errors"
)

// Repository 使用 Jet SQL Builder 实现 user 模块仓储。
type Repository struct {
	db    *sql.DB
	idgen idGenerator
}

// NewRepository 创建 user 仓储实例。
func NewRepository(db *sql.DB, idgen idGenerator) *Repository {
	return &Repository{
		db:    db,
		idgen: idgen,
	}
}

// Create 持久化一个新的用户实体及其角色关系。
func (r *Repository) Create(ctx context.Context, user *User) error {
	return platformdb.WithinTx(ctx, r.db, func(txCtx context.Context) error {
		executor := platformdb.ExecutorFromContext(txCtx, r.db)

		stmt := SysUser.
			INSERT(
				SysUser.ID,
				SysUser.Username,
				SysUser.PasswordHash,
				SysUser.Nickname,
				SysUser.Phone,
				SysUser.Email,
				SysUser.AvatarURL,
				SysUser.IsSuperAdmin,
				SysUser.IsEnable,
				SysUser.SortID,
				SysUser.CreatedAt,
				SysUser.UpdatedAt,
				SysUser.CreatedBy,
				SysUser.UpdatedBy,
			).
			VALUES(
				user.ID,
				user.Username,
				user.PasswordHash,
				user.Nickname,
				user.Phone,
				user.Email,
				user.AvatarURL,
				user.IsSuperAdmin,
				user.IsEnable,
				user.SortID,
				user.CreatedAt,
				user.UpdatedAt,
				user.CreatedBy,
				user.UpdatedBy,
			)

		query, args := stmt.Sql()
		if _, err := executor.ExecContext(txCtx, query, args...); err != nil {
			return apperrors.Wrap(err, apperrors.CodeDBError, "create user failed")
		}

		roleIDs, err := r.resolveRoleIDs(txCtx, user.RoleCodes)
		if err != nil {
			return err
		}
		if err := r.replaceUserRoles(txCtx, user.ID, roleIDs, user.CreatedBy, user.CreatedAt); err != nil {
			return err
		}

		return nil
	})
}

// FindByID 根据用户 ID 查询用户详情。
func (r *Repository) FindByID(ctx context.Context, id int64) (*User, error) {
	stmt := selectBase().
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

// FindByUsername 根据用户名查询用户详情。
func (r *Repository) FindByUsername(ctx context.Context, username string) (*User, error) {
	stmt := selectBase().
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

// List 分页查询用户列表。
func (r *Repository) List(ctx context.Context, input ListUsersParams) ([]*User, int64, error) {
	page := input.Page
	if page <= 0 {
		page = 1
	}
	pageSize := input.PageSize
	if pageSize <= 0 {
		pageSize = 20
	}
	offset := int64((page - 1) * pageSize)

	condition := SysUser.DeletedAt.IS_NULL()
	if keyword := strings.TrimSpace(input.Keyword); keyword != "" {
		pattern := "%" + keyword + "%"
		condition = AND(
			condition,
			OR(
				SysUser.Username.LIKE(String(pattern)),
				SysUser.Nickname.LIKE(String(pattern)),
				SysUser.Phone.LIKE(String(pattern)),
				SysUser.Email.LIKE(String(pattern)),
			),
		)
	}

	stmt := selectBase().
		FROM(SysUser).
		WHERE(condition).
		ORDER_BY(SysUser.SortID.ASC(), SysUser.ID.ASC()).
		LIMIT(int64(pageSize)).
		OFFSET(offset)

	executor := platformdb.ExecutorFromContext(ctx, r.db)
	query, args := stmt.Sql()
	rows, err := executor.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, 0, apperrors.Wrap(err, apperrors.CodeDBError, "list users failed")
	}
	defer rows.Close()

	items := make([]*User, 0)
	userIDs := make([]int64, 0)
	for rows.Next() {
		record, err := scanUser(rows)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, record)
		userIDs = append(userIDs, record.ID)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, apperrors.Wrap(err, apperrors.CodeDBError, "iterate users failed")
	}

	roleCodesByUserID, err := r.loadRoleCodesByUserIDs(ctx, userIDs)
	if err != nil {
		return nil, 0, err
	}
	for _, item := range items {
		item.RoleCodes = append([]string(nil), roleCodesByUserID[item.ID]...)
		if item.RoleCodes == nil {
			item.RoleCodes = []string{}
		}
	}

	countStmt := SELECT(COUNT(SysUser.ID).AS("total")).
		FROM(SysUser).
		WHERE(condition)

	countQuery, countArgs := countStmt.Sql()
	var total int64
	if err := executor.QueryRowContext(ctx, countQuery, countArgs...).Scan(&total); err != nil {
		return nil, 0, apperrors.Wrap(err, apperrors.CodeDBError, "count users failed")
	}

	return items, total, nil
}

// Update 更新用户基础信息与角色配置。
func (r *Repository) Update(ctx context.Context, user *User) error {
	return platformdb.WithinTx(ctx, r.db, func(txCtx context.Context) error {
		executor := platformdb.ExecutorFromContext(txCtx, r.db)

		stmt := SysUser.
			UPDATE(
				SysUser.Nickname,
				SysUser.Phone,
				SysUser.Email,
				SysUser.AvatarURL,
				SysUser.IsSuperAdmin,
				SysUser.IsEnable,
				SysUser.SortID,
				SysUser.UpdatedAt,
				SysUser.UpdatedBy,
			).
			SET(
				user.Nickname,
				user.Phone,
				user.Email,
				user.AvatarURL,
				user.IsSuperAdmin,
				user.IsEnable,
				user.SortID,
				user.UpdatedAt,
				user.UpdatedBy,
			).
			WHERE(
				AND(
					SysUser.ID.EQ(Int64(user.ID)),
					SysUser.DeletedAt.IS_NULL(),
				),
			)

		query, args := stmt.Sql()
		if _, err := executor.ExecContext(txCtx, query, args...); err != nil {
			return apperrors.Wrap(err, apperrors.CodeDBError, "update user failed")
		}

		roleIDs, err := r.resolveRoleIDs(txCtx, user.RoleCodes)
		if err != nil {
			return err
		}
		if err := r.replaceUserRoles(txCtx, user.ID, roleIDs, user.UpdatedBy, user.UpdatedAt); err != nil {
			return err
		}

		return nil
	})
}

// Delete 通过软删除方式移除用户，并清理角色关系。
func (r *Repository) Delete(ctx context.Context, id int64, operatorID int64) error {
	return platformdb.WithinTx(ctx, r.db, func(txCtx context.Context) error {
		executor := platformdb.ExecutorFromContext(txCtx, r.db)
		now := time.Now()

		stmt := SysUser.
			UPDATE(
				SysUser.DeletedAt,
				SysUser.UpdatedAt,
				SysUser.UpdatedBy,
			).
			SET(
				now,
				now,
				operatorID,
			).
			WHERE(
				AND(
					SysUser.ID.EQ(Int64(id)),
					SysUser.DeletedAt.IS_NULL(),
				),
			)

		query, args := stmt.Sql()
		if _, err := executor.ExecContext(txCtx, query, args...); err != nil {
			return apperrors.Wrap(err, apperrors.CodeDBError, "delete user failed")
		}

		roleStmt := SysUserRole.
			DELETE().
			WHERE(SysUserRole.UserID.EQ(Int64(id)))
		roleQuery, roleArgs := roleStmt.Sql()
		if _, err := executor.ExecContext(txCtx, roleQuery, roleArgs...); err != nil {
			return apperrors.Wrap(err, apperrors.CodeDBError, "delete user roles failed")
		}

		return nil
	})
}

type rowScanner interface {
	Scan(dest ...any) error
}

type userRecord struct {
	ID           int64
	Username     string
	PasswordHash string
	Nickname     sql.NullString
	Phone        sql.NullString
	Email        sql.NullString
	AvatarURL    sql.NullString
	IsSuperAdmin bool
	IsEnable     bool
	SortID       int
	CreatedAt    time.Time
	UpdatedAt    time.Time
	DeletedAt    sql.NullTime
	CreatedBy    sql.NullInt64
	UpdatedBy    sql.NullInt64
}

func selectBase() SelectStatement {
	return SELECT(
		SysUser.ID.AS("id"),
		SysUser.Username.AS("username"),
		SysUser.PasswordHash.AS("password_hash"),
		SysUser.Nickname.AS("nickname"),
		SysUser.Phone.AS("phone"),
		SysUser.Email.AS("email"),
		SysUser.AvatarURL.AS("avatar_url"),
		SysUser.IsSuperAdmin.AS("is_super_admin"),
		SysUser.IsEnable.AS("is_enable"),
		SysUser.SortID.AS("sort_id"),
		SysUser.CreatedAt.AS("created_at"),
		SysUser.UpdatedAt.AS("updated_at"),
		SysUser.DeletedAt.AS("deleted_at"),
		SysUser.CreatedBy.AS("created_by"),
		SysUser.UpdatedBy.AS("updated_by"),
	)
}

func (r *Repository) queryOne(ctx context.Context, stmt SelectStatement) (*User, error) {
	executor := platformdb.ExecutorFromContext(ctx, r.db)
	query, args := stmt.Sql()
	row := executor.QueryRowContext(ctx, query, args...)

	record := userRecord{}
	if err := row.Scan(
		&record.ID,
		&record.Username,
		&record.PasswordHash,
		&record.Nickname,
		&record.Phone,
		&record.Email,
		&record.AvatarURL,
		&record.IsSuperAdmin,
		&record.IsEnable,
		&record.SortID,
		&record.CreatedAt,
		&record.UpdatedAt,
		&record.DeletedAt,
		&record.CreatedBy,
		&record.UpdatedBy,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, apperrors.Wrap(err, apperrors.CodeDBError, "query user failed")
	}

	user := record.toDomain()
	roleCodes, err := r.loadRoleCodes(ctx, user.ID)
	if err != nil {
		return nil, err
	}
	user.RoleCodes = roleCodes
	return user, nil
}

func scanUser(scanner rowScanner) (*User, error) {
	record := userRecord{}
	if err := scanner.Scan(
		&record.ID,
		&record.Username,
		&record.PasswordHash,
		&record.Nickname,
		&record.Phone,
		&record.Email,
		&record.AvatarURL,
		&record.IsSuperAdmin,
		&record.IsEnable,
		&record.SortID,
		&record.CreatedAt,
		&record.UpdatedAt,
		&record.DeletedAt,
		&record.CreatedBy,
		&record.UpdatedBy,
	); err != nil {
		return nil, apperrors.Wrap(err, apperrors.CodeDBError, "scan user failed")
	}
	return record.toDomain(), nil
}

func (r userRecord) toDomain() *User {
	user := &User{
		ID:           r.ID,
		Username:     r.Username,
		PasswordHash: r.PasswordHash,
		Nickname:     r.Nickname.String,
		Phone:        r.Phone.String,
		Email:        r.Email.String,
		AvatarURL:    r.AvatarURL.String,
		RoleCodes:    []string{},
		IsSuperAdmin: r.IsSuperAdmin,
		IsEnable:     r.IsEnable,
		SortID:       r.SortID,
		CreatedAt:    r.CreatedAt,
		UpdatedAt:    r.UpdatedAt,
	}
	if r.DeletedAt.Valid {
		deletedAt := r.DeletedAt.Time
		user.DeletedAt = &deletedAt
	}
	if r.CreatedBy.Valid {
		user.CreatedBy = r.CreatedBy.Int64
	}
	if r.UpdatedBy.Valid {
		user.UpdatedBy = r.UpdatedBy.Int64
	}
	return user
}

func (r *Repository) resolveRoleIDs(ctx context.Context, roleCodes []string) ([]int64, error) {
	normalizedCodes := normalizeCodes(roleCodes)
	if len(normalizedCodes) == 0 {
		return []int64{}, nil
	}

	executor := platformdb.ExecutorFromContext(ctx, r.db)
	stmt := SELECT(
		SysRole.ID.AS("id"),
		SysRole.Code.AS("code"),
	).
		FROM(SysRole).
		WHERE(
			AND(
				SysRole.Code.IN(stringExpressions(normalizedCodes)...),
				SysRole.DeletedAt.IS_NULL(),
				SysRole.IsEnable.EQ(Bool(true)),
			),
		).
		ORDER_BY(SysRole.SortID.ASC(), SysRole.ID.ASC())

	query, args := stmt.Sql()
	rows, err := executor.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, apperrors.Wrap(err, apperrors.CodeDBError, "query roles failed")
	}
	defer rows.Close()

	roleIDsByCode := make(map[string]int64, len(normalizedCodes))
	for rows.Next() {
		var (
			roleID   int64
			roleCode string
		)
		if err := rows.Scan(&roleID, &roleCode); err != nil {
			return nil, apperrors.Wrap(err, apperrors.CodeDBError, "scan roles failed")
		}
		roleIDsByCode[roleCode] = roleID
	}
	if err := rows.Err(); err != nil {
		return nil, apperrors.Wrap(err, apperrors.CodeDBError, "iterate roles failed")
	}

	if len(roleIDsByCode) != len(normalizedCodes) {
		missingCodes := make([]string, 0)
		for _, roleCode := range normalizedCodes {
			if _, ok := roleIDsByCode[roleCode]; !ok {
				missingCodes = append(missingCodes, roleCode)
			}
		}
		return nil, apperrors.New(
			apperrors.CodeBadRequest,
			"role code is invalid",
			apperrors.WithDetail("missing_role_codes", missingCodes),
		)
	}

	roleIDs := make([]int64, 0, len(normalizedCodes))
	for _, roleCode := range normalizedCodes {
		roleIDs = append(roleIDs, roleIDsByCode[roleCode])
	}
	return roleIDs, nil
}

func (r *Repository) replaceUserRoles(ctx context.Context, userID int64, roleIDs []int64, operatorID int64, createdAt time.Time) error {
	executor := platformdb.ExecutorFromContext(ctx, r.db)

	deleteStmt := SysUserRole.
		DELETE().
		WHERE(SysUserRole.UserID.EQ(Int64(userID)))
	deleteQuery, deleteArgs := deleteStmt.Sql()
	if _, err := executor.ExecContext(ctx, deleteQuery, deleteArgs...); err != nil {
		return apperrors.Wrap(err, apperrors.CodeDBError, "replace user roles failed")
	}

	if len(roleIDs) == 0 {
		return nil
	}
	if r.idgen == nil {
		return apperrors.New(apperrors.CodeInternalError, "user role id generator is missing")
	}

	insertStmt := SysUserRole.INSERT(
		SysUserRole.ID,
		SysUserRole.UserID,
		SysUserRole.RoleID,
		SysUserRole.CreatedBy,
		SysUserRole.CreatedAt,
	)
	for _, roleID := range roleIDs {
		insertStmt = insertStmt.VALUES(r.idgen.NextID(), userID, roleID, operatorID, createdAt)
	}

	query, args := insertStmt.Sql()
	if _, err := executor.ExecContext(ctx, query, args...); err != nil {
		return apperrors.Wrap(err, apperrors.CodeDBError, "insert user roles failed")
	}
	return nil
}

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
		return nil, apperrors.Wrap(err, apperrors.CodeDBError, "query user roles failed")
	}
	defer rows.Close()

	roleCodes := make([]string, 0)
	for rows.Next() {
		var roleCode string
		if err := rows.Scan(&roleCode); err != nil {
			return nil, apperrors.Wrap(err, apperrors.CodeDBError, "scan user roles failed")
		}
		roleCodes = append(roleCodes, roleCode)
	}
	if err := rows.Err(); err != nil {
		return nil, apperrors.Wrap(err, apperrors.CodeDBError, "iterate user roles failed")
	}
	return roleCodes, nil
}

func (r *Repository) loadRoleCodesByUserIDs(ctx context.Context, userIDs []int64) (map[int64][]string, error) {
	result := make(map[int64][]string, len(userIDs))
	if len(userIDs) == 0 {
		return result, nil
	}

	executor := platformdb.ExecutorFromContext(ctx, r.db)
	stmt := SELECT(
		SysUserRole.UserID.AS("user_id"),
		SysRole.Code.AS("code"),
	).
		FROM(
			SysUserRole.
				INNER_JOIN(SysRole, SysRole.ID.EQ(SysUserRole.RoleID)),
		).
		WHERE(
			AND(
				SysUserRole.UserID.IN(int64Expressions(userIDs)...),
				SysUserRole.DeletedAt.IS_NULL(),
				SysRole.DeletedAt.IS_NULL(),
				SysRole.IsEnable.EQ(Bool(true)),
			),
		).
		ORDER_BY(SysUserRole.UserID.ASC(), SysRole.SortID.ASC(), SysRole.ID.ASC())

	query, args := stmt.Sql()
	rows, err := executor.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, apperrors.Wrap(err, apperrors.CodeDBError, "query users roles failed")
	}
	defer rows.Close()

	for rows.Next() {
		var (
			userID   int64
			roleCode string
		)
		if err := rows.Scan(&userID, &roleCode); err != nil {
			return nil, apperrors.Wrap(err, apperrors.CodeDBError, "scan users roles failed")
		}
		result[userID] = append(result[userID], roleCode)
	}
	if err := rows.Err(); err != nil {
		return nil, apperrors.Wrap(err, apperrors.CodeDBError, "iterate users roles failed")
	}
	return result, nil
}

func normalizeCodes(values []string) []string {
	if len(values) == 0 {
		return nil
	}
	seen := make(map[string]struct{}, len(values))
	result := make([]string, 0, len(values))
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" {
			continue
		}
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		result = append(result, value)
	}
	return result
}

func stringExpressions(values []string) []Expression {
	expressions := make([]Expression, 0, len(values))
	for _, value := range values {
		expressions = append(expressions, String(value))
	}
	return expressions
}

func int64Expressions(values []int64) []Expression {
	expressions := make([]Expression, 0, len(values))
	for _, value := range values {
		expressions = append(expressions, Int64(value))
	}
	return expressions
}
