package conf

import (
	"fmt"
)

// Snowflake 雪花发号器相关配置
type Snowflake struct {
	NodeID int64 `yaml:"node_id"` // NodeID 节点ID 取值范围: NodeID ∈ [1, 1023]
}

// validate 本方法用于检查配置文件中 snowflake 配置段中的 node_id 是否为合法值
func (s *Snowflake) validate() error {
	if s.NodeID < 1 || s.NodeID > 1023 {
		return fmt.Errorf("snowflake.node_id 必须在范围 [1, 1023] 内, 当前值: %d", s.NodeID)
	}

	return nil
}