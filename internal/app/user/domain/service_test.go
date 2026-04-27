package domain

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// fakeUserRepository 使用内存 map 模拟用户仓储，便于测试 service 业务编排。
type fakeUserRepository struct {
	created *User
	byID    map[int64]*User
}

// Create 保存用户副本。
func (f *fakeUserRepository) Create(_ context.Context, user *User) error {
	cloned := *user
	f.created = &cloned
	if f.byID == nil {
		f.byID = map[int64]*User{}
	}
	f.byID[user.ID] = &cloned
	return nil
}

// FindByID 根据用户 ID 返回用户副本。
func (f *fakeUserRepository) FindByID(_ context.Context, id int64) (*User, error) {
	if user, ok := f.byID[id]; ok {
		cloned := *user
		return &cloned, nil
	}
	return nil, nil
}

// FindByUsername 根据用户名返回用户副本。
func (f *fakeUserRepository) FindByUsername(_ context.Context, username string) (*User, error) {
	for _, user := range f.byID {
		if user.Username == username {
			cloned := *user
			return &cloned, nil
		}
	}
	return nil, nil
}

// List 返回仓储中的全部用户副本。
func (f *fakeUserRepository) List(_ context.Context, _ ListUsersInput) ([]*User, int64, error) {
	items := make([]*User, 0, len(f.byID))
	for _, user := range f.byID {
		cloned := *user
		items = append(items, &cloned)
	}
	return items, int64(len(items)), nil
}

// Update 覆盖保存用户副本。
func (f *fakeUserRepository) Update(_ context.Context, user *User) error {
	cloned := *user
	f.byID[user.ID] = &cloned
	return nil
}

// Delete 从内存仓储中移除用户。
func (f *fakeUserRepository) Delete(_ context.Context, id int64, _ int64) error {
	delete(f.byID, id)
	return nil
}

// fakeUserCache 记录缓存读写状态，验证 service 的回填和失效逻辑。
type fakeUserCache struct {
	stored *User
}

// Get 按预设结果模拟缓存命中或未命中。
func (f *fakeUserCache) Get(_ context.Context, _ int64) (*User, bool, error) {
	if f.stored == nil {
		return nil, false, nil
	}
	cloned := *f.stored
	return &cloned, true, nil
}

// Set 记录被回填的用户。
func (f *fakeUserCache) Set(_ context.Context, user *User) error {
	cloned := *user
	f.stored = &cloned
	return nil
}

// Delete 记录缓存删除动作。
func (f *fakeUserCache) Delete(_ context.Context, _ int64) error {
	f.stored = nil
	return nil
}

// fakeIDGenerator 返回可预测 ID，方便断言创建用户结果。
type fakeIDGenerator struct {
	id int64
}

// NextID 递增并返回下一个测试 ID。
func (f *fakeIDGenerator) NextID() int64 {
	return f.id
}

// fakePasswordManager 用固定规则模拟密码哈希和校验。
type fakePasswordManager struct{}

// Hash 返回测试哈希格式。
func (fakePasswordManager) Hash(password string) (string, error) {
	return "hashed:" + password, nil
}

// Compare 校验测试哈希格式。
func (fakePasswordManager) Compare(hash string, password string) error {
	if hash != "hashed:"+password {
		return ErrInvalidPassword
	}
	return nil
}

// TestServiceCreateAssignsIDAndHashesPassword 验证创建用户会生成 ID、哈希密码并设置默认角色。
func TestServiceCreateAssignsIDAndHashesPassword(t *testing.T) {
	now := time.Date(2026, 4, 21, 10, 0, 0, 0, time.UTC)
	repo := &fakeUserRepository{}
	cache := &fakeUserCache{}

	service := NewService(repo, cache, &fakeIDGenerator{id: 1001}, fakePasswordManager{}, func() time.Time {
		return now
	})

	user, err := service.Create(context.Background(), CreateUserInput{
		Username:     "alice",
		Password:     "secret-123",
		Nickname:     "Alice",
		Phone:        "13800000000",
		Email:        "alice@example.com",
		AvatarURL:    "https://example.com/avatar.png",
		RoleCodes:    []string{"admin"},
		IsSuperAdmin: true,
		OperatorID:   9001,
	})
	require.NoError(t, err)
	require.Equal(t, int64(1001), user.ID)
	require.Equal(t, "hashed:secret-123", repo.created.PasswordHash)
	require.Equal(t, "Alice", repo.created.Nickname)
	require.Equal(t, "13800000000", repo.created.Phone)
	require.Equal(t, "https://example.com/avatar.png", repo.created.AvatarURL)
	require.Equal(t, []string{"admin"}, repo.created.RoleCodes)
	require.True(t, repo.created.IsSuperAdmin)
	require.True(t, repo.created.IsEnable)
	require.Equal(t, now, repo.created.CreatedAt)
	require.Equal(t, int64(9001), repo.created.CreatedBy)
}

// TestServiceGetBackfillsCacheOnMiss 验证缓存未命中时会从仓储读取并回填缓存。
func TestServiceGetBackfillsCacheOnMiss(t *testing.T) {
	repo := &fakeUserRepository{
		byID: map[int64]*User{
			1001: {
				ID:        1001,
				Username:  "alice",
				Nickname:  "Alice",
				Phone:     "13800000000",
				Email:     "alice@example.com",
				RoleCodes: []string{"admin"},
				IsEnable:  true,
			},
		},
	}
	cache := &fakeUserCache{}

	service := NewService(repo, cache, &fakeIDGenerator{id: 1002}, fakePasswordManager{}, time.Now)

	user, err := service.Get(context.Background(), 1001)
	require.NoError(t, err)
	require.Equal(t, int64(1001), user.ID)
	require.NotNil(t, cache.stored)
	require.Equal(t, int64(1001), cache.stored.ID)
	require.Equal(t, []string{"admin"}, cache.stored.RoleCodes)
}
