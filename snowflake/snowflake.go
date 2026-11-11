package snowflake

import (
	"fmt"

	"github.com/bwmarrin/snowflake"
)

// 包级别发号器实例 (Tips: snowflake.node是并发安全的)
var node *snowflake.Node

// InitSnowflake 本函数用于根据给定的节点ID初始化发号器,在服务启动时调用1次即可
func InitSnowflake(nodeID int64) error {
	n, err := snowflake.NewNode(nodeID)
	if err != nil {
		return fmt.Errorf("初始化雪花发号器失败: %w", err)
	}

	node = n
	return nil
}

// Next 本函数用于生成一个雪花ID
func Next() int64 {
	return node.Generate().Int64()
}
