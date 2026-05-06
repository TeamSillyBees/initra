package user

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/teamsillybees/initra/pkg/pagination"
)

var errInvalidPassword = errors.New("invalid password")

type fakeUserRepository struct {
	created   *User
	byID      map[int64]*User
	lastPage  PageUsersDTO
	pageTotal int64
}

func (f *fakeUserRepository) Create(_ context.Context, user *User) error {
	cloned := *user
	f.created = &cloned
	if f.byID == nil {
		f.byID = map[int64]*User{}
	}
	f.byID[user.ID] = &cloned
	return nil
}

func (f *fakeUserRepository) FindByID(_ context.Context, id int64) (*User, error) {
	if user, ok := f.byID[id]; ok {
		cloned := *user
		return &cloned, nil
	}
	return nil, nil
}

func (f *fakeUserRepository) FindByUsername(_ context.Context, username string) (*User, error) {
	for _, user := range f.byID {
		if user.Username == username {
			cloned := *user
			return &cloned, nil
		}
	}
	return nil, nil
}

func (f *fakeUserRepository) Page(_ context.Context, input PageUsersDTO) ([]*User, int64, error) {
	f.lastPage = input
	items := make([]*User, 0, len(f.byID))
	for _, user := range f.byID {
		cloned := *user
		items = append(items, &cloned)
	}
	total := int64(len(items))
	if f.pageTotal > 0 {
		total = f.pageTotal
	}
	return items, total, nil
}

func (f *fakeUserRepository) Update(_ context.Context, user *User) error {
	cloned := *user
	f.byID[user.ID] = &cloned
	return nil
}

func (f *fakeUserRepository) Delete(_ context.Context, id int64, _ int64) error {
	delete(f.byID, id)
	return nil
}

type fakeUserCache struct {
	stored *User
}

func (f *fakeUserCache) Get(_ context.Context, _ int64) (*User, bool, error) {
	if f.stored == nil {
		return nil, false, nil
	}
	cloned := *f.stored
	return &cloned, true, nil
}

func (f *fakeUserCache) Set(_ context.Context, user *User) error {
	cloned := *user
	f.stored = &cloned
	return nil
}

func (f *fakeUserCache) Delete(_ context.Context, _ int64) error {
	f.stored = nil
	return nil
}

type fakeIDGenerator struct {
	id int64
}

func (f *fakeIDGenerator) NextID() int64 {
	return f.id
}

type fakePasswordManager struct{}

func (fakePasswordManager) Hash(password string) (string, error) {
	return "hashed:" + password, nil
}

func (fakePasswordManager) Compare(hash string, password string) error {
	if hash != "hashed:"+password {
		return errInvalidPassword
	}
	return nil
}

func TestServiceCreateAssignsIDAndHashesPassword(t *testing.T) {
	now := time.Date(2026, 4, 21, 10, 0, 0, 0, time.UTC)
	repo := &fakeUserRepository{}
	cache := &fakeUserCache{}

	service := NewService(repo, cache, &fakeIDGenerator{id: 1001}, fakePasswordManager{}, func() time.Time {
		return now
	})

	user, err := service.Create(context.Background(), CreateUserDTO{
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

func TestServicePageReturnsPaginationMeta(t *testing.T) {
	repo := &fakeUserRepository{
		byID: map[int64]*User{
			1001: {ID: 1001, Username: "alice", IsEnable: true},
		},
		pageTotal: 42,
	}
	service := NewService(repo, &fakeUserCache{}, &fakeIDGenerator{id: 1002}, fakePasswordManager{}, time.Now)

	result, err := service.Page(context.Background(), PageUsersDTO{})
	require.NoError(t, err)

	require.Equal(t, pagination.PageDTO{Page: 1, PageSize: 20}, repo.lastPage.Page)
	require.Equal(t, int64(42), result.Total)
	require.Equal(t, 1, result.Page.Page)
	require.Equal(t, 20, result.Page.PageSize)
	require.Equal(t, 3, pagination.NewPageMetaVO(result.Total, result.Page).TotalPages)
}
