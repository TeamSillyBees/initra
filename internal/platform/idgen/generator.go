package idgen

import "github.com/bwmarrin/snowflake"

// Generator 对 snowflake 节点做了一层薄封装，避免业务模块直接依赖第三方细节。
type Generator struct {
	node *snowflake.Node
}

// NewGenerator 初始化指定节点编号的雪花算法生成器。
func NewGenerator(node int64) (*Generator, error) {
	n, err := snowflake.NewNode(node)
	if err != nil {
		return nil, err
	}
	return &Generator{node: n}, nil
}

// NextID 返回一个全局唯一的 int64 主键。
func (g *Generator) NextID() int64 {
	return g.node.Generate().Int64()
}
