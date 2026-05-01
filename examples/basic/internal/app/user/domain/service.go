package domain

import (
	"context"
	"strings"
	"time"

	"github.com/teamsillybees/initra/examples/basic/internal/app/bizerrors"
	apperrors "github.com/teamsillybees/initra/pkg/errors"
)

// Repository 定义 user 模块访问持久化层的最小能力集合。
type Repository interface {
	Create(ctx context.Context, user *User) error
	FindByID(ctx context.Context, id int64) (*User, error)
	FindByUsername(ctx context.Context, username string) (*User, error)
	List(ctx context.Context, input ListUsersInput) ([]*User, int64, error)
	Update(ctx context.Context, user *User) error
	Delete(ctx context.Context, id int64, operatorID int64) error
}

// Cache 定义用户详情缓存的最小契约。
type Cache interface {
	Get(ctx context.Context, id int64) (*User, bool, error)
	Set(ctx context.Context, user *User) error
	Delete(ctx context.Context, id int64) error
}

// IDGenerator 抽象了雪花算法生成器，便于 service 在测试中替换。
type IDGenerator interface {
	NextID() int64
}

// PasswordManager 定义密码哈希与校验能力。
type PasswordManager interface {
	Hash(password string) (string, error)
	Compare(hash string, password string) error
}

// Service 是 user 模块的应用服务，实现 CRUD 编排与缓存回填。
type Service struct {
	repo      Repository
	cache     Cache
	idgen     IDGenerator
	passwords PasswordManager
	now       func() time.Time
}

// NewService 构造 user 模块应用服务。
func NewService(
	repo Repository,
	cache Cache,
	idgen IDGenerator,
	passwords PasswordManager,
	now func() time.Time,
) *Service {
	if now == nil {
		now = time.Now
	}
	return &Service{
		repo:      repo,
		cache:     cache,
		idgen:     idgen,
		passwords: passwords,
		now:       now,
	}
}

// Create 创建用户，并在应用层统一完成主键生成、密码哈希和审计字段赋值。
func (s *Service) Create(ctx context.Context, input CreateUserInput) (*User, error) {
	if strings.TrimSpace(input.Username) == "" {
		return nil, apperrors.New(apperrors.CodeBadRequest, "username is required")
	}
	if strings.TrimSpace(input.Password) == "" {
		return nil, apperrors.New(apperrors.CodeBadRequest, "password is required")
	}

	passwordHash, err := s.passwords.Hash(input.Password)
	if err != nil {
		return nil, apperrors.Wrap(err, apperrors.CodeInternalError, "hash password failed")
	}

	now := s.now()
	isEnable := true
	if input.IsEnable != nil {
		isEnable = *input.IsEnable
	}

	user := &User{
		ID:           s.idgen.NextID(),
		Username:     strings.TrimSpace(input.Username),
		PasswordHash: passwordHash,
		Nickname:     strings.TrimSpace(input.Nickname),
		Phone:        strings.TrimSpace(input.Phone),
		Email:        strings.TrimSpace(input.Email),
		AvatarURL:    strings.TrimSpace(input.AvatarURL),
		RoleCodes:    normalizeRoleCodes(input.RoleCodes),
		IsSuperAdmin: input.IsSuperAdmin,
		IsEnable:     isEnable,
		SortID:       input.SortID,
		CreatedAt:    now,
		UpdatedAt:    now,
		CreatedBy:    input.OperatorID,
		UpdatedBy:    input.OperatorID,
	}

	if err := s.repo.Create(ctx, user); err != nil {
		return nil, err
	}

	return cloneUser(user), nil
}

// Get 获取用户详情，并在缓存未命中时自动做回填。
func (s *Service) Get(ctx context.Context, id int64) (*User, error) {
	if cached, found, err := s.cache.Get(ctx, id); err != nil {
		return nil, apperrors.Wrap(err, apperrors.CodeCacheError, "load user from cache failed")
	} else if found {
		return cloneUser(cached), nil
	}

	user, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, bizerrors.UserNotFound(id)
	}

	if err := s.cache.Set(ctx, user); err != nil {
		return nil, apperrors.Wrap(err, apperrors.CodeCacheError, "set user cache failed")
	}

	return cloneUser(user), nil
}

// List 返回分页用户列表。
func (s *Service) List(ctx context.Context, input ListUsersInput) (*ListUsersResult, error) {
	items, total, err := s.repo.List(ctx, input)
	if err != nil {
		return nil, err
	}
	return &ListUsersResult{
		Items:    items,
		Total:    total,
		Page:     normalizePage(input.Page),
		PageSize: normalizePageSize(input.PageSize),
	}, nil
}

// Update 更新用户公开资料与角色配置，并同步清理缓存。
func (s *Service) Update(ctx context.Context, input UpdateUserInput) (*User, error) {
	user, err := s.repo.FindByID(ctx, input.ID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, bizerrors.UserNotFound(input.ID)
	}

	if input.Nickname != nil {
		user.Nickname = strings.TrimSpace(*input.Nickname)
	}
	if input.Phone != nil {
		user.Phone = strings.TrimSpace(*input.Phone)
	}
	if input.Email != nil {
		user.Email = strings.TrimSpace(*input.Email)
	}
	if input.AvatarURL != nil {
		user.AvatarURL = strings.TrimSpace(*input.AvatarURL)
	}
	if input.RoleCodes != nil {
		user.RoleCodes = normalizeRoleCodes(*input.RoleCodes)
	}
	if input.IsSuperAdmin != nil {
		user.IsSuperAdmin = *input.IsSuperAdmin
	}
	if input.IsEnable != nil {
		user.IsEnable = *input.IsEnable
	}
	if input.SortID != nil {
		user.SortID = *input.SortID
	}
	user.UpdatedAt = s.now()
	user.UpdatedBy = input.OperatorID

	if err := s.repo.Update(ctx, user); err != nil {
		return nil, err
	}
	if err := s.cache.Delete(ctx, user.ID); err != nil {
		return nil, apperrors.Wrap(err, apperrors.CodeCacheError, "delete user cache failed")
	}

	return cloneUser(user), nil
}

// Delete 删除用户并清理缓存。
func (s *Service) Delete(ctx context.Context, id int64, operatorID int64) error {
	if err := s.repo.Delete(ctx, id, operatorID); err != nil {
		return err
	}
	if err := s.cache.Delete(ctx, id); err != nil {
		return apperrors.Wrap(err, apperrors.CodeCacheError, "delete user cache failed")
	}
	return nil
}

// cloneUser 返回用户实体副本，避免缓存对象或仓储对象被上层调用者意外修改。
func cloneUser(user *User) *User {
	if user == nil {
		return nil
	}
	cloned := *user
	cloned.RoleCodes = append([]string(nil), user.RoleCodes...)
	if user.DeletedAt != nil {
		deletedAt := *user.DeletedAt
		cloned.DeletedAt = &deletedAt
	}
	return &cloned
}

// normalizeRoleCodes 清理空角色、去重，并在未显式指定时赋予 viewer 默认角色。
func normalizeRoleCodes(roleCodes []string) []string {
	if len(roleCodes) == 0 {
		return []string{"viewer"}
	}

	seen := make(map[string]struct{}, len(roleCodes))
	result := make([]string, 0, len(roleCodes))
	for _, roleCode := range roleCodes {
		roleCode = strings.TrimSpace(roleCode)
		if roleCode == "" {
			continue
		}
		if _, ok := seen[roleCode]; ok {
			continue
		}
		seen[roleCode] = struct{}{}
		result = append(result, roleCode)
	}
	if len(result) == 0 {
		return []string{"viewer"}
	}
	return result
}

// normalizePage 兜底修正非法页码。
func normalizePage(page int) int {
	if page <= 0 {
		return 1
	}
	return page
}

// normalizePageSize 兜底修正非法分页大小。
func normalizePageSize(pageSize int) int {
	if pageSize <= 0 {
		return 20
	}
	return pageSize
}
