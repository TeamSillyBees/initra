package domain

import "errors"

// auth 模块领域错误常量用于测试替身和业务逻辑表达一致的失败语义。
var (
	// ErrLoginFailed 暴露一个稳定的错误语义，方便测试替身和业务层统一表达。
	ErrLoginFailed = errors.New("login failed")
)
