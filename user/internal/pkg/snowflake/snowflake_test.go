package snowflake

import (
	"github.com/go-kratos/kratos/v2/log"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestSnowflakeGenerator(t *testing.T) {
	// 创建一个空日志记录器用于测试
	logger := log.DefaultLogger

	// 测试默认配置
	config := DefaultSnowflakeConfig()
	assert.Equal(t, int64(1), config.NodeID)

	// 测试创建生成器
	gen, err := NewSnowflakeGenerator(config, logger)
	require.NoError(t, err)
	assert.NotNil(t, gen)

	// 测试生成ID
	id1 := gen.GenerateID()
	id2 := gen.GenerateID()

	t.Logf("Generated ID1: %d", id1)
	t.Logf("Generated ID2: %d", id2)

	// ID应该是唯一的
	assert.NotEqual(t, id1, id2)

	// ID应该大于0
	assert.Greater(t, id1, int64(0))
	assert.Greater(t, id2, int64(0))

	// 测试解析ID
	nodeID, step, timestamp := ParseSnowflakeID(id1)
	t.Logf("Parsed ID1 - Node: %d, Step: %d, Time: %d", nodeID, step, timestamp)

	// 解析出的节点ID应该匹配配置
	assert.Equal(t, config.NodeID, nodeID)

	// 时间戳应该是合理的（雪花算法使用毫秒时间戳）
	// 由于雪花算法有自己的epoch，我们需要检查时间戳是否合理
	assert.Greater(t, timestamp, int64(0))          // 时间戳应该大于0
	assert.Less(t, timestamp, int64(2200000000000)) // 不应该超过2200年
}

func TestSnowflakeGeneratorFromEnv(t *testing.T) {
	// 设置环境变量测试
	t.Setenv("SNOWFLAKE_NODE_ID", "42")

	logger := log.DefaultLogger
	config := DefaultSnowflakeConfig()

	// 环境变量应该覆盖默认值
	gen, err := NewSnowflakeGenerator(config, logger)
	require.NoError(t, err)

	// 验证节点ID是否被环境变量覆盖
	id := gen.GenerateID()
	nodeID, _, _ := ParseSnowflakeID(id)
	assert.Equal(t, int64(42), nodeID)
}

func TestSnowflakeGeneratorInvalidNodeID(t *testing.T) {
	logger := log.DefaultLogger

	// 测试无效的节点ID（超过1023）
	config := &SnowflakeConfig{NodeID: 1024}
	_, err := NewSnowflakeGenerator(config, logger)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "node ID must be between 0 and 1023")
}

func TestSnowflakeGeneratorInvalidNodeIDNegative(t *testing.T) {
	logger := log.DefaultLogger

	// 测试负数的节点ID
	config := &SnowflakeConfig{NodeID: -1}
	_, err := NewSnowflakeGenerator(config, logger)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "node ID must be between 0 and 1023")
}

// currentTimeMillis 获取当前时间戳（毫秒）
func currentTimeMillis() int64 {
	return 1732048000000 // 简化实现，使用一个固定的时间戳用于测试
}
