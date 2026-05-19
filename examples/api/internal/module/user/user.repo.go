package user

import (
	"context"
	"strings"
	"time"

	"github.com/teamsillybees/initra/examples/api/internal/data"
	appent "github.com/teamsillybees/initra/examples/api/internal/ent"
	"github.com/teamsillybees/initra/examples/api/internal/ent/sysrole"
	"github.com/teamsillybees/initra/examples/api/internal/ent/sysuser"
	"github.com/teamsillybees/initra/examples/api/internal/ent/sysuserrole"
	"github.com/teamsillybees/initra/examples/api/internal/module/bizerrors"
	"github.com/teamsillybees/initra/pkg/entx"
)

// Repository 使用 Ent Client 实现 user 模块仓储。
type Repository struct {
	client *appent.Client
}

// NewRepository 创建 user 仓储实例。
func NewRepository(client *appent.Client) *Repository {
	return &Repository{client: client}
}

// Create 持久化一个新的用户实体及其角色关系。
func (r *Repository) Create(ctx context.Context, user *User) error {
	return data.WithinTx(ctx, r.client, func(txCtx context.Context, txClient *appent.Client) error {
		roleIDs, err := r.resolveRoleIDs(txCtx, txClient, user.RoleCodes)
		if err != nil {
			return err
		}

		record, err := txClient.SysUser.Create().
			SetUsername(user.Username).
			SetPasswordHash(user.PasswordHash).
			SetNillableNickname(nullableString(user.Nickname)).
			SetNillablePhone(nullableString(user.Phone)).
			SetNillableEmail(nullableString(user.Email)).
			SetNillableAvatarURL(nullableString(user.AvatarURL)).
			SetIsSuperAdmin(user.IsSuperAdmin).
			SetIsEnable(user.IsEnable).
			SetSortID(user.SortID).
			Save(txCtx)
		if err != nil {
			return mapEntWriteError(err, "create user failed")
		}

		fillDomainFromEnt(user, record)
		user.RoleCodes = normalizeCodes(user.RoleCodes)
		if err := r.replaceUserRoles(txCtx, txClient, record.ID, roleIDs); err != nil {
			return err
		}
		return nil
	})
}

// FindByID 根据用户 ID 查询用户详情。
func (r *Repository) FindByID(ctx context.Context, id int64) (*User, error) {
	record, err := r.client.SysUser.Query().
		Where(
			sysuser.ID(id),
			sysuser.DeletedAtIsNil(),
		).
		Only(ctx)
	if appent.IsNotFound(err) {
		return nil, nil
	}
	if err != nil {
		return nil, bizerrors.WrapDB(err, "query user failed")
	}
	return r.toDomain(ctx, record)
}

// FindByUsername 根据用户名查询用户详情。
func (r *Repository) FindByUsername(ctx context.Context, username string) (*User, error) {
	record, err := r.client.SysUser.Query().
		Where(
			sysuser.Username(strings.TrimSpace(username)),
			sysuser.DeletedAtIsNil(),
		).
		Only(ctx)
	if appent.IsNotFound(err) {
		return nil, nil
	}
	if err != nil {
		return nil, bizerrors.WrapDB(err, "query user failed")
	}
	return r.toDomain(ctx, record)
}

// Page 分页查询用户列表。
func (r *Repository) Page(ctx context.Context, input PageUsersDTO) ([]*User, int64, error) {
	pageDTO := input.Page
	query := r.client.SysUser.Query().Where(sysuser.DeletedAtIsNil())
	if keyword := strings.TrimSpace(input.Keyword); keyword != "" {
		query.Where(sysuser.Or(
			sysuser.UsernameContainsFold(keyword),
			sysuser.NicknameContainsFold(keyword),
			sysuser.PhoneContainsFold(keyword),
			sysuser.EmailContainsFold(keyword),
		))
	}

	total, err := query.Clone().Count(ctx)
	if err != nil {
		return nil, 0, bizerrors.WrapDB(err, "count users failed")
	}

	records, err := query.
		Order(appent.Asc(sysuser.FieldSortID), appent.Asc(sysuser.FieldID)).
		Limit(int(pageDTO.Limit())).
		Offset(int(pageDTO.Offset())).
		All(ctx)
	if err != nil {
		return nil, 0, bizerrors.WrapDB(err, "list users failed")
	}

	userIDs := make([]int64, 0, len(records))
	items := make([]*User, 0, len(records))
	for _, record := range records {
		item := userFromEnt(record)
		items = append(items, item)
		userIDs = append(userIDs, item.ID)
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

	return items, int64(total), nil
}

// Update 更新用户基础信息与角色配置。
func (r *Repository) Update(ctx context.Context, user *User) error {
	return data.WithinTx(ctx, r.client, func(txCtx context.Context, txClient *appent.Client) error {
		update := txClient.SysUser.UpdateOneID(user.ID).
			Where(sysuser.DeletedAtIsNil()).
			SetIsSuperAdmin(user.IsSuperAdmin).
			SetIsEnable(user.IsEnable).
			SetSortID(user.SortID)
		setOptionalStrings(update, user)

		record, err := update.Save(txCtx)
		if appent.IsNotFound(err) {
			return nil
		}
		if err != nil {
			return mapEntWriteError(err, "update user failed")
		}

		roleIDs, err := r.resolveRoleIDs(txCtx, txClient, user.RoleCodes)
		if err != nil {
			return err
		}
		if err := r.replaceUserRoles(txCtx, txClient, user.ID, roleIDs); err != nil {
			return err
		}

		refreshed, err := r.findByIDWithClient(txCtx, txClient, record.ID)
		if err != nil {
			return err
		}
		if refreshed != nil {
			*user = *refreshed
		}
		return nil
	})
}

// Delete 通过软删除方式移除用户，并清理角色关系。
func (r *Repository) Delete(ctx context.Context, id int64, operatorID int64) error {
	ctx = entx.WithOperatorID(ctx, operatorID)
	return data.WithinTx(ctx, r.client, func(txCtx context.Context, txClient *appent.Client) error {
		deletedAt := data.SoftDeleteTime()
		if _, err := txClient.SysUser.Update().
			Where(
				sysuser.ID(id),
				sysuser.DeletedAtIsNil(),
			).
			SetDeletedAt(deletedAt).
			Save(txCtx); err != nil {
			return mapEntWriteError(err, "delete user failed")
		}

		if _, err := txClient.SysUserRole.Update().
			Where(
				sysuserrole.UserID(id),
				sysuserrole.DeletedAtIsNil(),
			).
			SetDeletedAt(deletedAt).
			Save(txCtx); err != nil {
			return mapEntWriteError(err, "delete user roles failed")
		}
		return nil
	})
}

func (r *Repository) findByIDWithClient(ctx context.Context, client *appent.Client, id int64) (*User, error) {
	record, err := client.SysUser.Query().
		Where(
			sysuser.ID(id),
			sysuser.DeletedAtIsNil(),
		).
		Only(ctx)
	if appent.IsNotFound(err) {
		return nil, nil
	}
	if err != nil {
		return nil, bizerrors.WrapDB(err, "query user failed")
	}
	return r.toDomainWithClient(ctx, client, record)
}

func (r *Repository) toDomain(ctx context.Context, record *appent.SysUser) (*User, error) {
	return r.toDomainWithClient(ctx, r.client, record)
}

func (r *Repository) toDomainWithClient(ctx context.Context, client *appent.Client, record *appent.SysUser) (*User, error) {
	user := userFromEnt(record)
	roleCodes, err := r.loadRoleCodes(ctx, client, user.ID)
	if err != nil {
		return nil, err
	}
	user.RoleCodes = roleCodes
	return user, nil
}

func userFromEnt(record *appent.SysUser) *User {
	user := &User{}
	fillDomainFromEnt(user, record)
	user.RoleCodes = []string{}
	return user
}

func fillDomainFromEnt(user *User, record *appent.SysUser) {
	user.ID = record.ID
	user.Username = record.Username
	user.PasswordHash = record.PasswordHash
	user.Nickname = stringValue(record.Nickname)
	user.Phone = stringValue(record.Phone)
	user.Email = stringValue(record.Email)
	user.AvatarURL = stringValue(record.AvatarURL)
	user.IsSuperAdmin = record.IsSuperAdmin
	user.IsEnable = record.IsEnable
	user.SortID = record.SortID
	user.CreatedAt = record.CreatedAt
	user.UpdatedAt = record.UpdatedAt
	user.DeletedAt = cloneTimePtr(record.DeletedAt)
	user.CreatedBy = int64Value(record.CreatedBy)
	user.UpdatedBy = int64Value(record.UpdatedBy)
}

func (r *Repository) resolveRoleIDs(ctx context.Context, client *appent.Client, roleCodes []string) ([]int64, error) {
	normalizedCodes := normalizeCodes(roleCodes)
	if len(normalizedCodes) == 0 {
		return []int64{}, nil
	}

	var rows []struct {
		ID   int64  `json:"id"`
		Code string `json:"code"`
	}
	err := client.SysRole.Query().
		Where(
			sysrole.CodeIn(normalizedCodes...),
			sysrole.DeletedAtIsNil(),
			sysrole.IsEnable(true),
		).
		Order(appent.Asc(sysrole.FieldSortID), appent.Asc(sysrole.FieldID)).
		Select(sysrole.FieldID, sysrole.FieldCode).
		Scan(ctx, &rows)
	if err != nil {
		return nil, bizerrors.WrapDB(err, "query roles failed")
	}

	roleIDsByCode := make(map[string]int64, len(rows))
	for _, row := range rows {
		roleIDsByCode[row.Code] = row.ID
	}
	if len(roleIDsByCode) != len(normalizedCodes) {
		missingCodes := make([]string, 0)
		for _, roleCode := range normalizedCodes {
			if _, ok := roleIDsByCode[roleCode]; !ok {
				missingCodes = append(missingCodes, roleCode)
			}
		}
		return nil, bizerrors.BadRequest(
			"role code is invalid",
			bizerrors.WithDetail("missing_role_codes", missingCodes),
		)
	}

	roleIDs := make([]int64, 0, len(normalizedCodes))
	for _, roleCode := range normalizedCodes {
		roleIDs = append(roleIDs, roleIDsByCode[roleCode])
	}
	return roleIDs, nil
}

func (r *Repository) replaceUserRoles(ctx context.Context, client *appent.Client, userID int64, roleIDs []int64) error {
	deletedAt := data.SoftDeleteTime()
	if _, err := client.SysUserRole.Update().
		Where(
			sysuserrole.UserID(userID),
			sysuserrole.DeletedAtIsNil(),
		).
		SetDeletedAt(deletedAt).
		Save(ctx); err != nil {
		return mapEntWriteError(err, "replace user roles failed")
	}

	for _, roleID := range roleIDs {
		relation, err := client.SysUserRole.Query().
			Where(
				sysuserrole.UserID(userID),
				sysuserrole.RoleID(roleID),
			).
			Only(ctx)
		switch {
		case appent.IsNotFound(err):
			if _, err := client.SysUserRole.Create().
				SetUserID(userID).
				SetRoleID(roleID).
				Save(ctx); err != nil {
				return mapEntWriteError(err, "insert user roles failed")
			}
		case err != nil:
			return bizerrors.WrapDB(err, "query user role relation failed")
		default:
			if _, err := client.SysUserRole.UpdateOne(relation).
				ClearDeletedAt().
				Save(ctx); err != nil {
				return mapEntWriteError(err, "restore user role relation failed")
			}
		}
	}
	return nil
}

func (r *Repository) loadRoleCodes(ctx context.Context, client *appent.Client, userID int64) ([]string, error) {
	var rows []struct {
		Code string `json:"code"`
	}
	err := client.SysRole.Query().
		Where(
			sysrole.DeletedAtIsNil(),
			sysrole.IsEnable(true),
			sysrole.HasUserRolesWith(
				sysuserrole.UserID(userID),
				sysuserrole.DeletedAtIsNil(),
			),
		).
		Order(appent.Asc(sysrole.FieldSortID), appent.Asc(sysrole.FieldID)).
		Select(sysrole.FieldCode).
		Scan(ctx, &rows)
	if err != nil {
		return nil, bizerrors.WrapDB(err, "query user roles failed")
	}

	roleCodes := make([]string, 0, len(rows))
	for _, row := range rows {
		roleCodes = append(roleCodes, row.Code)
	}
	return roleCodes, nil
}

func (r *Repository) loadRoleCodesByUserIDs(ctx context.Context, userIDs []int64) (map[int64][]string, error) {
	result := make(map[int64][]string, len(userIDs))
	if len(userIDs) == 0 {
		return result, nil
	}

	roles, err := r.client.SysRole.Query().
		Where(
			sysrole.DeletedAtIsNil(),
			sysrole.IsEnable(true),
			sysrole.HasUserRolesWith(
				sysuserrole.UserIDIn(userIDs...),
				sysuserrole.DeletedAtIsNil(),
			),
		).
		WithUserRoles(func(query *appent.SysUserRoleQuery) {
			query.Where(
				sysuserrole.UserIDIn(userIDs...),
				sysuserrole.DeletedAtIsNil(),
			)
		}).
		Order(appent.Asc(sysrole.FieldSortID), appent.Asc(sysrole.FieldID)).
		All(ctx)
	if err != nil {
		return nil, bizerrors.WrapDB(err, "query users roles failed")
	}

	for _, role := range roles {
		for _, relation := range role.Edges.UserRoles {
			result[relation.UserID] = append(result[relation.UserID], role.Code)
		}
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

func setOptionalStrings(update *appent.SysUserUpdateOne, user *User) {
	setOrClearString(update.SetNickname, update.ClearNickname, user.Nickname)
	setOrClearString(update.SetPhone, update.ClearPhone, user.Phone)
	setOrClearString(update.SetEmail, update.ClearEmail, user.Email)
	setOrClearString(update.SetAvatarURL, update.ClearAvatarURL, user.AvatarURL)
}

func setOrClearString(set func(string) *appent.SysUserUpdateOne, clear func() *appent.SysUserUpdateOne, value string) {
	if strings.TrimSpace(value) == "" {
		clear()
		return
	}
	set(value)
}

func nullableString(value string) *string {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil
	}
	return &value
}

func stringValue(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}

func int64Value(value *int64) int64 {
	if value == nil {
		return 0
	}
	return *value
}

func cloneTimePtr(value *time.Time) *time.Time {
	if value == nil {
		return nil
	}
	return new(*value)
}

func mapEntWriteError(err error, message string) error {
	if appent.IsConstraintError(err) {
		return bizerrors.WrapBadRequest(err, message)
	}
	return bizerrors.WrapDB(err, message)
}
