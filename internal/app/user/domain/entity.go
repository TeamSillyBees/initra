package domain

import "time"

// User 是 user 模块围绕 sys_user 及其角色关系构建的核心领域实体。
// 该实体只保留应用层真正需要感知的字段，避免把底层表结构细节直接泄漏到上层。
type User struct {
	ID           int64
	Username     string
	PasswordHash string
	Nickname     string
	Phone        string
	Email        string
	AvatarURL    string
	RoleCodes    []string
	IsSuperAdmin bool
	IsEnable     bool
	SortID       int
	CreatedAt    time.Time
	UpdatedAt    time.Time
	DeletedAt    *time.Time
	CreatedBy    int64
	UpdatedBy    int64
}

// CreateUserInput 描述创建用户时所需的输入。
// 这里直接使用当前系统用户模型字段，避免旧示例字段继续扩散到新的业务代码中。
type CreateUserInput struct {
	Username     string
	Password     string
	Nickname     string
	Phone        string
	Email        string
	AvatarURL    string
	RoleCodes    []string
	IsSuperAdmin bool
	IsEnable     *bool
	SortID       int
	OperatorID   int64
}

// UpdateUserInput 描述更新用户时允许变更的字段。
// 采用指针字段表示“是否要修改该值”，以便同时支持显式置空和保持不变。
type UpdateUserInput struct {
	ID           int64
	Nickname     *string
	Phone        *string
	Email        *string
	AvatarURL    *string
	RoleCodes    *[]string
	IsSuperAdmin *bool
	IsEnable     *bool
	SortID       *int
	OperatorID   int64
}

// ListUsersInput 描述分页查询用户列表的输入参数。
type ListUsersInput struct {
	Page     int
	PageSize int
	Keyword  string
}

// ListUsersResult 描述分页查询结果。
type ListUsersResult struct {
	Items    []*User
	Total    int64
	Page     int
	PageSize int
}
