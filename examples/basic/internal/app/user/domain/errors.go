package domain

import "errors"

// user 模块领域错误常量用于测试替身和业务逻辑表达一致的失败语义。
var (
	// ErrInvalidPassword 用于测试替身与业务逻辑共享同一语义。
	ErrInvalidPassword = errors.New("invalid password")
)
