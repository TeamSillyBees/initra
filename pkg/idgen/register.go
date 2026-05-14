package idgen

import "github.com/samber/do"

// Register 将雪花算法 ID 生成器注册到 DI 容器。
func Register(injector *do.Injector, node int64) {
	do.Provide(injector, func(i *do.Injector) (*Generator, error) {
		return NewGenerator(node)
	})
}
