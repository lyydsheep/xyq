package snowflake

import (
	"fmt"
	"os"
	"strconv"

	"github.com/bwmarrin/snowflake"
	"github.com/go-kratos/kratos/v2/log"
)

// SnowflakeConfig 雪花算法配置
type SnowflakeConfig struct {
	NodeID    int64 // 节点ID (0-1023)
	StartTime int64 // 起始时间戳，可选，默认使用库的默认时间
}

// DefaultSnowflakeConfig 默认配置
func DefaultSnowflakeConfig() *SnowflakeConfig {
	return &SnowflakeConfig{
		NodeID: 1, // 默认节点ID为1
	}
}

// SnowflakeGenerator 雪花算法生成器
type SnowflakeGenerator struct {
	node *snowflake.Node
	log  *log.Helper
}

// NewSnowflakeGenerator 创建雪花算法生成器
func NewSnowflakeGenerator(config *SnowflakeConfig, logger log.Logger) (*SnowflakeGenerator, error) {
	if config == nil {
		config = DefaultSnowflakeConfig()
	}

	// 从环境变量获取节点ID，如果设置了的话
	if envNodeID := os.Getenv("SNOWFLAKE_NODE_ID"); envNodeID != "" {
		if nodeID, err := strconv.ParseInt(envNodeID, 10, 64); err == nil {
			config.NodeID = nodeID
		}
	}

	// 验证节点ID范围
	if config.NodeID < 0 || config.NodeID > 1023 {
		return nil, fmt.Errorf("node ID must be between 0 and 1023, got: %d", config.NodeID)
	}

	node, err := snowflake.NewNode(config.NodeID)
	if err != nil {
		return nil, fmt.Errorf("failed to create snowflake node: %w", err)
	}

	log := log.NewHelper(logger)
	log.Infof("Snowflake generator initialized with node ID: %d", config.NodeID)

	return &SnowflakeGenerator{
		node: node,
		log:  log,
	}, nil
}

// GenerateID 生成雪花ID
func (s *SnowflakeGenerator) GenerateID() int64 {
	id := s.node.Generate().Int64()
	s.log.Debugf("Generated snowflake ID: %d", id)
	return id
}

// GenerateIDString 生成雪花ID字符串
func (s *SnowflakeGenerator) GenerateIDString() string {
	id := s.node.Generate().String()
	s.log.Debugf("Generated snowflake ID string: %s", id)
	return id
}

// ParseSnowflakeID 解析雪花ID（用于调试）
func ParseSnowflakeID(id int64) (nodeID int64, sequence int64, timestamp int64) {
	sfID := snowflake.ParseInt64(id)
	return int64(sfID.Node()), int64(sfID.Step()), sfID.Time()
}
